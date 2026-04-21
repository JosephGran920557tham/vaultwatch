package tokenwatch

import (
	"fmt"
	"sync"
	"time"

	"github.com/vaultwatch/internal/alert"
)

// DefaultQuorumConfig returns sensible defaults for QuorumDetector.
func DefaultQuorumConfig() QuorumConfig {
	return QuorumConfig{
		MinVoters:   3,
		Threshold:   0.6,
		WindowSize:  5 * time.Minute,
		CriticalMin: 0.9,
	}
}

// QuorumConfig controls quorum detection behaviour.
type QuorumConfig struct {
	MinVoters   int
	Threshold   float64
	WindowSize  time.Duration
	CriticalMin float64
}

// QuorumDetector raises an alert when a token's health votes fail to reach quorum.
type QuorumDetector struct {
	cfg  QuorumConfig
	mu   sync.Mutex
	votes map[string][]quorumVote
}

type quorumVote struct {
	healthy bool
	at      time.Time
}

// NewQuorumDetector constructs a QuorumDetector, applying defaults for zero values.
func NewQuorumDetector(cfg QuorumConfig) *QuorumDetector {
	def := DefaultQuorumConfig()
	if cfg.MinVoters <= 0 {
		cfg.MinVoters = def.MinVoters
	}
	if cfg.Threshold <= 0 {
		cfg.Threshold = def.Threshold
	}
	if cfg.WindowSize <= 0 {
		cfg.WindowSize = def.WindowSize
	}
	if cfg.CriticalMin <= 0 {
		cfg.CriticalMin = def.CriticalMin
	}
	return &QuorumDetector{cfg: cfg, votes: make(map[string][]quorumVote)}
}

// Vote records a health vote for the given token ID.
func (q *QuorumDetector) Vote(tokenID string, healthy bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	cutoff := time.Now().Add(-q.cfg.WindowSize)
	votes := q.votes[tokenID]
	filtered := votes[:0]
	for _, v := range votes {
		if v.at.After(cutoff) {
			filtered = append(filtered, v)
		}
	}
	filtered = append(filtered, quorumVote{healthy: healthy, at: time.Now()})
	q.votes[tokenID] = filtered
}

// Check returns an alert if the token's recent votes fail to reach quorum.
func (q *QuorumDetector) Check(tokenID string) *alert.Alert {
	q.mu.Lock()
	defer q.mu.Unlock()
	votes := q.votes[tokenID]
	if len(votes) < q.cfg.MinVoters {
		return nil
	}
	healthy := 0
	for _, v := range votes {
		if v.healthy {
			healthy++
		}
	}
	ratio := float64(healthy) / float64(len(votes))
	if ratio >= q.cfg.Threshold {
		return nil
	}
	lvl := alert.LevelWarning
	if ratio < (1.0 - q.cfg.CriticalMin) {
		lvl = alert.LevelCritical
	}
	return &alert.Alert{
		LeaseID: tokenID,
		Level:   lvl,
		Message: fmt.Sprintf("quorum not reached: %.0f%% healthy (%d/%d votes)", ratio*100, healthy, len(votes)),
		Labels:  map[string]string{"detector": "quorum"},
	}
}
