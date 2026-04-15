package tokenwatch

import (
	"log"

	"github.com/vaultwatch/internal/alert"
)

// ForecastScanner scans all registered tokens and emits forecast alerts.
type ForecastScanner struct {
	registry  *Registry
	detectors map[string]*ForecastDetector
	cfg       ForecastConfig
	logger    *log.Logger
}

// NewForecastScanner creates a ForecastScanner backed by the given registry.
// Panics if registry is nil.
func NewForecastScanner(registry *Registry, cfg ForecastConfig, logger *log.Logger) *ForecastScanner {
	if registry == nil {
		panic("tokenwatch: ForecastScanner requires a non-nil registry")
	}
	if logger == nil {
		logger = log.Default()
	}
	return &ForecastScanner{
		registry:  registry,
		detectors: make(map[string]*ForecastDetector),
		cfg:       cfg,
		logger:    logger,
	}
}

// Scan records the current TTL for each token and returns any forecast alerts.
func (s *ForecastScanner) Scan(lookup func(tokenID string) (TokenInfo, error)) []alert.Alert {
	tokens := s.registry.List()
	var alerts []alert.Alert
	for _, id := range tokens {
		info, err := lookup(id)
		if err != nil {
			s.logger.Printf("forecast: lookup error for token %s: %v", id, err)
			continue
		}
		det := s.detectorFor(id)
		det.Record(info.TTL)
		if a := det.Check(id); a != nil {
			alerts = append(alerts, *a)
		}
	}
	return alerts
}

func (s *ForecastScanner) detectorFor(id string) *ForecastDetector {
	if d, ok := s.detectors[id]; ok {
		return d
	}
	d := NewForecastDetector(s.cfg)
	s.detectors[id] = d
	return d
}
