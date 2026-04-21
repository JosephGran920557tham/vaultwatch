package tokenwatch

import (
	"github.com/vaultwatch/internal/alert"
)

// QuorumScanner iterates the registry, casts votes, and emits alerts for tokens
// that fail to reach quorum.
type QuorumScanner struct {
	registry *Registry
	detector *QuorumDetector
	lookup   func(tokenID string) (TokenInfo, error)
}

// NewQuorumScanner constructs a QuorumScanner. All arguments are required.
func NewQuorumScanner(registry *Registry, detector *QuorumDetector, lookup func(string) (TokenInfo, error)) *QuorumScanner {
	if registry == nil {
		panic("quorum scanner: registry must not be nil")
	}
	if detector == nil {
		panic("quorum scanner: detector must not be nil")
	}
	if lookup == nil {
		panic("quorum scanner: lookup must not be nil")
	}
	return &QuorumScanner{registry: registry, detector: detector, lookup: lookup}
}

// Scan checks every registered token and returns any quorum alerts.
func (s *QuorumScanner) Scan() []alert.Alert {
	tokens := s.registry.List()
	var alerts []alert.Alert
	for _, id := range tokens {
		info, err := s.lookup(id)
		if err != nil {
			continue
		}
		healthy := info.TTL > 0
		s.detector.Vote(id, healthy)
		if a := s.detector.Check(id); a != nil {
			alerts = append(alerts, *a)
		}
	}
	return alerts
}
