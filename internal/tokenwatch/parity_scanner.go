package tokenwatch

import (
	"context"
	"log"

	"github.com/vaultwatch/internal/alert"
)

// parityLookup fetches the current TTL for a token by ID.
type parityLookup func(ctx context.Context, tokenID string) (TokenInfo, error)

// ParityScanner iterates registered token pairs and emits parity alerts.
type ParityScanner struct {
	registry *Registry
	detector *ParityDetector
	lookup   parityLookup
	logger   *log.Logger
}

// NewParityScanner constructs a ParityScanner. Panics if any argument is nil.
func NewParityScanner(registry *Registry, detector *ParityDetector, lookup parityLookup) *ParityScanner {
	if registry == nil {
		panic("tokenwatch: ParityScanner requires non-nil registry")
	}
	if detector == nil {
		panic("tokenwatch: ParityScanner requires non-nil detector")
	}
	if lookup == nil {
		panic("tokenwatch: ParityScanner requires non-nil lookup")
	}
	return &ParityScanner{
		registry: registry,
		detector: detector,
		lookup:   lookup,
		logger:   log.Default(),
	}
}

// Scan checks all registered token pairs and returns any parity alerts.
func (s *ParityScanner) Scan(ctx context.Context) []alert.Alert {
	tokens := s.registry.List()
	var alerts []alert.Alert

	s.detector.mu.Lock()
	pairs := make(map[string]string, len(s.detector.pairs))
	for k, v := range s.detector.pairs {
		pairs[k] = v
	}
	s.detector.mu.Unlock()

	for _, tokenID := range tokens {
		peerID, ok := pairs[tokenID]
		if !ok {
			continue
		}
		info, err := s.lookup(ctx, tokenID)
		if err != nil {
			s.logger.Printf("parity_scanner: lookup %s: %v", tokenID, err)
			continue
		}
		peerInfo, err := s.lookup(ctx, peerID)
		if err != nil {
			s.logger.Printf("parity_scanner: lookup peer %s: %v", peerID, err)
			continue
		}
		if a := s.detector.Check(tokenID, info.TTL, peerInfo.TTL); a != nil {
			alerts = append(alerts, *a)
		}
	}
	return alerts
}
