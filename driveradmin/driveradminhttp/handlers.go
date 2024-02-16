package driveradminhttp

import (
	"errors"
	"net/http"

	// cfhttphandlers module is imported solely for the WriteJSONResponse function
	cfhttphandlers "code.cloudfoundry.org/cfhttp/handlers"
	"code.cloudfoundry.org/dockerdriver/driverhttp"
	"code.cloudfoundry.org/lager/v3"
	"code.cloudfoundry.org/nfsv3driver/driveradmin"
	"github.com/tedsuo/rata"
)

func NewHandler(logger lager.Logger, client driveradmin.DriverAdmin) (http.Handler, error) {
	logger = logger.Session("server")
	logger.Info("start")
	defer logger.Info("end")

	var handlers = rata.Handlers{
		driveradmin.EvacuateRoute: newEvacuateHandler(logger, client),
		driveradmin.PingRoute:     newPingHandler(logger, client),
	}

	return rata.NewRouter(driveradmin.Routes, handlers)
}

func newEvacuateHandler(logger lager.Logger, client driveradmin.DriverAdmin) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		logger := logger.Session("handle-evacuate")
		logger.Info("start")
		defer logger.Info("end")

		env := driverhttp.EnvWithMonitor(logger, req.Context(), w)

		response := client.Evacuate(env)
		if response.Err != "" {
			logger.Error("failed-evacuating", errors.New(response.Err))
			cfhttphandlers.WriteJSONResponse(w, http.StatusInternalServerError, response)
			return
		}

		cfhttphandlers.WriteJSONResponse(w, http.StatusOK, response)
	}
}

func newPingHandler(logger lager.Logger, client driveradmin.DriverAdmin) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		logger := logger.Session("handle-ping")
		logger.Info("start")
		defer logger.Info("end")

		env := driverhttp.EnvWithMonitor(logger, req.Context(), w)

		response := client.Ping(env)
		if response.Err != "" {
			logger.Error("failed-pinging", errors.New(response.Err))
			cfhttphandlers.WriteJSONResponse(w, http.StatusInternalServerError, response)
			return
		}

		cfhttphandlers.WriteJSONResponse(w, http.StatusOK, response)
	}
}
