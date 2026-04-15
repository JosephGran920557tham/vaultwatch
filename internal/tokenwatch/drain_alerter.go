package tokenwatch

import (
	"context"
	"fmt"
	"time"

	"github.com/your-org/vaultwatch/internal/alert"
)

// DrainAlerter wraps a DrainDetector and emits alerts when drain is detected.
type DrainAlerter struct {
	detector *DrainDetector
	cooldown *Cooldown
}

// NewDrainAlerter constructs a DrainAlerter using the provided detector.
// A default cooldown is applied so repeated alerts are suppressed.
func NewDrainAlerter(detector *DrainDetector) (*DrainAlerter, error) {
	if detector == nil {
		return nil, fmt.Errorf("drain alerter: detector must not be nil")
	}
	cd, err := NewCooldown(DefaultCooldownConfig())
	if err != nil {
		return nil, fmt.Errorf("drain alerter: %w", err)
	}
	return &DrainAlerter{detector: detector, cooldown: cd}, nil
}

// Check evaluates the drain state for tokenID and returns an alert if the
// token is draining faster than the configured threshold and the cooldown
// window has elapsed. Returns nil when no alert should be raised.
func (a *DrainAlerter) Check(ctx context.Context, tokenID string) (*alert.Alert, error) {
	draining, rate, err := a.detector.IsDraining(ctx, tokenID)
	if err != nil {
		return nil, fmt.Errorf("drain check %s: %w", tokenID, err)
	}
	if !draining {
		return nil, nil
	}
	if !a.cooldown.Allow(tokenID) {
		return nil, nil
	}
	a := &alert.Alert{
		LeaseID:   tokenID,
		Level:     alert.Warning,
		Message:   fmt.Sprintf("token %s TTL draining at %.2f s/s (threshold %.2f)", tokenID, rate, a.detector.cfg.DrainThreshold),
		ExpiresAt: time.Now().Add(time.Duration(rate) * time.Second),
	}
	return a, nil
}
