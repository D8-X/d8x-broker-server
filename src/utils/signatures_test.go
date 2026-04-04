package utils

import (
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

	chConfig, err := LoadBrokerConfig("../../config/chainConfig.json")
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
	perpOrder := contracts.IPerpetualOrderOrder{
		BrokerFeeTbps: 10,
		TraderAddr:    common.HexToAddress("9d5aaB428e98678d0E645ea4AeBd25f744341a05"),
		BrokerAddr:    addr,
		IDeadline:     1742381717,
		IPerpetualId:  big.NewInt(int64(10000)),
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
