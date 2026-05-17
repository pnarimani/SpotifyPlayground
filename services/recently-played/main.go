package main

import (
	"context"
	"log/slog"
	"os/signal"
	"recently_played/api"
	"recently_played/consumer"
	"recently_played/db"
	"syscall"

	"golang.org/x/sync/errgroup"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	db.Init([]string{})
	defer db.Close()

	g, gCtx := errgroup.WithContext(ctx)
	g.Go(func() error { return api.StartServer(gCtx) } )
	g.Go(func() error { return consumer.ConsumeEvents(gCtx) } )
	err := g.Wait()

	if err != nil {
		slog.Error(err.Error())
	}
}
