package tokenwatch

import (
	"fmt"
	"time"

	"github.com/vaultwatch/internal/alert"
)

// HedgeScanner uses hedged requests to look up token info and produce alerts
// when the underlying lookup is slow or flaky.
type HedgeScanner struct {
	registry *Registry
	lookup   func(tokenID string) (TokenInfo, error)
	cfg      HedgeConfig
}

// NewHedgeScanner creates a HedgeScanner. Panics if registry or lookup is nil.
func NewHedgeScanner(registry *Registry, lookup func(string) (TokenInfo, error), cfg HedgeConfig) *HedgeScanner {
	if registry == nil {
		panic("hedge scanner: registry must not be nil")
	}
	if lookup == nil {
		panic("hedge scanner: lookup must not be nil")
	}
	if cfg.Delay <= 0 {
		cfg = DefaultHedgeConfig()
	}
	return &HedgeScanner{registry: registry, lookup: lookup, cfg: cfg}
}

// Scan iterates all registered tokens, fetches each via a hedged call, and
// returns one alert per token that could not be resolved within the hedge window.
func (s *HedgeScanner) Scan() []alert.Alert {
	tokens := s.registry.List()
	var alerts []alert.Alert
	for _, id := range tokens {
		_, err := Hedge(s.cfg, func() (interface{}, error) {
			return s.lookup(id)
		})
		if err != nil {
			alerts = append(alerts, alert.Alert{
				LeaseID:   id,
				Level:     alert.LevelWarning,
				Message:   fmt.Sprintf("hedge lookup failed for token %s: %v", id, err),
				ExpiresAt: time.Now(),
			})
		}
	}
	return alerts
}
