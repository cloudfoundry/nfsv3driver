package nfsv3driver

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"code.cloudfoundry.org/dockerdriver"
	"code.cloudfoundry.org/dockerdriver/driverhttp"
	"code.cloudfoundry.org/goshims/ioutilshim"
	"code.cloudfoundry.org/goshims/osshim"
	"code.cloudfoundry.org/goshims/syscallshim"
	"code.cloudfoundry.org/lager/v3"
	vmo "code.cloudfoundry.org/volume-mount-options"
	"code.cloudfoundry.org/volumedriver"
	"code.cloudfoundry.org/volumedriver/invoker"
	"code.cloudfoundry.org/volumedriver/mountchecker"
)

const MapfsDirectorySuffix = "_mapfs"
const MapfsMountTimeout = time.Minute * 5
const NobodyId = uint32(65534)
const UnknownId = uint32(4294967294)
const InvalidUidValueErrorMessage = "Invalid 'uid' option (0, negative, or non-integer)"
const InvalidGidValueErrorMessage = "Invalid 'gid' option (0, negative, or non-integer)"

type mapfsMounter struct {
	invoker           invoker.Invoker
	backgroundInvoker BackgroundInvoker
	osshim            osshim.Os
	syscallshim       syscallshim.Syscall
	ioutilshim        ioutilshim.Ioutil
	mountChecker      mountchecker.MountChecker
	fstype            string
	defaultOpts       string
	resolver          IdResolver
	mask              vmo.MountOptsMask
	mapfsPath         string
}

var legacyNfsSharePattern *regexp.Regexp

var PurgeTimeToSleep = time.Millisecond * 100

func init() {
	legacyNfsSharePattern, _ = regexp.Compile("^nfs://([^/]+)(/.*)?$")
}

func NewMapfsMounter(invoker invoker.Invoker, bgInvoker BackgroundInvoker, osshim osshim.Os, syscallshim syscallshim.Syscall, ioutilshim ioutilshim.Ioutil, mountChecker mountchecker.MountChecker, fstype, defaultOpts string, resolver IdResolver, mask vmo.MountOptsMask, mapfsPath string) volumedriver.Mounter {
	return &mapfsMounter{invoker, bgInvoker, osshim, syscallshim, ioutilshim, mountChecker, fstype, defaultOpts, resolver, mask, mapfsPath}
}

