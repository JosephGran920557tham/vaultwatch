package tokenwatch

import (
	"context"
	"time"

	"github.com/your-org/vaultwatch/internal/alert"
)

// GraceLookup retrieves the remaining TTL for a token.
type GraceLookup func(ctx context.Context, tokenID string) (time.Duration, error)

// GraceScanner scans all registered tokens for grace period violations.
type GraceScanner struct {
	registry *Registry
	detector *GraceDetector
	lookup   GraceLookup
}

// NewGraceScanner constructs a GraceScanner. Panics if any argument is nil.
func NewGraceScanner(registry *Registry, detector *GraceDetector, lookup GraceLookup) *GraceScanner {
	if registry == nil {
		panic("grace scanner: registry must not be nil")
	}
	if detector == nil {
		panic("grace scanner: detector must not be nil")
	}
	if lookup == nil {
		panic("grace scanner: lookup must not be nil")
	}
	return &GraceScanner{registry: registry, detector: detector, lookup: lookup}
}

// Scan checks all tokens and returns any grace period alerts.
func (s *GraceScanner) Scan(ctx context.Context) []alert.Alert {
	tokens := s.registry.List()
	var results []alert.Alert
	for _, id := range tokens {
		ttl, err := s.lookup(ctx, id)
		if err != nil {
			continue
		}
		if a := s.detector.Check(id, ttl); a != nil {
			results = append(results, *a)
		}
	}
	return results
}
