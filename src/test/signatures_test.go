package test

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"testing"

	"github.com/D8-X/d8x-broker-server/src/config"
	"github.com/D8-X/d8x-broker-server/src/env"
	"github.com/D8-X/d8x-broker-server/src/utils"
	d8x_futures "github.com/D8-X/d8x-futures-go-sdk"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/viper"
)

func TestSignOrder(t *testing.T) {

	// Generate a new private key
	privateKey, err := crypto.GenerateKey()
	//privateKey, err := crypto.HexToECDSA("yourprivatekey")
	if err != nil {
		log.Fatal(err)
	}
	// Derive the Ethereum address from the private key
	addr := crypto.PubkeyToAddress(privateKey.PublicKey)

	config, err := config.LoadChainConfig("../../config/shared.chainConfig.json")
	if err != nil {
		t.Errorf("loading deploymentconfig: %v", err)
		return
	}
	pk := fmt.Sprintf("%x", privateKey.D)
	pen, err := utils.NewSignaturePen(pk, config)
	var perpOrder = d8x_futures.IPerpetualOrderOrder{
		BrokerFeeTbps: 410,
		TraderAddr:    common.HexToAddress("0x9d5aaB428e98678d0E645ea4AeBd25f744341a05"),
		BrokerAddr:    addr,
		IDeadline:     1691249493,
		IPerpetualId:  big.NewInt(int64(10001)),
	}
	digest, sig, err := pen.SignOrder(perpOrder, 80001)
	if err != nil {
		t.Errorf("signing order: %v", err)
	}
	sigBytes, err := d8x_futures.BytesFromHexString(sig)
	if err != nil {
		t.Errorf("decoding signature: %v", err)
	}
	digestBytes, err := d8x_futures.BytesFromHexString(digest)
	if err != nil {
		t.Errorf("decoding signature: %v", err)
	}
	fmt.Println("digest = ", digestBytes)
	addrRecovered, err := d8x_futures.RecoverEvmAddress(digestBytes, sigBytes)
	v := addrRecovered.String()
	v0 := addr.String()
	if err != nil {
		t.Errorf("recovering address: %v", err)
	} else {
		t.Logf("recovered address")
		t.Logf(v)
	}

	t.Log("recovered addr = ", v)
	t.Log("signer    addr = ", v0)
	if v == v0 {
		t.Logf("recovered address correct")
	} else {
		t.Errorf("recovering address incorrect")
	}
}

func generateKey() (common.Address, *ecdsa.PrivateKey, error) {
	// Generate a new private key
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		log.Fatal(err)
		return common.Address{}, nil, err
	}
	// Derive the Ethereum address from the private key
	addr := crypto.PubkeyToAddress(privateKey.PublicKey)
	return addr, privateKey, err
}

func TestSignPayment(t *testing.T) {
	brokerAddr, brokerPk, err := generateKey()
	if err != nil {
		log.Fatal(err)
	}
	execAddr, execPk, err := generateKey()
	if err != nil {
		log.Fatal(err)
	}
	multiPayCtrctAddr := common.HexToAddress("0x30b55550e02B663E15A95B50850ebD20363c2AD5")
	summary := d8x_futures.PaySummary{
		Payer:         brokerAddr,
		Executor:      execAddr,
		Token:         common.HexToAddress("0x2d10075E54356E16Ebd5C6BB5194290709B69C1e"),
		Timestamp:     1691249493,
		Id:            1,
		TotalAmount:   big.NewInt(1e18),
		ChainId:       80001,
		MultiPayCtrct: multiPayCtrctAddr,
	}
	brokerAddrStr := brokerAddr.String()
	execAddrStr := execAddr.String()
	t.Log("brokerAddr = ", brokerAddrStr)
	t.Log("execAddr = ", execAddrStr)
	var execWallet d8x_futures.Wallet
	pk := fmt.Sprintf("%x", execPk.D)
	err = execWallet.NewWallet(pk, 80001, nil)
	if err != nil {
		t.Errorf("error creating wallet")
	}
	_, sg, err := d8x_futures.CreatePaymentBrokerSignature(summary, execWallet)

	data := d8x_futures.BrokerPaySignatureReq{
		Payment:           summary,
		ExecutorSignature: sg,
	}

	config, err := config.LoadChainConfig("../../config/example.chainConfig.json")
	if err != nil {
		t.Errorf("loading deploymentconfig: %v", err)
		return
	}
	pkBrker := fmt.Sprintf("%x", brokerPk.D)
	pen, err := utils.NewSignaturePen(pkBrker, config)
	jsonRes, err := pen.GetBrokerPaymentSignatureResponse(data)
	if err != nil {
		t.Errorf("GetBrokerPaymentSignatureResponse: %v", err)
		return
	}
	type SignatureData struct {
		BrokerSignature string `json:"brokerSignature"`
	}
	var brokerSig SignatureData
	err = json.Unmarshal(jsonRes, &brokerSig)
	if err != nil {
		t.Errorf("Unmarshal GetBrokerPaymentSignatureResponse: %v", err)
		return
	}
	// recover again
	sigBytes, err := d8x_futures.BytesFromHexString(brokerSig.BrokerSignature)
	if err != nil {
		t.Errorf("decoding signature: %v", err)
	}
	addr, err := d8x_futures.RecoverPaymentSignatureAddr(sigBytes, summary)
	if err != nil {
		t.Errorf("error RecoverPaymentSignatureAddr")
	}
	t.Log("recovered addr = ", addr.String())
	t.Log("signer    addr = ", brokerAddr.String())
	if addr != brokerAddr {
		t.Errorf("error wrong address recovered")
	} else {
		t.Logf("recovered address correct")
	}
}

func loadEnv() {

	viper.SetConfigFile("../../.env")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("could not load .env file")
		log.Fatalf(err.Error())
	}

	viper.SetDefault(env.API_BIND_ADDR, "")
	viper.SetDefault(env.API_PORT, "8000")

	requiredEnvs := []string{
		env.BROKER_KEY,
		env.BROKER_FEE_TBPS,
	}

	for _, e := range requiredEnvs {
		if !viper.IsSet(e) {
			log.Fatalf("required environment variable not set variable")
			log.Fatalf(e)
		}
	}
}
