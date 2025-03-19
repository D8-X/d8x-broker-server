package api

import (
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/D8-X/d8x-broker-server/src/utils"
)

func TestRestGetVip3Level(t *testing.T) {
	l, err := RestGetVip3Level("0x0aB6527027EcFF1144dEc3d78154fce309ac838c")
	if err != nil {
		fmt.Print(err.Error())
		t.FailNow()
	}
	fmt.Print(l)
	l, err = RestGetVip3Level("0xeddcfdf45d384f7b4e6722e55a92acd7c7dd27e1")
	if err != nil {
		fmt.Print(err.Error())
		t.FailNow()
	}
	fmt.Print(l)
}

func TestGetVip3Level(t *testing.T) {
	chConf, err := utils.LoadChainConfig("../../config/chainConfig.json")
	if err != nil {
		slog.Error("loading chain config: " + err.Error())
		return
	}
	rpcConf, err := utils.LoadRpcConfig("../../config/rpc.json")
	if err != nil {
		slog.Error("loading rpc config: " + err.Error())
		return
	}
	a, err := NewApp(
		"c3aadd4417f0f918fe7a53d7c6c75fa65352a1ef5c29097f0ce5ba8dbf05e08c",
		"8001",
		"127.0.0.1",
		"localhost:6379",
		"23_*PAejOanJma",
		"",
		chConf,
		rpcConf,
		400,
	)
	if err != nil {
		slog.Error("new app creation failed: " + err.Error())
		t.FailNow()
	}
	l := a.GetVip3Level("0x0aB6527027EcFF1144dEc3d78154fce309ac838c")

	fmt.Print(l)
	t0 := time.Now()
	l = a.GetVip3Level("0xabe292b291A18699b09608de86888D77aD6BAf23")

	t1 := time.Now()
	fmt.Print("Found level ", l, " in ", t1.Sub(t0))
	l = a.GetVip3Level("0xabe292b291A18699b09608de86888D77aD6BAf23")

	t0 = time.Now()
	fmt.Print("Found level ", l, " in ", t0.Sub(t1))
}

func TestGetBrokerFeeTbps(t *testing.T) {
	chConf, _ := utils.LoadChainConfig("../../config/chainConfig.json")
	rpcConf, _ := utils.LoadRpcConfig("../../config/rpc.json")
	conf := "1101:50,75,90"
	//conf := ""
	a, err := NewApp(
		"c3aadd4417f0f918fe7a53d7c6c75fa65352a1ef5c29097f0ce5ba8dbf05e08c",
		"8001",
		"127.0.0.1",
		"localhost:6379",
		"23_*PAejOanJma",
		conf,
		chConf,
		rpcConf,
		400,
	)
	if err != nil {
		slog.Error(err.Error())
		t.FailNow()
	}
	fee1 := a.getBrokerFeeTbps("0xB8aAEC178f5b30B6Bcd75740fB64F0369010faDF", 196)
	fee2 := a.getBrokerFeeTbps("0xdA5b972BdA66112E0D9035425AbAda4DaC933C30", 1101)
	fmt.Println("Fee ", fee1)
	fmt.Println("Fee ", fee2)
}

func TestStrToFeeArray(t *testing.T) {
	v := vip3ToFeeMap("50,75,90,100", 60)
	fmt.Println(v)
	v = vip3ToFeeMap("", 60)
	fmt.Println(v)
}
