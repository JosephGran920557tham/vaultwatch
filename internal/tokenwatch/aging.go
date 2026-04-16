package tokenwatch

import (
	"fmt"
	"time"

	"github.com/vaultwatch/internal/alert"
)

// DefaultAgingConfig returns sensible defaults for aging detection.
func DefaultAgingConfig() AgingConfig {
	return AgingConfig{
		WarnAfter:     72 * time.Hour,
		CriticalAfter: 168 * time.Hour, // 7 days
	}
}

// AgingConfig controls thresholds for token age alerts.
type AgingConfig struct {
	WarnAfter     time.Duration
	CriticalAfter time.Duration
}

// AgingDetector alerts when a token has been alive beyond acceptable thresholds.
type AgingDetector struct {
	cfg AgingConfig
	now func() time.Time
}

// NewAgingDetector creates an AgingDetector with the given config.
// Zero values are replaced with defaults.
func NewAgingDetector(cfg AgingConfig) *AgingDetector {
	def := DefaultAgingConfig()
	if cfg.WarnAfter <= 0 {
		cfg.WarnAfter = def.WarnAfter
	}
	if cfg.CriticalAfter <= 0 {
		cfg.CriticalAfter = def.CriticalAfter
	}
	return &AgingDetector{cfg: cfg, now: time.Now}
}

// Check returns an alert if the token was issued before the warn or critical threshold.
// Returns nil if the token age is within acceptable bounds.
func (d *AgingDetector) Check(tokenID string, issuedAt time.Time) *alert.Alert {
	age := d.now().Sub(issuedAt)
	var level alert.Level
	switch {
	case age >= d.cfg.CriticalAfter:
		level = alert.LevelCritical
	case age >= d.cfg.WarnAfter:
		level = alert.LevelWarning
	default:
		return nil
	}
	return &alert.Alert{
		LeaseID:  tokenID,
		Level:    level,
		Message:  fmt.Sprintf("token %s is %.1f hours old", tokenID, age.Hours()),
		Labels:   map[string]string{"detector": "aging"},
	}
}
