package tokenwatch

import (
	"fmt"
	"time"

	"github.com/vaultwatch/internal/alert"
)

// BurstScanner watches the token registry and emits alerts when a token's
// renewal or check events exceed the configured burst threshold.
type BurstScanner struct {
	registry *Registry
	detector *BurstDetector
}

// NewBurstScanner constructs a BurstScanner. Panics if registry or detector
// are nil.
func NewBurstScanner(registry *Registry, detector *BurstDetector) *BurstScanner {
	if registry == nil {
		panic("tokenwatch: BurstScanner requires a non-nil Registry")
	}
	if detector == nil {
		panic("tokenwatch: BurstScanner requires a non-nil BurstDetector")
	}
	return &BurstScanner{registry: registry, detector: detector}
}

// Record registers an event for tokenID at the current time.
func (s *BurstScanner) Record(tokenID string) {
	s.detector.Record(tokenID, time.Now())
}

// Scan checks all registered tokens and returns one alert per token that is
// currently bursting.
func (s *BurstScanner) Scan() []alert.Alert {
	now := time.Now()
	tokens := s.registry.List()
	var alerts []alert.Alert
	for _, id := range tokens {
		if s.detector.IsBursting(id, now) {
			count := s.detector.Count(id, now)
			alerts = append(alerts, alert.Alert{
				LeaseID:  id,
				Level:    alert.LevelWarning,
				Message:  fmt.Sprintf("token %s is bursting: %d events in window", id, count),
				IssuedAt: now,
			})
		}
	}
	return alerts
}
