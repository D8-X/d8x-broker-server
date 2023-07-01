package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Struct to represent the JSON argument for the /sign-order endpoint
type Order struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func GetBrokerAddress(w http.ResponseWriter, r *http.Request) {
	// Get the query parameter "id"
	id := r.URL.Query().Get("id")

	// Return the broker address for the given ID
	fmt.Fprintf(w, "Broker Address for ID %s", id)
}

func GetBrokerFee(w http.ResponseWriter, r *http.Request) {
	// Get the query parameter "perpetualId"
	perpetualID := r.URL.Query().Get("perpetualId")

	// Return the broker fee for the given perpetual ID
	fmt.Fprintf(w, "Broker Fee for Perpetual ID %s", perpetualID)
}

func SignOrder(w http.ResponseWriter, r *http.Request) {
	// Parse the JSON payload
	var order Order
	err := json.NewDecoder(r.Body).Decode(&order)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Process the order and return a response
	response := fmt.Sprintf("Order with ID %d and Name %s signed successfully", order.ID, order.Name)
	fmt.Fprintf(w, response)
}