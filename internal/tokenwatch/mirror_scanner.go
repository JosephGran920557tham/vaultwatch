package tokenwatch

import (
	"fmt"
	"time"
)

// MirrorScanner observes token TTLs into a Mirror and optionally surfaces
// alerts when the mirrored value diverges significantly from a fresh lookup.
type MirrorScanner struct {
	registry *Registry
	mirror   *Mirror
	lookup   func(tokenID string) (TokenInfo, error)
	delta    time.Duration // alert when |fresh - mirrored| > delta
}

// NewMirrorScanner constructs a MirrorScanner. Panics if any required
// dependency is nil.
func NewMirrorScanner(
	reg *Registry,
	mirror *Mirror,
	lookup func(tokenID string) (TokenInfo, error),
	delta time.Duration,
) *MirrorScanner {
	if reg == nil {
		panic("tokenwatch: MirrorScanner requires non-nil Registry")
	}
	if mirror == nil {
		panic("tokenwatch: MirrorScanner requires non-nil Mirror")
	}
	if lookup == nil {
		panic("tokenwatch: MirrorScanner requires non-nil lookup")
	}
	if delta <= 0 {
		delta = 30 * time.Second
	}
	return &MirrorScanner{
		registry: reg,
		mirror:   mirror,
		lookup:   lookup,
		delta:    delta,
	}
}

// Scan iterates all registered tokens, refreshes the mirror, and returns
// alerts for tokens whose TTL has drifted beyond the configured delta.
func (s *MirrorScanner) Scan() ([]Alert, error) {
	tokens := s.registry.List()
	var alerts []Alert
	for _, id := range tokens {
		info, err := s.lookup(id)
		if err != nil {
			continue
		}
		freshTTL := info.TTL
		mirrored, ok := s.mirror.Get(id)
		s.mirror.Observe(id, freshTTL)
		if !ok {
			continue
		}
		diff := freshTTL - mirrored
		if diff < 0 {
			diff = -diff
		}
		if diff > s.delta {
			alerts = append(alerts, Alert{
				LeaseID: id,
				Level:   LevelWarning,
				Message: fmt.Sprintf("mirror TTL drift of %s detected (fresh=%s mirrored=%s)", diff, freshTTL, mirrored),
				Labels:  info.Labels,
			})
		}
	}
	return alerts, nil
}
