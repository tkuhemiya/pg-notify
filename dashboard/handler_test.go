package dashboard

import (
	"bufio"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"themiyadk/pg-notify/metrics"
	"time"
)

func TestAPIMetricsReturnsSnapshot(t *testing.T) {
	store := metrics.NewStore(5 * time.Minute)
	hub := NewHub()
	now := time.Date(2026, 3, 8, 10, 0, 0, 0, time.UTC)
	store.Add("orders_inserted", 25, now)

	ts := httptest.NewServer(NewHandler(store, hub))
	defer ts.Close()

	res, err := http.Get(ts.URL + "/api/metrics")
	if err != nil {
		t.Fatalf("failed request: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}

	var snap metrics.Snapshot
	if err := json.NewDecoder(res.Body).Decode(&snap); err != nil {
		t.Fatalf("failed decode: %v", err)
	}
	if snap.WindowSeconds != 300 {
		t.Fatalf("expected window_seconds=300, got %d", snap.WindowSeconds)
	}
}

func TestSSEEventsEmitsMetricsEvent(t *testing.T) {
	store := metrics.NewStore(5 * time.Minute)
	hub := NewHub()
	now := time.Date(2026, 3, 8, 10, 0, 0, 0, time.UTC)
	store.Add("orders_inserted", 10, now)

	ts := httptest.NewServer(NewHandler(store, hub))
	defer ts.Close()

	client := &http.Client{Timeout: 3 * time.Second}
	res, err := client.Get(ts.URL + "/events")
	if err != nil {
		t.Fatalf("failed connect sse: %v", err)
	}
	defer res.Body.Close()

	reader := bufio.NewReader(res.Body)
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		line, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("failed read sse line: %v", err)
		}
		if strings.HasPrefix(line, "event: metrics") {
			return
		}
	}
	t.Fatal("did not receive metrics event")
}
