package tokenwatch

import (
	"errors"
	"time"
)

// RetryConfig holds configuration for the token check retry mechanism.
type RetryConfig struct {
	// MaxAttempts is the maximum number of attempts before giving up.
	MaxAttempts int
	// InitialDelay is the wait time before the first retry.
	InitialDelay time.Duration
	// MaxDelay caps the exponential backoff delay.
	MaxDelay time.Duration
	// Multiplier scales the delay on each attempt.
	Multiplier float64
}

// DefaultRetryConfig returns sensible defaults for token check retries.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 200 * time.Millisecond,
		MaxDelay:     5 * time.Second,
		Multiplier:   2.0,
	}
}

// Retry holds state for retrying a token check operation.
type Retry struct {
	cfg RetryConfig
}

// NewRetry creates a Retry with the given config. If MaxAttempts < 1 or
// Multiplier < 1, defaults are applied.
func NewRetry(cfg RetryConfig) (*Retry, error) {
	if cfg.MaxAttempts < 1 {
		return nil, errors.New("retry: MaxAttempts must be at least 1")
	}
	if cfg.Multiplier < 1.0 {
		return nil, errors.New("retry: Multiplier must be >= 1.0")
	}
	if cfg.InitialDelay <= 0 {
		cfg.InitialDelay = DefaultRetryConfig().InitialDelay
	}
	if cfg.MaxDelay <= 0 {
		cfg.MaxDelay = DefaultRetryConfig().MaxDelay
	}
	return &Retry{cfg: cfg}, nil
}

// Do executes fn up to MaxAttempts times, backing off between attempts.
// It returns the last error if all attempts fail, or nil on success.
func (r *Retry) Do(fn func() error) error {
	delay := r.cfg.InitialDelay
	var lastErr error
	for attempt := 0; attempt < r.cfg.MaxAttempts; attempt++ {
		if err := fn(); err == nil {
			return nil
		} else {
			lastErr = err
		}
		if attempt < r.cfg.MaxAttempts-1 {
			time.Sleep(delay)
			delay = time.Duration(float64(delay) * r.cfg.Multiplier)
			if delay > r.cfg.MaxDelay {
				delay = r.cfg.MaxDelay
			}
		}
	}
	return lastErr
}

// Attempts returns the configured maximum number of attempts.
func (r *Retry) Attempts() int {
	return r.cfg.MaxAttempts
}
