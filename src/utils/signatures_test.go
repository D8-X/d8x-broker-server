package utils

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"testing"

	"github.com/D8-X/d8x-broker-server/src/env"

	"github.com/D8-X/d8x-futures-go-sdk/pkg/contracts"
	"github.com/D8-X/d8x-futures-go-sdk/pkg/d8x_futures"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/viper"
)

func TestSignOrder(t *testing.T) {

	loadEnv()
	privateKey, err := crypto.HexToECDSA(viper.GetString("PK_TEST"))
	// instead generate a new private key
	// privateKey, err := crypto.GenerateKey()

	if err != nil {
		log.Fatal(err)
	}
	// Derive the Ethereum address from the private key
	addr := crypto.PubkeyToAddress(privateKey.PublicKey)

	chConfig, err := LoadChainConfig("../../config/chainConfig.json")
	if err != nil {
		fmt.Printf("loading deploymentconfig: %v", err)
		return
	}
	rpcConfig, err := LoadRpcConfig("../../config/rpc.json")
	if err != nil {
		fmt.Printf("loading deploymentconfig: %v", err)
		return
	}
	pk := fmt.Sprintf("%x", privateKey.D)
	pen, err := NewSignaturePen(pk, chConfig, rpcConfig)
	if err != nil {
		fmt.Printf("NewSignaturePen: %v\n", err)
		t.FailNow()
	}
	fmt.Printf("broker = %s\n", addr.String())
	var perpOrder = contracts.IPerpetualOrderOrder{
		BrokerFeeTbps: 40,
		TraderAddr:    common.HexToAddress("def43cf2dd024abc5447c1dcdc2fe3fe58547b84"),
		BrokerAddr:    addr,
		IDeadline:     1743101927,
		IPerpetualId:  big.NewInt(int64(100000)),
	}
	digest, sig, err := pen.SignOrder(perpOrder, chConfig[80094].ProxyAddr, 80094)
	if err != nil {
		t.Errorf("signing order: %v", err)
		t.FailNow()
	}
	fmt.Printf("\nsignature = %s\n", sig)
	sigBytes, err := d8x_futures.BytesFromHexString(sig)
	if err != nil {
		t.Errorf("decoding signature: %v", err)
		t.FailNow()
	}
	digestBytes, err := d8x_futures.BytesFromHexString(digest)
	if err != nil {
		t.Errorf("decoding signature: %v", err)
		t.FailNow()
	}
	fmt.Println("digest = ", digest)
	fmt.Println("digest bytes = ", digestBytes)
	addrRecovered, err := d8x_futures.RecoverEvmAddress(digestBytes, sigBytes)
	v := addrRecovered.String()
	v0 := addr.String()
	if err != nil {
		t.Errorf("recovering address: %v", err)
	} else {
		fmt.Println("recovered address")
		fmt.Println(v)
	}

	t.Log("recovered addr = ", v)
	t.Log("signer    addr = ", v0)
	if v == v0 {
		fmt.Println("recovered address correct")
	} else {
		fmt.Println("recovering address incorrect")
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
	//brokerAddr, brokerPk, err := getAddrPkFromString("key")
	if err != nil {
		log.Fatal(err)
	}
	execAddr, execPk, err := generateKey()
	//execAddr, execPk, err := getAddrPkFromString("key")
	if err != nil {
		log.Fatal(err)
	}

	multiPayCtrctAddr := common.HexToAddress("0xfCBE2f332b1249cDE226DFFE8b2435162426AfE5")
	summary := d8x_futures.PaySummary{
		Payer:         brokerAddr,
		Executor:      execAddr,
		Token:         common.HexToAddress("0x2d10075E54356E16Ebd5C6BB5194290709B69C1e"),
		Timestamp:     1697025629,
		Id:            1,
		TotalAmount:   big.NewInt(1e18),
		ChainId:       1442,
		MultiPayCtrct: multiPayCtrctAddr,
	}
	brokerAddrStr := brokerAddr.String()
	execAddrStr := execAddr.String()
	t.Log("brokerAddr = ", brokerAddrStr)
	t.Log("execAddr = ", execAddrStr)

	pk := fmt.Sprintf("%x", execPk.D)
	execWallet, err := d8x_futures.NewWallet(pk, 1442, nil)
	if err != nil {
		t.Errorf("error creating wallet")
	}
	_, sg, err := d8x_futures.RawCreatePaymentBrokerSignature(&summary, execWallet)
	if err != nil {
		fmt.Println(err.Error())
		t.FailNow()
	}
	data := d8x_futures.BrokerPaySignatureReq{
		Payment:           summary,
		ExecutorSignature: sg,
	}
	fmt.Println(data)
	chConfig, err := LoadChainConfig("../../config/chainConfig.json")
	if err != nil {
		t.Errorf("loading deploymentconfig: %v", err)
		return
	}
	rpcConfig, err := LoadRpcConfig("../../config/rpc.json")
	if err != nil {
		t.Errorf("loading deploymentconfig: %v", err)
		return
	}
	pkBrker := fmt.Sprintf("%x", brokerPk.D)
	pen, err := NewSignaturePen(pkBrker, chConfig, rpcConfig)
	if err != nil {
		fmt.Println(err.Error())
		t.FailNow()
	}
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
	addr, err := d8x_futures.RecoverPaymentSignatureAddr(sigBytes, &summary)
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
		return
	}

	viper.SetDefault(env.API_BIND_ADDR, "")
	viper.SetDefault(env.API_PORT, "8000")

	requiredEnvs := []string{
		env.BROKER_FEE_TBPS,
	}

	for _, e := range requiredEnvs {
		if !viper.IsSet(e) {
			log.Fatalf("required environment variable not set variable")
			return
		}
	}
}
