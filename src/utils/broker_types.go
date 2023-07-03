package utils

import (
	"fmt"

	d8x_futures "github.com/D8-X/d8x-futures-go-sdk"
	"github.com/ethereum/go-ethereum/common"
)

type DeploymentConfig struct {
	ChainId					  int64 `json:"chainId"`
	Name                      string `json:"name"`
	PerpetualManagerProxyAddr common.Address `json:"perpetualManagerProxyAddr"`
}

type APIBrokerSignatureReq struct {
	Order d8x_futures.IPerpetualOrderOrder `json:"order"`
	ChainId int `json:"chainId"`
}

func (req *APIBrokerSignatureReq) CheckData() error {
	if req.ChainId==0 {
		return fmt.Errorf("chainId not provided")
	}
	if req.Order.IDeadline==0 {
		return fmt.Errorf("request requires order with IDeadline")
	}
	zeroAddr := common.Address{}
	if req.Order.TraderAddr==zeroAddr {
		return fmt.Errorf("order requires order with non-zero TraderAddr")
	}
	if req.Order.IPerpetualId==nil {
		return fmt.Errorf("request requires order with IPerpetualId")
	}

	return nil
}

type APIBrokerSignatureRes struct {
	BrokerFeeTbps       uint16
	BrokerAddr          string
	Deadline            uint32
	BrokerSignature     string
}

type APIBrokerFeeRes struct {
	BrokerFeeTbps       uint16
}