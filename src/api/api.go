package api

import (
	"errors"
	"log/slog"
	"net"
	"net/http"

	"github.com/D8-X/d8x-broker-server/src/utils"
	"github.com/go-chi/chi/v5"
)

// App is dependency container for API server
type App struct {
	Port          string
	BindAddr      string
	Pen           utils.SignaturePen
	BrokerFeeTbps uint16
}

// StartApiServer initializes and starts the api server. This func is blocking
func (a *App) StartApiServer() error {
	if len(a.Port) == 0 {
		return errors.New("could not start the API server, Port must be provided")
	}

	router := chi.NewRouter()
	a.RegisterGlobalMiddleware(router)
	a.RegisterRoutes(router)

	addr := net.JoinHostPort(
		a.BindAddr,
		a.Port,
	)
	slog.Info("starting api server host_port " + addr)
	err := http.ListenAndServe(
		addr,
		router,
	)
	return errors.New("api server is shutting down" + err.Error())
}
