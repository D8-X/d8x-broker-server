package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/D8-X/d8x-broker-server/src/utils"
	"github.com/ethereum/go-ethereum/common"
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
	var req utils.APIBrokerOrderSignatureReq
	err := json.Unmarshal([]byte(jsonData), &req)
	if err != nil {
		errMsg := `Wrong argument types. Usage: 
			{'order': {'traderAddr': '0xABCD..', 'iDeadline': 1688347462, 'iPerpetualId': 10001},
			'chainId': 80001}`
		errMsg = strings.ReplaceAll(errMsg, "\t", "")
		errMsg = strings.ReplaceAll(errMsg, "\n", "")
		http.Error(w, string(formatError(errMsg)), http.StatusBadRequest)
		return
	}
	err = req.CheckData()
	if err != nil {
		http.Error(w, string(formatError(err.Error())), http.StatusBadRequest)
		return
	}
	req.Order.BrokerFeeTbps = feeTbps
	jsonResponse, err := pen.GetBrokerOrderSignatureResponse(req.Order, int64(req.ChainId))
	if err != nil {
		response := string(formatError(err.Error()))
		fmt.Fprintf(w, response)
		return
	}
	// Set the Content-Type header to application/json
	w.Header().Set("Content-Type", "application/json")
	// Write the JSON response
	w.Write(jsonResponse)
}

func SignPayment(w http.ResponseWriter, r *http.Request, pen utils.SignaturePen) {
	// Read the JSON data from the request body
	var jsonData []byte
	if r.Body != nil {
		defer r.Body.Close()
		jsonData, _ = ioutil.ReadAll(r.Body)
	}

	// Parse the JSON payload
	var req utils.APIBrokerPaySignatureReq
	err := json.Unmarshal([]byte(jsonData), &req)
	if err != nil {

		http.Error(w, string(formatError(err.Error())), http.StatusBadRequest)
		return
	}
	addr, err := pen.RecoverPaymentSignerAddr(req)
	if err != nil {
		response := string(formatError(err.Error()))
		fmt.Fprintf(w, response)
		return
	}
	if addr != req.Payment.Executor {
		response := string(formatError("wrong signature"))
		fmt.Fprintf(w, response)
		return
	}
	// signature correct, check if this is a registered payment executor
	if !findExecutor(pen, req.Payment.ChainId, addr) {
		response := string(formatError("executor not allowed"))
		fmt.Fprintf(w, response)
		return
	}
	// allowed executor, we can sign
	jsonResponse, err := pen.GetBrokerPaymentSignatureResponse(req)
	if err != nil {
		response := string(formatError(err.Error()))
		fmt.Fprintf(w, response)
		return
	}
	// Set the Content-Type header to application/json
	w.Header().Set("Content-Type", "application/json")
	// Write the JSON response
	w.Write(jsonResponse)

}

func findExecutor(pen utils.SignaturePen, chainId int64, executor common.Address) bool {
	config := pen.Config[chainId]
	for _, addr := range config.AllowedExecutors {
		if addr == executor {
			return true
		}
	}
	return false
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
