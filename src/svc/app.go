package svc

import (
	"log"

	"github.com/D8-X/d8x-broker-server/src/api"
	"github.com/D8-X/d8x-broker-server/src/env"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// Struct to represent the JSON argument for the /sign-order endpoint
type Order struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func Run() {
	l, err := GetDefaultLogger()
	if err != nil {
		log.Fatalf("creating logger: %v", err)
	}
	loadEnv(l);
	l.Info("starting REST API server");
	// Start the rest api
	app := &api.App{
		Logger:   l,
		Port:     viper.GetString(env.API_PORT),
		BindAddr: viper.GetString(env.API_BIND_ADDR),
	}
	app.StartApiServer()
	
}


func loadEnv(l *zap.Logger) {

	viper.SetConfigFile(".env")
	if err := viper.ReadInConfig(); err != nil {
		l.Warn("could not load .env file", zap.Error(err))
	}

	viper.SetDefault(env.API_BIND_ADDR, "")
	viper.SetDefault(env.API_PORT, "8000")

	requiredEnvs := []string{
		env.BROKER_KEY,
	}

	for _, e := range requiredEnvs {
		if !viper.IsSet(e) {
			l.Fatal("required environment variable not set", zap.String("variable", e))
		}
	}
}