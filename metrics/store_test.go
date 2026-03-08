package metrics

import (
	"testing"
	"time"
)

func TestStoreSnapshotRollingWindowAndPercentiles(t *testing.T) {
	window := 5 * time.Minute
	store := NewStore(window)
	base := time.Date(2026, 3, 8, 10, 0, 0, 0, time.UTC)

	store.Add("orders_inserted", 100, base.Add(-6*time.Minute))
	store.Add("orders_inserted", 10, base.Add(-4*time.Minute))
	store.Add("orders_inserted", 20, base.Add(-3*time.Minute))
	store.Add("orders_inserted", 30, base.Add(-2*time.Minute))
	store.Add("orders_inserted", 40, base.Add(-1*time.Minute))

	snap := store.Snapshot(base)
	if snap.Count != 4 {
		t.Fatalf("expected 4 events in window, got %d", snap.Count)
	}
	if snap.P90DelayMS != 40 {
		t.Fatalf("expected p90=40, got %.0f", snap.P90DelayMS)
	}
	if snap.P99DelayMS != 40 {
		t.Fatalf("expected p99=40, got %.0f", snap.P99DelayMS)
	}

	expectedRate := float64(4) / 300.0
	if snap.RatePerSec != expectedRate {
		t.Fatalf("expected rate %.6f, got %.6f", expectedRate, snap.RatePerSec)
	}

	ch := snap.Channels["orders_inserted"]
	if ch.Count != 4 {
		t.Fatalf("expected channel count 4, got %d", ch.Count)
	}
}

func TestStoreEmptyWindow(t *testing.T) {
	store := NewStore(5 * time.Minute)
	now := time.Date(2026, 3, 8, 10, 0, 0, 0, time.UTC)

	snap := store.Snapshot(now)
	if snap.Count != 0 {
		t.Fatalf("expected zero count, got %d", snap.Count)
	}
	if snap.P90DelayMS != 0 || snap.P99DelayMS != 0 {
		t.Fatalf("expected zero percentiles, got p90=%.0f p99=%.0f", snap.P90DelayMS, snap.P99DelayMS)
	}
	if snap.LastEventAt != nil {
		t.Fatalf("expected nil last_event_at")
	}
}
