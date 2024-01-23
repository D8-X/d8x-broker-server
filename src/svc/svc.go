package svc

import (
	"embed"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/D8-X/d8x-broker-server/src/api"
	"github.com/D8-X/d8x-broker-server/src/config"
	"github.com/D8-X/d8x-broker-server/src/env"
	"github.com/D8-X/d8x-broker-server/src/executorws"
	"github.com/D8-X/d8x-broker-server/src/utils"
	"github.com/spf13/viper"
)

//go:embed ranky.txt
var embedFS embed.FS
var abc []byte

func init() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
	})))
}

func RunExecutorWs() {
	requiredEnvs := []string{
		env.CONFIG_PATH,
		env.REDIS_ADDR,
		env.REDIS_PW,
	}
	err := loadEnv(requiredEnvs)
	if err != nil {
		slog.Error("loading env: " + err.Error())
		return
	}
	config, err := config.LoadChainConfig(viper.GetString(env.CONFIG_PATH))
	if err != nil {
		slog.Error("loading chain config: " + err.Error())
		return
	}
	wsAddr := viper.GetString(env.WS_ADDR)
	redisAddr := viper.GetString(env.REDIS_ADDR)
	redisPw := viper.GetString(env.REDIS_PW)
	err = executorws.StartWSServer(config, wsAddr, redisAddr, redisPw)
	if err != nil {
		slog.Error("Executor WS server: " + err.Error())
	}
}

func RunBroker() {
	requiredEnvs := []string{
		env.BROKER_FEE_TBPS,
		env.CONFIG_PATH,
		env.REDIS_ADDR,
		env.REDIS_PW,
		env.KEYFILE_PATH,
		env.CONFIG_RPC_PATH,
	}
	err := loadEnv(requiredEnvs)
	if err != nil {
		slog.Error("loading env: " + err.Error())
		return
	}

	fmt.Println("Loading config file from " + viper.GetString(env.CONFIG_PATH))
	chConf, err := config.LoadChainConfig(viper.GetString(env.CONFIG_PATH))
	if err != nil {
		slog.Error("loading chain config: " + err.Error())
		return
	}
	fmt.Println("Loading rpc config file from " + viper.GetString(env.CONFIG_RPC_PATH))
	rpcConf, err := config.LoadRpcConfig(viper.GetString(env.CONFIG_RPC_PATH))
	if err != nil {
		slog.Error("loading rpc config: " + err.Error())
		return
	}
	pk := utils.LoadFromFile(viper.GetString(env.KEYFILE_PATH)+"keyfile.txt", abc)
	pen, err := utils.NewSignaturePen(pk, chConf, rpcConf)
	if err != nil {
		log.Fatalf("unable to create signature pen: %v", err)
	}
	fee := uint16(viper.GetInt32(env.BROKER_FEE_TBPS))
	slog.Info("starting REST API server")
	// Start the rest api
	app := &api.App{
		Port:            viper.GetString(env.API_PORT),
		BindAddr:        viper.GetString(env.API_BIND_ADDR),
		Pen:             pen,
		BrokerFeeTbps:   fee,
		TokenApprovalTs: make(map[string]int64),
	}
	err = app.StartApiServer(viper.GetString(env.REDIS_ADDR),
		viper.GetString(env.REDIS_PW))
	if err != nil {
		slog.Error("API server: " + err.Error())
	}
}

func loadEnv(requiredEnvs []string) error {

	viper.SetConfigFile(".env")
	if err := viper.ReadInConfig(); err != nil {
		slog.Info("could not load .env file using AutomaticEnv")
	}
	loadAbc()
	viper.AutomaticEnv()

	viper.SetDefault(env.API_BIND_ADDR, "")
	viper.SetDefault(env.API_PORT, "8001")
	viper.SetDefault(env.WS_ADDR, "executorws:8080")
	for _, e := range requiredEnvs {
		if !viper.IsSet(e) {
			return errors.New("required environment variable not set variable" + e)
		}
	}
	return nil
}

func loadAbc() {
	content, err := embedFS.ReadFile("ranky.txt")
	if err != nil {
		fmt.Println("Error reading embedded file:", err)
		return
	}
	abc = content
}
