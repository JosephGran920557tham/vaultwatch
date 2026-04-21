package tokenwatch

import (
	"context"
	"log"
)

// CensusScanner populates a Census by iterating over all registered tokens
// and calling Observe for each one found via the lookup function.
type CensusScanner struct {
	registry *Registry
	census   *Census
	lookup   func(ctx context.Context, tokenID string) (TokenInfo, error)
	logger   *log.Logger
}

// NewCensusScanner creates a CensusScanner. Panics if any required argument
// is nil.
func NewCensusScanner(
	registry *Registry,
	census *Census,
	lookup func(ctx context.Context, tokenID string) (TokenInfo, error),
) *CensusScanner {
	if registry == nil {
		panic("census_scanner: registry must not be nil")
	}
	if census == nil {
		panic("census_scanner: census must not be nil")
	}
	if lookup == nil {
		panic("census_scanner: lookup must not be nil")
	}
	return &CensusScanner{
		registry: registry,
		census:   census,
		lookup:   lookup,
		logger:   log.Default(),
	}
}

// Scan iterates over all registered tokens, looks each one up, and records
// it in the Census. Tokens that fail lookup are logged and skipped.
func (s *CensusScanner) Scan(ctx context.Context) {
	tokens := s.registry.List()
	for _, id := range tokens {
		info, err := s.lookup(ctx, id)
		if err != nil {
			s.logger.Printf("census_scanner: lookup failed for %s: %v", id, err)
			continue
		}
		s.census.Observe(id, info.Labels)
	}
}