func (m *mapfsMounter) Mount(env dockerdriver.Env, remote string, target string, opts map[string]interface{}) error {
	logger := env.Logger().Session("mount")
	logger.Info("mount-start")
	defer logger.Info("mount-end")

	if username, ok := opts["username"]; ok {
		if _, found := opts["uid"]; found {
			return dockerdriver.SafeError{SafeDescription: "Not allowed options"}
		}

		if _, found := opts["gid"]; found {
			return dockerdriver.SafeError{SafeDescription: "Not allowed options"}
		}

		if m.resolver == nil {
			return dockerdriver.SafeError{SafeDescription: "LDAP username is specified but LDAP is not configured"}
		}
		password, ok := opts["password"]
		if !ok {
			return dockerdriver.SafeError{SafeDescription: "LDAP username is specified but LDAP password is missing"}
		}

		uid, gid, err := m.resolver.Resolve(env, username.(string), password.(string))
		if err != nil {
			return err
		}

		opts["uid"] = uid
		opts["gid"] = gid
	}

	_, uidok := opts["uid"]
	_, gidok := opts["gid"]
	if uidok && !gidok {
		return dockerdriver.SafeError{SafeDescription: "required 'gid' option is missing"}
	}

	optsToUse, err := vmo.NewMountOpts(opts, m.mask)
	if err != nil {
		logger.Debug("mount-options-failed", lager.Data{
			"source":  remote,
			"target":  target,
			"options": opts,
		})
		return dockerdriver.SafeError{SafeDescription: err.Error()}
	}

	// check for legacy URL formatted mounts and rewrite to standard nfs format as necessary
	match := legacyNfsSharePattern.FindStringSubmatch(remote)

	if len(match) > 2 {
		if strings.TrimSpace(match[1]) == "" {
			return dockerdriver.SafeError{SafeDescription: "Invalid 'share' option"}
		}
		if match[2] == "" {
			remote = match[1] + ":/"
		} else {
			remote = match[1] + ":" + match[2]
		}
	}

	target = strings.TrimSuffix(target, "/")

	intermediateMount := target + MapfsDirectorySuffix
	orig := syscall.Umask(000)
	defer syscall.Umask(orig)
	err = m.osshim.MkdirAll(intermediateMount, os.ModePerm)
	if err != nil {
		logger.Error("mkdir-intermediate-failed", err)
		return dockerdriver.SafeError{SafeDescription: err.Error()}
	}

	cache := false
	mountOptions := m.defaultOpts

	if val, ok := opts["readonly"]; ok {
		cache, _ = strconv.ParseBool(fmt.Sprintf("%v", val))
	}

	if val, ok := opts["cache"]; ok {
		cache, err = strconv.ParseBool(fmt.Sprintf("%v", val))
		if err != nil {
			logger.Error("invalid-cache-option", err)
			return dockerdriver.SafeError{SafeDescription: "Invalid 'cache' option"}
		}
	}

	if cache == true {
		mountOptions = strings.ReplaceAll(mountOptions, ",actimeo=0", "")
	}

	if version, ok := opts["version"].(string); ok {
		versionFloat, err := strconv.ParseFloat(version, 64)
		if err != nil {
			return dockerdriver.SafeError{SafeDescription: "\"version\" must be a positive numeric value"}
		}

		if versionFloat <= 0 {
			return dockerdriver.SafeError{SafeDescription: "\"version\" must be a positive numeric value"}
		}

		mountOptions = mountOptions + ",vers=" + version
	}

	t := intermediateMount
	if !uidok {
		t = target
	}

	err = m.invoker.Invoke(env, "mount", []string{"-t", m.fstype, "-o", mountOptions, remote, t}).Wait()
	if err != nil {
		logger.Error("invoke-mount-failed", err)
		err1 := m.osshim.Remove(intermediateMount)
		if err1 != nil {
			logger.Error("remove-failed", err1)
		}
		return dockerdriver.SafeError{SafeDescription: err.Error()}
	}

	if uidok {
		// make sure the mapped user has read access to the directory before doing the mapfs mount
		// this check is best effort--root may not be able to stat the directory, or the server may
		// anonymize the owner UID.
		uid, err := strconv.Atoi(uniformData(opts["uid"]))
		if err != nil {
			return dockerdriver.SafeError{SafeDescription: InvalidUidValueErrorMessage}
		}
		if uid <= 0 {
			return dockerdriver.SafeError{SafeDescription: InvalidUidValueErrorMessage}
		}

		gid, err := strconv.Atoi(uniformData(opts["gid"]))
		if err != nil {
			return dockerdriver.SafeError{SafeDescription: InvalidGidValueErrorMessage}
		}
		if gid <= 0 {
			return dockerdriver.SafeError{SafeDescription: InvalidGidValueErrorMessage}
		}

		st := syscall.Stat_t{}
		err = m.syscallshim.Stat(intermediateMount, &st)
		if err != nil {
			logger.Error("unable-to-stat-new-mount", err)
			err = nil
		} else {
			if (st.Mode&04 == 0) &&
				((uint32(gid) != st.Gid && NobodyId != st.Gid && UnknownId != st.Gid) || st.Mode&040 == 0) &&
				((uint32(uid) != st.Uid && NobodyId != st.Uid && UnknownId != st.Uid) || st.Mode&0400 == 0) {
				err = errors.New("user lacks read access to share")
			}
		}
		if err != nil {
			logger.Error("mount-read-access-check-failed", err)
			err1 := m.invoker.Invoke(env, "umount", []string{intermediateMount}).Wait()

			if err1 != nil {
				logger.Error("intermediate-unmount-failed", err1)
			}

			if err1 == nil {
				err1 = m.osshim.Remove(intermediateMount)
				if err1 != nil {
					logger.Error("intermediate-remove-failed", err1)
				}
			}

			return dockerdriver.SafeError{SafeDescription: err.Error()}
		}

		args := mapfsOptions(optsToUse)
		args = append(args, target, intermediateMount)
		err, _ = m.backgroundInvoker.Invoke(env, m.mapfsPath, args, "Mounted!", MapfsMountTimeout)
		if err != nil {
			logger.Error("background-invoke-mount-failed", err)

			err1 := m.invoker.Invoke(env, "umount", []string{intermediateMount}).Wait()
			if err1 != nil {
				logger.Error("unmount-failed", err1)
			}

			if err1 == nil {
				err1 = m.osshim.Remove(intermediateMount)
				if err1 != nil {
					logger.Error("remove-failed", err1)
				}
			}
			return dockerdriver.SafeError{SafeDescription: err.Error()}
		}
	}

	return nil
}

func (m *mapfsMounter) Unmount(env dockerdriver.Env, target string) error {
	logger := env.Logger().Session("unmount")
	logger.Info("unmount-start")
	defer logger.Info("unmount-end")

	target = strings.TrimSuffix(target, "/")
	intermediateMount := target + MapfsDirectorySuffix

	if e := m.invoker.Invoke(env, "umount", []string{"-l", target}).Wait(); e != nil {
		return dockerdriver.SafeError{SafeDescription: e.Error()}
	}

	if exists, err := m.mountChecker.Exists(intermediateMount); exists {
		if e := m.invoker.Invoke(env, "umount", []string{"-l", intermediateMount}).Wait(); e != nil {
			logger.Error("warning-umount-intermediate-failed", e)
			return nil
		}
	} else if err != nil {
		logger.Error("warning-umount-check-intermediate-failed", err)
	}

	_, err := m.osshim.Stat(intermediateMount)
	if err == nil {
		if e := m.osshim.Remove(intermediateMount); e != nil {
			return dockerdriver.SafeError{SafeDescription: e.Error()}
		}
	}

	return nil
}

