// This file was generated by counterfeiter
package nfsdriverfakes

import (
	"sync"

	"code.cloudfoundry.org/nfsv3driver"
	"code.cloudfoundry.org/voldriver"
)

type FakeIdResolver struct {
	ResolveStub        func(env voldriver.Env, username string, password string) (uid string, gid string, err error)
	resolveMutex       sync.RWMutex
	resolveArgsForCall []struct {
		env      voldriver.Env
		username string
		password string
	}
	resolveReturns struct {
		result1 string
		result2 string
		result3 error
	}
}

func (fake *FakeIdResolver) Resolve(env voldriver.Env, username string, password string) (uid string, gid string, err error) {
	fake.resolveMutex.Lock()
	fake.resolveArgsForCall = append(fake.resolveArgsForCall, struct {
		env      voldriver.Env
		username string
		password string
	}{env, username, password})
	fake.resolveMutex.Unlock()
	if fake.ResolveStub != nil {
		return fake.ResolveStub(env, username, password)
	} else {
		return fake.resolveReturns.result1, fake.resolveReturns.result2, fake.resolveReturns.result3
	}
}

func (fake *FakeIdResolver) ResolveCallCount() int {
	fake.resolveMutex.RLock()
	defer fake.resolveMutex.RUnlock()
	return len(fake.resolveArgsForCall)
}

func (fake *FakeIdResolver) ResolveArgsForCall(i int) (voldriver.Env, string, string) {
	fake.resolveMutex.RLock()
	defer fake.resolveMutex.RUnlock()
	return fake.resolveArgsForCall[i].env, fake.resolveArgsForCall[i].username, fake.resolveArgsForCall[i].password
}

func (fake *FakeIdResolver) ResolveReturns(result1 string, result2 string, result3 error) {
	fake.ResolveStub = nil
	fake.resolveReturns = struct {
		result1 string
		result2 string
		result3 error
	}{result1, result2, result3}
}

var _ nfsv3driver.IdResolver = new(FakeIdResolver)
