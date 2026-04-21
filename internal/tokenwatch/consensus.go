package tokenwatch

import (
	"fmt"
	"sync"
	"time"

	"github.com/vaultwatch/internal/alert"
)

// DefaultConsensusConfig returns a ConsensusConfig with sensible defaults.
func DefaultConsensusConfig() ConsensusConfig {
	return ConsensusConfig{
		Quorum:  2,
		Window:  5 * time.Minute,
		MaxKeys: 1000,
	}
}

// ConsensusConfig controls quorum-based alert suppression.
type ConsensusConfig struct {
	// Quorum is the minimum number of independent sources that must agree
	// before an alert is forwarded.
	Quorum int
	// Window is how long votes are retained before expiry.
	Window time.Duration
	// MaxKeys caps the number of distinct alert keys tracked.
	MaxKeys int
}

type vote struct {
	sources map[string]time.Time
}

// Consensus suppresses alerts until a quorum of distinct sources agree.
type Consensus struct {
	cfg   ConsensusConfig
	mu    sync.Mutex
	votes map[string]*vote
	now   func() time.Time
}

// NewConsensus creates a Consensus with the given config.
// Zero values are replaced with defaults.
func NewConsensus(cfg ConsensusConfig) *Consensus {
	def := DefaultConsensusConfig()
	if cfg.Quorum <= 0 {
		cfg.Quorum = def.Quorum
	}
	if cfg.Window <= 0 {
		cfg.Window = def.Window
	}
	if cfg.MaxKeys <= 0 {
		cfg.MaxKeys = def.MaxKeys
	}
	return &Consensus{
		cfg:   cfg,
		votes: make(map[string]*vote),
		now:   time.Now,
	}
}

// Vote records a vote from source for the given alert.
// It returns true when the quorum threshold has been reached.
func (c *Consensus) Vote(source string, a alert.Alert) bool {
	key := fmt.Sprintf("%s:%s", a.LeaseID, a.Level)
	c.mu.Lock()
	defer c.mu.Unlock()

	now := c.now()
	c.evict(now)

	v, ok := c.votes[key]
	if !ok {
		if len(c.votes) >= c.cfg.MaxKeys {
			return false
		}
		v = &vote{sources: make(map[string]time.Time)}
		c.votes[key] = v
	}
	v.sources[source] = now
	return len(v.sources) >= c.cfg.Quorum
}

// evict removes expired votes. Must be called with c.mu held.
func (c *Consensus) evict(now time.Time) {
	cutoff := now.Add(-c.cfg.Window)
	for key, v := range c.votes {
		for src, ts := range v.sources {
			if ts.Before(cutoff) {
				delete(v.sources, src)
			}
		}
		if len(v.sources) == 0 {
			delete(c.votes, key)
		}
	}
}
