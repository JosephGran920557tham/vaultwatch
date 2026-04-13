package notify

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/vaultwatch/internal/alert"
)

func makePDAlert(sev alert.SeverityLevel) alert.Alert {
	return alert.Alert{
		LeaseID:   "secret/data/db/prod",
		Severity:  sev,
		Message:   "Lease expiring in 2h",
		ExpiresAt: time.Now().Add(2 * time.Hour),
	}
}

func TestNewPagerDutyNotifier_EmptyKey(t *testing.T) {
	_, err := NewPagerDutyNotifier("", 0)
	if err == nil {
		t.Fatal("expected error for empty integration key")
	}
}

func TestNewPagerDutyNotifier_DefaultTimeout(t *testing.T) {
	n, err := NewPagerDutyNotifier("test-key", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.client.Timeout != 10*time.Second {
		t.Errorf("expected 10s timeout, got %v", n.client.Timeout)
	}
}

func TestPagerDutyNotifier_Send_Success(t *testing.T) {
	var received pdPayload
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Errorf("failed to decode body: %v", err)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer ts.Close()

	n, _ := NewPagerDutyNotifier("my-key", 5*time.Second)
	n.eventsURL = ts.URL

	if err := n.Send(makePDAlert(alert.SeverityCritical)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received.RoutingKey != "my-key" {
		t.Errorf("expected routing key 'my-key', got %s", received.RoutingKey)
	}
	if received.Payload.Severity != "critical" {
		t.Errorf("expected severity 'critical', got %s", received.Payload.Severity)
	}
}

func TestPagerDutyNotifier_Send_NonOKStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	n, _ := NewPagerDutyNotifier("my-key", 5*time.Second)
	n.eventsURL = ts.URL

	if err := n.Send(makePDAlert(alert.SeverityWarning)); err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}
