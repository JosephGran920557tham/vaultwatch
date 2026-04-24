package tokenwatch

import (
	"fmt"

	"github.com/vaultwatch/internal/alert"
)

// EntropyScanner walks the token registry, records current TTLs, and emits
// alerts for tokens whose renewal entropy falls below configured thresholds.
type EntropyScanner struct {
	registry *Registry
	detector *EntropyDetector
	lookup   func(tokenID string) (TokenInfo, error)
}

// NewEntropyScanner constructs an EntropyScanner.
// Panics if any required dependency is nil.
func NewEntropyScanner(r *Registry, d *EntropyDetector, lookup func(string) (TokenInfo, error)) *EntropyScanner {
	if r == nil {
		panic("tokenwatch: EntropyScanner requires a non-nil Registry")
	}
	if d == nil {
		panic("tokenwatch: EntropyScanner requires a non-nil EntropyDetector")
	}
	if lookup == nil {
		panic("tokenwatch: EntropyScanner requires a non-nil lookup func")
	}
	return &EntropyScanner{registry: r, detector: d, lookup: lookup}
}

// Scan iterates over all registered tokens, records their current TTL, and
// returns any entropy-based alerts.
func (s *EntropyScanner) Scan() ([]alert.Alert, error) {
	tokens := s.registry.List()
	var alerts []alert.Alert

	for _, id := range tokens {
		info, err := s.lookup(id)
		if err != nil {
			// Non-fatal: skip this token but continue scanning others.
			continue
		}
		s.detector.Record(id, info.TTL)
		if a := s.detector.Check(id); a != nil {
			alerts = append(alerts, *a)
		}
	}

	return alerts, nil
}

// String implements fmt.Stringer for diagnostics.
func (s *EntropyScanner) String() string {
	return fmt.Sprintf("EntropyScanner(tokens=%d)", len(s.registry.List()))
}
