package tokenwatch

import (
	"fmt"
	"sync"
	"time"

	"github.com/vaultwatch/internal/alert"
)

// DefaultClusterConfig returns a ClusterConfig with sensible defaults.
func DefaultClusterConfig() ClusterConfig {
	return ClusterConfig{
		MinMembers:      2,
		WarnThreshold:   0.5,
		CritThreshold:   0.25,
		MemberTTL:       5 * time.Minute,
	}
}

// ClusterConfig controls cluster membership health detection.
type ClusterConfig struct {
	MinMembers    int
	WarnThreshold float64 // fraction of MinMembers below which warning fires
	CritThreshold float64 // fraction of MinMembers below which critical fires
	MemberTTL     time.Duration
}

// ClusterDetector tracks cluster membership and detects unhealthy states.
type ClusterDetector struct {
	mu      sync.Mutex
	cfg     ClusterConfig
	members map[string]time.Time
}

// NewClusterDetector creates a ClusterDetector with the given config.
// Zero values are replaced with defaults.
func NewClusterDetector(cfg ClusterConfig) *ClusterDetector {
	d := DefaultClusterConfig()
	if cfg.MinMembers > 0 {
		d.MinMembers = cfg.MinMembers
	}
	if cfg.WarnThreshold > 0 {
		d.WarnThreshold = cfg.WarnThreshold
	}
	if cfg.CritThreshold > 0 {
		d.CritThreshold = cfg.CritThreshold
	}
	if cfg.MemberTTL > 0 {
		d.MemberTTL = cfg.MemberTTL
	}
	return &ClusterDetector{cfg: d, members: make(map[string]time.Time)}
}

// Observe registers or refreshes a cluster member.
func (c *ClusterDetector) Observe(memberID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.members[memberID] = time.Now()
}

// Check evaluates cluster health and returns an alert if degraded.
func (c *ClusterDetector) Check(tokenID string) *alert.Alert {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	active := 0
	for _, seen := range c.members {
		if now.Sub(seen) <= c.cfg.MemberTTL {
			active++
		}
	}

	min := c.cfg.MinMembers
	ratio := float64(active) / float64(min)

	var level alert.Level
	switch {
	case ratio <= c.cfg.CritThreshold:
		level = alert.LevelCritical
	case ratio <= c.cfg.WarnThreshold:
		level = alert.LevelWarning
	default:
		return nil
	}

	return &alert.Alert{
		LeaseID: tokenID,
		Level:   level,
		Message: fmt.Sprintf("cluster degraded: %d/%d active members", active, min),
	}
}
