// This file was generated by counterfeiter
package nfsdriverfakes

import (
	"sync"

	"code.cloudfoundry.org/nfsv3driver/driveradmin"
	"code.cloudfoundry.org/voldriver"
)

type FakeDrainable struct {
	DrainStub        func(env voldriver.Env) error
	drainMutex       sync.RWMutex
	drainArgsForCall []struct {
		env voldriver.Env
	}
	drainReturns struct {
		result1 error
	}
}

func (fake *FakeDrainable) Drain(env voldriver.Env) error {
	fake.drainMutex.Lock()
	fake.drainArgsForCall = append(fake.drainArgsForCall, struct {
		env voldriver.Env
	}{env})
	fake.drainMutex.Unlock()
	if fake.DrainStub != nil {
		return fake.DrainStub(env)
	} else {
		return fake.drainReturns.result1
	}
}

func (fake *FakeDrainable) DrainCallCount() int {
	fake.drainMutex.RLock()
	defer fake.drainMutex.RUnlock()
	return len(fake.drainArgsForCall)
}

func (fake *FakeDrainable) DrainArgsForCall(i int) voldriver.Env {
	fake.drainMutex.RLock()
	defer fake.drainMutex.RUnlock()
	return fake.drainArgsForCall[i].env
}

func (fake *FakeDrainable) DrainReturns(result1 error) {
	fake.DrainStub = nil
	fake.drainReturns = struct {
		result1 error
	}{result1}
}

var _ driveradmin.Drainable = new(FakeDrainable)
