package tokenwatch

import (
	"time"

	"github.com/your-org/vaultwatch/internal/alert"
)

// DefaultGraceConfig returns sensible defaults for grace period detection.
func DefaultGraceConfig() GraceConfig {
	return GraceConfig{
		WarningBefore: 10 * time.Minute,
		CriticalBefore: 2 * time.Minute,
	}
}

// GraceConfig controls grace period alert thresholds.
type GraceConfig struct {
	WarningBefore  time.Duration
	CriticalBefore time.Duration
}

// GraceDetector alerts when a token is within its grace period before expiry.
type GraceDetector struct {
	cfg GraceConfig
}

// NewGraceDetector constructs a GraceDetector, applying defaults for zero values.
func NewGraceDetector(cfg GraceConfig) *GraceDetector {
	def := DefaultGraceConfig()
	if cfg.WarningBefore <= 0 {
		cfg.WarningBefore = def.WarningBefore
	}
	if cfg.CriticalBefore <= 0 {
		cfg.CriticalBefore = def.CriticalBefore
	}
	return &GraceDetector{cfg: cfg}
}

// Check returns an alert if the token TTL falls within the grace window, nil otherwise.
func (g *GraceDetector) Check(tokenID string, ttl time.Duration) *alert.Alert {
	switch {
	case ttl <= g.cfg.CriticalBefore:
		return &alert.Alert{
			LeaseID: tokenID,
			Level:   alert.Critical,
			Message: "token within critical grace period",
		}
	case ttl <= g.cfg.WarningBefore:
		return &alert.Alert{
			LeaseID: tokenID,
			Level:   alert.Warning,
			Message: "token within warning grace period",
		}
	default:
		return nil
	}
}
