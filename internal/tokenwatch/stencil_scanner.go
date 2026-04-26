package tokenwatch

import (
	"fmt"

	"github.com/vaultwatch/internal/alert"
)

// StencilScanner scans the registry and, for each token whose stencil template
// is present in the Stencil store, emits an informational alert rendered from
// that template.
type StencilScanner struct {
	registry *Registry
	stencil  *Stencil
	lookup   func(tokenID string) (TokenInfo, error)
}

// NewStencilScanner creates a StencilScanner. All arguments are required.
func NewStencilScanner(registry *Registry, stencil *Stencil, lookup func(string) (TokenInfo, error)) *StencilScanner {
	if registry == nil {
		panic("stencil scanner: registry must not be nil")
	}
	if stencil == nil {
		panic("stencil scanner: stencil must not be nil")
	}
	if lookup == nil {
		panic("stencil scanner: lookup must not be nil")
	}
	return &StencilScanner{registry: registry, stencil: stencil, lookup: lookup}
}

// Scan iterates over all registered tokens. For each token that has a stencil
// template stored, it emits an alert whose message is the rendered template
// (with the token ID substituted for the placeholder "TOKEN").
func (sc *StencilScanner) Scan() []alert.Alert {
	tokens := sc.registry.List()
	var alerts []alert.Alert
	for _, id := range tokens {
		tmpl, ok := sc.stencil.Get(id)
		if !ok {
			continue
		}
		_, err := sc.lookup(id)
		if err != nil {
			continue
		}
		msg := fmt.Sprintf(tmpl, id)
		alerts = append(alerts, alert.Alert{
			LeaseID: id,
			Level:   alert.LevelInfo,
			Message: msg,
			Labels:  map[string]string{"source": "stencil"},
		})
	}
	return alerts
}
