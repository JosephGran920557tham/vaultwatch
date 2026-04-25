package tokenwatch

import (
	"time"

	"github.com/vaultwatch/internal/alert"
)

// TenureScanner scans all registered tokens for long-lived credential alerts.
type TenureScanner struct {
	registry *Registry
	detector *TenureDetector
	lookup   func(tokenID string) (TokenInfo, error)
}

// NewTenureScanner constructs a TenureScanner.
// Panics if any argument is nil.
func NewTenureScanner(
	registry *Registry,
	detector *TenureDetector,
	lookup func(tokenID string) (TokenInfo, error),
) *TenureScanner {
	if registry == nil {
		panic("tokenwatch: TenureScanner requires non-nil registry")
	}
	if detector == nil {
		panic("tokenwatch: TenureScanner requires non-nil detector")
	}
	if lookup == nil {
		panic("tokenwatch: TenureScanner requires non-nil lookup")
	}
	return &TenureScanner{
		registry: registry,
		detector: detector,
		lookup:   lookup,
	}
}

// Scan iterates over all registered tokens, tracks their issue time via
// lookup, and returns any tenure alerts.
func (s *TenureScanner) Scan() []alert.Alert {
	tokens := s.registry.List()
	now := time.Now()
	var alerts []alert.Alert
	for _, id := range tokens {
		info, err := s.lookup(id)
		if err != nil {
			continue
		}
		if !info.IssueTime.IsZero() {
			s.detector.Track(id, info.IssueTime)
		}
		if a := s.detector.Check(id, now); a != nil {
			alerts = append(alerts, *a)
		}
	}
	return alerts
}
