package nfsv3driver

import (
	"context"
	"fmt"
	"time"

	"code.cloudfoundry.org/goshims/ioutilshim"
	"code.cloudfoundry.org/goshims/osshim"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/nfsdriver"
	"code.cloudfoundry.org/voldriver"
	"code.cloudfoundry.org/voldriver/driverhttp"
	"code.cloudfoundry.org/voldriver/invoker"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

const MAPFS_DIRECTORY_SUFFIX = "_mapfs"

type mapfsMounter struct {
	invoker     invoker.Invoker
	v3Mounter   nfsdriver.Mounter
	osshim      osshim.Os
	ioutilshim  ioutilshim.Ioutil
	fstype      string
	defaultOpts string
}

func NewMapfsMounter(invoker invoker.Invoker, v3Mounter nfsdriver.Mounter, osshim osshim.Os, ioutilshim ioutilshim.Ioutil, fstype, defaultOpts string) nfsdriver.Mounter {
	return &mapfsMounter{invoker, v3Mounter, osshim, ioutilshim, fstype, defaultOpts}
}

func (m *mapfsMounter) Mount(env voldriver.Env, remote string, target string, opts map[string]interface{}) error {
	logger := env.Logger().Session("mount")
	logger.Info("mount-start")
	defer logger.Info("mount-end")

	if opts["experimental"] == nil {
		return m.v3Mounter.Mount(env, remote, target, opts)
	}

	intermediateMount := target + MAPFS_DIRECTORY_SUFFIX
	orig := syscall.Umask(000)
	defer syscall.Umask(orig)
	err := m.osshim.MkdirAll(intermediateMount, os.ModePerm);
	if err != nil {
		logger.Error("mkdir-rootpath-failed", err)
		return err
	}

	_, err = m.invoker.Invoke(env, "mount", []string{"-t", m.fstype, "-o", m.defaultOpts, remote, intermediateMount})
	if err != nil {
		logger.Error("invoke-mount-failed", err)
		return err
	}

	return nil
}

func (m *mapfsMounter) Unmount(env voldriver.Env, target string) error {
	logger := env.Logger().Session("unmount")
	logger.Info("unmount-start")
	defer logger.Info("unmount-end")

	var err error
	_, err = m.osshim.Stat(target + "_mapfs")
	if err == nil {
		_, err = m.invoker.Invoke(env, "umount", []string{target})
	} else {
		err = m.v3Mounter.Unmount(env, target)
	}
	return err
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
		time.Sleep(time.Millisecond * 1) // TODO!!!!!!
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

			if err := m.osshim.RemoveAll(filepath.Join(path, realMountpoint)); err != nil {
				env.Logger().Error("purge-cannot-remove-directory", err, lager.Data{"name": realMountpoint, "path": path})
			}

			if err := m.osshim.RemoveAll(filepath.Join(path, fileInfo.Name())); err != nil {
				env.Logger().Error("purge-cannot-remove-directory", err, lager.Data{"name": fileInfo.Name(), "path": path})
			}
		}
	}

	// TODO -- when we remove this, replace it with something that just deletes all the remaining directories
	m.v3Mounter.Purge(env, path)
}
