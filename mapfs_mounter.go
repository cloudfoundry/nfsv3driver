package nfsv3driver

import (
	"context"
	"fmt"
	"time"

	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"

	"code.cloudfoundry.org/goshims/ioutilshim"
	"code.cloudfoundry.org/goshims/osshim"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/nfsdriver"
	"code.cloudfoundry.org/voldriver"
	"code.cloudfoundry.org/voldriver/driverhttp"
	"code.cloudfoundry.org/voldriver/invoker"
)

const MAPFS_DIRECTORY_SUFFIX = "_mapfs"
const MAPFS_MOUNT_TIMEOUT = time.Minute * 5

type mapfsMounter struct {
	invoker           invoker.Invoker
	backgroundInvoker BackgroundInvoker
	v3Mounter         nfsdriver.Mounter
	osshim            osshim.Os
	ioutilshim        ioutilshim.Ioutil
	fstype            string
	defaultOpts       string
	resolver          IdResolver
	config            Config
	mapfsPath         string
}

var legacyNfsSharePattern *regexp.Regexp

func init() {
	legacyNfsSharePattern, _ = regexp.Compile("^nfs://([^/]+)(/.*)$")
}
func NewMapfsMounter(invoker invoker.Invoker, bgInvoker BackgroundInvoker, v3Mounter nfsdriver.Mounter, osshim osshim.Os, ioutilshim ioutilshim.Ioutil, fstype, defaultOpts string, resolver IdResolver, config *Config, mapfsPath string) nfsdriver.Mounter {
	return &mapfsMounter{invoker, bgInvoker, v3Mounter, osshim, ioutilshim, fstype, defaultOpts, resolver, *config, mapfsPath}
}

func (m *mapfsMounter) Mount(env voldriver.Env, remote string, target string, opts map[string]interface{}) error {
	logger := env.Logger().Session("mount")
	logger.Info("mount-start")
	defer logger.Info("mount-end")

	if _, ok := opts["experimental"]; !ok {
		return m.v3Mounter.Mount(env, remote, target, opts)
	}

	// TODO--refactor the config object so that we don't have to make a local copy just to keep
	// TODO--it from leaking information between mounts.
	tempConfig := m.config.Copy()

	if err := tempConfig.SetEntries(remote, opts, []string{
		"source", "mount", "readonly", "username", "password", "experimental", "version",
	}); err != nil {
		logger.Debug("error-parse-entries", lager.Data{
			"given_source":  remote,
			"given_target":  target,
			"given_options": opts,
			"config_source": tempConfig.source,
			"config_mounts": tempConfig.mount,
			"config_sloppy": tempConfig.sloppyMount,
		})
		return err
	}

	if username, ok := opts["username"]; ok {
		if m.resolver == nil {
			return errors.New("LDAP username is specified but LDAP is not configured")
		}
		password, ok := opts["password"]
		if !ok {
			return errors.New("LDAP username is specified but LDAP password is missing")
		}

		uid, gid, err := m.resolver.Resolve(env, username.(string), password.(string))
		if err != nil {
			return err
		}

		opts["uid"] = uid
		opts["gid"] = gid
		tempConfig.source.Allowed = append(tempConfig.source.Allowed, "uid", "gid")
		if err := tempConfig.SetEntries(remote, opts, []string{
			"source", "mount", "readonly", "username", "password", "experimental",
		}); err != nil {
			return err
		}
	}

	if _, ok := opts["uid"]; !ok {
		return errors.New("required 'uid' option is missing")
	}
	if _, ok := opts["gid"]; !ok {
		return errors.New("required 'gid' option is missing")
	}

	// check for legacy URL formatted mounts and rewrite to standard nfs format as necessary
	match := legacyNfsSharePattern.FindStringSubmatch(remote)
	if len(match) > 2 {
		remote = match[1] + ":" + match[2]
	}

	target = strings.TrimSuffix(target, "/")

	intermediateMount := target + MAPFS_DIRECTORY_SUFFIX
	orig := syscall.Umask(000)
	defer syscall.Umask(orig)
	err := m.osshim.MkdirAll(intermediateMount, os.ModePerm)
	if err != nil {
		logger.Error("mkdir-rootpath-failed", err)
		return err
	}

	mountOptions := m.defaultOpts
	if _, ok := opts["readonly"]; ok {
		mountOptions = mountOptions + ",ro"
	}

	if _, ok := opts["version"]; ok {
		mountOptions = mountOptions + ",vers=" + opts["version"].(string)
	} else {
		mountOptions = mountOptions + ",vers=3"
	}

	_, err = m.invoker.Invoke(env, "mount", []string{"-t", m.fstype, "-o", mountOptions, remote, intermediateMount})
	if err != nil {
		logger.Error("invoke-mount-failed", err)
		m.osshim.RemoveAll(intermediateMount)
		return err
	}

	args := tempConfig.MapfsOptions()
	args = append(args, target, intermediateMount)
	err = m.backgroundInvoker.Invoke(env, m.mapfsPath, args, "Mounted!", MAPFS_MOUNT_TIMEOUT)
	if err != nil {
		logger.Error("background-invoke-mount-failed", err)
		m.invoker.Invoke(env, "umount", []string{intermediateMount})
		m.osshim.Remove(intermediateMount)
		return err
	}

	return nil
}

