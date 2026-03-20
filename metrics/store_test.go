package metrics

import (
	"testing"
	"time"
)

func TestPruneOutOfOrderAndPercentiles(t *testing.T) {
	s := NewStore(1 * time.Minute)
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	// Add a recent event (within window)
	s.Add("chan", 10, now.Add(-30*time.Second))

	// Add an older event (outside window) but append after the recent one (out-of-order)
	s.Add("chan", 1000, now.Add(-2*time.Minute))

	snap := s.Snapshot(now)
	if snap.Count != 1 {
		t.Fatalf("expected count 1 after prune, got %d", snap.Count)
	}
	if snap.P99DelayMS != 10 {
		t.Fatalf("expected P99 10, got %v", snap.P99DelayMS)
	}
}
