package tokenwatch

import "log"

// BaselineScanner scans all registered tokens, records their current TTL
// into the BaselineDetector, and surfaces alerts for significant deviations.
type BaselineScanner struct {
	registry *Registry
	detector *BaselineDetector
	lookup   func(tokenID string) (*TokenInfo, error)
}

// NewBaselineScanner creates a BaselineScanner.
// Panics if registry, detector, or lookup are nil.
func NewBaselineScanner(
	registry *Registry,
	detector *BaselineDetector,
	lookup func(tokenID string) (*TokenInfo, error),
) *BaselineScanner {
	if registry == nil {
		panic("tokenwatch: BaselineScanner requires a non-nil registry")
	}
	if detector == nil {
		panic("tokenwatch: BaselineScanner requires a non-nil detector")
	}
	if lookup == nil {
		panic("tokenwatch: BaselineScanner requires a non-nil lookup func")
	}
	return &BaselineScanner{
		registry: registry,
		detector: detector,
		lookup:   lookup,
	}
}

// Scan iterates every registered token, records its TTL sample, and
// returns any baseline-deviation alerts.
func (s *BaselineScanner) Scan() []Alert {
	tokens := s.registry.List()
	var alerts []Alert

	for _, id := range tokens {
		info, err := s.lookup(id)
		if err != nil {
			log.Printf("baseline_scanner: lookup error for %s: %v", id, err)
			continue
		}
		s.detector.Record(id, info.TTL)
		if a := s.detector.Check(id, info.TTL); a != nil {
			alerts = append(alerts, *a)
		}
	}
	return alerts
}
