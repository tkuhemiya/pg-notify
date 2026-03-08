package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"
	"themiyadk/pg-notify/config"
	"themiyadk/pg-notify/dashboard"
	"themiyadk/pg-notify/listener"
	"themiyadk/pg-notify/metrics"
	"time"

	"golang.org/x/sync/errgroup"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.New()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	store := metrics.NewStore(5 * time.Minute)
	hub := dashboard.NewHub()

	eventHandler := func(_ context.Context, _ uint32, channel string, payload string) {
		receivedAt := time.Now().UTC()
		delayMS := metrics.IngestDelayMS(payload, receivedAt)
		store.Add(channel, delayMS, receivedAt)
		hub.Notify()
	}

	l := listener.New(cfg, eventHandler)

	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return l.Start(gctx)
	})
	g.Go(func() error {
		return dashboard.StartServer(gctx, cfg.Port, store, hub)
	})

	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatalf("service exited: %v", err)
	}
}
