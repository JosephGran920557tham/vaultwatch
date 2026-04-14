package tokenwatch

import (
	"fmt"
	"time"

	"github.com/your-org/vaultwatch/internal/alert"
)

// ExpiryClassifier determines the alert level for a token based on its TTL
// and the configured warning/critical thresholds.
type ExpiryClassifier struct {
	WarnThreshold     time.Duration
	CriticalThreshold time.Duration
}

// DefaultExpiryClassifier returns an ExpiryClassifier with sensible defaults.
func DefaultExpiryClassifier() *ExpiryClassifier {
	return &ExpiryClassifier{
		WarnThreshold:     24 * time.Hour,
		CriticalThreshold: 4 * time.Hour,
	}
}

// Classify returns the alert level appropriate for the given remaining TTL.
// Returns alert.LevelInfo if the TTL is above the warning threshold.
func (e *ExpiryClassifier) Classify(ttl time.Duration) alert.Level {
	switch {
	case ttl <= e.CriticalThreshold:
		return alert.LevelCritical
	case ttl <= e.WarnThreshold:
		return alert.LevelWarning
	default:
		return alert.LevelInfo
	}
}

// Summary returns a human-readable description of the token's expiry state.
func (e *ExpiryClassifier) Summary(tokenID string, ttl time.Duration) string {
	level := e.Classify(ttl)
	expiry := time.Now().Add(ttl).UTC().Format(time.RFC3339)
	return fmt.Sprintf("[%s] token %s expires at %s (TTL: %s)",
		level, tokenID, expiry, ttl.Round(time.Second))
}

// Validate ensures thresholds are logically consistent and non-negative.
func (e *ExpiryClassifier) Validate() error {
	if e.CriticalThreshold < 0 {
		return fmt.Errorf("critical threshold must be non-negative")
	}
	if e.WarnThreshold < 0 {
		return fmt.Errorf("warn threshold must be non-negative")
	}
	if e.CriticalThreshold >= e.WarnThreshold {
		return fmt.Errorf("critical threshold (%s) must be less than warn threshold (%s)",
			e.CriticalThreshold, e.WarnThreshold)
	}
	return nil
}
