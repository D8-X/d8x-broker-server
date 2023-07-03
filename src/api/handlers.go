package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/D8-X/d8x-broker-server/src/utils"
)

var pen utils.SignaturePen

func setVariables(_pen utils.SignaturePen) {
	pen = _pen
}

func GetBrokerAddress(w http.ResponseWriter, r *http.Request, pen utils.SignaturePen) {
	var brokerAddr string
	for _, v := range pen.Wallets {
		brokerAddr = v.Address.String()
		break
	}
	response := struct {
		BrokerAddr string `json:"brokerAddr"`
	}{
		BrokerAddr: brokerAddr,
	}
	// Marshal the struct into JSON
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set the Content-Type header to application/json
	w.Header().Set("Content-Type", "application/json")

	// Write the JSON response
	w.Write(jsonResponse)
}

func GetBrokerFee(w http.ResponseWriter, r *http.Request, fee uint16) {
	res := utils.APIBrokerFeeRes{
		BrokerFeeTbps: fee,
	}
	// Marshal the struct into JSON
	jsonResponse, err := json.Marshal(res)
	if err != nil {
		http.Error(w, string(formatError(err.Error())), http.StatusBadRequest)
	}
	// Set the Content-Type header to application/json
	w.Header().Set("Content-Type", "application/json")
	// Write the JSON response
	w.Write(jsonResponse)
}

// SignOrder signs an order with the broker key and sets the fee
// Additional to the order, the broker needs to know the
// chainId
func SignOrder(w http.ResponseWriter, r *http.Request, pen utils.SignaturePen, feeTbps uint16) {
	// Read the JSON data from the request body
	var jsonData []byte
	if r.Body != nil {
		defer r.Body.Close()
		jsonData, _ = ioutil.ReadAll(r.Body)
	}

	// Parse the JSON payload
	var req utils.APIBrokerSignatureReq
	err := json.Unmarshal([]byte(jsonData), &req)
	if err != nil {
		http.Error(w, string(formatError(err.Error())), http.StatusBadRequest)
		return
	}
	err = req.CheckData()
	if err != nil {
		http.Error(w, string(formatError(err.Error())), http.StatusBadRequest)
		return
	}
	req.Order.BrokerFeeTbps = feeTbps
	jsonResponse, error := pen.GetBrokerSignatureResponse(req.Order, int64(req.ChainId))
	if error != nil {
		response := string(formatError(error.Error()))
		fmt.Fprintf(w, response)
		return
	}
	// Set the Content-Type header to application/json
	w.Header().Set("Content-Type", "application/json")
	// Write the JSON response
	w.Write(jsonResponse)
}

func formatError(errorMsg string) []byte {
	response := struct {
		Error string `json:"error"`
	}{
		Error: errorMsg,
	}
	// Marshal the struct into JSON
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		return []byte(err.Error())
	}
	return jsonResponse
}
