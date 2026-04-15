package tokenwatch

import (
	"log"

	"github.com/vaultwatch/internal/alert"
)

// TrendScanner periodically records TTL observations and emits alerts for
// tokens whose TTL is declining faster than the configured threshold.
type TrendScanner struct {
	registry *Registry
	detector *TrendDetector
	lookup   func(tokenID string) (TokenInfo, error)
	log      *log.Logger
}

// NewTrendScanner creates a TrendScanner.
// Panics if registry, detector, or lookup are nil.
func NewTrendScanner(
	registry *Registry,
	detector *TrendDetector,
	lookup func(tokenID string) (TokenInfo, error),
	logger *log.Logger,
) *TrendScanner {
	if registry == nil {
		panic("tokenwatch: TrendScanner requires a non-nil Registry")
	}
	if detector == nil {
		panic("tokenwatch: TrendScanner requires a non-nil TrendDetector")
	}
	if lookup == nil {
		panic("tokenwatch: TrendScanner requires a non-nil lookup func")
	}
	if logger == nil {
		logger = log.Default()
	}
	return &TrendScanner{
		registry: registry,
		detector: detector,
		lookup:   lookup,
		log:      logger,
	}
}

// Scan records the current TTL for every registered token and returns any
// trend alerts produced by the detector.
func (s *TrendScanner) Scan() []alert.Alert {
	tokens := s.registry.List()
	var alerts []alert.Alert
	for _, id := range tokens {
		info, err := s.lookup(id)
		if err != nil {
			s.log.Printf("trend_scanner: lookup error for token %s: %v", id, err)
			continue
		}
		s.detector.Record(id, info.TTL)
		if a := s.detector.Check(id); a != nil {
			alerts = append(alerts, *a)
		}
	}
	return alerts
}