func (m *mapfsMounter) Unmount(env voldriver.Env, target string) error {
	logger := env.Logger().Session("unmount")
	logger.Info("unmount-start")
	defer logger.Info("unmount-end")

	target = strings.TrimSuffix(target, "/")

	intermediateMount := target + MAPFS_DIRECTORY_SUFFIX
	if _, e := m.osshim.Stat(intermediateMount); e != nil {
		return m.v3Mounter.Unmount(env, target)
	}

	if _, e := m.invoker.Invoke(env, "umount", []string{target}); e != nil {
		return e
	}
	if _, e := m.invoker.Invoke(env, "umount", []string{intermediateMount}); e != nil {
		return e
	}
	if e := m.osshim.Remove(intermediateMount); e != nil {
		return e
	}

	return nil
}

func (m *mapfsMounter) Check(env voldriver.Env, name, mountPoint string) bool {
	logger := env.Logger().Session("check")
	logger.Info("check-start")
	defer logger.Info("check-end")

	ctx, _ := context.WithDeadline(context.TODO(), time.Now().Add(time.Second*5))
	env = driverhttp.EnvWithContext(ctx, env)
	_, err := m.invoker.Invoke(env, "mountpoint", []string{"-q", mountPoint})

	if err != nil {
		env.Logger().Info(fmt.Sprintf("unable to verify volume %s (%s)", name, err.Error()))
		return false
	}
	return true
}

func (m *mapfsMounter) Purge(env voldriver.Env, path string) {
	logger := env.Logger().Session("purge")
	logger.Info("purge-start")
	defer logger.Info("purge-end")

	output, err := m.invoker.Invoke(env, "pkill", []string{"mapfs"})
	logger.Info("pkill", lager.Data{"output": output, "err": err})

	for i := 0; i < 30 && err == nil; i++ {
		logger.Info("waiting-for-kill")
		time.Sleep(PurgeTimeToSleep)
		output, err = m.invoker.Invoke(env, "pgrep", []string{"mapfs"})
		logger.Info("pgrep", lager.Data{"output": output, "err": err})
	}

	fileInfos, err := m.ioutilshim.ReadDir(path)
	if err != nil {
		env.Logger().Error("purge-readdir-failed", err, lager.Data{"path": path})
		return
	}

	for _, fileInfo := range fileInfos {
		if fileInfo.IsDir() && strings.HasSuffix(fileInfo.Name(), "_mapfs") {
			realMountpoint := strings.TrimSuffix(fileInfo.Name(), "_mapfs")

			m.invoker.Invoke(env, "umount", []string{"-f", filepath.Join(path, realMountpoint)})

			if err := m.osshim.Remove(filepath.Join(path, realMountpoint)); err != nil {
				env.Logger().Error("purge-cannot-remove-directory", err, lager.Data{"name": realMountpoint, "path": path})
			}

			if err := m.osshim.Remove(filepath.Join(path, fileInfo.Name())); err != nil {
				env.Logger().Error("purge-cannot-remove-directory", err, lager.Data{"name": fileInfo.Name(), "path": path})
			}
		}
	}

	// TODO -- when we remove the legacy mounter, replace this with something that just deletes all the remaining
	// TODO -- directories
	m.v3Mounter.Purge(env, path)
}
