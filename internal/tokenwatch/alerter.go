package tokenwatch

import (
	"context"
	"fmt"
	"time"

	"github.com/vaultwatch/internal/alert"
)

// Alerter combines a Watcher, a Registry, and a Throttle to produce
// deduplicated alerts for all tracked tokens.
type Alerter struct {
	registry *Registry
	watcher  *Watcher
	throttle *Throttle
}

// NewAlerter constructs an Alerter from the given components.
func NewAlerter(registry *Registry, watcher *Watcher, throttle *Throttle) (*Alerter, error) {
	if registry == nil {
		return nil, fmt.Errorf("tokenwatch: registry must not be nil")
	}
	if watcher == nil {
		return nil, fmt.Errorf("tokenwatch: watcher must not be nil")
	}
	if throttle == nil {
		return nil, fmt.Errorf("tokenwatch: throttle must not be nil")
	}
	return &Alerter{
		registry: registry,
		watcher:  watcher,
		throttle: throttle,
	}, nil
}

// CheckAll inspects every registered token and returns throttled alerts.
func (a *Alerter) CheckAll(ctx context.Context) ([]alert.Alert, error) {
	tokens := a.registry.List()
	var results []alert.Alert

	for _, id := range tokens {
		select {
		case <-ctx.Done():
			return results, ctx.Err()
		default:
		}

		alerts, err := a.watcher.Check(ctx, id)
		if err != nil {
			continue
		}
		for _, al := range alerts {
			if a.throttle.Allow(id) {
				results = append(results, al)
			}
		}
	}
	return results, nil
}

// buildTokenAlert is a helper used by Watcher.Check to construct an alert.
func buildTokenAlert(id string, ttl time.Duration, level alert.Level) alert.Alert {
	return alert.Alert{
		LeaseID:   id,
		Level:     level,
		Message:   fmt.Sprintf("token %s expires in %s", id, ttl.Round(time.Second)),
		ExpiresAt: time.Now().Add(ttl),
	}
}
