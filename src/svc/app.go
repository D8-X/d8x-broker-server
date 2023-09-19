package svc

import (
	"errors"
	"log"
	"log/slog"

	"github.com/D8-X/d8x-broker-server/src/api"
	"github.com/D8-X/d8x-broker-server/src/config"
	"github.com/D8-X/d8x-broker-server/src/env"
	"github.com/D8-X/d8x-broker-server/src/utils"
	"github.com/spf13/viper"
)

func Run() {

	err := loadEnv()
	if err != nil {
		slog.Error("loading env: " + err.Error())
		return
	}
	config, err := config.LoadChainConfig(viper.GetString(env.CONFIG_PATH))
	if err != nil {
		slog.Error("loading deploymentconfig: " + err.Error())
		return
	}
	pk := viper.GetString(env.BROKER_KEY)
	pen, err := utils.NewSignaturePen(pk, config)
	if err != nil {
		log.Fatalf("unable to create signature pen: %v", err)
	}
	fee := uint16(viper.GetInt32(env.BROKER_FEE_TBPS))
	slog.Info("starting REST API server")
	// Start the rest api
	app := &api.App{
		Port:          viper.GetString(env.API_PORT),
		BindAddr:      viper.GetString(env.API_BIND_ADDR),
		Pen:           pen,
		BrokerFeeTbps: fee,
	}
	app.StartApiServer()

}

func loadEnv() error {

	viper.SetConfigFile(".env")
	if err := viper.ReadInConfig(); err != nil {
		return errors.New("could not load .env file" + err.Error())
	}

	viper.AutomaticEnv()

	viper.SetDefault(env.API_BIND_ADDR, "")
	viper.SetDefault(env.API_PORT, "8000")

	requiredEnvs := []string{
		env.BROKER_KEY,
		env.BROKER_FEE_TBPS,
		env.CONFIG_PATH,
	}

	for _, e := range requiredEnvs {
		if !viper.IsSet(e) {
			return errors.New("required environment variable not set variable" + e)
		}
	}
	return nil
}
