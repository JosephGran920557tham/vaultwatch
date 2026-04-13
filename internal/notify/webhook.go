package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/your-org/vaultwatch/internal/alert"
)

// WebhookNotifier sends alert notifications to an HTTP endpoint.
type WebhookNotifier struct {
	URL    string
	client *http.Client
}

// webhookPayload is the JSON body sent to the webhook endpoint.
type webhookPayload struct {
	LeaseID   string    `json:"lease_id"`
	Severity  string    `json:"severity"`
	Message   string    `json:"message"`
	ExpiresAt time.Time `json:"expires_at"`
	SentAt    time.Time `json:"sent_at"`
}

// NewWebhookNotifier creates a WebhookNotifier that posts to the given URL.
func NewWebhookNotifier(url string, timeout time.Duration) *WebhookNotifier {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &WebhookNotifier{
		URL: url,
		client: &http.Client{Timeout: timeout},
	}
}

// Send marshals the alert and POSTs it to the configured webhook URL.
func (w *WebhookNotifier) Send(a alert.Alert) error {
	payload := webhookPayload{
		LeaseID:   a.LeaseID,
		Severity:  string(a.Severity),
		Message:   a.Message,
		ExpiresAt: a.ExpiresAt,
		SentAt:    time.Now().UTC(),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("webhook: marshal payload: %w", err)
	}

	resp, err := w.client.Post(w.URL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("webhook: post to %s: %w", w.URL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook: unexpected status %d from %s", resp.StatusCode, w.URL)
	}
	return nil
}
