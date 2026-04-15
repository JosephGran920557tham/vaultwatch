package tokenwatch

import (
	"fmt"
	"time"

	"github.com/vaultwatch/internal/alert"
)

// AnomalyConfig holds tuning parameters for TTL anomaly detection.
type AnomalyConfig struct {
	// MinTTL is the floor below which a token TTL is considered anomalously low.
	MinTTL time.Duration
	// MaxTTL is the ceiling above which a token TTL is considered anomalously high.
	MaxTTL time.Duration
}

// DefaultAnomalyConfig returns sensible defaults for anomaly detection.
func DefaultAnomalyConfig() AnomalyConfig {
	return AnomalyConfig{
		MinTTL: 30 * time.Second,
		MaxTTL: 30 * 24 * time.Hour,
	}
}

// AnomalyDetector checks whether a token's TTL falls outside expected bounds.
type AnomalyDetector struct {
	cfg AnomalyConfig
}

// NewAnomalyDetector creates a new AnomalyDetector. Zero-value fields in cfg
// are replaced with defaults.
func NewAnomalyDetector(cfg AnomalyConfig) *AnomalyDetector {
	def := DefaultAnomalyConfig()
	if cfg.MinTTL <= 0 {
		cfg.MinTTL = def.MinTTL
	}
	if cfg.MaxTTL <= 0 {
		cfg.MaxTTL = def.MaxTTL
	}
	return &AnomalyDetector{cfg: cfg}
}

// Check returns an alert if the given TTL is outside the configured bounds,
// or nil when the TTL is within the expected range.
func (d *AnomalyDetector) Check(tokenID string, ttl time.Duration) *alert.Alert {
	switch {
	case ttl < d.cfg.MinTTL:
		return &alert.Alert{
			LeaseID:   tokenID,
			Level:     alert.LevelCritical,
			Message:   fmt.Sprintf("token TTL %s is below minimum threshold %s", ttl, d.cfg.MinTTL),
			ExpiresAt: time.Now().Add(ttl),
		}
	case ttl > d.cfg.MaxTTL:
		return &alert.Alert{
			LeaseID:   tokenID,
			Level:     alert.LevelWarning,
			Message:   fmt.Sprintf("token TTL %s exceeds maximum threshold %s", ttl, d.cfg.MaxTTL),
			ExpiresAt: time.Now().Add(ttl),
		}
	}
	return nil
}
