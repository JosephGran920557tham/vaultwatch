package renew

import (
	"fmt"
	"time"
)

// Policy defines when and how a lease should be renewed.
type Policy struct {
	// RenewBefore is how far before expiry to trigger renewal.
	RenewBefore time.Duration
	// Increment is the requested TTL extension.
	Increment time.Duration
	// MaxRetries is the number of times to retry on failure.
	MaxRetries int
}

// DefaultPolicy returns a sensible default renewal policy.
func DefaultPolicy() Policy {
	return Policy{
		RenewBefore: 10 * time.Minute,
		Increment:   1 * time.Hour,
		MaxRetries:  3,
	}
}

// Validate checks that policy fields are logically consistent.
func (p Policy) Validate() error {
	if p.RenewBefore <= 0 {
		return fmt.Errorf("RenewBefore must be positive, got %v", p.RenewBefore)
	}
	if p.Increment <= 0 {
		return fmt.Errorf("Increment must be positive, got %v", p.Increment)
	}
	if p.MaxRetries < 0 {
		return fmt.Errorf("MaxRetries must be non-negative, got %d", p.MaxRetries)
	}
	return nil
}

// ShouldRenew returns true if the lease expiry is within the RenewBefore window.
func (p Policy) ShouldRenew(expiry time.Time) bool {
	return time.Until(expiry) <= p.RenewBefore
}
