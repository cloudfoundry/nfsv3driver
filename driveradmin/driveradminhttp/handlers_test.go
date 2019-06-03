package driveradminhttp_test

import (
	"github.com/tedsuo/rata"
	"net/http"
	"net/http/httptest"

	"fmt"

	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/nfsv3driver/driveradmin"
	"code.cloudfoundry.org/nfsv3driver/driveradmin/driveradminhttp"
	"code.cloudfoundry.org/nfsv3driver/nfsdriverfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Volman Driver Handlers", func() {

	Context("when generating http handlers", func() {
		var (
			testLogger = lagertest.NewTestLogger("HandlersTest")
			driverAdmin = &nfsdriverfakes.FakeDriverAdmin{}
			handler http.Handler
			httpRequest *http.Request
			httpResponseRecorder *httptest.ResponseRecorder
			route rata.Route

		)

		BeforeEach(func() {
			var err error
			handler, err = driveradminhttp.NewHandler(testLogger, driverAdmin)
			Expect(err).NotTo(HaveOccurred())
		})

		JustBeforeEach(func() {
			var err error
			path := fmt.Sprintf("http://0.0.0.0%s", route.Path)
			httpRequest, err = http.NewRequest("GET", path, nil)
			Expect(err).NotTo(HaveOccurred())

			httpResponseRecorder = httptest.NewRecorder()
			handler.ServeHTTP(httpResponseRecorder, httpRequest)
		})

		Context("Evacuate", func() {
			BeforeEach(func() {
				driverAdmin.EvacuateReturns(driveradmin.ErrorResponse{})

				var found bool
				route, found = driveradmin.Routes.FindRouteByName(driveradmin.EvacuateRoute)
				Expect(found).To(BeTrue())
			})

			It("should produce a handler with an evacuate route", func() {
				Expect(httpResponseRecorder.Code).To(Equal(200))
				Expect(httpResponseRecorder.Body).Should(MatchJSON(`{"Err":""}`))
			})

		})

		Context("Ping", func() {
			BeforeEach(func() {
				driverAdmin.EvacuateReturns(driveradmin.ErrorResponse{})

				var found bool
				route, found = driveradmin.Routes.FindRouteByName(driveradmin.PingRoute)
				Expect(found).To(BeTrue())
			})

			It("should produce a handler with an evacuate route", func() {
				Expect(httpResponseRecorder.Code).To(Equal(200))
				Expect(httpResponseRecorder.Body).Should(MatchJSON(`{"Err":""}`))
			})

		})

	})
})
