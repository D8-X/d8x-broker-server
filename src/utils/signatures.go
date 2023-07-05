package utils

import (
	"encoding/json"
	"fmt"
	"math/big"

	d8x_futures "github.com/D8-X/d8x-futures-go-sdk"
	"github.com/ethereum/go-ethereum/common"
)

// SignaturePen stores the chainId <-> deployment address mappings,
// and the wallet struct for the broker
type SignaturePen struct {
	Config  map[int64]ChainConfig
	Wallets map[int64]d8x_futures.Wallet
}

func NewSignaturePen(privateKeyHex string, config []ChainConfig) (SignaturePen, error) {
	wallets, err := createWalletMap(config, privateKeyHex)
	if err != nil {
		return SignaturePen{}, err
	}
	pen := SignaturePen{
		Config:  createConfigMap(config),
		Wallets: wallets,
	}
	return pen, nil
}

func (p *SignaturePen) RecoverPaymentSignerAddr(ps APIBrokerPaySignatureReq) (common.Address, error) {
	sig, err := d8x_futures.BytesFromHexString(ps.ExecutorSignature)
	if err != nil {
		return common.Address{}, err
	}
	ctrct := p.Config[ps.ChainId].MultiPayCtrctAddr
	addr, err := d8x_futures.RecoverPaymentSignatureAddr(sig, ctrct, ps.Payment, ps.ChainId)
	if err != nil {
		return common.Address{}, err
	}
	return addr, nil
}
func (p *SignaturePen) GetBrokerPaymentSignatureResponse(ps APIBrokerPaySignatureReq) ([]byte, error) {
	ctrct := p.Config[ps.ChainId].MultiPayCtrctAddr
	_, sig, err := d8x_futures.CreatePaymentBrokerSignature(ctrct, ps.Payment, ps.ChainId, p.Wallets[ps.ChainId])
	if err != nil {
		return nil, err
	}
	response := struct {
		BrokerSignature string `json:"brokerSignature"`
	}{
		BrokerSignature: sig,
	}
	// Marshal the struct into JSON
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}
	return jsonResponse, nil
}

func (p *SignaturePen) GetBrokerOrderSignatureResponse(order APIOrderSig, chainId int64) ([]byte, error) {
	var perpOrder = d8x_futures.IPerpetualOrderOrder{
		BrokerFeeTbps: order.BrokerFeeTbps,
		TraderAddr:    common.HexToAddress(order.TraderAddr),
		BrokerAddr:    common.HexToAddress(order.BrokerAddr),
		IDeadline:     order.Deadline,
		IPerpetualId:  big.NewInt(int64(order.PerpetualId)),
	}
	_, sig, err := p.SignOrder(perpOrder, chainId)
	if err != nil {
		return nil, err
	}
	order.BrokerAddr = p.Wallets[chainId].Address.String()
	res := APIBrokerSignatureRes{
		Order:           order,
		ChainId:         chainId,
		BrokerSignature: sig,
	}
	// Marshal the struct into JSON
	jsonResponse, err := json.Marshal(res)
	if err != nil {
		return nil, err
	}
	return jsonResponse, nil
}

func (p *SignaturePen) SignOrder(order d8x_futures.IPerpetualOrderOrder, chainId int64) (string, string, error) {
	//
	proxyAddr := p.Config[chainId].PerpetualManagerProxyAddr
	wallet := p.Wallets[chainId]
	if wallet.PrivateKey == nil {
		return "", "", fmt.Errorf("No broker key defined for chain %d", chainId)
	}
	digest, sig, err := d8x_futures.CreateOrderBrokerSignature(
		proxyAddr, chainId, wallet, int32(order.IPerpetualId.Int64()), uint32(order.BrokerFeeTbps),
		order.TraderAddr.String(), order.IDeadline)
	//proxyAddr common.Address, chainId int64, brokerWallet Wallet,
	//iPerpetualId int32, brokerFeeTbps uint32, traderAddr string, iDeadline uint3

	return digest, sig, err
}

func createConfigMap(configList []ChainConfig) map[int64]ChainConfig {
	config := make(map[int64]ChainConfig)
	for _, c := range configList {
		config[c.ChainId] = c
	}
	return config
}

func createWalletMap(configList []ChainConfig, privateKeyHex string) (map[int64]d8x_futures.Wallet, error) {
	walletMap := make(map[int64]d8x_futures.Wallet)
	for _, c := range configList {
		var wallet d8x_futures.Wallet
		err := wallet.NewWallet(privateKeyHex, c.ChainId, nil)
		if err != nil {
			return nil, fmt.Errorf("error casting public key to ECDSA")
		}
		walletMap[c.ChainId] = wallet
	}
	return walletMap, nil
}
