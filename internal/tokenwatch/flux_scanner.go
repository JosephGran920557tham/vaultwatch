package tokenwatch

import (
	"context"
	"time"
)

// FluxScanner periodically scans all registered tokens for TTL flux
// (rapid oscillation) and emits alerts when instability is detected.
type FluxScanner struct {
	registry *Registry
	detector *FluxDetector
	lookup   func(ctx context.Context, tokenID string) (TokenInfo, error)
}

// NewFluxScanner creates a FluxScanner. All arguments are required.
func NewFluxScanner(
	registry *Registry,
	detector *FluxDetector,
	lookup func(ctx context.Context, tokenID string) (TokenInfo, error),
) *FluxScanner {
	if registry == nil {
		panic("tokenwatch: FluxScanner requires a non-nil registry")
	}
	if detector == nil {
		panic("tokenwatch: FluxScanner requires a non-nil detector")
	}
	if lookup == nil {
		panic("tokenwatch: FluxScanner requires a non-nil lookup func")
	}
	return &FluxScanner{
		registry: registry,
		detector: detector,
		lookup:   lookup,
	}
}

// Scan checks every token in the registry for flux and returns alerts.
func (s *FluxScanner) Scan(ctx context.Context) ([]Alert, error) {
	tokens := s.registry.List()
	var alerts []Alert
	for _, id := range tokens {
		info, err := s.lookup(ctx, id)
		if err != nil {
			continue
		}
		s.detector.Record(id, info.TTL)
		flux := s.detector.Flux(id)
		if flux <= 0 {
			continue
		}
		level := LevelInfo
		if flux >= s.detector.cfg.CriticalFlux {
			level = LevelCritical
		} else if flux >= s.detector.cfg.WarningFlux {
			level = LevelWarning
		}
		alerts = append(alerts, Alert{
			LeaseID:   id,
			Level:     level,
			Message:   "token TTL flux detected",
			ExpiresAt: time.Now().Add(info.TTL),
		})
	}
	return alerts, nil
}
