package api

import (
	"context"
	"errors"
	"log/slog"
	"math/big"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/D8-X/d8x-broker-server/src/contracts"
	"github.com/D8-X/d8x-broker-server/src/utils"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-chi/chi/v5"
	"github.com/redis/rueidis"
)

// App is dependency container for API server
type App struct {
	Port           string
	BindAddr       string
	Pen            utils.SignaturePen
	BrokerFeeTbps  uint16
	RedisClient    *utils.RueidisClient
	ApprovedTokens map[string]bool
}

// StartApiServer initializes and starts the api server. This func is blocking
func (a *App) StartApiServer(REDIS_ADDR string, REDIS_PW string) error {
	if len(a.Port) == 0 {
		return errors.New("could not start the API server, Port must be provided")
	}

	client, err := rueidis.NewClient(
		rueidis.ClientOption{InitAddress: []string{REDIS_ADDR}, Password: REDIS_PW})
	if err != nil {
		return err
	}
	a.RedisClient = &utils.RueidisClient{
		Client: &client,
		Ctx:    context.Background(),
	}
	router := chi.NewRouter()
	a.RegisterGlobalMiddleware(router)
	a.RegisterRoutes(router)

	addr := net.JoinHostPort(
		a.BindAddr,
		a.Port,
	)
	slog.Info("starting api server host_port " + addr)
	err = http.ListenAndServe(
		addr,
		router,
	)
	return errors.New("api server is shutting down" + err.Error())
}

func (a *App) ApproveToken(chainId int64, tokenAddr common.Address) error {
	chainIdBI := new(big.Int).SetInt64(chainId)
	key := chainIdBI.String() + "." + tokenAddr.Hex()
	if a.ApprovedTokens[key] {
		// already approved
		slog.Info("Token already approved for key " + key)
		return nil
	}
	config := a.Pen.ChainConfig[chainId]
	rpcUrls := a.Pen.RpcConfig[chainId]
	if len(rpcUrls) == 0 {
		return errors.New("No rpc url defined for chain " + strconv.Itoa(int(chainId)))
	}
	client, err := createRpcClient(rpcUrls)
	if err != nil {
		return errors.New("Error creating rpc cliet for chain " + strconv.Itoa(int(chainId)) + ": " + err.Error())
	}
	defer client.Close()

	tknInstance, err := contracts.NewErc20(tokenAddr, client)
	if err != nil {
		return errors.New("Error creating token instance " + tokenAddr.String() + " for chain " + strconv.Itoa(int(chainId)) + ": " + err.Error())
	}

	auth, err := bind.NewKeyedTransactorWithChainID(a.Pen.Wallets[chainId].PrivateKey, chainIdBI)
	if err != nil {
		return errors.New("Error creating NewKeyedTransactorWithChainID for chain " + strconv.Itoa(int(chainId)) + ": " + err.Error())
	}
	nonce, err := getNonce(client, a.Pen.Wallets[chainId].Address)
	if err != nil {
		return errors.New("Error getting nonce for chain " + strconv.Itoa(int(chainId)) + ": " + err.Error())
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.GasLimit = uint64(300000)
	auth.GasPrice = big.NewInt(1000000000)
	approvalTx, err := tknInstance.Approve(auth, config.MultiPayCtrctAddr, getMaxUint256())
	if err != nil {
		return errors.New("Error approving token for chain " + strconv.Itoa(int(chainId)) + ": " + err.Error())
	}
	// Wait for the transaction to be mined
	receipt, err := bind.WaitMined(context.Background(), client, approvalTx)
	if err != nil {
		return err
	}
	slog.Info("Approval transaction hash: " + receipt.TxHash.Hex())
	a.ApprovedTokens[key] = true
	return nil
}

func getNonce(rpc *ethclient.Client, a common.Address) (uint64, error) {
	nonce, err := rpc.PendingNonceAt(context.Background(), a)
	if err != nil {
		return 0, err
	}
	return nonce, nil
}

func getMaxUint256() *big.Int {
	maxUint256 := new(big.Int)
	maxUint256.Exp(big.NewInt(2), big.NewInt(256), nil)
	maxUint256.Sub(maxUint256, big.NewInt(1))
	return maxUint256
}

func createRpcClient(rpcUrl []string) (*ethclient.Client, error) {
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
