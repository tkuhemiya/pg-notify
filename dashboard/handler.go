package dashboard

import (
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
	_, _ = w.Write([]byte(indexHTML))
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

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	sub, unsubscribe := h.hub.Subscribe()
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
		case <-sub:
			if !sendSnapshot() {
				return
			}
		}
	}
}

var indexHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>PG Notify Metrics</title>
  <script src="https://code.jquery.com/jquery-3.7.1.min.js" integrity="sha256-/JqT3SQfawRcv/BIHPThkBvs0OEvtFFmqPF/lYI/Cxo=" crossorigin="anonymous"></script>
  <style>
    :root {
      --bg: #f7f9fc;
      --surface: #ffffff;
      --ink: #13253a;
      --muted: #5a6b7c;
      --border: #dbe3ec;
      --accent: #1070ff;
    }
    * { box-sizing: border-box; }
    body {
      margin: 0;
      font-family: "Segoe UI", Tahoma, sans-serif;
      color: var(--ink);
      background: radial-gradient(circle at top right, #e7f0ff 0, var(--bg) 35%);
    }
    .wrap { max-width: 980px; margin: 0 auto; padding: 20px; }
    h1 { margin: 0 0 4px; }
    .sub { color: var(--muted); margin-bottom: 20px; }
    .status { font-weight: 600; margin-bottom: 14px; }
    .grid {
      display: grid;
      grid-template-columns: repeat(4, minmax(0, 1fr));
      gap: 12px;
      margin-bottom: 16px;
    }
    .card {
      background: var(--surface);
      border: 1px solid var(--border);
      border-radius: 10px;
      padding: 12px;
    }
    .label { color: var(--muted); font-size: 13px; }
    .value { font-size: 26px; font-weight: 700; margin-top: 4px; }
    table {
      width: 100%;
      border-collapse: collapse;
      background: var(--surface);
      border: 1px solid var(--border);
      border-radius: 10px;
      overflow: hidden;
    }
    th, td {
      text-align: left;
      padding: 10px;
      border-bottom: 1px solid var(--border);
      font-size: 14px;
    }
    th { background: #eef4ff; }
    tr:last-child td { border-bottom: none; }
    @media (max-width: 760px) {
      .grid { grid-template-columns: 1fr 1fr; }
    }
    @media (max-width: 420px) {
      .grid { grid-template-columns: 1fr; }
    }
  </style>
</head>
<body>
  <div class="wrap">
    <h1>Notification Metrics</h1>
    <div class="sub">Rolling 5-minute operational view</div>
    <div class="status">SSE status: <span id="conn">connecting</span></div>

    <div class="grid">
      <div class="card"><div class="label">Total Rate (/s)</div><div class="value" id="rate">0.00</div></div>
      <div class="card"><div class="label">P90 Delay (ms)</div><div class="value" id="p90">0</div></div>
      <div class="card"><div class="label">P99 Delay (ms)</div><div class="value" id="p99">0</div></div>
      <div class="card"><div class="label">Count (5m)</div><div class="value" id="count">0</div></div>
    </div>

    <div class="card" style="margin-bottom: 16px;">
      <div class="label">Last Event At</div>
      <div id="last_event">n/a</div>
    </div>

    <table>
      <thead>
        <tr>
          <th>Channel</th>
          <th>Count</th>
          <th>Rate (/s)</th>
          <th>P90 (ms)</th>
          <th>P99 (ms)</th>
          <th>Last Event</th>
        </tr>
      </thead>
      <tbody id="channel_rows"></tbody>
    </table>
  </div>

  <script>
    function fmtNum(v, d) {
      if (v === null || v === undefined) return '0';
      return Number(v).toFixed(d);
    }

    function fmtTS(v) {
      if (!v) return 'n/a';
      const dt = new Date(v);
      if (Number.isNaN(dt.getTime())) return 'n/a';
      return dt.toLocaleString();
    }

    function render(snapshot) {
      $('#rate').text(fmtNum(snapshot.rate_per_sec, 2));
      $('#p90').text(fmtNum(snapshot.p90_delay_ms, 0));
      $('#p99').text(fmtNum(snapshot.p99_delay_ms, 0));
      $('#count').text(snapshot.count || 0);
      $('#last_event').text(fmtTS(snapshot.last_event_at));

      const channels = snapshot.channels || {};
      const names = Object.keys(channels).sort();
      const rows = names.map(function(name) {
        const c = channels[name] || {};
        return '<tr>' +
          '<td>' + name + '</td>' +
          '<td>' + (c.count || 0) + '</td>' +
          '<td>' + fmtNum(c.rate_per_sec, 2) + '</td>' +
          '<td>' + fmtNum(c.p90_delay_ms, 0) + '</td>' +
          '<td>' + fmtNum(c.p99_delay_ms, 0) + '</td>' +
          '<td>' + fmtTS(c.last_event_at) + '</td>' +
          '</tr>';
      }).join('');

      $('#channel_rows').html(rows || '<tr><td colspan="6">No data in rolling window</td></tr>');
    }

    function setStatus(s) {
      $('#conn').text(s);
    }

    function connect() {
      const es = new EventSource('/events');
      es.addEventListener('metrics', function(ev) {
        try {
          const payload = JSON.parse(ev.data);
          render(payload);
        } catch (_) {}
      });
      es.onopen = function() { setStatus('connected'); };
      es.onerror = function() { setStatus('reconnecting'); };
      return es;
    }

    $(function() {
      $.getJSON('/api/metrics').done(render).fail(function() {
        setStatus('api unavailable');
      });
      connect();
    });
  </script>
</body>
</html>`
