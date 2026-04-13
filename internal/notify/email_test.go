package notify

import (
	"errors"
	"testing"
	"time"

	"github.com/vaultwatch/internal/alert"
)

func makeEmailAlert() alert.Alert {
	return alert.Alert{
		LeaseID:   "lease/email/001",
		Severity:  alert.SeverityWarning,
		Message:   "Lease expires soon",
		ExpiresAt: time.Now().Add(48 * time.Hour),
	}
}

func TestNewEmailNotifier_MissingHost(t *testing.T) {
	_, err := NewEmailNotifier(EmailConfig{To: []string{"ops@example.com"}})
	if err == nil {
		t.Fatal("expected error for missing host")
	}
}

func TestNewEmailNotifier_MissingRecipients(t *testing.T) {
	_, err := NewEmailNotifier(EmailConfig{Host: "smtp.example.com"})
	if err == nil {
		t.Fatal("expected error for missing recipients")
	}
}

func TestNewEmailNotifier_DefaultPort(t *testing.T) {
	n, err := NewEmailNotifier(EmailConfig{
		Host: "smtp.example.com",
		To:   []string{"ops@example.com"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.cfg.Port != 587 {
		t.Errorf("expected default port 587, got %d", n.cfg.Port)
	}
}

func TestEmailNotifier_Send_Success(t *testing.T) {
	n, _ := NewEmailNotifier(EmailConfig{
		Host: "smtp.example.com",
		From: "vault@example.com",
		To:   []string{"ops@example.com"},
	})
	n.send = func(addr, from string, to []string, msg []byte) error {
		if from != "vault@example.com" {
			t.Errorf("unexpected from: %s", from)
		}
		if len(to) != 1 || to[0] != "ops@example.com" {
			t.Errorf("unexpected to: %v", to)
		}
		return nil
	}
	if err := n.Send(makeEmailAlert()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEmailNotifier_Send_Error(t *testing.T) {
	n, _ := NewEmailNotifier(EmailConfig{
		Host: "smtp.example.com",
		To:   []string{"ops@example.com"},
	})
	n.send = func(addr, from string, to []string, msg []byte) error {
		return errors.New("connection refused")
	}
	err := n.Send(makeEmailAlert())
	if err == nil {
		t.Fatal("expected error from failed send")
	}
}
