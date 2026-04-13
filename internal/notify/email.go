package notify

import (
	"fmt"
	"net/smtp"
	"strings"
	"time"

	"github.com/vaultwatch/internal/alert"
)

// EmailConfig holds SMTP configuration for sending email alerts.
type EmailConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	To       []string
	Timeout  time.Duration
}

// EmailNotifier sends alert notifications via SMTP email.
type EmailNotifier struct {
	cfg  EmailConfig
	send func(addr, from string, to []string, msg []byte) error
}

// NewEmailNotifier creates a new EmailNotifier with the given config.
// Returns an error if required fields are missing.
func NewEmailNotifier(cfg EmailConfig) (*EmailNotifier, error) {
	if cfg.Host == "" {
		return nil, fmt.Errorf("email notifier: SMTP host is required")
	}
	if len(cfg.To) == 0 {
		return nil, fmt.Errorf("email notifier: at least one recipient is required")
	}
	if cfg.Port == 0 {
		cfg.Port = 587
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 10 * time.Second
	}
	return &EmailNotifier{
		cfg:  cfg,
		send: smtp.SendMail,
	}, nil
}

// Send delivers an alert notification via email.
func (e *EmailNotifier) Send(a alert.Alert) error {
	subject := fmt.Sprintf("[VaultWatch] %s lease expiring: %s", a.Severity, a.LeaseID)
	body := fmt.Sprintf(
		"Lease ID: %s\nSeverity: %s\nExpires: %s\nMessage: %s",
		a.LeaseID, a.Severity, a.ExpiresAt.Format(time.RFC3339), a.Message,
	)
	msg := []byte(fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		e.cfg.From,
		strings.Join(e.cfg.To, ", "),
		subject,
		body,
	))
	addr := fmt.Sprintf("%s:%d", e.cfg.Host, e.cfg.Port)
	var auth smtp.Auth
	if e.cfg.Username != "" {
		auth = smtp.PlainAuth("", e.cfg.Username, e.cfg.Password, e.cfg.Host)
	}
	_ = auth // smtp.SendMail accepts nil auth
	if err := e.send(addr, e.cfg.From, e.cfg.To, msg); err != nil {
		return fmt.Errorf("email notifier: failed to send: %w", err)
	}
	return nil
}
