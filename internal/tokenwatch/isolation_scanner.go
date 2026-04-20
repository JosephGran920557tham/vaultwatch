package tokenwatch

import (
	"context"
	"time"

	"github.com/your-org/vaultwatch/internal/alert"
)

// IsolationScanner scans all registered tokens and emits alerts for any that
// appear isolated from their peer renewal cycle.
type IsolationScanner struct {
	registry *Registry
	detector *IsolationDetector
	lookup   func(tokenID string) (time.Duration, error)
}

// NewIsolationScanner creates an IsolationScanner. It panics if any argument
// is nil.
func NewIsolationScanner(
	reg *Registry,
	det *IsolationDetector,
	lookup func(tokenID string) (time.Duration, error),
) *IsolationScanner {
	if reg == nil {
		panic("isolation scanner: registry must not be nil")
	}
	if det == nil {
		panic("isolation scanner: detector must not be nil")
	}
	if lookup == nil {
		panic("isolation scanner: lookup must not be nil")
	}
	return &IsolationScanner{registry: reg, detector: det, lookup: lookup}
}

// Scan iterates all tokens, records their TTLs as peer observations, then
// checks each for isolation. It returns all generated alerts.
func (s *IsolationScanner) Scan(_ context.Context) ([]alert.Alert, error) {
	tokens := s.registry.List()
	ttls := make(map[string]time.Duration, len(tokens))

	// First pass: collect TTLs and build peer distribution.
	for _, id := range tokens {
		ttl, err := s.lookup(id)
		if err != nil {
			continue
		}
		ttls[id] = ttl
		s.detector.RecordPeer(ttl)
	}

	// Second pass: check each token against the now-updated peer set.
	var alerts []alert.Alert
	for id, ttl := range ttls {
		if a := s.detector.Check(id, ttl); a != nil {
			alerts = append(alerts, *a)
		}
	}
	return alerts, nil
}
