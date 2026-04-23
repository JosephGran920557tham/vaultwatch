package tokenwatch

import (
	"fmt"

	"github.com/your-org/vaultwatch/internal/alert"
)

// AffinityScanner scans registered tokens and emits an alert when a token's
// affinity score drops below a configured threshold, indicating reduced activity.
type AffinityScanner struct {
	registry  *Registry
	detector  *AffinityDetector
	lookup    func(tokenID string) (TokenInfo, error)
	threshold float64
}

// NewAffinityScanner creates an AffinityScanner.
// Panics if registry, detector, or lookup are nil.
func NewAffinityScanner(r *Registry, d *AffinityDetector, lookup func(string) (TokenInfo, error), threshold float64) *AffinityScanner {
	if r == nil {
		panic("affinity scanner: registry is nil")
	}
	if d == nil {
		panic("affinity scanner: detector is nil")
	}
	if lookup == nil {
		panic("affinity scanner: lookup is nil")
	}
	if threshold <= 0 {
		threshold = 0.2
	}
	return &AffinityScanner{
		registry:  r,
		detector:  d,
		lookup:    lookup,
		threshold: threshold,
	}
}

// Scan iterates over all registered tokens, records an observation, and
// returns warning alerts for tokens whose affinity score is below the threshold.
func (s *AffinityScanner) Scan() ([]alert.Alert, error) {
	tokens := s.registry.List()
	var alerts []alert.Alert

	for _, id := range tokens {
		info, err := s.lookup(id)
		if err != nil {
			continue
		}

		score := s.detector.Observe(id)
		if score < s.threshold {
			alerts = append(alerts, alert.Alert{
				LeaseID:  id,
				Level:    alert.LevelWarning,
				Message:  fmt.Sprintf("token affinity score %.2f below threshold %.2f", score, s.threshold),
				Metadata: map[string]string{"ttl": fmt.Sprintf("%d", int(info.TTL.Seconds()))},
			})
		}
	}

	return alerts, nil
}
