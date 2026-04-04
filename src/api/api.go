package api

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"github.com/D8-X/d8x-broker-server/src/globalrpc"
	"github.com/D8-X/d8x-broker-server/src/utils"
	"github.com/go-chi/chi/v5"
	"github.com/redis/rueidis"
)

// App is dependency container for API server
type App struct {
	Port        string
	BindAddr    string
	Pen         utils.SignaturePen
	BrokerFeeTbps uint16
	RedisClient *utils.RueidisClient
	BrokerConf  map[int64]utils.BrokerConfig
	GlblRpc     map[int64]*globalrpc.GlobalRpc
}

func NewApp(pk, port, bindAddr, REDIS_ADDR, REDIS_PW string, brkrConf map[int64]utils.BrokerConfig, rpcConf string, feeTbps uint16) (*App, error) {
	a := App{
		Port:          port,
		BindAddr:      bindAddr,
		BrokerFeeTbps: feeTbps,
		BrokerConf:    brkrConf,
		GlblRpc:       make(map[int64]*globalrpc.GlobalRpc),
	}
	for cId := range brkrConf {
		r, err := globalrpc.NewGlobalRpc(int(cId), rpcConf, REDIS_ADDR, REDIS_PW)
		if err != nil {
			return nil, fmt.Errorf("unable to create global rpc for chain %d: %v", cId, err)
		}
		a.GlblRpc[cId] = r
	}
	rpcs, err := utils.LoadRpcConfig(rpcConf)
	if err != nil {
		return nil, err
	}
	pen, err := utils.NewSignaturePen(pk, brkrConf, rpcs)
	if err != nil {
		return nil, errors.New("Unable to create signature pen:" + err.Error())
	}
	a.Pen = pen

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

