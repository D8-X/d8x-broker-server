package svc

import "go.uber.org/zap"

func GetDefaultLogger() (*zap.Logger, error) {
	return zap.NewDevelopment()
}