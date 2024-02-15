package nfsv3driver

import (
	"bufio"
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"code.cloudfoundry.org/dockerdriver"
	"code.cloudfoundry.org/goshims/execshim"
	"code.cloudfoundry.org/lager/v3"
)

//go:generate counterfeiter -o nfsdriverfakes/fake_background_invoker.go . BackgroundInvoker

type BackgroundInvoker interface {
	Invoke(env dockerdriver.Env, executable string, cmdArgs []string, waitFor string, timeout time.Duration) (error, context.CancelFunc)
}

type backgroundInvoker struct {
	exec execshim.Exec
}

func NewBackgroundInvoker(useExec execshim.Exec) BackgroundInvoker {
	return &backgroundInvoker{useExec}
}

func (r *backgroundInvoker) Invoke(env dockerdriver.Env, executable string, cmdArgs []string, waitFor string, timeout time.Duration) (error, context.CancelFunc) {
	logger := env.Logger().Session("invoking-command", lager.Data{"executable": executable, "args": cmdArgs})
	logger.Info("start")
	defer logger.Info("end")

	ctx, cancel := context.WithCancel(context.Background())
	cmdHandle := r.exec.CommandContext(ctx, executable, cmdArgs...)
	stdout, err := cmdHandle.StdoutPipe()
	if err != nil {
		cancel()
		logger.Error("error-getting-pipe", err)
		return err, nil
	}

	if err := cmdHandle.Start(); err != nil {
		logger.Error("error-starting-command", err)
		cancel()
		return err, nil
	}
	defer func() { go cmdHandle.Wait() }()

	var mutex sync.Mutex
	cancelled := false
	timer := time.AfterFunc(timeout, func() {
		mutex.Lock()
		defer mutex.Unlock()
		cancelled = true
		cancel()
	})

	// wait for the process to report the string we are waiting for
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), waitFor) {
			timer.Stop()
			return nil, cancel
		}
	}

	err = scanner.Err()
	if err == nil {
		mutex.Lock()
		c := cancelled
		mutex.Unlock()
		if c {
			err = errors.New("command timed out")
		} else {
			err = errors.New("command exited")
		}
	}

	timer.Stop()
	logger.Error("operation failed to report success", err, lager.Data{"waitFor": waitFor})
	cancel()
	return err, nil
}
