package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// RegisterRoutes registers all API routes for D8X-Backend application
func (a *App) RegisterRoutes(router chi.Router) {
	// Endpoint: /broker-address?id={id}
	router.Get("/broker-address", func(w http.ResponseWriter, r *http.Request) {
        GetBrokerAddress(w, r, a.Pen) // Pass fee here
    })

	// Endpoint: /broker-fee?perpetualId={perpetualId}
	router.Get("/broker-fee", func(w http.ResponseWriter, r *http.Request) {
        GetBrokerFee(w, r, a.BrokerFeeTbps) // Pass fee here
    })

	// Endpoint: /sign-order
	router.Post("/sign-order", func(w http.ResponseWriter, r *http.Request) {
        SignOrder(w, r, a.Pen, a.BrokerFeeTbps) // Pass `a.Pen` and fee here
    })
}
	
