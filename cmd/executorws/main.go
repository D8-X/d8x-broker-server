package main

import (
	"log/slog"

	"github.com/D8-X/d8x-broker-server/src/svc"
)

// Injected via -ldflags -X
var VERSION = "broker-executor-ws-development"

func main() {
	slog.Info("starting service",
		slog.String("name", "broker-executor-ws"),
		slog.String("version", VERSION),
	)
	svc.RunExecutorWs()
}
