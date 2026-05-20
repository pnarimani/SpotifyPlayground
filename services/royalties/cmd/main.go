package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"royalties/internal/config"
	"royalties/internal/consumer"
	"royalties/internal/postgres"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := config.ReadFromEnv()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.LogLevel,
	}))

	logger.InfoContext(ctx, "service starting")

	pg, err := postgres.New(ctx, cfg, logger.With("Component", "Postgres"))
	if err != nil {
		logger.ErrorContext(ctx, "postgres initialization failed", "err", err)
		return
	}

	con := consumer.New(cfg, logger.With("Component", "Consumer"), pg)
	if err := con.ConsumeEvents(ctx); err != nil {
		logger.ErrorContext(ctx, "consume event failed", "err", err)
		return
	}
}
