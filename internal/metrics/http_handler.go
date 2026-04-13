package metrics

import (
	"net/http"
	"strings"
)

// Handler serves metrics over HTTP in either JSON or plain text format.
// The format is selected via the Accept header or a `format` query parameter.
type Handler struct {
	exporter *Exporter
}

// NewHandler creates a new HTTP handler backed by the given Exporter.
func NewHandler(e *Exporter) *Handler {
	if e == nil {
		panic("metrics: NewHandler requires a non-nil Exporter")
	}
	return &Handler{exporter: e}
}

// ServeHTTP handles GET /metrics requests.
// Responds with JSON when Accept or format=json is requested; plain text otherwise.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	format := r.URL.Query().Get("format")
	accept := r.Header.Get("Accept")

	if format == "json" || strings.Contains(accept, "application/json") {
		w.Header().Set("Content-Type", "application/json")
		if err := h.exporter.WriteJSON(w); err != nil {
			http.Error(w, "failed to write metrics", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if err := h.exporter.WriteText(w); err != nil {
		http.Error(w, "failed to write metrics", http.StatusInternalServerError)
	}
}

// Mux returns an http.ServeMux with the metrics handler registered at /metrics.
func (h *Handler) Mux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/metrics", h)
	return mux
}
