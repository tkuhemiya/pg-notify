package listener

import (
	"context"
	"fmt"
	"themiyadk/pg-notify/config"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Listener struct {
	config   config.Config
	callback func(context.Context, uint32, string, string)
	sem      chan struct{}
}

func New(config config.Config, callback func(context.Context, uint32, string, string)) *Listener {
	return &Listener{
		config:   config,
		callback: callback,
		sem:      make(chan struct{}, 255),
	}
}

func (l *Listener) Start(ctx context.Context) error {
	fmt.Println("Starting Server")
	conn, err := pgx.Connect(ctx, l.config.DatabaseURL)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)
	for _, ch := range l.config.EventName {
		ident := pgx.Identifier{ch}.Sanitize()
		if _, err := conn.Exec(ctx, "LISTEN "+ident); err != nil {
			return err
		}
	}
	for {
		n, err := conn.WaitForNotification(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}
		select {
		case l.sem <- struct{}{}: // acquire
			go func(n *pgconn.Notification) {
				defer func() { <-l.sem }() // release
				l.callback(ctx, n.PID, n.Channel, n.Payload)
			}(n)
		case <-ctx.Done():
			return nil
		}
	}
}
