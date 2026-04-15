package tokenwatch

import (
	"math"
	"time"
)

// BackoffConfig holds parameters for exponential backoff.
type BackoffConfig struct {
	// InitialInterval is the starting delay.
	InitialInterval time.Duration
	// MaxInterval caps the computed delay.
	MaxInterval time.Duration
	// Multiplier is applied on each failure.
	Multiplier float64
	// MaxRetries is the maximum number of attempts before giving up (0 = unlimited).
	MaxRetries int
}

// DefaultBackoffConfig returns sensible defaults.
func DefaultBackoffConfig() BackoffConfig {
	return BackoffConfig{
		InitialInterval: 500 * time.Millisecond,
		MaxInterval:     30 * time.Second,
		Multiplier:      2.0,
		MaxRetries:      5,
	}
}

// Backoff tracks retry state for a single operation.
type Backoff struct {
	cfg     BackoffConfig
	attempt int
}

// NewBackoff creates a Backoff using the provided config.
// It returns an error if the config is invalid.
func NewBackoff(cfg BackoffConfig) (*Backoff, error) {
	if cfg.InitialInterval <= 0 {
		return nil, errorf("backoff: InitialInterval must be positive")
	}
	if cfg.MaxInterval < cfg.InitialInterval {
		return nil, errorf("backoff: MaxInterval must be >= InitialInterval")
	}
	if cfg.Multiplier <= 1.0 {
		return nil, errorf("backoff: Multiplier must be > 1.0")
	}
	return &Backoff{cfg: cfg}, nil
}

// Next returns the delay for the current attempt and advances the counter.
// The second return value is false when MaxRetries has been exceeded.
func (b *Backoff) Next() (time.Duration, bool) {
	if b.cfg.MaxRetries > 0 && b.attempt >= b.cfg.MaxRetries {
		return 0, false
	}
	delay := float64(b.cfg.InitialInterval) * math.Pow(b.cfg.Multiplier, float64(b.attempt))
	if delay > float64(b.cfg.MaxInterval) {
		delay = float64(b.cfg.MaxInterval)
	}
	b.attempt++
	return time.Duration(delay), true
}

// Reset sets the attempt counter back to zero.
func (b *Backoff) Reset() {
	b.attempt = 0
}

// Attempt returns the current attempt index (zero-based).
func (b *Backoff) Attempt() int {
	return b.attempt
}

// errorf is a thin wrapper so the package avoids importing fmt at the top level.
func errorf(msg string) error {
	return &backoffError{msg: msg}
}

type backoffError struct{ msg string }

func (e *backoffError) Error() string { return e.msg }
