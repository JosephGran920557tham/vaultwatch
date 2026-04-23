package tokenwatch

import (
	"fmt"
	"sync"
	"time"

	"github.com/vaultwatch/internal/alert"
)

// DefaultTopologyConfig returns sensible defaults for TopologyDetector.
func DefaultTopologyConfig() TopologyConfig {
	return TopologyConfig{
		MaxNeighbors: 32,
		StaleAfter:   5 * time.Minute,
		MinLinks:     2,
	}
}

// TopologyConfig controls topology-aware detection behaviour.
type TopologyConfig struct {
	MaxNeighbors int
	StaleAfter   time.Duration
	MinLinks     int
}

type topologyEntry struct {
	neighbors map[string]time.Time
}

// TopologyDetector tracks token co-occurrence relationships and alerts
// when a token becomes isolated (too few active neighbours).
type TopologyDetector struct {
	mu      sync.Mutex
	cfg     TopologyConfig
	nodes   map[string]*topologyEntry
	now     func() time.Time
}

// NewTopologyDetector constructs a TopologyDetector. Zero-value config
// fields are replaced with defaults.
func NewTopologyDetector(cfg TopologyConfig) *TopologyDetector {
	def := DefaultTopologyConfig()
	if cfg.MaxNeighbors <= 0 {
		cfg.MaxNeighbors = def.MaxNeighbors
	}
	if cfg.StaleAfter <= 0 {
		cfg.StaleAfter = def.StaleAfter
	}
	if cfg.MinLinks <= 0 {
		cfg.MinLinks = def.MinLinks
	}
	return &TopologyDetector{
		cfg:   cfg,
		nodes: make(map[string]*topologyEntry),
		now:   time.Now,
	}
}

// Link records a co-occurrence between tokenA and tokenB.
func (t *TopologyDetector) Link(tokenA, tokenB string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.ensureNode(tokenA).neighbors[tokenB] = t.now()
	t.ensureNode(tokenB).neighbors[tokenA] = t.now()
}

func (t *TopologyDetector) ensureNode(id string) *topologyEntry {
	if _, ok := t.nodes[id]; !ok {
		t.nodes[id] = &topologyEntry{neighbors: make(map[string]time.Time)}
	}
	return t.nodes[id]
}

// Check returns a warning alert when tokenID has fewer active links than
// MinLinks, or nil when the topology looks healthy.
func (t *TopologyDetector) Check(tokenID string) *alert.Alert {
	t.mu.Lock()
	defer t.mu.Unlock()
	node, ok := t.nodes[tokenID]
	if !ok {
		return nil
	}
	cutoff := t.now().Add(-t.cfg.StaleAfter)
	active := 0
	for _, seen := range node.neighbors {
		if seen.After(cutoff) {
			active++
		}
	}
	if active < t.cfg.MinLinks {
		return &alert.Alert{
			LeaseID:  tokenID,
			Level:    alert.LevelWarning,
			Message:  fmt.Sprintf("token %s has only %d active topology link(s) (min %d)", tokenID, active, t.cfg.MinLinks),
		}
	}
	return nil
}
