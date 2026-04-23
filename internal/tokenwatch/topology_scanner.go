package tokenwatch

import (
	"fmt"

	"github.com/vaultwatch/internal/alert"
)

// TopologyScanner iterates registered tokens, links neighbours by shared
// labels, and emits alerts for isolated tokens.
type TopologyScanner struct {
	registry *Registry
	detector *TopologyDetector
	lookup   func(id string) (TokenInfo, error)
}

// NewTopologyScanner constructs a TopologyScanner. All arguments are
// required; the function panics on nil input.
func NewTopologyScanner(r *Registry, d *TopologyDetector, lookup func(string) (TokenInfo, error)) *TopologyScanner {
	if r == nil {
		panic("tokenwatch: TopologyScanner requires non-nil Registry")
	}
	if d == nil {
		panic("tokenwatch: TopologyScanner requires non-nil TopologyDetector")
	}
	if lookup == nil {
		panic("tokenwatch: TopologyScanner requires non-nil lookup")
	}
	return &TopologyScanner{registry: r, detector: d, lookup: lookup}
}

// Scan links tokens that share at least one label value, then checks
// each token for topology isolation. It returns one alert per isolated
// token and never returns an error for individual lookup failures.
func (s *TopologyScanner) Scan() ([]alert.Alert, error) {
	ids := s.registry.List()
	if len(ids) == 0 {
		return nil, nil
	}

	type infoEntry struct {
		id   string
		info TokenInfo
	}
	entries := make([]infoEntry, 0, len(ids))
	for _, id := range ids {
		info, err := s.lookup(id)
		if err != nil {
			continue
		}
		entries = append(entries, infoEntry{id: id, info: info})
	}

	// Build label-value → token index for O(n) linking.
	index := make(map[string][]string)
	for _, e := range entries {
		for k, v := range e.info.Labels {
			key := fmt.Sprintf("%s=%s", k, v)
			index[key] = append(index[key], e.id)
		}
	}
	for _, peers := range index {
		for i := 0; i < len(peers); i++ {
			for j := i + 1; j < len(peers); j++ {
				s.detector.Link(peers[i], peers[j])
			}
		}
	}

	var alerts []alert.Alert
	for _, e := range entries {
		if a := s.detector.Check(e.id); a != nil {
			alerts = append(alerts, *a)
		}
	}
	return alerts, nil
}
