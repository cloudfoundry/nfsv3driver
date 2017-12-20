package nfsv3driver

import (
	"code.cloudfoundry.org/voldriver"
	"code.cloudfoundry.org/goshims/execshim"
	"code.cloudfoundry.org/lager"
	"bufio"
	"strings"
	"errors"
)

//go:generate counterfeiter -o errnfsdriverfakes/fake_background_invoker.go . BackgroundInvoker

type BackgroundInvoker interface {
	Invoke(env voldriver.Env, executable string, cmdArgs []string, waitFor string) error
}

type backgroundInvoker struct {
	exec execshim.Exec
}

func NewBackgroundInvoker(useExec execshim.Exec) BackgroundInvoker {
	return &backgroundInvoker{useExec}
}

func (r *backgroundInvoker) Invoke(env voldriver.Env, executable string, cmdArgs []string, waitFor string) error {
	logger := env.Logger().Session("invoking-command", lager.Data{"executable": executable, "args": cmdArgs})
	logger.Info("start")
	defer logger.Info("end")

	// TODO--use context?
	cmdHandle := r.exec.Command(executable, cmdArgs...)
	stdout, err := cmdHandle.StdoutPipe()
	if err != nil {
		logger.Error("error-getting-pipe", err)
		return err
	}

	if err := cmdHandle.Start(); err != nil {
		logger.Error("error-starting-command", err)
		return err
	}

	if waitFor == "" {
		return nil
	}

	// wait for the process to report the string we are waiting for
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), waitFor) {
			return nil
		}
	}

	err = scanner.Err()
	if err == nil {
		err = errors.New("command exited")
	}
	logger.Error("operation failed to report success", err, lager.Data{"waitFor": waitFor})
	return err
}
