package driveradminlocal

import (
	"code.cloudfoundry.org/nfsv3driver/driveradmin"
	"code.cloudfoundry.org/voldriver"
)

type DriverAdminLocal struct {
}

func NewDriverAdminLocal() *DriverAdminLocal {
	d := &DriverAdminLocal{}

	return d
}

func (d *DriverAdminLocal) Evacuate(env voldriver.Env) driveradmin.ErrorResponse {
	logger := env.Logger().Session("evacuate")
	logger.Info("start")
	defer logger.Info("end")

	return driveradmin.ErrorResponse{}
}

func (d *DriverAdminLocal) Ping(env voldriver.Env) driveradmin.ErrorResponse {
	logger := env.Logger().Session("ping")
	logger.Info("start")
	defer logger.Info("end")

	return driveradmin.ErrorResponse{}
}