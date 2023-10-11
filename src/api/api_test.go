package api

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"testing"

	"github.com/D8-X/d8x-broker-server/src/utils"
	"github.com/ethereum/go-ethereum/common"

	"github.com/D8-X/d8x-broker-server/src/config"
)

func TestApproveTokenAmount(t *testing.T) {
	pk := os.Getenv("PK")

	CONFIG_PATH := "../../config/chainConfig.json"
	CONFIG_RPC_PATH := "../../config/rpc.json"
	if pk == "" {
		fmt.Printf("Private key not set in environment (export PK)")
		t.Fail()
	}
	chConf, err := config.LoadChainConfig(CONFIG_PATH)
	if err != nil {
		slog.Error("loading chain config: " + err.Error())
		return
	}
	rpcConf, err := config.LoadRpcConfig(CONFIG_RPC_PATH)
	if err != nil {
		slog.Error("loading rpc config: " + err.Error())
		t.Fail()
		return
	}
	pen, err := utils.NewSignaturePen(pk, chConf, rpcConf)
	if err != nil {
		log.Fatalf("unable to create signature pen: %v", err)
		t.Fail()
	}
	app := &App{
		Port:           "80001",
		BindAddr:       "0.0.0.0",
		Pen:            pen,
		BrokerFeeTbps:  60,
		ApprovedTokens: make(map[string]bool),
	}
	if err != nil {
		log.Fatalf("unable to create app: " + err.Error())
		t.Fail()
	}
	mockTkn := common.HexToAddress("0x37D97d1FFc09587EA9BDF88Ea77ec4aFAA911260")
	err = app.ApproveToken(1442, mockTkn)
	if err != nil {
		t.Fail()
	}
	fmt.Println("success")
}
