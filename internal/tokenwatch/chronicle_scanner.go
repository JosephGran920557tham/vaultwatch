package tokenwatch

import (
	"context"

	"github.com/vaultwatch/internal/alert"
)

// ChronicleScanner records every alert produced by a source Scanner into a
// Chronicle, then forwards them unchanged to the caller.
type ChronicleScanner struct {
	registry  *Registry
	chronicle *Chronicle
	source    func(ctx context.Context, tokenID string) ([]alert.Alert, error)
}

// NewChronicleScanner creates a ChronicleScanner. All three arguments are required.
func NewChronicleScanner(
	registry *Registry,
	chronicle *Chronicle,
	source func(ctx context.Context, tokenID string) ([]alert.Alert, error),
) *ChronicleScanner {
	if registry == nil {
		panic("chronicle_scanner: registry must not be nil")
	}
	if chronicle == nil {
		panic("chronicle_scanner: chronicle must not be nil")
	}
	if source == nil {
		panic("chronicle_scanner: source must not be nil")
	}
	return &ChronicleScanner{
		registry:  registry,
		chronicle: chronicle,
		source:    source,
	}
}

// Scan iterates over all registered tokens, invokes the source for each, records
// results in the Chronicle, and returns the full set of alerts.
func (s *ChronicleScanner) Scan(ctx context.Context) ([]alert.Alert, error) {
	tokens := s.registry.List()
	var all []alert.Alert
	for _, tokenID := range tokens {
		alerts, err := s.source(ctx, tokenID)
		if err != nil {
			continue
		}
		for _, a := range alerts {
			s.chronicle.Record(tokenID, a)
		}
		all = append(all, alerts...)
	}
	return all, nil
}
