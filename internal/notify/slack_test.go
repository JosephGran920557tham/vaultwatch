package notify

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/vaultwatch/internal/alert"
)

func makeSlackAlert() alert.Alert {
	return alert.Alert{
		LeaseID:  "lease/test/001",
		Path:     "secret/data/myapp",
		TTL:      2 * time.Hour,
		Severity: alert.SeverityWarning,
	}
}

func TestNewSlackNotifier_EmptyURL(t *testing.T) {
	_, err := NewSlackNotifier("", 0)
	if err == nil {
		t.Fatal("expected error for empty webhook URL")
	}
}

func TestNewSlackNotifier_DefaultTimeout(t *testing.T) {
	n, err := NewSlackNotifier("https://hooks.slack.com/test", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.client.Timeout != 10*time.Second {
		t.Errorf("expected default timeout 10s, got %v", n.client.Timeout)
	}
}

func TestSlackNotifier_Send_Success(t *testing.T) {
	var received slackPayload
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Errorf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	n, _ := NewSlackNotifier(server.URL, 5*time.Second)
	if err := n.Send(makeSlackAlert()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(received.Text, "lease/test/001") {
		t.Errorf("expected lease ID in message, got: %s", received.Text)
	}
	if !strings.Contains(received.Text, "WARNING") {
		t.Errorf("expected severity in message, got: %s", received.Text)
	}
}

func TestSlackNotifier_Send_NonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	n, _ := NewSlackNotifier(server.URL, 5*time.Second)
	if err := n.Send(makeSlackAlert()); err == nil {
		t.Fatal("expected error for non-OK status")
	}
}

func TestSlackNotifier_Send_InvalidURL(t *testing.T) {
	n, _ := NewSlackNotifier("http://127.0.0.1:0/no-server", time.Second)
	if err := n.Send(makeSlackAlert()); err == nil {
		t.Fatal("expected error for unreachable URL")
	}
}
