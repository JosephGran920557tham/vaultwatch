package tokenwatch

import "fmt"

// DecayScanner iterates over all registered tokens, records current decay, and
// returns alerts for tokens whose normalised TTL score is below threshold.
type DecayScanner struct {
	registry *Registry
	detector *DecayDetector
	lookup   func(tokenID string) (TokenInfo, error)
}

// NewDecayScanner constructs a DecayScanner. All arguments are required.
func NewDecayScanner(r *Registry, d *DecayDetector, lookup func(string) (TokenInfo, error)) *DecayScanner {
	if r == nil {
		panic("tokenwatch: DecayScanner requires non-nil Registry")
	}
	if d == nil {
		panic("tokenwatch: DecayScanner requires non-nil DecayDetector")
	}
	if lookup == nil {
		panic("tokenwatch: DecayScanner requires non-nil lookup func")
	}
	return &DecayScanner{registry: r, detector: d, lookup: lookup}
}

// Scan checks all registered tokens and returns any decay alerts.
func (s *DecayScanner) Scan() []Alert {
	tokens := s.registry.List()
	var alerts []Alert
	for _, id := range tokens {
		info, err := s.lookup(id)
		if err != nil {
			continue
		}
		if info.InitialTTL <= 0 {
			continue
		}
		s.detector.Record(id, info.TTL, info.InitialTTL)
		if a := s.detector.Check(id); a != nil {
			a.Labels["detail"] = fmt.Sprintf("score=%.2f", float64(info.TTL)/float64(info.InitialTTL))
			alerts = append(alerts, *a)
		}
	}
	return alerts
}
