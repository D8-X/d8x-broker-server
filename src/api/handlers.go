package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"log/slog"

	"github.com/D8-X/d8x-broker-server/src/utils"
	"github.com/D8-X/d8x-futures-go-sdk/pkg/d8x_futures"
	"github.com/ethereum/go-ethereum/common"
)

func (a *App) GetChainConfig(w http.ResponseWriter, r *http.Request) {
	config := make([]utils.ChainConfig, len(a.Pen.ChainConfig))
	var k int
	for _, conf := range a.Pen.ChainConfig {
		config[k] = conf
		k++
	}
	// Marshal the struct into JSON
	jsonResponse, err := json.Marshal(config)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set the Content-Type header to application/json
	w.Header().Set("Content-Type", "application/json")

	// Write the JSON response
	w.Write(jsonResponse)
}

func (a *App) BrokerAddress() string {
	// same address for all chains
	for _, wallet := range a.Pen.Wallets {
		return wallet.Address.Hex()
	}
	return ""
}

func (a *App) GetBrokerAddress(w http.ResponseWriter, r *http.Request) {
	brokerAddr := a.BrokerAddress()
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

func (a *App) GetBrokerFee(w http.ResponseWriter, r *http.Request) {

	addr := r.URL.Query().Get("addr")
	chainIdStr := r.URL.Query().Get("chain")
	var chainId int
	if chainIdStr == "" {
		chainId = -1
	} else {
		var err error
		chainId, err = strconv.Atoi(chainIdStr)
		if err != nil {
			chainId = -1
		}
	}
	fee := a.getBrokerFeeTbps(addr, chainId)
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
func (a *App) SignOrder(w http.ResponseWriter, r *http.Request) {
	pen := a.Pen
	redis := a.RedisClient
	// Read the JSON data from the request body
	var jsonData []byte
	if r.Body != nil {
		defer r.Body.Close()
		jsonData, _ = io.ReadAll(r.Body)
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
	//if req.Order.BrokerAddr == (common.Address{}).String() || len(req.Order.BrokerAddr) != len((common.Address{})) {
	//	req.Order.BrokerAddr = a.BrokerAddress()
	//}
	err = req.CheckData()
	if err != nil {
		http.Error(w, string(formatError(err.Error())), http.StatusBadRequest)
		return
	}
	slog.Info(fmt.Sprintf("Order signature request: trader %s... Perpetual %d Chain %d broker %s... deadline %d fee Tbps %d",
		string(req.Order.TraderAddr[0:8]),
		int(req.Order.PerpetualId),
		int(req.ChainId),
		string(req.Order.BrokerAddr[0:8]),
		int(req.Order.Deadline),
		int(req.Order.BrokerFeeTbps)))
	req.Order.BrokerFeeTbps = a.getBrokerFeeTbps(req.Order.TraderAddr, int(req.ChainId))

	jsonResponse, err := pen.GetBrokerOrderSignatureResponse(req.Order, int64(req.ChainId), redis)
	if err != nil {
		slog.Error("Error in signature request: " + err.Error())
		response := string(formatError(err.Error()))
		fmt.Fprint(w, response)
		return
	}
	// Set the Content-Type header to application/json
	w.Header().Set("Content-Type", "application/json")
	// Write the JSON response
	w.Write(jsonResponse)
}

// OrdersSubmitted marks an order ID as being submitted to the chain,
// adds it to the queue of order that are then published by the
// websocket
func (a *App) OrdersSubmitted(w http.ResponseWriter, r *http.Request) {
	// Read the JSON data from the request body
	var jsonData []byte
	if r.Body != nil {
		defer r.Body.Close()
		jsonData, _ = io.ReadAll(r.Body)
	}
	type Post struct {
		OrderIds []string `json:"orderIds"`
	}
	var req Post
	err := json.Unmarshal([]byte(jsonData), &req)
	if err != nil || len(req.OrderIds) == 0 {
		errMsg := `Wrong argument types. Usage: { "orderIds": "[0xABCE...,...]"}`
		http.Error(w, string(formatError(errMsg)), http.StatusBadRequest)
		return
	}
	for k := range req.OrderIds {
		req.OrderIds[k] = strings.TrimPrefix(req.OrderIds[k], "0x")
	}
	err = a.RedisClient.OrderSubmission(req.OrderIds)
	if err != nil {
		slog.Error(err.Error())
		response := string(formatError(err.Error()))
		fmt.Fprint(w, response)
		return
	}
	fmt.Fprint(w, `{"orders-submitted": "success"}`)
}

func (a *App) SignPayment(w http.ResponseWriter, r *http.Request) {
	pen := a.Pen
	// Read the JSON data from the request body
	var jsonData []byte
	if r.Body != nil {
		defer r.Body.Close()
		jsonData, _ = io.ReadAll(r.Body)
	}

	// Parse the JSON payload
	var req d8x_futures.BrokerPaySignatureReq
	err := req.UnmarshalJSON([]byte(jsonData))
	if err != nil {
		slog.Error("Error in payment signature request: " + err.Error())
		errMsg := `Wrong argument types. Usage: {
			'payment': {
				'payer': '0x4Fdc785fe2C6812960C93CA2F9D12b5Bd21ea2a1', 
				'executor': '0xDa47a0CAc77D50114F2725D06a2Ce887cF9f4D98', 
				'token': '0x2d10075E54356E16Ebd5C6BB5194290709B69C1e', 
				'timestamp': 1691249493, 
				'id': 1,
				'totalAmount': '1000000000000000000',
				'chainId': 80001,
				'multiPayCtrct': '0x30b55550e02B663E15A95B50850ebD20363c2AD5'
			},
			'signature': '0xABCE...'
		}`
		errMsg = strings.ReplaceAll(errMsg, "\t", "")
		errMsg = strings.ReplaceAll(errMsg, "\n", "")
		http.Error(w, string(formatError(errMsg)), http.StatusBadRequest)
		return
	}
	addr, err := pen.RecoverPaymentSignerAddr(req)
	if err != nil {
		slog.Error("SignPayment RecoverPaymentSignerAddr:" + err.Error())
		response := string(formatError(err.Error()))
		fmt.Fprint(w, response)
		return
	}
	if addr != req.Payment.Executor {
		slog.Error("SignPayment: wrong referrer signature")
		response := string(formatError("wrong signature"))
		fmt.Fprint(w, response)
		return
	}
	// signature correct, check if this is a registered payment executor
	if !findExecutor(pen, req.Payment.ChainId, addr) {
		slog.Error("SignPayment: executor not whitelisted")
		response := string(formatError("executor not allowed"))
		fmt.Fprint(w, response)
		return
	}
	// ensure token is approved to be spent
	err = a.ApproveToken(req.Payment.ChainId, req.Payment.Token)
	if err != nil {
		msg := fmt.Sprintf("error approving token for chain %d %s", req.Payment.ChainId, err.Error())
		slog.Error(msg)
		response := string(formatError("error approving token spending"))
		fmt.Fprint(w, response)
		return
	}
	// allowed executor, token approved, we can sign
	jsonResponse, err := pen.GetBrokerPaymentSignatureResponse(req)
	if err != nil {
		response := string(formatError(err.Error()))
		fmt.Fprint(w, response)
		return
	}
	// Set the Content-Type header to application/json
	w.Header().Set("Content-Type", "application/json")
	// Write the JSON response
	w.Write(jsonResponse)

}

func findExecutor(pen utils.SignaturePen, chainId int64, executor common.Address) bool {
	config := pen.ChainConfig[chainId]
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
