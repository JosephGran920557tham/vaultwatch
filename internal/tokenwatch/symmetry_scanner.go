package tokenwatch

import (
	"fmt"

	"github.com/vaultwatch/internal/alert"
)

// SymmetryScanner scans all registered tokens for TTL symmetry violations.
type SymmetryScanner struct {
	registry *Registry
	detector *SymmetryDetector
	lookup   func(tokenID string) (TokenInfo, error)
}

// NewSymmetryScanner creates a SymmetryScanner.
// Panics if registry, detector, or lookup is nil.
func NewSymmetryScanner(
	registry *Registry,
	detector *SymmetryDetector,
	lookup func(tokenID string) (TokenInfo, error),
) *SymmetryScanner {
	if registry == nil {
		panic("symmetry scanner: registry must not be nil")
	}
	if detector == nil {
		panic("symmetry scanner: detector must not be nil")
	}
	if lookup == nil {
		panic("symmetry scanner: lookup must not be nil")
	}
	return &SymmetryScanner{
		registry: registry,
		detector: detector,
		lookup:   lookup,
	}
}

// Scan observes all tokens then checks each one for symmetry violations.
// Errors from lookup are skipped with the token omitted from results.
func (s *SymmetryScanner) Scan() ([]alert.Alert, error) {
	tokens := s.registry.List()

	// First pass: observe all TTLs so the detector has a full peer picture.
	for _, id := range tokens {
		info, err := s.lookup(id)
		if err != nil {
			continue
		}
		s.detector.Observe(id, info.TTL)
	}

	// Second pass: check each token against its peers.
	var alerts []alert.Alert
	for _, id := range tokens {
		a := s.detector.Check(id)
		if a != nil {
			alerts = append(alerts, *a)
		}
	}

	_ = fmt.Sprintf // keep import
	return alerts, nil
}
