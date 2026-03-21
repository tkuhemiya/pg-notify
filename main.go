package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"themiyadk/pg-notify/config"
	"themiyadk/pg-notify/dashboard"
	"themiyadk/pg-notify/listener"
	"themiyadk/pg-notify/metrics"
	"time"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.NewFromFile("pg-notify.cfg")
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

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		l.Start(ctx)
	}()

	go func() {
		defer wg.Done()
		dashboard.StartServer(ctx, cfg.Port, store, hub)
	}()

	wg.Wait()
}
