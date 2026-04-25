package tokenwatch

import (
	"fmt"
	"sync"
	"time"

	"github.com/vaultwatch/internal/alert"
)

// DefaultSymmetryConfig returns a SymmetryConfig with sensible defaults.
func DefaultSymmetryConfig() SymmetryConfig {
	return SymmetryConfig{
		MaxSkew:  30 * time.Second,
		MinPeers: 2,
	}
}

// SymmetryConfig controls how the SymmetryDetector behaves.
type SymmetryConfig struct {
	// MaxSkew is the maximum acceptable TTL difference between peers.
	MaxSkew time.Duration
	// MinPeers is the minimum number of peers required to evaluate symmetry.
	MinPeers int
}

// SymmetryDetector detects when a token's TTL diverges significantly from
// its peer group, which may indicate a misconfiguration or renewal issue.
type SymmetryDetector struct {
	cfg   SymmetryConfig
	mu    sync.Mutex
	peers map[string]time.Duration // tokenID -> last observed TTL
}

// NewSymmetryDetector creates a SymmetryDetector with the given config.
// Zero values are replaced with defaults.
func NewSymmetryDetector(cfg SymmetryConfig) *SymmetryDetector {
	def := DefaultSymmetryConfig()
	if cfg.MaxSkew <= 0 {
		cfg.MaxSkew = def.MaxSkew
	}
	if cfg.MinPeers <= 0 {
		cfg.MinPeers = def.MinPeers
	}
	return &SymmetryDetector{
		cfg:   cfg,
		peers: make(map[string]time.Duration),
	}
}

// Observe records the current TTL for a token.
func (d *SymmetryDetector) Observe(tokenID string, ttl time.Duration) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.peers[tokenID] = ttl
}

// Check evaluates whether tokenID's TTL is symmetric with its peers.
// Returns nil if there are insufficient peers or TTL is within MaxSkew.
func (d *SymmetryDetector) Check(tokenID string) *alert.Alert {
	d.mu.Lock()
	defer d.mu.Unlock()

	ttl, ok := d.peers[tokenID]
	if !ok || len(d.peers) < d.cfg.MinPeers {
		return nil
	}

	var sum time.Duration
	count := 0
	for id, t := range d.peers {
		if id != tokenID {
			sum += t
			count++
		}
	}
	if count == 0 {
		return nil
	}

	avg := sum / time.Duration(count)
	skew := ttl - avg
	if skew < 0 {
		skew = -skew
	}
	if skew <= d.cfg.MaxSkew {
		return nil
	}

	return &alert.Alert{
		LeaseID: tokenID,
		Level:   alert.LevelWarning,
		Message: fmt.Sprintf("token TTL skew %.1fs exceeds max %.1fs vs peer avg",
			skew.Seconds(), d.cfg.MaxSkew.Seconds()),
		Labels:  map[string]string{"detector": "symmetry"},
	}
}
