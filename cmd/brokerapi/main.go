package main

import (
	"log/slog"

	"github.com/D8-X/d8x-broker-server/src/svc"
)

// Injected via -ldflags -X
var VERSION = "broker-api-development"

func main() {
	slog.Info("starting service",
		slog.String("name", "broker-api"),
		slog.String("version", VERSION),
	)
	svc.RunBroker()
}
