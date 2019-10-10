// Code generated by counterfeiter. DO NOT EDIT.
package dockerdriverfakes

import (
	"sync"

	"code.cloudfoundry.org/dockerdriver"
)

type FakeDriver struct {
	ActivateStub        func(env dockerdriver.Env) dockerdriver.ActivateResponse
	activateMutex       sync.RWMutex
	activateArgsForCall []struct {
		env dockerdriver.Env
	}
	activateReturns struct {
		result1 dockerdriver.ActivateResponse
	}
	activateReturnsOnCall map[int]struct {
		result1 dockerdriver.ActivateResponse
	}
	GetStub        func(env dockerdriver.Env, getRequest dockerdriver.GetRequest) dockerdriver.GetResponse
	getMutex       sync.RWMutex
	getArgsForCall []struct {
		env        dockerdriver.Env
		getRequest dockerdriver.GetRequest
	}
	getReturns struct {
		result1 dockerdriver.GetResponse
	}
	getReturnsOnCall map[int]struct {
		result1 dockerdriver.GetResponse
	}
	ListStub        func(env dockerdriver.Env) dockerdriver.ListResponse
	listMutex       sync.RWMutex
	listArgsForCall []struct {
		env dockerdriver.Env
	}
	listReturns struct {
		result1 dockerdriver.ListResponse
	}
	listReturnsOnCall map[int]struct {
		result1 dockerdriver.ListResponse
	}
	MountStub        func(env dockerdriver.Env, mountRequest dockerdriver.MountRequest) dockerdriver.MountResponse
	mountMutex       sync.RWMutex
	mountArgsForCall []struct {
		env          dockerdriver.Env
		mountRequest dockerdriver.MountRequest
	}
	mountReturns struct {
		result1 dockerdriver.MountResponse
	}
	mountReturnsOnCall map[int]struct {
		result1 dockerdriver.MountResponse
	}
	PathStub        func(env dockerdriver.Env, pathRequest dockerdriver.PathRequest) dockerdriver.PathResponse
	pathMutex       sync.RWMutex
	pathArgsForCall []struct {
		env         dockerdriver.Env
		pathRequest dockerdriver.PathRequest
	}
	pathReturns struct {
		result1 dockerdriver.PathResponse
	}
	pathReturnsOnCall map[int]struct {
		result1 dockerdriver.PathResponse
	}
	UnmountStub        func(env dockerdriver.Env, unmountRequest dockerdriver.UnmountRequest) dockerdriver.ErrorResponse
	unmountMutex       sync.RWMutex
	unmountArgsForCall []struct {
		env            dockerdriver.Env
		unmountRequest dockerdriver.UnmountRequest
	}
	unmountReturns struct {
		result1 dockerdriver.ErrorResponse
	}
	unmountReturnsOnCall map[int]struct {
		result1 dockerdriver.ErrorResponse
	}
	CapabilitiesStub        func(env dockerdriver.Env) dockerdriver.CapabilitiesResponse
	capabilitiesMutex       sync.RWMutex
	capabilitiesArgsForCall []struct {
		env dockerdriver.Env
	}
	capabilitiesReturns struct {
		result1 dockerdriver.CapabilitiesResponse
	}
	capabilitiesReturnsOnCall map[int]struct {
		result1 dockerdriver.CapabilitiesResponse
	}
	CreateStub        func(env dockerdriver.Env, createRequest dockerdriver.CreateRequest) dockerdriver.ErrorResponse
	createMutex       sync.RWMutex
	createArgsForCall []struct {
		env           dockerdriver.Env
		createRequest dockerdriver.CreateRequest
	}
	createReturns struct {
		result1 dockerdriver.ErrorResponse
	}
	createReturnsOnCall map[int]struct {
		result1 dockerdriver.ErrorResponse
	}
	RemoveStub        func(env dockerdriver.Env, removeRequest dockerdriver.RemoveRequest) dockerdriver.ErrorResponse
	removeMutex       sync.RWMutex
	removeArgsForCall []struct {
		env           dockerdriver.Env
		removeRequest dockerdriver.RemoveRequest
	}
	removeReturns struct {
		result1 dockerdriver.ErrorResponse
	}
	removeReturnsOnCall map[int]struct {
		result1 dockerdriver.ErrorResponse
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeDriver) Activate(env dockerdriver.Env) dockerdriver.ActivateResponse {
	fake.activateMutex.Lock()
	ret, specificReturn := fake.activateReturnsOnCall[len(fake.activateArgsForCall)]
	fake.activateArgsForCall = append(fake.activateArgsForCall, struct {
		env dockerdriver.Env
	}{env})
	fake.recordInvocation("Activate", []interface{}{env})
	fake.activateMutex.Unlock()
	if fake.ActivateStub != nil {
		return fake.ActivateStub(env)
	}
	if specificReturn {
		return ret.result1
	}
	return fake.activateReturns.result1
}

func (fake *FakeDriver) ActivateCallCount() int {
	fake.activateMutex.RLock()
	defer fake.activateMutex.RUnlock()
	return len(fake.activateArgsForCall)
}

func (fake *FakeDriver) ActivateArgsForCall(i int) dockerdriver.Env {
	fake.activateMutex.RLock()
	defer fake.activateMutex.RUnlock()
	return fake.activateArgsForCall[i].env
}

func (fake *FakeDriver) ActivateReturns(result1 dockerdriver.ActivateResponse) {
	fake.ActivateStub = nil
	fake.activateReturns = struct {
		result1 dockerdriver.ActivateResponse
	}{result1}
}

func (fake *FakeDriver) ActivateReturnsOnCall(i int, result1 dockerdriver.ActivateResponse) {
	fake.ActivateStub = nil
	if fake.activateReturnsOnCall == nil {
		fake.activateReturnsOnCall = make(map[int]struct {
			result1 dockerdriver.ActivateResponse
		})
	}
	fake.activateReturnsOnCall[i] = struct {
		result1 dockerdriver.ActivateResponse
	}{result1}
}

func (fake *FakeDriver) Get(env dockerdriver.Env, getRequest dockerdriver.GetRequest) dockerdriver.GetResponse {
	fake.getMutex.Lock()
	ret, specificReturn := fake.getReturnsOnCall[len(fake.getArgsForCall)]
	fake.getArgsForCall = append(fake.getArgsForCall, struct {
		env        dockerdriver.Env
		getRequest dockerdriver.GetRequest
	}{env, getRequest})
	fake.recordInvocation("Get", []interface{}{env, getRequest})
	fake.getMutex.Unlock()
	if fake.GetStub != nil {
		return fake.GetStub(env, getRequest)
	}
	if specificReturn {
		return ret.result1
	}
	return fake.getReturns.result1
}

func (fake *FakeDriver) GetCallCount() int {
	fake.getMutex.RLock()
	defer fake.getMutex.RUnlock()
	return len(fake.getArgsForCall)
}

func (fake *FakeDriver) GetArgsForCall(i int) (dockerdriver.Env, dockerdriver.GetRequest) {
	fake.getMutex.RLock()
	defer fake.getMutex.RUnlock()
	return fake.getArgsForCall[i].env, fake.getArgsForCall[i].getRequest
}

func (fake *FakeDriver) GetReturns(result1 dockerdriver.GetResponse) {
	fake.GetStub = nil
	fake.getReturns = struct {
		result1 dockerdriver.GetResponse
	}{result1}
}

func (fake *FakeDriver) GetReturnsOnCall(i int, result1 dockerdriver.GetResponse) {
	fake.GetStub = nil
	if fake.getReturnsOnCall == nil {
		fake.getReturnsOnCall = make(map[int]struct {
			result1 dockerdriver.GetResponse
		})
	}
	fake.getReturnsOnCall[i] = struct {
		result1 dockerdriver.GetResponse
	}{result1}
}

func (fake *FakeDriver) List(env dockerdriver.Env) dockerdriver.ListResponse {
	fake.listMutex.Lock()
	ret, specificReturn := fake.listReturnsOnCall[len(fake.listArgsForCall)]
	fake.listArgsForCall = append(fake.listArgsForCall, struct {
		env dockerdriver.Env
	}{env})
	fake.recordInvocation("List", []interface{}{env})
	fake.listMutex.Unlock()
	if fake.ListStub != nil {
		return fake.ListStub(env)
	}
	if specificReturn {
		return ret.result1
	}
	return fake.listReturns.result1
}

func (fake *FakeDriver) ListCallCount() int {
	fake.listMutex.RLock()
	defer fake.listMutex.RUnlock()
	return len(fake.listArgsForCall)
}

func (fake *FakeDriver) ListArgsForCall(i int) dockerdriver.Env {
	fake.listMutex.RLock()
	defer fake.listMutex.RUnlock()
	return fake.listArgsForCall[i].env
}

func (fake *FakeDriver) ListReturns(result1 dockerdriver.ListResponse) {
	fake.ListStub = nil
	fake.listReturns = struct {
		result1 dockerdriver.ListResponse
	}{result1}
}

func (fake *FakeDriver) ListReturnsOnCall(i int, result1 dockerdriver.ListResponse) {
	fake.ListStub = nil
	if fake.listReturnsOnCall == nil {
		fake.listReturnsOnCall = make(map[int]struct {
			result1 dockerdriver.ListResponse
		})
	}
	fake.listReturnsOnCall[i] = struct {
		result1 dockerdriver.ListResponse
	}{result1}
}

func (fake *FakeDriver) Mount(env dockerdriver.Env, mountRequest dockerdriver.MountRequest) dockerdriver.MountResponse {
	fake.mountMutex.Lock()
	ret, specificReturn := fake.mountReturnsOnCall[len(fake.mountArgsForCall)]
	fake.mountArgsForCall = append(fake.mountArgsForCall, struct {
		env          dockerdriver.Env
		mountRequest dockerdriver.MountRequest
	}{env, mountRequest})
	fake.recordInvocation("Mount", []interface{}{env, mountRequest})
	fake.mountMutex.Unlock()
	if fake.MountStub != nil {
		return fake.MountStub(env, mountRequest)
	}
	if specificReturn {
		return ret.result1
	}
	return fake.mountReturns.result1
}

func (fake *FakeDriver) MountCallCount() int {
	fake.mountMutex.RLock()
	defer fake.mountMutex.RUnlock()
	return len(fake.mountArgsForCall)
}

func (fake *FakeDriver) MountArgsForCall(i int) (dockerdriver.Env, dockerdriver.MountRequest) {
	fake.mountMutex.RLock()
	defer fake.mountMutex.RUnlock()
	return fake.mountArgsForCall[i].env, fake.mountArgsForCall[i].mountRequest
}

func (fake *FakeDriver) MountReturns(result1 dockerdriver.MountResponse) {
	fake.MountStub = nil
	fake.mountReturns = struct {
		result1 dockerdriver.MountResponse
	}{result1}
}

func (fake *FakeDriver) MountReturnsOnCall(i int, result1 dockerdriver.MountResponse) {
	fake.MountStub = nil
	if fake.mountReturnsOnCall == nil {
		fake.mountReturnsOnCall = make(map[int]struct {
			result1 dockerdriver.MountResponse
		})
	}
	fake.mountReturnsOnCall[i] = struct {
		result1 dockerdriver.MountResponse
	}{result1}
}

func (fake *FakeDriver) Path(env dockerdriver.Env, pathRequest dockerdriver.PathRequest) dockerdriver.PathResponse {
	fake.pathMutex.Lock()
	ret, specificReturn := fake.pathReturnsOnCall[len(fake.pathArgsForCall)]
	fake.pathArgsForCall = append(fake.pathArgsForCall, struct {
		env         dockerdriver.Env
		pathRequest dockerdriver.PathRequest
	}{env, pathRequest})
	fake.recordInvocation("Path", []interface{}{env, pathRequest})
	fake.pathMutex.Unlock()
	if fake.PathStub != nil {
		return fake.PathStub(env, pathRequest)
	}
	if specificReturn {
		return ret.result1
	}
	return fake.pathReturns.result1
}

func (fake *FakeDriver) PathCallCount() int {
	fake.pathMutex.RLock()
	defer fake.pathMutex.RUnlock()
	return len(fake.pathArgsForCall)
}

func (fake *FakeDriver) PathArgsForCall(i int) (dockerdriver.Env, dockerdriver.PathRequest) {
	fake.pathMutex.RLock()
	defer fake.pathMutex.RUnlock()
	return fake.pathArgsForCall[i].env, fake.pathArgsForCall[i].pathRequest
}

func (fake *FakeDriver) PathReturns(result1 dockerdriver.PathResponse) {
	fake.PathStub = nil
	fake.pathReturns = struct {
		result1 dockerdriver.PathResponse
	}{result1}
}

func (fake *FakeDriver) PathReturnsOnCall(i int, result1 dockerdriver.PathResponse) {
	fake.PathStub = nil
	if fake.pathReturnsOnCall == nil {
		fake.pathReturnsOnCall = make(map[int]struct {
			result1 dockerdriver.PathResponse
		})
	}
	fake.pathReturnsOnCall[i] = struct {
		result1 dockerdriver.PathResponse
	}{result1}
}

func (fake *FakeDriver) Unmount(env dockerdriver.Env, unmountRequest dockerdriver.UnmountRequest) dockerdriver.ErrorResponse {
	fake.unmountMutex.Lock()
	ret, specificReturn := fake.unmountReturnsOnCall[len(fake.unmountArgsForCall)]
	fake.unmountArgsForCall = append(fake.unmountArgsForCall, struct {
		env            dockerdriver.Env
		unmountRequest dockerdriver.UnmountRequest
	}{env, unmountRequest})
	fake.recordInvocation("Unmount", []interface{}{env, unmountRequest})
	fake.unmountMutex.Unlock()
	if fake.UnmountStub != nil {
		return fake.UnmountStub(env, unmountRequest)
	}
	if specificReturn {
		return ret.result1
	}
	return fake.unmountReturns.result1
}

func (fake *FakeDriver) UnmountCallCount() int {
	fake.unmountMutex.RLock()
	defer fake.unmountMutex.RUnlock()
	return len(fake.unmountArgsForCall)
}

func (fake *FakeDriver) UnmountArgsForCall(i int) (dockerdriver.Env, dockerdriver.UnmountRequest) {
	fake.unmountMutex.RLock()
	defer fake.unmountMutex.RUnlock()
	return fake.unmountArgsForCall[i].env, fake.unmountArgsForCall[i].unmountRequest
}

func (fake *FakeDriver) UnmountReturns(result1 dockerdriver.ErrorResponse) {
	fake.UnmountStub = nil
	fake.unmountReturns = struct {
		result1 dockerdriver.ErrorResponse
	}{result1}
}

func (fake *FakeDriver) UnmountReturnsOnCall(i int, result1 dockerdriver.ErrorResponse) {
	fake.UnmountStub = nil
	if fake.unmountReturnsOnCall == nil {
		fake.unmountReturnsOnCall = make(map[int]struct {
			result1 dockerdriver.ErrorResponse
		})
	}
	fake.unmountReturnsOnCall[i] = struct {
		result1 dockerdriver.ErrorResponse
	}{result1}
}

func (fake *FakeDriver) Capabilities(env dockerdriver.Env) dockerdriver.CapabilitiesResponse {
	fake.capabilitiesMutex.Lock()
	ret, specificReturn := fake.capabilitiesReturnsOnCall[len(fake.capabilitiesArgsForCall)]
	fake.capabilitiesArgsForCall = append(fake.capabilitiesArgsForCall, struct {
		env dockerdriver.Env
	}{env})
	fake.recordInvocation("Capabilities", []interface{}{env})
	fake.capabilitiesMutex.Unlock()
	if fake.CapabilitiesStub != nil {
		return fake.CapabilitiesStub(env)
	}
	if specificReturn {
		return ret.result1
	}
	return fake.capabilitiesReturns.result1
}

func (fake *FakeDriver) CapabilitiesCallCount() int {
	fake.capabilitiesMutex.RLock()
	defer fake.capabilitiesMutex.RUnlock()
	return len(fake.capabilitiesArgsForCall)
}

func (fake *FakeDriver) CapabilitiesArgsForCall(i int) dockerdriver.Env {
	fake.capabilitiesMutex.RLock()
	defer fake.capabilitiesMutex.RUnlock()
	return fake.capabilitiesArgsForCall[i].env
}

func (fake *FakeDriver) CapabilitiesReturns(result1 dockerdriver.CapabilitiesResponse) {
	fake.CapabilitiesStub = nil
	fake.capabilitiesReturns = struct {
		result1 dockerdriver.CapabilitiesResponse
	}{result1}
}

func (fake *FakeDriver) CapabilitiesReturnsOnCall(i int, result1 dockerdriver.CapabilitiesResponse) {
	fake.CapabilitiesStub = nil
	if fake.capabilitiesReturnsOnCall == nil {
		fake.capabilitiesReturnsOnCall = make(map[int]struct {
			result1 dockerdriver.CapabilitiesResponse
		})
	}
	fake.capabilitiesReturnsOnCall[i] = struct {
		result1 dockerdriver.CapabilitiesResponse
	}{result1}
}

func (fake *FakeDriver) Create(env dockerdriver.Env, createRequest dockerdriver.CreateRequest) dockerdriver.ErrorResponse {
	fake.createMutex.Lock()
	ret, specificReturn := fake.createReturnsOnCall[len(fake.createArgsForCall)]
	fake.createArgsForCall = append(fake.createArgsForCall, struct {
		env           dockerdriver.Env
		createRequest dockerdriver.CreateRequest
	}{env, createRequest})
	fake.recordInvocation("Create", []interface{}{env, createRequest})
	fake.createMutex.Unlock()
	if fake.CreateStub != nil {
		return fake.CreateStub(env, createRequest)
	}
	if specificReturn {
		return ret.result1
	}
	return fake.createReturns.result1
}

func (fake *FakeDriver) CreateCallCount() int {
	fake.createMutex.RLock()
	defer fake.createMutex.RUnlock()
	return len(fake.createArgsForCall)
}

func (fake *FakeDriver) CreateArgsForCall(i int) (dockerdriver.Env, dockerdriver.CreateRequest) {
	fake.createMutex.RLock()
	defer fake.createMutex.RUnlock()
	return fake.createArgsForCall[i].env, fake.createArgsForCall[i].createRequest
}

func (fake *FakeDriver) CreateReturns(result1 dockerdriver.ErrorResponse) {
	fake.CreateStub = nil
	fake.createReturns = struct {
		result1 dockerdriver.ErrorResponse
	}{result1}
}

func (fake *FakeDriver) CreateReturnsOnCall(i int, result1 dockerdriver.ErrorResponse) {
	fake.CreateStub = nil
	if fake.createReturnsOnCall == nil {
		fake.createReturnsOnCall = make(map[int]struct {
			result1 dockerdriver.ErrorResponse
		})
	}
	fake.createReturnsOnCall[i] = struct {
		result1 dockerdriver.ErrorResponse
	}{result1}
}

func (fake *FakeDriver) Remove(env dockerdriver.Env, removeRequest dockerdriver.RemoveRequest) dockerdriver.ErrorResponse {
	fake.removeMutex.Lock()
	ret, specificReturn := fake.removeReturnsOnCall[len(fake.removeArgsForCall)]
	fake.removeArgsForCall = append(fake.removeArgsForCall, struct {
		env           dockerdriver.Env
		removeRequest dockerdriver.RemoveRequest
	}{env, removeRequest})
	fake.recordInvocation("Remove", []interface{}{env, removeRequest})
	fake.removeMutex.Unlock()
	if fake.RemoveStub != nil {
		return fake.RemoveStub(env, removeRequest)
	}
	if specificReturn {
		return ret.result1
	}
	return fake.removeReturns.result1
}

func (fake *FakeDriver) RemoveCallCount() int {
	fake.removeMutex.RLock()
	defer fake.removeMutex.RUnlock()
	return len(fake.removeArgsForCall)
}

func (fake *FakeDriver) RemoveArgsForCall(i int) (dockerdriver.Env, dockerdriver.RemoveRequest) {
	fake.removeMutex.RLock()
	defer fake.removeMutex.RUnlock()
	return fake.removeArgsForCall[i].env, fake.removeArgsForCall[i].removeRequest
}

func (fake *FakeDriver) RemoveReturns(result1 dockerdriver.ErrorResponse) {
	fake.RemoveStub = nil
	fake.removeReturns = struct {
		result1 dockerdriver.ErrorResponse
	}{result1}
}

func (fake *FakeDriver) RemoveReturnsOnCall(i int, result1 dockerdriver.ErrorResponse) {
	fake.RemoveStub = nil
	if fake.removeReturnsOnCall == nil {
		fake.removeReturnsOnCall = make(map[int]struct {
			result1 dockerdriver.ErrorResponse
		})
	}
	fake.removeReturnsOnCall[i] = struct {
		result1 dockerdriver.ErrorResponse
	}{result1}
}

func (fake *FakeDriver) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.activateMutex.RLock()
	defer fake.activateMutex.RUnlock()
	fake.getMutex.RLock()
	defer fake.getMutex.RUnlock()
	fake.listMutex.RLock()
	defer fake.listMutex.RUnlock()
	fake.mountMutex.RLock()
	defer fake.mountMutex.RUnlock()
	fake.pathMutex.RLock()
	defer fake.pathMutex.RUnlock()
	fake.unmountMutex.RLock()
	defer fake.unmountMutex.RUnlock()
	fake.capabilitiesMutex.RLock()
	defer fake.capabilitiesMutex.RUnlock()
	fake.createMutex.RLock()
	defer fake.createMutex.RUnlock()
	fake.removeMutex.RLock()
	defer fake.removeMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeDriver) recordInvocation(key string, args []interface{}) {
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

var _ dockerdriver.Driver = new(FakeDriver)
