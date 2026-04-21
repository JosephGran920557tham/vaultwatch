package tokenwatch

import (
	"fmt"
	"time"

	"github.com/your-org/vaultwatch/internal/alert"
)

// FingerprintScanner checks each registered token for identity changes
// by comparing its current label fingerprint against the last known hash.
type FingerprintScanner struct {
	registry    *Registry
	fingerprint *Fingerprint
	lookup      func(tokenID string) (map[string]string, error)
}

// NewFingerprintScanner constructs a FingerprintScanner.
// It panics if any required dependency is nil.
func NewFingerprintScanner(r *Registry, fp *Fingerprint, lookup func(string) (map[string]string, error)) *FingerprintScanner {
	if r == nil {
		panic("fingerprint scanner: registry is nil")
	}
	if fp == nil {
		panic("fingerprint scanner: fingerprint tracker is nil")
	}
	if lookup == nil {
		panic("fingerprint scanner: lookup func is nil")
	}
	return &FingerprintScanner{registry: r, fingerprint: fp, lookup: lookup}
}

// Scan iterates all registered tokens and emits a critical alert for any
// token whose label fingerprint has changed since the last scan.
func (s *FingerprintScanner) Scan() []alert.Alert {
	tokens := s.registry.List()
	var alerts []alert.Alert
	for _, id := range tokens {
		labels, err := s.lookup(id)
		if err != nil {
			continue
		}
		hash := Compute(labels)
		if s.fingerprint.Track(id, hash) {
			// Only emit an alert when the token is not brand new (i.e. a change
			// was detected, not just the first observation). We can't distinguish
			// first-seen from changed here without extra state, so we emit for
			// all changes and let downstream deduplication suppress first-seen noise.
			alerts = append(alerts, alert.Alert{
				LeaseID:   id,
				Level:     alert.LevelCritical,
				Message:   fmt.Sprintf("token %s: identity fingerprint changed", id),
				ExpiresAt: time.Now().Add(time.Hour),
			})
		}
	}
	return alerts
}
