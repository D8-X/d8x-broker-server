package utils

import (
	"log/slog"
	"math/rand"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
)

func CreateRpcClient(rpcUrl []string) (*ethclient.Client, error) {
	rnd := rand.Intn(len(rpcUrl))
	var rpc *ethclient.Client
	var err error
	for trial := 0; ; trial++ {
		rpc, err = ethclient.Dial(rpcUrl[rnd])
		if err != nil {
			if trial == 5 {
				return nil, err
			}
			slog.Info("Rpc error" + err.Error() + " retrying " + strconv.Itoa(5-trial))
			time.Sleep(time.Duration(2) * time.Second)
		} else {
			break
		}
	}
	return rpc, nil
}
