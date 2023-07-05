package svc

import (
	"log"

	"github.com/D8-X/d8x-broker-server/src/api"
	"github.com/D8-X/d8x-broker-server/src/config"
	"github.com/D8-X/d8x-broker-server/src/env"
	"github.com/D8-X/d8x-broker-server/src/utils"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func Run() {
	l, err := GetDefaultLogger()
	if err != nil {
		log.Fatalf("creating logger: %v", err)
	}
	loadEnv(l)
	config, err := config.LoadChainConfig("../config/chainConfig.json")
	if err != nil {
		log.Fatalf("loading deploymentconfig: %v", err)
	}
	pk := viper.GetString(env.BROKER_KEY)
	pen, err := utils.NewSignaturePen(pk, config)
	if err != nil {
		log.Fatalf("unable to create signature pen: %v", err)
	}
	fee := uint16(viper.GetInt32(env.BROKER_FEE_TBPS))
	l.Info("starting REST API server")
	// Start the rest api
	app := &api.App{
		Logger:        l,
		Port:          viper.GetString(env.API_PORT),
		BindAddr:      viper.GetString(env.API_BIND_ADDR),
		Pen:           pen,
		BrokerFeeTbps: fee,
	}
	app.StartApiServer()

}

func loadEnv(l *zap.Logger) {

	viper.SetConfigFile("../.env")
	if err := viper.ReadInConfig(); err != nil {
		l.Warn("could not load .env file", zap.Error(err))
	}

	viper.SetDefault(env.API_BIND_ADDR, "")
	viper.SetDefault(env.API_PORT, "8000")

	requiredEnvs := []string{
		env.BROKER_KEY,
		env.BROKER_FEE_TBPS,
	}

	for _, e := range requiredEnvs {
		if !viper.IsSet(e) {
			l.Fatal("required environment variable not set", zap.String("variable", e))
		}
	}
}
