package metrics

import (
	"encoding/json"
	"time"
)

type payloadCreatedAt struct {
	CreatedAt string `json:"created_at"`
}

func IngestDelayMS(payload string, receivedAt time.Time) float64 {
	var parsed payloadCreatedAt
	if err := json.Unmarshal([]byte(payload), &parsed); err != nil {
		return 0
	}
	if parsed.CreatedAt == "" {
		return 0
	}

	eventAt, err := time.Parse(time.RFC3339Nano, parsed.CreatedAt)
	if err != nil {
		return 0
	}
	delay := receivedAt.Sub(eventAt.UTC()).Milliseconds()
	if delay < 0 {
		return 0
	}
	return float64(delay)
}
