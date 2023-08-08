package utils

import (
	"fmt"

	d8x_futures "github.com/D8-X/d8x-futures-go-sdk"
	"github.com/ethereum/go-ethereum/common"
)

type ChainConfig struct {
	ChainId                   int64            `json:"chainId"`
	Name                      string           `json:"name"`
	PerpetualManagerProxyAddr common.Address   `json:"perpetualManagerProxyAddr"`
	MultiPayCtrctAddr         common.Address   `json:"multiPayCtrctAddr"`
	AllowedExecutors          []common.Address `json:"allowedExecutors"`
}

type APIBrokerOrderSignatureReq struct {
	Order     APIOrderSig `json:"order"`
	ChainId   int64       `json:"chainId"`
	Signature string      `json:"signature"`
}

func (req *APIBrokerOrderSignatureReq) CheckData() error {
	if req.ChainId == 0 {
		return fmt.Errorf("chainId not provided")
	}
	if req.Order.Deadline == 0 {
		return fmt.Errorf("request requires order with iDeadline")
	}
	zeroAddr := common.Address{}.Hex()
	if req.Order.TraderAddr == zeroAddr ||
		req.Order.TraderAddr == "" {
		return fmt.Errorf("order requires order with non-zero traderAddr")
	}
	if req.Order.PerpetualId == 0 {
		return fmt.Errorf("request requires order with iPerpetualId")
	}

	return nil
}

type APIBrokerPaySignatureReq struct {
	Payment           d8x_futures.PaySummary `json:"payment"`
	ChainId           int64                  `json:"chainId"`
	ExecutorSignature string                 `json:"signature"`
}

// Required data to sign: iPerpetualId: number, brokerFeeTbps: number, traderAddr: string, iDeadline: number,
//
//	chainId: number, proxyAddress: string
type APIOrderSig struct {
	PerpetualId   int32  `json:"iPerpetualId"`
	BrokerFeeTbps uint16 `json:"brokerFeeTbps"`
	BrokerAddr    string `json:"brokerAddr"`
	TraderAddr    string `json:"traderAddr"`
	Deadline      uint32 `json:"iDeadline"`
}

type APIBrokerSignatureRes struct {
	Order           APIOrderSig `json:"orderFields"`
	ChainId         int64       `json:"chainId"`
	BrokerSignature string      `json:"brokerSignature"`
}

type APIBrokerFeeRes struct {
	BrokerFeeTbps uint16
}
