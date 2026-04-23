package tokenwatch

import (
	"context"
	"time"
)

// DriftScanner scans all registered tokens for TTL drift relative to
// an established baseline and emits alerts when drift exceeds thresholds.
type DriftScanner struct {
	registry *Registry
	detector *DriftDetector
	lookup   func(ctx context.Context, tokenID string) (TokenInfo, error)
}

// NewDriftScanner creates a DriftScanner. All arguments are required.
func NewDriftScanner(
	registry *Registry,
	detector *DriftDetector,
	lookup func(ctx context.Context, tokenID string) (TokenInfo, error),
) *DriftScanner {
	if registry == nil {
		panic("tokenwatch: DriftScanner requires a non-nil registry")
	}
	if detector == nil {
		panic("tokenwatch: DriftScanner requires a non-nil detector")
	}
	if lookup == nil {
		panic("tokenwatch: DriftScanner requires a non-nil lookup func")
	}
	return &DriftScanner{
		registry: registry,
		detector: detector,
		lookup:   lookup,
	}
}

// Scan checks every token in the registry for TTL drift and returns alerts.
func (s *DriftScanner) Scan(ctx context.Context) ([]Alert, error) {
	tokens := s.registry.List()
	var alerts []Alert
	for _, id := range tokens {
		info, err := s.lookup(ctx, id)
		if err != nil {
			continue
		}
		a := s.detector.Check(id, info.TTL)
		if a == nil {
			continue
		}
		a.ExpiresAt = time.Now().Add(info.TTL)
		alerts = append(alerts, *a)
	}
	return alerts, nil
}
