package tokenwatch

import (
	"context"
	"log"

	"github.com/vaultwatch/internal/alert"
)

// BudgetScanner scans all registered tokens and emits alerts for those whose
// renewal budgets are exhausted or near exhaustion.
type BudgetScanner struct {
	registry *Registry
	detector *BudgetDetector
	lookup   func(ctx context.Context, tokenID string) (TokenInfo, error)
	logger   *log.Logger
}

// NewBudgetScanner constructs a BudgetScanner. All arguments are required.
func NewBudgetScanner(
	reg *Registry,
	det *BudgetDetector,
	lookup func(ctx context.Context, tokenID string) (TokenInfo, error),
) *BudgetScanner {
	if reg == nil {
		panic("BudgetScanner: registry must not be nil")
	}
	if det == nil {
		panic("BudgetScanner: detector must not be nil")
	}
	if lookup == nil {
		panic("BudgetScanner: lookup must not be nil")
	}
	return &BudgetScanner{
		registry: reg,
		detector: det,
		lookup:   lookup,
		logger:   log.Default(),
	}
}

// Scan iterates over all tokens in the registry, records a renewal observation
// via the detector, and collects any budget alerts.
func (s *BudgetScanner) Scan(ctx context.Context) []alert.Alert {
	tokens := s.registry.List()
	var alerts []alert.Alert
	for _, id := range tokens {
		_, err := s.lookup(ctx, id)
		if err != nil {
			s.logger.Printf("BudgetScanner: lookup failed for token %s: %v", id, err)
			continue
		}
		s.detector.Record(id)
		if a := s.detector.Check(id); a != nil {
			alerts = append(alerts, *a)
		}
	}
	return alerts
}
