package tokenwatch

import (
	"sync"
	"time"

	"github.com/vaultwatch/internal/alert"
)

// DefaultTenureConfig returns a TenureConfig with sensible defaults.
func DefaultTenureConfig() TenureConfig {
	return TenureConfig{
		WarningAge:  7 * 24 * time.Hour,
		CriticalAge: 30 * 24 * time.Hour,
	}
}

// TenureConfig controls thresholds for token age alerts.
type TenureConfig struct {
	WarningAge  time.Duration
	CriticalAge time.Duration
}

// TenureDetector alerts when a token has been alive longer than expected,
// which may indicate a forgotten long-lived credential.
type TenureDetector struct {
	mu      sync.Mutex
	cfg     TenureConfig
	issued  map[string]time.Time
}

// NewTenureDetector creates a TenureDetector with the given config.
// Zero values fall back to defaults.
func NewTenureDetector(cfg TenureConfig) *TenureDetector {
	def := DefaultTenureConfig()
	if cfg.WarningAge <= 0 {
		cfg.WarningAge = def.WarningAge
	}
	if cfg.CriticalAge <= 0 || cfg.CriticalAge <= cfg.WarningAge {
		cfg.CriticalAge = def.CriticalAge
	}
	return &TenureDetector{
		cfg:    cfg,
		issued: make(map[string]time.Time),
	}
}

// Track records the issue time for a token. If the token is already tracked
// the existing issue time is preserved.
func (d *TenureDetector) Track(tokenID string, issuedAt time.Time) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if _, ok := d.issued[tokenID]; !ok {
		d.issued[tokenID] = issuedAt
	}
}

// Check returns an alert if the token has exceeded a tenure threshold.
// Returns nil when the token is unknown or within acceptable age.
func (d *TenureDetector) Check(tokenID string, now time.Time) *alert.Alert {
	d.mu.Lock()
	issuedAt, ok := d.issued[tokenID]
	d.mu.Unlock()
	if !ok {
		return nil
	}
	age := now.Sub(issuedAt)
	switch {
	case age >= d.cfg.CriticalAge:
		return &alert.Alert{
			LeaseID: tokenID,
			Level:   alert.Critical,
			Message: "token tenure exceeds critical age threshold",
			Labels:  map[string]string{"age": age.String()},
		}
	case age >= d.cfg.WarningAge:
		return &alert.Alert{
			LeaseID: tokenID,
			Level:   alert.Warning,
			Message: "token tenure exceeds warning age threshold",
			Labels:  map[string]string{"age": age.String()},
		}
	default:
		return nil
	}
}
