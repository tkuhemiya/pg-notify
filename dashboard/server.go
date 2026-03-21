package dashboard

import (
	"context"
	"fmt"
	"net/http"
	"themiyadk/pg-notify/metrics"
)

func StartServer(ctx context.Context, port int, store *metrics.Store, hub *Hub) error {
	handler := NewHandler(store, hub)
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: handler,
	}

	errCh := make(chan error, 1)
	go func() {
		// server.Shutdown(ctx) will makes this error http.ErrServerClosed
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			errCh <- err
			return
		}
	}()

	select {
	case <-ctx.Done():
		err := server.Shutdown(ctx)
		return err
	case err := <-errCh:
		return err
	}
}
