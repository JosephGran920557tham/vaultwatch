package tokenwatch

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestHealthChecker(t *testing.T, ttl time.Duration) *HealthChecker {
	t.Helper()
	reg := NewRegistry()
	w, err := New(func(_ context.Context, _ string) (TokenInfo, error) {
		return TokenInfo{TTL: ttl}, nil
	}, 30*time.Second, time.Hour)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	a, err := NewAlerter(reg, w)
	if err != nil {
		t.Fatalf("NewAlerter: %v", err)
	}
	c := DefaultExpiryClassifier()
	hc, err := NewHealthChecker(a, c)
	if err != nil {
		t.Fatalf("NewHealthChecker: %v", err)
	}
	return hc
}

func TestHealthHandler_MethodNotAllowed(t *testing.T) {
	hc := newTestHealthChecker(t, time.Hour)
	h := HealthHandler(hc)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestHealthHandler_HealthyTokens_Returns200(t *testing.T) {
	hc := newTestHealthChecker(t, 24*time.Hour)
	h := HealthHandler(hc)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHealthHandler_ResponseIsValidJSON(t *testing.T) {
	hc := newTestHealthChecker(t, 24*time.Hour)
	h := HealthHandler(hc)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	h.ServeHTTP(rec, req)

	var resp healthResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if resp.Status == "" {
		t.Error("expected non-empty status field")
	}
	if resp.CheckedAt.IsZero() {
		t.Error("expected non-zero checked_at")
	}
}

func TestHealthHandler_ContentTypeJSON(t *testing.T) {
	hc := newTestHealthChecker(t, 24*time.Hour)
	h := HealthHandler(hc)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	h.ServeHTTP(rec, req)
	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected application/json, got %q", ct)
	}
}
