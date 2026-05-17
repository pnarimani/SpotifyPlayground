package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"recently_played/api"
	"recently_played/config"
	"recently_played/consumer"
	"recently_played/db"
	"syscall"

	"golang.org/x/sync/errgroup"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: config.GetLogLevel(),
	}))

	slog.SetDefault(logger)

	slog.Info("service starting")

	if err := db.Init(config.GetCassandraHosts()); err != nil {
		slog.ErrorContext(ctx, "db initialization failed", "err", err)
		return
	}
	defer db.Close()

	slog.Info("database initialized")

	g, gCtx := errgroup.WithContext(ctx)
	g.Go(func() error { return api.StartServer(gCtx) })
	g.Go(func() error { return consumer.ConsumeEvents(gCtx) })
	err := g.Wait()

	if err != nil {
		slog.ErrorContext(ctx, err.Error())
	}
}
