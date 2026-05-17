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
	"recently_played/reader"
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

	conf := config.ReadFromEnv()

	store, err := db.New(conf)
	if err != nil {
		slog.ErrorContext(ctx, "db initialization failed", "err", err)
		return
	}
	defer store.Close()

	slog.Info("database initialized")

	readService := reader.New(conf, store)

	serv := api.New(&readService)

	cons := consumer.New(conf, store)

	g, gCtx := errgroup.WithContext(ctx)
	g.Go(func() error { return serv.StartServer(gCtx) })
	g.Go(func() error { return cons.ConsumeEvents(gCtx) })
	err = g.Wait()

	if err != nil {
		slog.ErrorContext(ctx, err.Error())
	}
}
