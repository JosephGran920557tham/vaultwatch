package tokenwatch

import (
	"fmt"

	"github.com/your-org/vaultwatch/internal/alert"
)

// CapacityScanner checks registry utilisation and emits an alert when
// thresholds are breached.
type CapacityScanner struct {
	registry *Registry
	detector *CapacityDetector
}

// NewCapacityScanner creates a CapacityScanner. Panics on nil arguments.
func NewCapacityScanner(registry *Registry, detector *CapacityDetector) *CapacityScanner {
	if registry == nil {
		panic("tokenwatch: CapacityScanner requires non-nil registry")
	}
	if detector == nil {
		panic("tokenwatch: CapacityScanner requires non-nil detector")
	}
	return &CapacityScanner{registry: registry, detector: detector}
}

// Scan evaluates current registry size and returns an alert if a threshold is
// breached, or nil when utilisation is within acceptable bounds.
func (s *CapacityScanner) Scan() *alert.Alert {
	tokens := s.registry.List()
	res := s.detector.Check(len(tokens))

	if res.Level == "ok" {
		return nil
	}

	lvl := alert.LevelWarning
	if res.Level == "critical" {
		lvl = alert.LevelCritical
	}

	return &alert.Alert{
		LeaseID: "registry/capacity",
		Level:   lvl,
		Message: fmt.Sprintf(
			"token registry at %.0f%% capacity (%d tokens)",
			res.Utilisation*100, res.Count,
		),
	}
}
