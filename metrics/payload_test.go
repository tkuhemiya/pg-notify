package metrics

import (
	"testing"
	"time"
)

func TestIngestDelayMSValidCreatedAt(t *testing.T) {
	receivedAt := time.Date(2026, 3, 8, 10, 0, 5, 0, time.UTC)
	payload := `{"created_at":"2026-03-08T10:00:00Z"}`

	delay := IngestDelayMS(payload, receivedAt)
	if delay != 5000 {
		t.Fatalf("expected 5000ms delay, got %.0f", delay)
	}
}

func TestIngestDelayMSMissingOrInvalidCreatedAt(t *testing.T) {
	receivedAt := time.Date(2026, 3, 8, 10, 0, 5, 0, time.UTC)
	cases := []string{
		`{}`,
		`{"created_at":"bad"}`,
		`not-json`,
	}

	for _, c := range cases {
		if delay := IngestDelayMS(c, receivedAt); delay != 0 {
			t.Fatalf("expected 0ms for payload %q, got %.0f", c, delay)
		}
	}
}

func TestIngestDelayMSClampsNegative(t *testing.T) {
	receivedAt := time.Date(2026, 3, 8, 10, 0, 0, 0, time.UTC)
	payload := `{"created_at":"2026-03-08T10:00:05Z"}`

	delay := IngestDelayMS(payload, receivedAt)
	if delay != 0 {
		t.Fatalf("expected clamped 0ms delay, got %.0f", delay)
	}
}
