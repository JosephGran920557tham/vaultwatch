package tokenwatch

import (
	"time"

	"github.com/vaultwatch/internal/alert"
)

// Envelope wraps an alert with delivery metadata for pipeline tracking.
type Envelope struct {
	// Alert is the underlying alert being delivered.
	Alert alert.Alert
	// Token is the Vault token ID that triggered the alert.
	Token string
	// CreatedAt is when the envelope was created.
	CreatedAt time.Time
	// Attempt is the current delivery attempt number (1-based).
	Attempt int
	// MaxAttempts is the maximum number of delivery attempts allowed.
	MaxAttempts int
	// LastError holds the most recent delivery error, if any.
	LastError error
}

// NewEnvelope wraps an alert in an envelope with default delivery settings.
func NewEnvelope(token string, a alert.Alert, maxAttempts int) Envelope {
	if maxAttempts <= 0 {
		maxAttempts = 3
	}
	return Envelope{
		Alert:       a,
		Token:       token,
		CreatedAt:   time.Now(),
		Attempt:     0,
		MaxAttempts: maxAttempts,
	}
}

// Exhausted reports whether all delivery attempts have been consumed.
func (e *Envelope) Exhausted() bool {
	return e.Attempt >= e.MaxAttempts
}

// Increment advances the attempt counter and records the last error.
func (e *Envelope) Increment(err error) {
	e.Attempt++
	e.LastError = err
}

// Age returns how long ago the envelope was created.
func (e *Envelope) Age() time.Duration {
	return time.Since(e.CreatedAt)
}
