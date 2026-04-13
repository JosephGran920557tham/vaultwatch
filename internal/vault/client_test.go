package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newTestServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	return httptest.NewServer(handler)
}

func TestNewClient_InvalidAddress(t *testing.T) {
	_, err := NewClient("://bad-url", "token")
	if err == nil {
		t.Fatal("expected error for invalid address, got nil")
	}
}

func TestLookupLease_Success(t *testing.T) {
	ts := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/v1/sys/leases/lookup") {
			http.NotFound(w, r)
			return
		}
		resp := map[string]interface{}{
			"data": map[string]interface{}{
				"ttl":       float64(3600),
				"renewable": true,
				"id":        "database/creds/my-role/abc123",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer ts.Close()

	client, err := NewClient(ts.URL, "test-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	info, err := client.LookupLease("database/creds/my-role/abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if info.TTL != 3600*time.Second {
		t.Errorf("expected TTL 3600s, got %v", info.TTL)
	}
	if !info.Renewable {
		t.Error("expected lease to be renewable")
	}
	if info.ExpiresAt.Before(time.Now()) {
		t.Error("expected ExpiresAt to be in the future")
	}
}

func TestIsHealthy_Sealed(t *testing.T) {
	ts := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"sealed":      true,
			"initialized": true,
		}
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(resp)
	})
	defer ts.Close()

	client, err := NewClient(ts.URL, "test-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := client.IsHealthy(); err == nil {
		t.Fatal("expected error for sealed vault, got nil")
	}
}
