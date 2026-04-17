package tokenwatch

import "time"

// HeartbeatScanner iterates the registry and checks each token for missed
// heartbeats using a HeartbeatDetector.
type HeartbeatScanner struct {
	registry *Registry
	detector *HeartbeatDetector
	lookup   func(tokenID string) (*TokenInfo, error)
}

// NewHeartbeatScanner creates a HeartbeatScanner.
// Panics if any argument is nil.
func NewHeartbeatScanner(r *Registry, d *HeartbeatDetector, lookup func(string) (*TokenInfo, error)) *HeartbeatScanner {
	if r == nil {
		panic("heartbeat scanner: nil registry")
	}
	if d == nil {
		panic("heartbeat scanner: nil detector")
	}
	if lookup == nil {
		panic("heartbeat scanner: nil lookup")
	}
	return &HeartbeatScanner{registry: r, detector: d, lookup: lookup}
}

// Scan checks all registered tokens and returns any heartbeat alerts.
// Tokens that are successfully looked up have their heartbeat recorded.
func (s *HeartbeatScanner) Scan() []Alert {
	tokens := s.registry.List()
	now := time.Now()
	var alerts []Alert
	for _, id := range tokens {
		_, err := s.lookup(id)
		if err == nil {
			s.detector.Beat(id)
		}
		if a := s.detector.Check(id, now); a != nil {
			alerts = append(alerts, *a)
		}
	}
	return alerts
}
