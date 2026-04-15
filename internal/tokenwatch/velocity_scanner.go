package tokenwatch

import (
	"github.com/yourusername/vaultwatch/internal/alert"
)

// VelocityScanner feeds TTL observations into a VelocityDetector for every
// registered token and returns any resulting alerts.
type VelocityScanner struct {
	registry *Registry
	detector *VelocityDetector
}

// NewVelocityScanner creates a VelocityScanner.
// Panics if registry or detector is nil.
func NewVelocityScanner(registry *Registry, detector *VelocityDetector) *VelocityScanner {
	if registry == nil {
		panic("tokenwatch: VelocityScanner requires a non-nil Registry")
	}
	if detector == nil {
		panic("tokenwatch: VelocityScanner requires a non-nil VelocityDetector")
	}
	return &VelocityScanner{registry: registry, detector: detector}
}

// Scan records the current TTL for each token and returns velocity alerts.
func (s *VelocityScanner) Scan(lookup func(tokenID string) (TokenInfo, error)) []alert.Alert {
	tokens := s.registry.List()
	var alerts []alert.Alert
	for _, id := range tokens {
		info, err := lookup(id)
		if err != nil {
			continue
		}
		s.detector.Record(id, info.TTL)
		if a := s.detector.Check(id); a != nil {
			alerts = append(alerts, *a)
		}
	}
	return alerts
}
