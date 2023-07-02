package utils

import (
	d8x_futures "github.com/D8-X/d8x-futures-go-sdk"
	"github.com/ethereum/go-ethereum/common"
)

type DeploymentConfig struct {
	ChainId					  int64 `json:"chainId"`
	Name                      string `json:"name"`
	PerpetualManagerProxyAddr common.Address `json:"perpetualManagerProxyAddr"`
}

type APIBrokerSignatureReq struct {
	Order d8x_futures.Order `json:"order"`
	ChainId int `json:"chainId"`
}
