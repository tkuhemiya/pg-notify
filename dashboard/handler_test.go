package dashboard

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"themiyadk/pg-notify/metrics"
	"time"
)

func TestIndexEndpoint(t *testing.T) {
	store := metrics.NewStore(5 * time.Minute)
	hub := NewHub()
	handler := NewHandler(store, hub)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Header().Get("Content-Type"), "text/html") {
		t.Fatalf("expected html content type, got %q", rec.Header().Get("Content-Type"))
	}
	if !strings.Contains(rec.Body.String(), "Notification Metrics") {
		t.Fatalf("expected dashboard html in body")
	}
}

func TestMetricsEndpoint(t *testing.T) {
	store := metrics.NewStore(5 * time.Minute)
	now := time.Now().UTC()
	store.Add("orders_inserted", 10, now)
	store.Add("orders_inserted", 20, now)
	hub := NewHub()
	handler := NewHandler(store, hub)

	req := httptest.NewRequest(http.MethodGet, "/api/metrics", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Header().Get("Content-Type"), "application/json") {
		t.Fatalf("expected json content type, got %q", rec.Header().Get("Content-Type"))
	}

	var got metrics.Snapshot
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to decode json: %v", err)
	}
	if got.Count != 2 {
		t.Fatalf("expected count 2, got %d", got.Count)
	}
	if _, ok := got.Channels["orders_inserted"]; !ok {
		t.Fatalf("expected channel orders_inserted in response")
	}
}

func TestEventsEndpointSendsMetricsEvent(t *testing.T) {
	store := metrics.NewStore(5 * time.Minute)
	store.Add("orders_inserted", 15, time.Now().UTC())
	hub := NewHub()
	handler := NewHandler(store, hub)

	ctx, cancel := context.WithCancel(context.Background())
	req := httptest.NewRequest(http.MethodGet, "/events", nil).WithContext(ctx)
	rec := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		handler.ServeHTTP(rec, req)
		close(done)
	}()

	time.Sleep(20 * time.Millisecond)
	cancel()
	<-done

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); got != "text/event-stream" {
		t.Fatalf("expected text/event-stream content type, got %q", got)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "event: metrics") {
		t.Fatalf("expected SSE event name in body, got %q", body)
	}
	if !strings.Contains(body, "data: ") {
		t.Fatalf("expected SSE data payload in body, got %q", body)
	}
}
