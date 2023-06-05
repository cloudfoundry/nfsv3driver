// Code generated by counterfeiter. DO NOT EDIT.
package dockerdriverfakes

import (
	"sync"

	"code.cloudfoundry.org/dockerdriver"
	"code.cloudfoundry.org/dockerdriver/driverhttp"
)

type FakeRemoteClientFactory struct {
	NewRemoteClientStub        func(url string, tls *dockerdriver.TLSConfig) (dockerdriver.Driver, error)
	newRemoteClientMutex       sync.RWMutex
	newRemoteClientArgsForCall []struct {
		url string
		tls *dockerdriver.TLSConfig
	}
	newRemoteClientReturns struct {
		result1 dockerdriver.Driver
		result2 error
	}
	newRemoteClientReturnsOnCall map[int]struct {
		result1 dockerdriver.Driver
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeRemoteClientFactory) NewRemoteClient(url string, tls *dockerdriver.TLSConfig) (dockerdriver.Driver, error) {
	fake.newRemoteClientMutex.Lock()
	ret, specificReturn := fake.newRemoteClientReturnsOnCall[len(fake.newRemoteClientArgsForCall)]
	fake.newRemoteClientArgsForCall = append(fake.newRemoteClientArgsForCall, struct {
		url string
		tls *dockerdriver.TLSConfig
	}{url, tls})
	fake.recordInvocation("NewRemoteClient", []interface{}{url, tls})
	fake.newRemoteClientMutex.Unlock()
	if fake.NewRemoteClientStub != nil {
		return fake.NewRemoteClientStub(url, tls)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fake.newRemoteClientReturns.result1, fake.newRemoteClientReturns.result2
}

func (fake *FakeRemoteClientFactory) NewRemoteClientCallCount() int {
	fake.newRemoteClientMutex.RLock()
	defer fake.newRemoteClientMutex.RUnlock()
	return len(fake.newRemoteClientArgsForCall)
}

func (fake *FakeRemoteClientFactory) NewRemoteClientArgsForCall(i int) (string, *dockerdriver.TLSConfig) {
	fake.newRemoteClientMutex.RLock()
	defer fake.newRemoteClientMutex.RUnlock()
	return fake.newRemoteClientArgsForCall[i].url, fake.newRemoteClientArgsForCall[i].tls
}

func (fake *FakeRemoteClientFactory) NewRemoteClientReturns(result1 dockerdriver.Driver, result2 error) {
	fake.NewRemoteClientStub = nil
	fake.newRemoteClientReturns = struct {
		result1 dockerdriver.Driver
		result2 error
	}{result1, result2}
}

func (fake *FakeRemoteClientFactory) NewRemoteClientReturnsOnCall(i int, result1 dockerdriver.Driver, result2 error) {
	fake.NewRemoteClientStub = nil
	if fake.newRemoteClientReturnsOnCall == nil {
		fake.newRemoteClientReturnsOnCall = make(map[int]struct {
			result1 dockerdriver.Driver
			result2 error
		})
	}
	fake.newRemoteClientReturnsOnCall[i] = struct {
		result1 dockerdriver.Driver
		result2 error
	}{result1, result2}
}

func (fake *FakeRemoteClientFactory) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.newRemoteClientMutex.RLock()
	defer fake.newRemoteClientMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeRemoteClientFactory) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ driverhttp.RemoteClientFactory = new(FakeRemoteClientFactory)