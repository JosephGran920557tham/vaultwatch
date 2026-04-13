package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/vaultwatch/internal/alert"
)

const pagerDutyEventsURL = "https://events.pagerduty.com/v2/enqueue"

// PagerDutyNotifier sends alerts to PagerDuty Events API v2.
type PagerDutyNotifier struct {
	integrationKey string
	client         *http.Client
	eventsURL      string
}

type pdPayload struct {
	RoutingKey  string    `json:"routing_key"`
	EventAction string    `json:"event_action"`
	Payload     pdDetails `json:"payload"`
}

type pdDetails struct {
	Summary  string `json:"summary"`
	Source   string `json:"source"`
	Severity string `json:"severity"`
}

// NewPagerDutyNotifier creates a notifier that sends to PagerDuty.
func NewPagerDutyNotifier(integrationKey string, timeout time.Duration) (*PagerDutyNotifier, error) {
	if integrationKey == "" {
		return nil, fmt.Errorf("pagerduty notifier: integration key is required")
	}
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	return &PagerDutyNotifier{
		integrationKey: integrationKey,
		client:         &http.Client{Timeout: timeout},
		eventsURL:      pagerDutyEventsURL,
	}, nil
}

// Send triggers a PagerDuty alert for the given lease alert.
func (p *PagerDutyNotifier) Send(a alert.Alert) error {
	severity := "warning"
	if a.Severity == alert.SeverityCritical {
		severity = "critical"
	}
	body := pdPayload{
		RoutingKey:  p.integrationKey,
		EventAction: "trigger",
		Payload: pdDetails{
			Summary:  fmt.Sprintf("VaultWatch: %s", a.Message),
			Source:   a.LeaseID,
			Severity: severity,
		},
	}
	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("pagerduty notifier: marshal error: %w", err)
	}
	resp, err := p.client.Post(p.eventsURL, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("pagerduty notifier: request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("pagerduty notifier: unexpected status %d", resp.StatusCode)
	}
	return nil
}
