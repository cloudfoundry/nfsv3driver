package nfsv3driver

import (
	"time"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/voldriver"
)

type mockMounter struct {
	mountTime time.Duration
	logger    lager.Logger
}

func NewMockMounter(mountTime time.Duration, logger lager.Logger) *mockMounter {
	return &mockMounter{
		mountTime: mountTime,
		logger:    logger,
	}
}

func (m *mockMounter) Mount(env voldriver.Env, source string, target string, opts map[string]interface{}) error {
	m.logger.Info("start-mocking-mount")
	defer m.logger.Info("end-mocking-mount")
	time.Sleep(m.mountTime)
	return nil
}

func (m *mockMounter) Unmount(env voldriver.Env, target string) error {
	return nil
}

func (m *mockMounter) Check(env voldriver.Env, name, mountPoint string) bool {
	//always remount
	return false
}

func (m *mockMounter) Purge(env voldriver.Env, path string) {
}
