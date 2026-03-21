package dashboard

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"themiyadk/pg-notify/metrics"
	"time"
)

type Handler struct {
	store *metrics.Store
	hub   *Hub
}

func NewHandler(store *metrics.Store, hub *Hub) http.Handler {
	h := &Handler{store: store, hub: hub}
	mux := http.NewServeMux()
	mux.HandleFunc("/", h.handleIndex)
	mux.HandleFunc("/api/metrics", h.handleMetrics)
	mux.HandleFunc("/events", h.handleEvents)
	return mux
}

func (h *Handler) handleIndex(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(indexHTML))
}

func (h *Handler) handleMetrics(w http.ResponseWriter, _ *http.Request) {
	snapshot := h.store.Snapshot(time.Now().UTC())
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(snapshot); err != nil {
		http.Error(w, "failed to encode metrics", http.StatusInternalServerError)
	}
}

func (h *Handler) handleEvents(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	// SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	// w.Header().Set("X-Accel-Buffering", "no")

	subChannel, unsubscribe := h.hub.Subscribe()
	defer unsubscribe()

	sendSnapshot := func() bool {
		snapshot := h.store.Snapshot(time.Now().UTC())
		b, err := json.Marshal(snapshot)
		if err != nil {
			return false
		}
		if _, err := fmt.Fprintf(w, "event: metrics\ndata: %s\n\n", b); err != nil {
			return false
		}
		flusher.Flush()
		return true
	}

	if !sendSnapshot() {
		return
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			if !sendSnapshot() {
				return
			}
		case <-subChannel:
			if !sendSnapshot() {
				return
			}
		}
	}
}

// magic
//go:embed templates/index.html
var indexHTML string
