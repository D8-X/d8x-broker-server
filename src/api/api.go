package api

import (
	"net"
	"net/http"

	"github.com/D8-X/d8x-broker-server/src/utils"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// App is dependency container for API server
type App struct {
	Logger   *zap.Logger
	Port     string
	BindAddr string
	Pen utils.SignaturePen 
}

// StartApiServer initializes and starts the api server. This func is blocking
func (a *App) StartApiServer() {
	if len(a.Port) == 0 {
		a.Logger.Fatal("could not start start the API server, Port must be provided")
	}

	router := chi.NewRouter()
	a.RegisterGlobalMiddleware(router)
	a.RegisterRoutes(router)

	addr := net.JoinHostPort(
		a.BindAddr,
		a.Port,
	)
	a.Logger.Info("starting api server", zap.String("host_port", addr))
	err := http.ListenAndServe(
		addr,
		router,
	)
	a.Logger.Fatal("api server is shutting down", zap.Error(err))
}