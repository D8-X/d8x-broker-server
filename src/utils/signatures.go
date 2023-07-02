package utils

import (
	"fmt"

	d8x_futures "github.com/D8-X/d8x-futures-go-sdk"
)

// SignaturePen stores the chainId <-> deployment address mappings,
// and the wallet struct for the broker
type SignaturePen struct {
	Config map[int64]DeploymentConfig
	Wallets map[int64]d8x_futures.Wallet  
}

func NewSignaturePen(privateKeyHex string, config []DeploymentConfig) (SignaturePen, error) {
	wallets, err := createWalletMap(config, privateKeyHex)
	if err!=nil {
		return SignaturePen{}, err
	}
	pen := SignaturePen {
		Config: createConfigMap(config),
		Wallets: wallets,
	}
	return pen, nil
}

func (p *SignaturePen) SignOrder(order d8x_futures.IPerpetualOrderOrder, brokerFeeTbps uint32, chainId int64) (string, string, error) {
	//
	proxyAddr := p.Config[chainId].PerpetualManagerProxyAddr
	wallet := p.Wallets[chainId]
	digest, sig, err := d8x_futures.CreateBrokerSignature(
		proxyAddr, chainId, wallet, int32(order.IPerpetualId.Int64()), brokerFeeTbps, 
		order.TraderAddr.String(), order.IDeadline);
	//proxyAddr common.Address, chainId int64, brokerWallet Wallet, 
	//iPerpetualId int32, brokerFeeTbps uint32, traderAddr string, iDeadline uint3

	return digest, sig, err
}

func createConfigMap(configList []DeploymentConfig) map[int64]DeploymentConfig {
	config := make(map[int64]DeploymentConfig)
	for _, c := range configList {
		config[c.ChainId] = c;
	}
	return config
}

func createWalletMap(configList []DeploymentConfig, privateKeyHex string) (map[int64]d8x_futures.Wallet, error) {
	walletMap := make(map[int64]d8x_futures.Wallet)
	for _, c := range configList {
		var wallet d8x_futures.Wallet
		err := wallet.NewWallet(privateKeyHex, c.ChainId, nil);
		if (err!=nil) {
			return nil, fmt.Errorf("error casting public key to ECDSA")
		}
		walletMap[c.ChainId] = wallet;
	}
	return walletMap, nil
}