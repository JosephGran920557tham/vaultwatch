package metrics

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestHandler(t *testing.T) *Handler {
	t.Helper()
	reg := NewRegistry()
	c := reg.Counter("requests_total")
	c.Inc()
	c.Inc()
	g := reg.Gauge("active_leases")
	g.Set(5)
	return NewHandler(NewExporter(reg))
}

func TestHTTPHandler_PlainText(t *testing.T) {
	h := newTestHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "requests_total") {
		t.Errorf("expected requests_total in text output, got: %s", body)
	}
	if !strings.Contains(rr.Header().Get("Content-Type"), "text/plain") {
		t.Errorf("expected text/plain content type")
	}
}

func TestHTTPHandler_JSON_ViaQueryParam(t *testing.T) {
	h := newTestHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/metrics?format=json", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var out map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatalf("expected valid JSON: %v", err)
	}
	if _, ok := out["metrics"]; !ok {
		t.Errorf("expected 'metrics' key in JSON output")
	}
}

func TestHTTPHandler_JSON_ViaAcceptHeader(t *testing.T) {
	h := newTestHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.Header.Set("Accept", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if !strings.Contains(rr.Header().Get("Content-Type"), "application/json") {
		t.Errorf("expected application/json content type")
	}
}

func TestHTTPHandler_MethodNotAllowed(t *testing.T) {
	h := newTestHandler(t)
	req := httptest.NewRequest(http.MethodPost, "/metrics", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rr.Code)
	}
}

func TestNewHandler_NilExporter_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for nil exporter")
		}
	}()
	NewHandler(nil)
}

func TestHandler_Mux_RegistersRoute(t *testing.T) {
	h := newTestHandler(t)
	mux := h.Mux()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 from mux, got %d", rr.Code)
	}
}

func TestHTTPHandler_JSON_ContainsExpectedMetrics(t *testing.T) {
	h := newTestHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/metrics?format=json", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	var out map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatalf("expected valid JSON: %v", err)
	}

	metrics, ok := out["metrics"].([]interface{})
	if !ok {
		t.Fatalf("expected 'metrics' to be an array")
	}

	names := make(map[string]bool)
	for _, m := range metrics {
		if entry, ok := m.(map[string]interface{}); ok {
			if name, ok := entry["name"].(string); ok {
				names[name] = true
			}
		}
	}

	for _, expected := range []string{"requests_total", "active_leases"} {
		if !names[expected] {
			t.Errorf("expected metric %q in JSON output, got names: %v", expected, names)
		}
	}
}
