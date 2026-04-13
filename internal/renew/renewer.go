package renew

import (
	"context"
	"fmt"
	"time"
)

// LeaseRenewer defines the interface for renewing a Vault lease.
type LeaseRenewer interface {
	RenewLease(ctx context.Context, leaseID string, increment time.Duration) error
}

// RenewRequest holds the parameters for a single lease renewal.
type RenewRequest struct {
	LeaseID   string
	Increment time.Duration
}

// Result captures the outcome of a renewal attempt.
type Result struct {
	LeaseID   string
	RenewedAt time.Time
	Err       error
}

// Manager orchestrates lease renewals using a LeaseRenewer.
type Manager struct {
	renewer LeaseRenewer
}

// NewManager creates a new renewal Manager.
func NewManager(r LeaseRenewer) (*Manager, error) {
	if r == nil {
		return nil, fmt.Errorf("renewer must not be nil")
	}
	return &Manager{renewer: r}, nil
}

// Renew attempts to renew a single lease and returns a Result.
func (m *Manager) Renew(ctx context.Context, req RenewRequest) Result {
	err := m.renewer.RenewLease(ctx, req.LeaseID, req.Increment)
	return Result{
		LeaseID:   req.LeaseID,
		RenewedAt: time.Now().UTC(),
		Err:       err,
	}
}

// RenewAll attempts to renew all provided leases and returns results for each.
func (m *Manager) RenewAll(ctx context.Context, reqs []RenewRequest) []Result {
	results := make([]Result, 0, len(reqs))
	for _, req := range reqs {
		results = append(results, m.Renew(ctx, req))
	}
	return results
}
