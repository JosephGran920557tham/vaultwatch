package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/yourusername/vaultwatch/internal/alert"
)

const defaultOpsGenieTimeout = 10 * time.Second
const opsGenieAPIURL = "https://api.opsgenie.com/v2/alerts"

// OpsGenieNotifier sends alerts to OpsGenie.
type OpsGenieNotifier struct {
	apiKey  string
	apiURL  string
	client  *http.Client
}

// NewOpsGenieNotifier creates a new OpsGenieNotifier.
// Returns an error if apiKey is empty.
func NewOpsGenieNotifier(apiKey string, timeout time.Duration) (*OpsGenieNotifier, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("opsgenie: api key must not be empty")
	}
	if timeout <= 0 {
		timeout = defaultOpsGenieTimeout
	}
	return &OpsGenieNotifier{
		apiKey: apiKey,
		apiURL: opsGenieAPIURL,
		client: &http.Client{Timeout: timeout},
	}, nil
}

type opsGeniePayload struct {
	Message     string            `json:"message"`
	Description string            `json:"description"`
	Priority    string            `json:"priority"`
	Details     map[string]string `json:"details"`
}

func opsGeniePriority(level alert.Level) string {
	switch level {
	case alert.Critical:
		return "P1"
	case alert.Warning:
		return "P3"
	default:
		return "P5"
	}
}

// Send delivers an alert to OpsGenie.
func (o *OpsGenieNotifier) Send(a alert.Alert) error {
	payload := opsGeniePayload{
		Message:     fmt.Sprintf("Vault lease expiring: %s", a.LeaseID),
		Description: a.Message,
		Priority:    opsGeniePriority(a.Level),
		Details: map[string]string{
			"lease_id":   a.LeaseID,
			"expires_in": a.ExpiresIn.String(),
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("opsgenie: failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, o.apiURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("opsgenie: failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "GenieKey "+o.apiKey)

	resp, err := o.client.Do(req)
	if err != nil {
		return fmt.Errorf("opsgenie: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("opsgenie: unexpected status code %d", resp.StatusCode)
	}
	return nil
}
