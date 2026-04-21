package tokenwatch

import (
	"github.com/vaultwatch/internal/alert"
)

// CorrelationScanner iterates over the token registry and emits correlated
// alerts for tokens whose event density exceeds the detector threshold.
type CorrelationScanner struct {
	registry *Registry
	detector *CorrelationDetector
	lookup   func(tokenID string) (TokenInfo, error)
}

// NewCorrelationScanner creates a CorrelationScanner.
// Panics if any argument is nil.
func NewCorrelationScanner(
	registry *Registry,
	detector *CorrelationDetector,
	lookup func(tokenID string) (TokenInfo, error),
) *CorrelationScanner {
	if registry == nil {
		panic("tokenwatch: CorrelationScanner requires non-nil registry")
	}
	if detector == nil {
		panic("tokenwatch: CorrelationScanner requires non-nil detector")
	}
	if lookup == nil {
		panic("tokenwatch: CorrelationScanner requires non-nil lookup")
	}
	return &CorrelationScanner{
		registry: registry,
		detector: detector,
		lookup:   lookup,
	}
}

// Scan checks all registered tokens and returns correlated alerts.
func (s *CorrelationScanner) Scan() []alert.Alert {
	tokens := s.registry.List()
	var results []alert.Alert
	for _, id := range tokens {
		info, err := s.lookup(id)
		if err != nil {
			continue
		}
		s.detector.Record(id)
		level := alert.LevelInfo
		if info.TTL > 0 {
			level = alert.LevelWarning
		}
		a := s.detector.Check(id, level)
		if a != nil {
			results = append(results, *a)
		}
	}
	return results
}
