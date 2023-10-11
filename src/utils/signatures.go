package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"strings"

	d8x_futures "github.com/D8-X/d8x-futures-go-sdk"
	"github.com/ethereum/go-ethereum/common"
	solsha3 "github.com/miguelmota/go-solidity-sha3"
)

// SignaturePen stores the chainId <-> deployment address mappings,
// and the wallet struct for the broker
type SignaturePen struct {
	ChainConfig map[int64]ChainConfig
	RpcConfig   map[int64][]string
	Wallets     map[int64]d8x_futures.Wallet
}

func NewSignaturePen(privateKeyHex string, chConf []ChainConfig, rpcConf []RpcConfig) (SignaturePen, error) {
	wallets, err := createWalletMap(chConf, privateKeyHex)
	if err != nil {
		return SignaturePen{}, err
	}
	pen := SignaturePen{
		ChainConfig: createChainConfigMap(chConf),
		RpcConfig:   createRpcConfigMap(rpcConf),
		Wallets:     wallets,
	}
	return pen, nil
}

func (p *SignaturePen) RecoverPaymentSignerAddr(ps d8x_futures.BrokerPaySignatureReq) (common.Address, error) {
	sig, err := d8x_futures.BytesFromHexString(ps.ExecutorSignature)
	if err != nil {
		return common.Address{}, err
	}
	ctrct := p.ChainConfig[ps.Payment.ChainId].MultiPayCtrctAddr
	if ctrct != ps.Payment.MultiPayCtrct {
		return common.Address{}, fmt.Errorf("Multipay ctrct mismatch")
	}
	addr, err := d8x_futures.RecoverPaymentSignatureAddr(sig, ps.Payment)
	if err != nil {
		return common.Address{}, err
	}
	return addr, nil
}

func (p *SignaturePen) GetBrokerPaymentSignatureResponse(ps d8x_futures.BrokerPaySignatureReq) ([]byte, error) {
	ctrct := p.ChainConfig[ps.Payment.ChainId].MultiPayCtrctAddr
	if strings.ToLower(ctrct.String()) != strings.ToLower(ps.Payment.MultiPayCtrct.String()) {
		return nil, fmt.Errorf("Multipay ctrct mismatch, expect " + ctrct.String())
	}
	_, sig, err := d8x_futures.CreatePaymentBrokerSignature(ps.Payment, p.Wallets[ps.Payment.ChainId])
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

func (p *SignaturePen) GetBrokerOrderSignatureResponse(order APIOrderSig, chainId int64, redis *RueidisClient) ([]byte, error) {
	var perpOrder = d8x_futures.IPerpetualOrderOrder{
		// data for broker signature
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
	sigBytes, err := d8x_futures.BytesFromHexString(sig)
	if err != nil {
		return []byte{}, errors.New("decoding signature: " + err.Error())
	}
	// order digest
	order.BrokerSignature = sigBytes
	order.BrokerAddr = p.Wallets[chainId].Address.String()
	digest, orderId, err := p.createOrderDigest(order, chainId)
	if err != nil {
		return []byte{}, errors.New("decoding signature: " + err.Error())
	}
	res := APIBrokerSignatureRes{
		Order:           order,
		ChainId:         chainId,
		BrokerSignature: sig,
		OrderDigest:     digest,
		OrderId:         orderId,
	}
	redis.PubOrder(order, orderId, chainId)
	// Marshal the struct into JSON
	jsonResponse, err := json.Marshal(res)
	if err != nil {
		return nil, err
	}
	return jsonResponse, nil
}

func (p *SignaturePen) createOrderDigest(order APIOrderSig, chainId int64) (string, string, error) {
	perpId := new(big.Int).SetInt64(int64(order.PerpetualId))

	var co d8x_futures.IClientOrderClientOrder
	limitPrice, triggerPrice, amount := new(big.Int), new(big.Int), new(big.Int)
	_, s1 := limitPrice.SetString(order.FLimitPrice, 10)
	_, s2 := triggerPrice.SetString(order.FTriggerPrice, 10)
	_, s3 := amount.SetString(order.FAmount, 10)
	if !s1 || !s2 || !s3 {
		return "", "", errors.New("invalid big number")
	}

	co.BrokerAddr = common.HexToAddress(order.BrokerAddr)
	co.IPerpetualId = perpId
	co.FLimitPrice = limitPrice
	co.LeverageTDR = order.LeverageTDR
	co.ExecutionTimestamp = order.ExecutionTimestamp
	co.Flags = order.Flags
	co.IDeadline = order.Deadline
	co.BrokerAddr = common.HexToAddress(order.BrokerAddr)
	co.FTriggerPrice = triggerPrice
	co.FAmount = amount
	co.TraderAddr = common.HexToAddress(order.TraderAddr)
	co.BrokerFeeTbps = order.BrokerFeeTbps
	co.BrokerSignature = order.BrokerSignature
	d, err := d8x_futures.CreateOrderDigest(co, int(chainId), true, p.ChainConfig[chainId].PerpetualManagerProxyAddr.String())
	if err != nil {
		return "", "", err
	}
	digestBytes, err := hex.DecodeString(d)
	if err != nil {
		return "", "", err
	}
	dsol := solsha3.SoliditySHA3WithPrefix(digestBytes)
	orderId := hex.EncodeToString(dsol)
	return d, orderId, nil
}

func (p *SignaturePen) SignOrder(order d8x_futures.IPerpetualOrderOrder, chainId int64) (string, string, error) {
	//
	proxyAddr := p.ChainConfig[chainId].PerpetualManagerProxyAddr
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

func createChainConfigMap(configList []ChainConfig) map[int64]ChainConfig {
	config := make(map[int64]ChainConfig)
	for _, c := range configList {
		config[c.ChainId] = c
	}
	return config
}

func createRpcConfigMap(configList []RpcConfig) map[int64][]string {
	config := make(map[int64][]string)
	for _, c := range configList {
		config[c.ChainId] = c.Rpc
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

func Encrypt(plainText string, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	encrypted := gcm.Seal(nonce, nonce, []byte(plainText), nil)
	return hex.EncodeToString(encrypted), nil
}

func Decrypt(encryptedText string, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	encrypted, err := hex.DecodeString(encryptedText)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(encrypted) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, encrypted := encrypted[:nonceSize], encrypted[nonceSize:]
	plainText, err := gcm.Open(nil, nonce, encrypted, nil)
	if err != nil {
		return "", err
	}

	return string(plainText), nil
}
