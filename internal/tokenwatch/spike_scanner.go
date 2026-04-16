package tokenwatch

import (
	"log"

	"github.com/yourusername/vaultwatch/internal/alert"
)

// SpikeScanner scans all registered tokens for TTL spikes.
type SpikeScanner struct {
	registry *Registry
	detector *SpikeDetector
	lookup   func(tokenID string) (TokenInfo, error)
	logger   *log.Logger
}

// NewSpikeScanner constructs a SpikeScanner. Panics if any argument is nil.
func NewSpikeScanner(registry *Registry, detector *SpikeDetector, lookup func(string) (TokenInfo, error), logger *log.Logger) *SpikeScanner {
	if registry == nil {
		panic("spike scanner: registry must not be nil")
	}
	if detector == nil {
		panic("spike scanner: detector must not be nil")
	}
	if lookup == nil {
		panic("spike scanner: lookup must not be nil")
	}
	if logger == nil {
		logger = log.Default()
	}
	return &SpikeScanner{
		registry: registry,
		detector: detector,
		lookup:   lookup,
		logger:   logger,
	}
}

// Scan checks every registered token for TTL spikes and returns any alerts.
func (s *SpikeScanner) Scan() []alert.Alert {
	tokens := s.registry.List()
	var results []alert.Alert
	for _, id := range tokens {
		info, err := s.lookup(id)
		if err != nil {
			s.logger.Printf("spike scanner: lookup %s: %v", id, err)
			continue
		}
		s.detector.Record(id, info.TTL)
		if a := s.detector.Check(id, info.TTL); a != nil {
			results = append(results, *a)
		}
	}
	return results
}
