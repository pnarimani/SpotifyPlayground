package main

import (
	"auth/internals/api"
	"auth/internals/config"
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: config.GetLogLevel(),
	}))

	slog.SetDefault(logger)
	slog.Info("service starting")

	serv := api.New(logger.With("Component", "Api"))

	if err := serv.StartServer(ctx); err != nil {
		slog.ErrorContext(ctx, err.Error())
	}
}
