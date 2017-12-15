package driveradmin

import (
	"code.cloudfoundry.org/voldriver"
	"github.com/tedsuo/rata"
)

const (
	EvacuateRoute = "evacuate"
	PingRoute     = "ping"
)

var Routes = rata.Routes{
	{Path: "/evacuate", Method: "GET", Name: EvacuateRoute},
	{Path: "/ping", Method: "GET", Name: PingRoute},
}

//go:generate counterfeiter -o ../nfsdriverfakes/fake_driver_admin.go . DriverAdmin

type DriverAdmin interface {
	Evacuate(env voldriver.Env) ErrorResponse
	Ping(env voldriver.Env) ErrorResponse
}

type ErrorResponse struct {
	Err string
}

//go:generate counterfeiter -o ../nfsdriverfakes/fake_drainable.go . Drainable
type Drainable interface {
	Drain(env voldriver.Env) error
}
