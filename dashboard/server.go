package dashboard

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"themiyadk/pg-notify/metrics"
	"time"
)

func StartServer(ctx context.Context, port int, store *metrics.Store, hub *Hub) error {
	handler := NewHandler(store, hub)
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: handler,
	}

	errCh := make(chan error, 1)
	go func() {
		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
		return <-errCh
	case err := <-errCh:
		return err
	}
}
