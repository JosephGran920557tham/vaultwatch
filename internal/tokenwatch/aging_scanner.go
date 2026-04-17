package tokenwatch

import (
	"context"

	"github.com/vaultwatch/internal/alert"
)

// AgingScanner scans all registered tokens for aging alerts.
type AgingScanner struct {
	registry *Registry
	detector *AgingDetector
	lookup   func(ctx context.Context, tokenID string) (TokenInfo, error)
}

// NewAgingScanner creates a new AgingScanner. Panics if any argument is nil.
func NewAgingScanner(
	registry *Registry,
	detector *AgingDetector,
	lookup func(ctx context.Context, tokenID string) (TokenInfo, error),
) *AgingScanner {
	if registry == nil {
		panic("tokenwatch: AgingScanner requires a non-nil registry")
	}
	if detector == nil {
		panic("tokenwatch: AgingScanner requires a non-nil detector")
	}
	if lookup == nil {
		panic("tokenwatch: AgingScanner requires a non-nil lookup func")
	}
	return &AgingScanner{
		registry: registry,
		detector: detector,
		lookup:   lookup,
	}
}

// Scan checks all registered tokens and returns aging alerts.
func (s *AgingScanner) Scan(ctx context.Context) []alert.Alert {
	tokens := s.registry.List()
	var alerts []alert.Alert
	for _, id := range tokens {
		info, err := s.lookup(ctx, id)
		if err != nil {
			continue
		}
		if a := s.detector.Check(id, info); a != nil {
			alerts = append(alerts, *a)
		}
	}
	return alerts
}
