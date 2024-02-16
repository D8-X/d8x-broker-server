package api

import (
	"github.com/go-chi/chi/v5"
)

func (a *App) RegisterGlobalMiddleware(r chi.Router) {
	// cors is handled in nginx
}
