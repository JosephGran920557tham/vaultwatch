package tokenwatch

import (
	"fmt"
	"time"

	"github.com/vaultwatch/internal/alert"
)

// ShadowScanner compares current token TTLs against their shadow copies
// and emits alerts when a significant drop is detected.
type ShadowScanner struct {
	registry *Registry
	shadow   *ShadowRegistry
	lookup   func(tokenID string) (time.Duration, error)
	dropRatio float64
}

// NewShadowScanner creates a ShadowScanner.
// dropRatio is the fraction of TTL drop (0–1) that triggers a warning.
func NewShadowScanner(reg *Registry, shadow *ShadowRegistry, lookup func(string) (time.Duration, error), dropRatio float64) *ShadowScanner {
	if reg == nil {
		panic("shadow scanner: nil registry")
	}
	if shadow == nil {
		panic("shadow scanner: nil shadow registry")
	}
	if lookup == nil {
		panic("shadow scanner: nil lookup")
	}
	if dropRatio <= 0 || dropRatio > 1 {
		dropRatio = 0.5
	}
	return &ShadowScanner{
		registry:  reg,
		shadow:    shadow,
		lookup:    lookup,
		dropRatio: dropRatio,
	}
}

// Scan checks each registered token against its shadow entry.
func (s *ShadowScanner) Scan() []alert.Alert {
	tokens := s.registry.List()
	var alerts []alert.Alert
	for _, id := range tokens {
		current, err := s.lookup(id)
		if err != nil {
			continue
		}
		prev, ok := s.shadow.Get(id)
		if !ok {
			s.shadow.Set(id, current)
			continue
		}
		if prev.TTL > 0 && float64(current) < float64(prev.TTL)*(1-s.dropRatio) {
			alerts = append(alerts, alert.Alert{
				LeaseID:  id,
				Level:    alert.Warning,
				Message:  fmt.Sprintf("token %s TTL dropped from %v to %v", id, prev.TTL, current),
				Metadata: map[string]string{"prev_ttl": prev.TTL.String(), "current_ttl": current.String()},
			})
		}
		s.shadow.Set(id, current)
	}
	return alerts
}
