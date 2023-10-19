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
	d8x_futures "github.com/D8-X/d8x-futures-go-sdk"
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
func SignOrder(w http.ResponseWriter, r *http.Request, pen utils.SignaturePen, feeTbps uint16, redis *utils.RueidisClient) {
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
	err = req.CheckData()
	if err != nil {
		http.Error(w, string(formatError(err.Error())), http.StatusBadRequest)
		return
	}
	slog.Info("Order signature request: trader " + string(req.Order.TraderAddr[1:4]) + "... Perpetual " + strconv.Itoa(int(req.Order.PerpetualId)))
	req.Order.BrokerFeeTbps = feeTbps
	jsonResponse, err := pen.GetBrokerOrderSignatureResponse(req.Order, int64(req.ChainId), redis)
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

func SignPayment(w http.ResponseWriter, r *http.Request, a *App) {
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
	// ensure token is approved to be spent
	err = a.ApproveToken(req.Payment.ChainId, req.Payment.Token)
	if err != nil {
		slog.Error(err.Error())
		response := string(formatError("Error approving token spending"))
		fmt.Fprintf(w, response)
		return
	}
	// allowed executor, token approved, we can sign
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
