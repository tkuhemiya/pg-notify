package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"themiyadk/pg-notify/config"
	"themiyadk/pg-notify/listener"
)

func eventHandler(ctx context.Context, PID uint32, channel string, payload string) {
	fmt.Println(PID, channel, payload)
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.New()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	l := listener.New(cfg, eventHandler)
	if err := l.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatalf("listener exited: %v", err)
	}
}
