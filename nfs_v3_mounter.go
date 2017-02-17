package nfsv3driver

import (
	"context"
	"fmt"
	"time"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/nfsdriver"
	"code.cloudfoundry.org/voldriver"
	"code.cloudfoundry.org/voldriver/driverhttp"
	"code.cloudfoundry.org/voldriver/invoker"
	"strings"
)

type nfsV3Mounter struct {
	invoker invoker.Invoker
	config  Config
}

func NewNfsV3Mounter(invoker invoker.Invoker, config *Config) nfsdriver.Mounter {
	return &nfsV3Mounter{invoker: invoker, config: *config}
}

func (m *nfsV3Mounter) Mount(env voldriver.Env, source string, target string, opts map[string]interface{}) error {
	logger := env.Logger().Session("fuse-nfs-mount")
	logger.Info("start")
	defer logger.Info("end")

	if err := m.config.SetEntries(source, opts, []string{
		"share", "mount", "kerberosPrincipal", "kerberosKeytab", "readonly",
	}); err != nil {
		logger.Debug("error-parse-entries", lager.Data{
			"given_source":  source,
			"given_target":  target,
			"given_options": opts,
			"config_source": m.config.source,
			"config_mounts": m.config.mount,
			"config_sloppy": m.config.sloppyMount,
		})
		return err
	}

	mountOptions := append([]string{
		"-n", m.config.Share(source),
		"-m", target,
	}, m.config.Mount()...)

	logger.Debug("parse-mount", lager.Data{
		"given_source":  source,
		"given_target":  target,
		"given_options": opts,
		"config_source": m.config.source,
		"config_mounts": m.config.mount,
		"config_sloppy": m.config.sloppyMount,
		"mountOptions":  mountOptions,
	})

	logger.Debug("exec-mount", lager.Data{"params": strings.Join(mountOptions, ",")})
	_, err := m.invoker.Invoke(env, "fuse-nfs", mountOptions)
	return err
}

func (m *nfsV3Mounter) Unmount(env voldriver.Env, target string) error {
	_, err := m.invoker.Invoke(env, "fusermount", []string{"-u", target})
	return err
}

func (m *nfsV3Mounter) Check(env voldriver.Env, name, mountPoint string) bool {
	ctx, _ := context.WithDeadline(context.TODO(), time.Now().Add(time.Second*5))
	env = driverhttp.EnvWithContext(ctx, env)
	_, err := m.invoker.Invoke(env, "mountpoint", []string{"-q", mountPoint})

	if err != nil {
		// Note: Created volumes (with no mounts) will be removed
		//       since VolumeInfo.Mountpoint will be an empty string
		env.Logger().Info(fmt.Sprintf("unable to verify volume %s (%s)", name, err.Error()))
		return false
	}
	return true
}
