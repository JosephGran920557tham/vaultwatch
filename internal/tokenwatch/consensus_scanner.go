package tokenwatch

import (
	"fmt"

	"github.com/vaultwatch/internal/alert"
)

// ConsensusScanner wraps multiple alert sources and only forwards alerts
// that reach quorum across those sources.
type ConsensusScanner struct {
	consensus *Consensus
	sources   []AlertSource
}

// AlertSource is anything that can produce a slice of alerts for a token.
type AlertSource interface {
	Scan(tokenID string) ([]alert.Alert, error)
}

// NewConsensusScanner creates a ConsensusScanner.
// Panics if consensus or sources are nil/empty.
func NewConsensusScanner(c *Consensus, sources ...AlertSource) *ConsensusScanner {
	if c == nil {
		panic("tokenwatch: ConsensusScanner requires non-nil Consensus")
	}
	if len(sources) == 0 {
		panic("tokenwatch: ConsensusScanner requires at least one AlertSource")
	}
	return &ConsensusScanner{consensus: c, sources: sources}
}

// Scan collects alerts from all sources and returns only those that
// have reached the configured quorum.
func (cs *ConsensusScanner) Scan(tokenID string) ([]alert.Alert, error) {
	var agreed []alert.Alert
	seen := make(map[string]bool)

	for i, src := range cs.sources {
		alerts, err := src.Scan(tokenID)
		if err != nil {
			// Non-fatal: skip this source.
			continue
		}
		sourceName := fmt.Sprintf("source-%d", i)
		for _, a := range alerts {
			key := fmt.Sprintf("%s:%s", a.LeaseID, a.Level)
			if cs.consensus.Vote(sourceName, a) && !seen[key] {
				seen[key] = true
				agreed = append(agreed, a)
			}
		}
	}
	return agreed, nil
}
