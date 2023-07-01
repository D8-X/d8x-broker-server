package api

import (
	"github.com/go-chi/chi/v5"
)

// RegisterRoutes registers all API routes for D8X-Backend application
func (a *App) RegisterRoutes(router chi.Router) {
	// Endpoint: /broker-address?id={id}
	router.Get("/broker-address", GetBrokerAddress)

	// Endpoint: /broker-fee?perpetualId={perpetualId}
	router.Get("/broker-fee", GetBrokerFee)

	// Endpoint: /sign-order
	router.Post("/sign-order", SignOrder)
}
	
