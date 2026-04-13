package notify

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/alert"
)

func makeOGAlert(level alert.Level) alert.Alert {
	return alert.Alert{
		LeaseID:   "secret/data/db#abc123",
		Message:   "Lease expiring soon",
		Level:     level,
		ExpiresIn: 30 * time.Minute,
	}
}

func TestNewOpsGenieNotifier_EmptyKey(t *testing.T) {
	_, err := NewOpsGenieNotifier("", 0)
	if err == nil {
		t.Fatal("expected error for empty api key")
	}
}

func TestNewOpsGenieNotifier_DefaultTimeout(t *testing.T) {
	n, err := NewOpsGenieNotifier("test-key", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.client.Timeout != defaultOpsGenieTimeout {
		t.Errorf("expected default timeout %v, got %v", defaultOpsGenieTimeout, n.client.Timeout)
	}
}

func TestOpsGenieNotifier_Send_Success(t *testing.T) {
	var received opsGeniePayload
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header")
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Errorf("failed to decode body: %v", err)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer ts.Close()

	n, _ := NewOpsGenieNotifier("test-key", 5*time.Second)
	n.apiURL = ts.URL

	a := makeOGAlert(alert.Critical)
	if err := n.Send(a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received.Priority != "P1" {
		t.Errorf("expected priority P1 for critical, got %s", received.Priority)
	}
	if received.Details["lease_id"] != a.LeaseID {
		t.Errorf("expected lease_id %s, got %s", a.LeaseID, received.Details["lease_id"])
	}
}

func TestOpsGenieNotifier_Send_NonOKStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer ts.Close()

	n, _ := NewOpsGenieNotifier("bad-key", 5*time.Second)
	n.apiURL = ts.URL

	if err := n.Send(makeOGAlert(alert.Warning)); err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}

func TestOpsGenieNotifier_Priority_Mapping(t *testing.T) {
	cases := []struct {
		level    alert.Level
		wantPrio string
	}{
		{alert.Critical, "P1"},
		{alert.Warning, "P3"},
		{alert.Info, "P5"},
	}
	for _, tc := range cases {
		got := opsGeniePriority(tc.level)
		if got != tc.wantPrio {
			t.Errorf("level %v: expected %s, got %s", tc.level, tc.wantPrio, got)
		}
	}
}
