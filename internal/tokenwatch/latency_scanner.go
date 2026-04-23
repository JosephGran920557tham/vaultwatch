package tokenwatch

import (
	"context"
	"time"

	"github.com/vaultwatch/internal/alert"
)

// LatencyScanner periodically measures token lookup latency and emits
// alerts when latency thresholds are breached.
type LatencyScanner struct {
	registry *Registry
	detector *LatencyDetector
	lookup   func(ctx context.Context, tokenID string) (TokenInfo, error)
}

// NewLatencyScanner constructs a LatencyScanner.
// Panics if any argument is nil.
func NewLatencyScanner(reg *Registry, det *LatencyDetector, lookup func(context.Context, string) (TokenInfo, error)) *LatencyScanner {
	if reg == nil {
		panic("latency scanner: registry is nil")
	}
	if det == nil {
		panic("latency scanner: detector is nil")
	}
	if lookup == nil {
		panic("latency scanner: lookup is nil")
	}
	return &LatencyScanner{registry: reg, detector: det, lookup: lookup}
}

// Scan iterates over all registered tokens, measures lookup latency, records
// the sample in the detector, and returns any resulting alerts.
func (s *LatencyScanner) Scan(ctx context.Context) []alert.Alert {
	tokens := s.registry.List()
	var alerts []alert.Alert
	for _, id := range tokens {
		start := time.Now()
		_, err := s.lookup(ctx, id)
		latency := time.Since(start)
		if err != nil {
			continue
		}
		if a := s.detector.Check(id, latency); a != nil {
			alerts = append(alerts, *a)
		}
	}
	return alerts
}
