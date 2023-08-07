package test

import (
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
	if err != nil {
		log.Fatal(err)
	}
	// Derive the Ethereum address from the private key
	addr := crypto.PubkeyToAddress(privateKey.PublicKey)

	config, err := config.LoadChainConfig("../../config/chainConfig.json")
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
	if err != nil {
		t.Errorf("recovering address: %v", err)
	} else {
		t.Logf("recovered address")
		t.Logf(addrRecovered.String())
	}
	t.Log("recovered addr = ", addrRecovered.String())
	t.Log("signer    addr = ", addr.String())
	if addrRecovered.String() == addr.String() {
		t.Logf("recovered address correct")
	} else {
		t.Errorf("recovering address incorrect")
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
