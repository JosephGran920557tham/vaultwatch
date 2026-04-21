package tokenwatch

import (
	"context"
	"fmt"

	"github.com/your-org/vaultwatch/internal/alert"
)

// HorizonScanner scans all registered tokens for horizon-based TTL alerts,
// flagging tokens whose remaining TTL falls within a configurable look-ahead
// window. It complements the expiry classifier by projecting forward in time
// rather than reacting to the current TTL alone.
type HorizonScanner struct {
	registry *Registry
	detector *HorizonDetector
	lookup   func(ctx context.Context, tokenID string) (TokenInfo, error)
}

// NewHorizonScanner constructs a HorizonScanner. All three dependencies are
// required; the function panics if any is nil.
func NewHorizonScanner(
	registry *Registry,
	detector *HorizonDetector,
	lookup func(ctx context.Context, tokenID string) (TokenInfo, error),
) *HorizonScanner {
	if registry == nil {
		panic("tokenwatch: HorizonScanner requires a non-nil Registry")
	}
	if detector == nil {
		panic("tokenwatch: HorizonScanner requires a non-nil HorizonDetector")
	}
	if lookup == nil {
		panic("tokenwatch: HorizonScanner requires a non-nil lookup function")
	}
	return &HorizonScanner{
		registry: registry,
		detector: detector,
		lookup:   lookup,
	}
}

// Scan iterates over every token in the registry, fetches its current TTL via
// the lookup function, and runs the HorizonDetector against it. Tokens that
// cannot be looked up are skipped with a warning-level alert so the caller is
// still notified of the failure without aborting the entire scan.
func (s *HorizonScanner) Scan(ctx context.Context) []alert.Alert {
	tokens := s.registry.List()
	if len(tokens) == 0 {
		return nil
	}

	var alerts []alert.Alert

	for _, tokenID := range tokens {
		info, err := s.lookup(ctx, tokenID)
		if err != nil {
			alerts = append(alerts, alert.Alert{
				LeaseID:  tokenID,
				Level:    alert.LevelWarning,
				Message:  fmt.Sprintf("horizon scanner: lookup failed for token %s: %v", tokenID, err),
				Metadata: map[string]string{"scanner": "horizon"},
			})
			continue
		}

		if a := s.detector.Check(info); a != nil {
			alerts = append(alerts, *a)
		}
	}

	return alerts
}
