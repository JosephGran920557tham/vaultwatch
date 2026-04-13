package notify_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/your-org/vaultwatch/internal/alert"
	"github.com/your-org/vaultwatch/internal/notify"
)

func makeAlert(severity alert.Severity) alert.Alert {
	return alert.Alert{
		LeaseID:   "secret/data/myapp#abc123",
		Severity:  severity,
		Message:   "lease expires soon",
		ExpiresAt: time.Now().Add(2 * time.Hour),
	}
}

func TestWebhookNotifier_Send_Success(t *testing.T) {
	var received map[string]interface{}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &received)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	n := notify.NewWebhookNotifier(ts.URL, 5*time.Second)
	err := n.Send(makeAlert(alert.SeverityWarning))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if received["lease_id"] != "secret/data/myapp#abc123" {
		t.Errorf("unexpected lease_id: %v", received["lease_id"])
	}
	if received["severity"] != "warning" {
		t.Errorf("unexpected severity: %v", received["severity"])
	}
}

func TestWebhookNotifier_Send_NonOKStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	n := notify.NewWebhookNotifier(ts.URL, 5*time.Second)
	err := n.Send(makeAlert(alert.SeverityCritical))
	if err == nil {
		t.Fatal("expected error for non-2xx status, got nil")
	}
}

func TestWebhookNotifier_Send_InvalidURL(t *testing.T) {
	n := notify.NewWebhookNotifier("http://127.0.0.1:0/no-server", 1*time.Second)
	err := n.Send(makeAlert(alert.SeverityInfo))
	if err == nil {
		t.Fatal("expected connection error, got nil")
	}
}

func TestNewWebhookNotifier_DefaultTimeout(t *testing.T) {
	n := notify.NewWebhookNotifier("http://example.com", 0)
	if n == nil {
		t.Fatal("expected non-nil notifier")
	}
}
