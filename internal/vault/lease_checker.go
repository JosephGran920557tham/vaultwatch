package vault

import (
	"fmt"
	"time"
)

// ExpiryAlert represents a lease that is approaching or past expiration.
type ExpiryAlert struct {
	LeaseID   string
	TTL       time.Duration
	ExpiresAt time.Time
	Renewable bool
	Severity  string // "warning" or "critical"
}

// LeaseChecker evaluates leases against configured thresholds.
type LeaseChecker struct {
	Client          *Client
	WarningThreshold  time.Duration
	CriticalThreshold time.Duration
}

// NewLeaseChecker creates a LeaseChecker with the given thresholds.
func NewLeaseChecker(client *Client, warning, critical time.Duration) (*LeaseChecker, error) {
	if critical >= warning {
		return nil, fmt.Errorf("critical threshold (%v) must be less than warning threshold (%v)", critical, warning)
	}
	return &LeaseChecker{
		Client:            client,
		WarningThreshold:  warning,
		CriticalThreshold: critical,
	}, nil
}

// CheckLeases looks up each lease ID and returns alerts for those near expiry.
func (lc *LeaseChecker) CheckLeases(leaseIDs []string) ([]ExpiryAlert, error) {
	var alerts []ExpiryAlert

	for _, id := range leaseIDs {
		info, err := lc.Client.LookupLease(id)
		if err != nil {
			return nil, fmt.Errorf("checking lease %q: %w", id, err)
		}

		severity := lc.classify(info.TTL)
		if severity == "" {
			continue
		}

		alerts = append(alerts, ExpiryAlert{
			LeaseID:   info.LeaseID,
			TTL:       info.TTL,
			ExpiresAt: info.ExpiresAt,
			Renewable: info.Renewable,
			Severity:  severity,
		})
	}

	return alerts, nil
}

// classify returns the severity level for a given TTL, or empty string if healthy.
func (lc *LeaseChecker) classify(ttl time.Duration) string {
	switch {
	case ttl <= lc.CriticalThreshold:
		return "critical"
	case ttl <= lc.WarningThreshold:
		return "warning"
	default:
		return ""
	}
}
