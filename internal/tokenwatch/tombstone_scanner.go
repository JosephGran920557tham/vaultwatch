package tokenwatch

import "github.com/vaultwatch/internal/alert"

// TombstoneScanner wraps an inner scanner and suppresses alerts for
// tokens that have already been revoked according to a Tombstone.
type TombstoneScanner struct {
	inner     Scanner
	tombstone *Tombstone
}

// Scanner is a local interface satisfied by any scan source that
// returns a slice of alerts.
type Scanner interface {
	Scan() ([]alert.Alert, error)
}

// NewTombstoneScanner creates a TombstoneScanner.
// Panics if inner or tombstone is nil.
func NewTombstoneScanner(inner Scanner, tombstone *Tombstone) *TombstoneScanner {
	if inner == nil {
		panic("tombstone_scanner: inner scanner must not be nil")
	}
	if tombstone == nil {
		panic("tombstone_scanner: tombstone must not be nil")
	}
	return &TombstoneScanner{inner: inner, tombstone: tombstone}
}

// Scan delegates to the inner scanner and filters out any alerts whose
// LeaseID (used as the token identifier) is recorded in the Tombstone.
func (s *TombstoneScanner) Scan() ([]alert.Alert, error) {
	alerts, err := s.inner.Scan()
	if err != nil {
		return nil, err
	}
	filtered := alerts[:0]
	for _, a := range alerts {
		if !s.tombstone.IsRevoked(a.LeaseID) {
			filtered = append(filtered, a)
		}
	}
	return filtered, nil
}
