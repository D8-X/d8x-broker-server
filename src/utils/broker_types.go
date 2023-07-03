package utils

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
)

type DeploymentConfig struct {
	ChainId                   int64          `json:"chainId"`
	Name                      string         `json:"name"`
	PerpetualManagerProxyAddr common.Address `json:"perpetualManagerProxyAddr"`
}

type APIBrokerSignatureReq struct {
	Order   APIOrderSig `json:"order"`
	ChainId int64       `json:"chainId"`
}

func (req *APIBrokerSignatureReq) CheckData() error {
	if req.ChainId == 0 {
		return fmt.Errorf("chainId not provided")
	}
	if req.Order.Deadline == 0 {
		return fmt.Errorf("request requires order with IDeadline")
	}
	zeroAddr := common.Address{}.Hex()
	if req.Order.TraderAddr == zeroAddr {
		return fmt.Errorf("order requires order with non-zero TraderAddr")
	}
	if req.Order.PerpetualId == 0 {
		return fmt.Errorf("request requires order with IPerpetualId")
	}

	return nil
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
