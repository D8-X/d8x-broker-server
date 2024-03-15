package api

import (
	"context"
	"errors"
	"log/slog"
	"math/big"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/D8-X/d8x-broker-server/src/contracts"
	"github.com/D8-X/d8x-broker-server/src/utils"
	"github.com/D8-X/d8x-futures-go-sdk/pkg/d8x_futures"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-chi/chi/v5"
	"github.com/redis/rueidis"
)

const APPROVAL_EXPIRY_SEC int64 = 86400 * 7

// App is dependency container for API server
type App struct {
	Port              string
	BindAddr          string
	Pen               utils.SignaturePen
	BrokerFeeTbps     uint16
	BrokerFeeLvlsTbps []uint16
	RedisClient       *utils.RueidisClient
	TokenApprovalTs   map[string]int64
}

func NewApp(pk, port, bindAddr, REDIS_ADDR, REDIS_PW, FeeRed string, chConf []utils.ChainConfig, rpcConf []utils.RpcConfig, feeTbps uint16) (*App, error) {
	pen, err := utils.NewSignaturePen(pk, chConf, rpcConf)
	if err != nil {
		return nil, errors.New("Unable to create signature pen:" + err.Error())
	}
	feeRed := strToFeeArray(FeeRed, feeTbps)
	if len(feeRed) > 0 {
		slog.Info("VIP3 reduction enabled")
	}
	a := App{
		Port:              port,
		BindAddr:          bindAddr,
		Pen:               pen,
		BrokerFeeTbps:     feeTbps,
		BrokerFeeLvlsTbps: feeRed,
		TokenApprovalTs:   make(map[string]int64),
	}
	client, err := rueidis.NewClient(
		rueidis.ClientOption{InitAddress: []string{REDIS_ADDR}, Password: REDIS_PW})
	if err != nil {
		return nil, err
	}
	a.RedisClient = &utils.RueidisClient{
		Client: &client,
		Ctx:    context.Background(),
	}
	return &a, nil
}

// StartApiServer initializes and starts the api server. This func is blocking
func (a *App) StartApiServer() error {
	router := chi.NewRouter()
	a.RegisterGlobalMiddleware(router)
	a.RegisterRoutes(router)

	addr := net.JoinHostPort(
		a.BindAddr,
		a.Port,
	)
	slog.Info("starting api server host_port " + addr)
	err := http.ListenAndServe(
		addr,
		router,
	)
	return errors.New("api server is shutting down" + err.Error())
}

func (a *App) ApproveToken(chainId int64, tokenAddr common.Address) error {
	chainIdBI := new(big.Int).SetInt64(chainId)
	key := chainIdBI.String() + "." + tokenAddr.Hex()
	now := time.Now().Unix()
	if now-a.TokenApprovalTs[key] < APPROVAL_EXPIRY_SEC {
		// already approved
		slog.Info("Token already approved for key " + key)
		return nil
	}
	config := a.Pen.ChainConfig[chainId]
	rpcUrls := a.Pen.RpcUrl[chainId]
	if len(rpcUrls) == 0 {
		return errors.New("No rpc url defined for chain " + strconv.Itoa(int(chainId)))
	}
	client, err := utils.CreateRpcClient(rpcUrls)
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
	auth.GasLimit = uint64(300_000)
	g, err := d8x_futures.GetGasPrice(client)
	if err != nil {
		slog.Error("Could not get gas price:" + err.Error())
		return err
	}
	// mark up gas price
	g.Mul(g, big.NewInt(15))
	g.Div(g, big.NewInt(10))
	auth.GasPrice = g
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
	a.TokenApprovalTs[key] = time.Now().Unix()
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
