package tokenwatch

import (
	"fmt"
	"time"

	"github.com/vaultwatch/internal/alert"
)

// StalenessConfig holds thresholds for detecting stale token checks.
type StalenessConfig struct {
	// WarnAfter is the duration after which a token not checked is considered stale.
	WarnAfter time.Duration
	// CriticalAfter is the duration after which a stale token is considered critical.
	CriticalAfter time.Duration
}

// DefaultStalenessConfig returns sensible defaults.
func DefaultStalenessConfig() StalenessConfig {
	return StalenessConfig{
		WarnAfter:     10 * time.Minute,
		CriticalAfter: 30 * time.Minute,
	}
}

// StalenessDetector checks whether a token's last-seen time indicates staleness.
type StalenessDetector struct {
	cfg StalenessConfig
	now func() time.Time
}

// NewStalenessDetector creates a StalenessDetector with the provided config.
// Zero-value durations in cfg fall back to defaults.
func NewStalenessDetector(cfg StalenessConfig) *StalenessDetector {
	def := DefaultStalenessConfig()
	if cfg.WarnAfter <= 0 {
		cfg.WarnAfter = def.WarnAfter
	}
	if cfg.CriticalAfter <= 0 {
		cfg.CriticalAfter = def.CriticalAfter
	}
	return &StalenessDetector{cfg: cfg, now: time.Now}
}

// Check returns an alert if the token identified by tokenID has not been seen
// since lastSeen and the elapsed time exceeds a configured threshold.
// Returns nil if the token is fresh.
func (d *StalenessDetector) Check(tokenID string, lastSeen time.Time) *alert.Alert {
	elapsed := d.now().Sub(lastSeen)
	var level alert.Level
	switch {
	case elapsed >= d.cfg.CriticalAfter:
		level = alert.LevelCritical
	case elapsed >= d.cfg.WarnAfter:
		level = alert.LevelWarning
	default:
		return nil
	}
	return &alert.Alert{
		LeaseID:  tokenID,
		Level:    level,
		Message:  fmt.Sprintf("token %s has not been checked for %s", tokenID, elapsed.Round(time.Second)),
		Metadata: map[string]string{"last_seen": lastSeen.UTC().Format(time.RFC3339)},
	}
}
