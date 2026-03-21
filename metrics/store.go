package metrics

import (
	"log"
	"math"
	"sort"
	"sync"
	"time"
)

type Snapshot struct {
	WindowSeconds int                        `json:"window_seconds"`
	Count         int                        `json:"count"`
	RatePerSec    float64                    `json:"rate_per_sec"`
	P90DelayMS    float64                    `json:"p90_delay_ms"`
	P99DelayMS    float64                    `json:"p99_delay_ms"`
	LastEventAt   *time.Time                 `json:"last_event_at"`
	Channels      map[string]ChannelSnapshot `json:"channels"`
}

type ChannelSnapshot struct {
	Count       int        `json:"count"`
	RatePerSec  float64    `json:"rate_per_sec"`
	P90DelayMS  float64    `json:"p90_delay_ms"`
	P99DelayMS  float64    `json:"p99_delay_ms"`
	LastEventAt *time.Time `json:"last_event_at"`
}

type eventRecord struct {
	receivedAt time.Time
	delayMS    float64
	channel    string
}

type Store struct {
	window time.Duration

	mu     sync.Mutex
	events []eventRecord
}

func NewStore(window time.Duration) *Store {
	return &Store{window: window}
}

func (s *Store) Add(channel string, delayMS float64, receivedAt time.Time) {
	if delayMS < 0 {
		delayMS = 0
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.events = append(s.events, eventRecord{
		receivedAt: receivedAt.UTC(),
		delayMS:    delayMS,
		channel:    channel,
	})
	s.removeOldEvents(receivedAt.UTC())
}

func (s *Store) Snapshot(now time.Time) Snapshot {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.removeOldEvents(now)

	res := Snapshot{
		WindowSeconds: int(s.window.Seconds()),
		Channels:      make(map[string]ChannelSnapshot),
	}
	res.Count = len(s.events)
	res.RatePerSec = float64(res.Count) / s.window.Seconds()

	if res.Count == 0 {
		res.P90DelayMS = 0
		res.P99DelayMS = 0
		return res
	}

	totalDelays := make([]float64, 0, len(s.events))
	perChannelDelays := map[string][]float64{}
	perChannelLastAt := map[string]time.Time{}
	var latest time.Time

	for _, ev := range s.events {
		totalDelays = append(totalDelays, ev.delayMS)
		perChannelDelays[ev.channel] = append(perChannelDelays[ev.channel], ev.delayMS)

		if ev.receivedAt.After(latest) {
			latest = ev.receivedAt
		}
		if ev.receivedAt.After(perChannelLastAt[ev.channel]) {
			perChannelLastAt[ev.channel] = ev.receivedAt
		}
	}

	res.P90DelayMS = nearestRankPercentile(totalDelays, 90)
	res.P99DelayMS = nearestRankPercentile(totalDelays, 99)
	latestCopy := latest
	res.LastEventAt = &latestCopy

	for channel, delays := range perChannelDelays {
		channelLast := perChannelLastAt[channel]
		channelLastCopy := channelLast
		count := len(delays)
		res.Channels[channel] = ChannelSnapshot{
			Count:       count,
			RatePerSec:  float64(count) / s.window.Seconds(),
			P90DelayMS:  nearestRankPercentile(delays, 90),
			P99DelayMS:  nearestRankPercentile(delays, 99),
			LastEventAt: &channelLastCopy,
		}
	}

	// Emit a diagnostic log if tail percentiles are large so we can inspect contributing events.
	const warnThresholdMS = 2000.0
	if res.P99DelayMS > warnThresholdMS {
		// copy events and sort by delay desc to show top offenders
		evs := append([]eventRecord(nil), s.events...)
		sort.Slice(evs, func(i, j int) bool { return evs[i].delayMS > evs[j].delayMS })
		n := 10
		if len(evs) < n {
			n = len(evs)
		}
		log.Printf("metrics: high tail percentiles: p90=%.0fms p99=%.0fms count=%d (showing top %d events):", res.P90DelayMS, res.P99DelayMS, res.Count, n)
		for i := 0; i < n; i++ {
			e := evs[i]
			log.Printf("  event[%d] channel=%s receivedAt=%s delay=%.0fms", i, e.channel, e.receivedAt.UTC().Format(time.RFC3339Nano), e.delayMS)
		}
	}

	return res
}

func (s *Store) removeOldEvents(now time.Time) {
	cutoff := now.Add(-s.window)
	// keep only events with receivedAt >= cutoff
	j := 0
	for _, ev := range s.events {
		if !ev.receivedAt.Before(cutoff) {
			s.events[j] = ev
			j++
		}
	}
	s.events = s.events[:j]
}

func nearestRankPercentile(values []float64, percentile float64) float64 {
	if len(values) == 0 {
		return 0
	}
	copied := append([]float64(nil), values...)
	sort.Float64s(copied)

	rank := int(math.Ceil((percentile / 100.0) * float64(len(copied))))
	if rank < 1 {
		rank = 1
	}
	if rank > len(copied) {
		rank = len(copied)
	}
	return copied[rank-1]
}
