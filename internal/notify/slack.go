package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/vaultwatch/internal/alert"
)

// SlackNotifier sends alert notifications to a Slack webhook URL.
type SlackNotifier struct {
	webhookURL string
	client     *http.Client
}

type slackPayload struct {
	Text string `json:"text"`
}

// NewSlackNotifier creates a SlackNotifier with the given webhook URL.
// An optional timeout may be provided; defaults to 10 seconds.
func NewSlackNotifier(webhookURL string, timeout time.Duration) (*SlackNotifier, error) {
	if webhookURL == "" {
		return nil, fmt.Errorf("slack webhook URL must not be empty")
	}
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &SlackNotifier{
		webhookURL: webhookURL,
		client:     &http.Client{Timeout: timeout},
	}, nil
}

// Send posts the alert to the configured Slack webhook.
func (s *SlackNotifier) Send(a alert.Alert) error {
	message := fmt.Sprintf("[%s] Lease *%s* expires in %s — path: %s",
		a.Severity, a.LeaseID, a.TTL.Round(time.Second), a.Path)

	payload := slackPayload{Text: message}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("slack: marshal payload: %w", err)
	}

	resp, err := s.client.Post(s.webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("slack: post request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("slack: unexpected status %d", resp.StatusCode)
	}
	return nil
}
