package nfsv3driver

import (
	"context"
	"errors"
	"fmt"
	"time"

	"strings"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/nfsdriver"
	"code.cloudfoundry.org/voldriver"
	"code.cloudfoundry.org/voldriver/driverhttp"
	"code.cloudfoundry.org/voldriver/invoker"
)

type nfsV3Mounter struct {
	invoker  invoker.Invoker
	config   Config
	resolver IdResolver
}

func NewNfsV3Mounter(invoker invoker.Invoker, config *Config, resolver IdResolver) nfsdriver.Mounter {
	return &nfsV3Mounter{invoker: invoker, config: *config, resolver: resolver}
}

func (m *nfsV3Mounter) Mount(env voldriver.Env, source string, target string, opts map[string]interface{}) error {
	logger := env.Logger().Session("fuse-nfs-mount")
	logger.Info("start")
	defer logger.Info("end")

	// TODO--refactor the config object so that we don't have to make a local copy just to keep
	// TODO--it from leaking information between mounts.
  tempConfig := m.config.Copy()

	if err := tempConfig.SetEntries(source, opts, []string{
		"source", "mount", "kerberosPrincipal", "kerberosKeytab", "readonly", "username", "password",
	}); err != nil {
		logger.Debug("error-parse-entries", lager.Data{
			"given_source":  source,
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

		tempConfig.source.Allowed = append(tempConfig.source.Allowed, "uid", "gid")

		uid, gid, err := m.resolver.Resolve(env, username.(string), password.(string))
		if err != nil {
			return err
		}

		opts["uid"] = uid
		opts["gid"] = gid
		err = tempConfig.SetEntries(source, opts, []string{
			"source", "mount", "kerberosPrincipal", "kerberosKeytab", "readonly", "username", "password",
		})
	  if err != nil {
			return err
		}
	}

	mountOptions := append([]string{
		"-a",
		"-n", tempConfig.Share(source),
		"-m", target,
	}, tempConfig.Mount()...)

	if _, ok := opts["readonly"]; ok {
		mountOptions = append(mountOptions, "-O")
	}

	logger.Debug("parse-mount", lager.Data{
		"given_source":  source,
		"given_target":  target,
		"given_options": opts,
		"config_source": tempConfig.source,
		"config_mounts": tempConfig.mount,
		"config_sloppy": tempConfig.sloppyMount,
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

func (m *nfsV3Mounter) Purge(env voldriver.Env, path string) {
	output, err := m.invoker.Invoke(env, "pkill", []string{"-f", "fuse-nfs"})
	env.Logger().Info("purge", lager.Data{"output": output, "err": err})
}

