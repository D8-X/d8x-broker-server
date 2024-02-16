package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// RegisterRoutes registers all API routes for D8X-Backend application
func (a *App) RegisterRoutes(router chi.Router) {
	// Endpoint: /broker-address?id={id}
	router.Get("/broker-address", func(w http.ResponseWriter, r *http.Request) {
		a.GetBrokerAddress(w, r)
	})

	// Endpoint: /broker-fee?perpetualId={perpetualId}
	router.Get("/broker-fee", func(w http.ResponseWriter, r *http.Request) {
		a.GetBrokerFee(w, r)
	})

	// Endpoint: /sign-order
	router.Post("/sign-order", func(w http.ResponseWriter, r *http.Request) {
		a.SignOrder(w, r)
	})

	// Endpoint: /order-submitted
	router.Post("/order-submitted", func(w http.ResponseWriter, r *http.Request) {
		a.OrderSubmitted(w, r)
	})

	// Endpoint: /payment-signature
	router.Post("/sign-payment", func(w http.ResponseWriter, r *http.Request) {
		a.SignPayment(w, r)
	})
}