func (m *mapfsMounter) Check(env dockerdriver.Env, name, mountPoint string) bool {
	logger := env.Logger().Session("check")
	logger.Info("check-start")
	defer logger.Info("check-end")

	ctx, cancel := context.WithDeadline(context.TODO(), time.Now().Add(time.Second*5))
	defer cancel()
	env = driverhttp.EnvWithContext(ctx, env)
	err := m.invoker.Invoke(env, "mountpoint", []string{"-q", mountPoint}).Wait()

	if err != nil {
		logger.Info(fmt.Sprintf("unable to verify volume %s (%s)", name, err.Error()))
		return false
	}
	return true
}

func (m *mapfsMounter) Purge(env dockerdriver.Env, path string) {
	logger := env.Logger().Session("purge")
	logger.Info("purge-start")
	defer logger.Info("purge-end")

	invokeResult := m.invoker.Invoke(env, "pkill", []string{"mapfs"})
	err := invokeResult.Wait()
	logger.Info("pkill", lager.Data{"stdout": invokeResult.StdOutput(), "stderr": invokeResult.StdError(), "err": err})

	for i := 0; i < 30 && err == nil; i++ {
		logger.Info("waiting-for-kill")
		time.Sleep(PurgeTimeToSleep)
		invokeResult = m.invoker.Invoke(env, "pgrep", []string{"mapfs"})
		err = invokeResult.Wait()
		logger.Info("pgrep", lager.Data{"stdout": invokeResult.StdOutput(), "stderr": invokeResult.StdError(), "err": err})
	}

	mountPattern, err := regexp.Compile("^" + path + ".*" + MapfsDirectorySuffix + "$")
	if err != nil {
		logger.Error("unable-to-list-mounts", err)
		return
	}

	mounts, err := m.mountChecker.List(mountPattern)
	if err != nil {
		logger.Error("check-proc-mounts-failed", err, lager.Data{"path": path})
		return
	}

	logger.Info("mount-directory-list", lager.Data{"mounts": mounts})

	for _, mountDir := range mounts {
		realMountpoint := strings.TrimSuffix(mountDir, MapfsDirectorySuffix)

		invokeResult = m.invoker.Invoke(env, "umount", []string{"-l", "-f", realMountpoint})
		err = invokeResult.Wait()
		if err != nil {
			logger.Error("warning-umount-intermediate-failed", err)
		}

		logger.Info("unmount-successful", lager.Data{"path": realMountpoint})

		if err := m.osshim.Remove(realMountpoint); err != nil {
			logger.Error("purge-cannot-remove-directory", err, lager.Data{"name": realMountpoint, "path": path})
		}

		logger.Info("remove-directory-successful", lager.Data{"path": realMountpoint})

		invokeResult = m.invoker.Invoke(env, "umount", []string{"-l", "-f", mountDir})
		err = invokeResult.Wait()
		if err != nil {
			logger.Error("warning-umount-mapfs-failed", err)
		}

		logger.Info("unmount-successful", lager.Data{"path": mountDir})

		if err := m.osshim.Remove(mountDir); err != nil {
			logger.Error("purge-cannot-remove-directory", err, lager.Data{"name": mountDir, "path": path})
		}

		logger.Info("remove-directory-successful", lager.Data{"path": mountDir})
	}
}

func NewMapFsVolumeMountMask(allowedMountOptions string, defaultMountOptions string) (vmo.MountOptsMask, error) {
	allowed := []string{"auto_cache", "mount", "source", "experimental", "uid", "gid", "username", "password", "readonly", "version", "cache"}
	allowed = append(allowed, strings.Split(allowedMountOptions, ",")...)

	defaultMap := map[string]interface{}{}
	for _, value := range strings.Split(defaultMountOptions, ",") {
		split := strings.Split(value, ":")
		if len(split) == 2 {
			defaultMap[split[0]] = split[1]
		}
	}

	return vmo.NewMountOptsMask(
		allowed,
		defaultMap,
		nil,
		[]string{},
		[]string{},
	)

}

func uniformData(data interface{}) string {
	switch data.(type) {
	case int:
		return strconv.FormatInt(int64(data.(int)), 10)

	case string:
		return data.(string)
	}

	return ""
}

func mapfsOptions(opts vmo.MountOpts) []string {
	var ret []string
	if uid, ok := opts["uid"]; ok {
		ret = append(ret, "-uid", uniformData(uid))
	}
	if gid, ok := opts["gid"]; ok {
		ret = append(ret, "-gid", uniformData(gid))
	}
	if _, ok := opts["auto_cache"]; ok {
		ret = append(ret, "-auto_cache")
	}
	if fsname, ok := opts["fsname"]; ok {
		ret = append(ret, "-fsname", uniformData(fsname))
	}
	return ret
}
