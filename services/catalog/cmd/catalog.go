package main

import (
	"catalog/internal/albums"
	"catalog/internal/api"
	"catalog/internal/cache"
	"catalog/internal/config"
	"catalog/internal/postgres"
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.ReadFromEnv()
	if err != nil {
		slog.Error("invalid configuration", "err", err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.LogLevel,
	}))

	logger.InfoContext(ctx, "service starting")

	pg, err := postgres.New(ctx, cfg, logger.With("component", "postgres"))
	if err != nil {
		logger.ErrorContext(ctx, "postgres initialization failed", "err", err)
		os.Exit(1)
	}
	defer pg.Close()

	cachedRepo, err := cache.New(ctx, pg, logger.With("component", "cache"), cfg)
	if err != nil {
		logger.ErrorContext(ctx, "redis initialization failed", "err", err)
		os.Exit(1)
	}
	defer cachedRepo.Close()

	albumsService := albums.New(cachedRepo)

	server := api.New(logger.With("component", "server"), albumsService)
	if err := server.StartServer(ctx); err != nil {
		logger.ErrorContext(ctx, "server failed", "err", err)
		os.Exit(1)
	}
}
