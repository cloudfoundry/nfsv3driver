package driveradminlocal_test

import (
	"context"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/voldriver"
	"code.cloudfoundry.org/voldriver/driverhttp"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"code.cloudfoundry.org/nfsv3driver/driveradmin/driveradminlocal"
	"code.cloudfoundry.org/nfsv3driver/driveradmin"
)

var _ = Describe("Driver Admin Local", func() {
	var logger lager.Logger
	var ctx context.Context
	var env voldriver.Env
	var driverAdminLocal *driveradminlocal.DriverAdminLocal
	var err driveradmin.ErrorResponse

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("driveradminlocal")
		ctx = context.TODO()
		env = driverhttp.NewHttpDriverEnv(logger, ctx)
	})

	Context("created", func() {
		BeforeEach(func() {
			driverAdminLocal = driveradminlocal.NewDriverAdminLocal()
		})

		Describe("Evacuate", func() {
			Context("when the driver evacuates", func() {
				BeforeEach(func() {
					err = driverAdminLocal.Evacuate(env)
				})

				It("should not fail", func() {
					Expect(err.Err).To(Equal(""))
				})
			})
		})

		Describe("Ping", func() {
			Context("when the driver pings", func() {
				BeforeEach(func() {
					err = driverAdminLocal.Ping(env)
				})

				It("should not fail", func() {
					Expect(err.Err).To(Equal(""))
				})
			})
		})
	})
})
