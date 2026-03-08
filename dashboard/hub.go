package dashboard

import "sync"

type Hub struct {
	mu   sync.Mutex
	subs map[chan struct{}]struct{}
}

func NewHub() *Hub {
	return &Hub{subs: make(map[chan struct{}]struct{})}
}

func (h *Hub) Subscribe() (<-chan struct{}, func()) {
	ch := make(chan struct{}, 1)

	h.mu.Lock()
	h.subs[ch] = struct{}{}
	h.mu.Unlock()

	unsubscribe := func() {
		h.mu.Lock()
		if _, ok := h.subs[ch]; ok {
			delete(h.subs, ch)
			close(ch)
		}
		h.mu.Unlock()
	}

	return ch, unsubscribe
}

func (h *Hub) Notify() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for ch := range h.subs {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}
