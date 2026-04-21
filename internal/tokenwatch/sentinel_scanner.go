package tokenwatch

import (
	"time"

	"github.com/vaultwatch/internal/alert"
)

// SentinelScanner runs the SentinelDetector across all registered tokens.
type SentinelScanner struct {
	registry *Registry
	detector *SentinelDetector
	lookup   func(tokenID string) (TokenInfo, error)
}

// NewSentinelScanner constructs a SentinelScanner. All arguments are required.
func NewSentinelScanner(registry *Registry, detector *SentinelDetector, lookup func(string) (TokenInfo, error)) *SentinelScanner {
	if registry == nil {
		panic("sentinel_scanner: registry must not be nil")
	}
	if detector == nil {
		panic("sentinel_scanner: detector must not be nil")
	}
	if lookup == nil {
		panic("sentinel_scanner: lookup must not be nil")
	}
	return &SentinelScanner{
		registry: registry,
		detector: detector,
		lookup:   lookup,
	}
}

// Scan iterates over all registered tokens, pings those that are reachable,
// and returns alerts for any that exceed the miss threshold.
func (s *SentinelScanner) Scan() []alert.Alert {
	tokens := s.registry.List()
	now := time.Now()
	var alerts []alert.Alert

	for _, id := range tokens {
		_, err := s.lookup(id)
		if err == nil {
			s.detector.Ping(id)
			continue
		}
		if a := s.detector.Check(id, now); a != nil {
			alerts = append(alerts, *a)
		}
	}
	return alerts
}
