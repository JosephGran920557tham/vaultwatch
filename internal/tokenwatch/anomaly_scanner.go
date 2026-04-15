package tokenwatch

import (
	"context"
	"fmt"
	"time"

	"github.com/vaultwatch/internal/alert"
)

// TTLSource is a function that returns the remaining TTL for a given token ID.
type TTLSource func(ctx context.Context, tokenID string) (time.Duration, error)

// AnomalyScanner scans all registered tokens for TTL anomalies.
type AnomalyScanner struct {
	registry *Registry
	detector *AnomalyDetector
	source   TTLSource
}

// NewAnomalyScanner creates a new AnomalyScanner.
// Panics if registry, detector, or source is nil.
func NewAnomalyScanner(registry *Registry, detector *AnomalyDetector, source TTLSource) *AnomalyScanner {
	if registry == nil {
		panic("anomaly scanner: registry must not be nil")
	}
	if detector == nil {
		panic("anomaly scanner: detector must not be nil")
	}
	if source == nil {
		panic("anomaly scanner: ttl source must not be nil")
	}
	return &AnomalyScanner{
		registry: registry,
		detector: detector,
		source:   source,
	}
}

// Scan iterates over all registered tokens, fetches their TTLs via the source,
// and returns any anomaly alerts detected. Lookup errors are logged but do not
// abort the scan.
func (s *AnomalyScanner) Scan(ctx context.Context) ([]alert.Alert, error) {
	tokens := s.registry.List()
	var alerts []alert.Alert

	for _, tokenID := range tokens {
		ttl, err := s.source(ctx, tokenID)
		if err != nil {
			// Non-fatal: skip this token and continue scanning.
			_ = fmt.Errorf("anomaly scanner: lookup %s: %w", tokenID, err)
			continue
		}
		if a := s.detector.Check(tokenID, ttl); a != nil {
			alerts = append(alerts, *a)
		}
	}
	return alerts, nil
}
