package tokenwatch

import (
	"sync"
	"time"
)

// DefaultCohortConfig returns a CohortConfig with sensible defaults.
func DefaultCohortConfig() CohortConfig {
	return CohortConfig{
		MaxAge:   30 * time.Minute,
		MaxSize:  256,
		GroupKey: "env",
	}
}

// CohortConfig controls cohort grouping behaviour.
type CohortConfig struct {
	MaxAge   time.Duration
	MaxSize  int
	GroupKey string
}

// cohortEntry holds a token ID and the time it was admitted.
type cohortEntry struct {
	tokenID  string
	admitted time.Time
}

// Cohort groups token IDs by a label value and tracks membership age.
type Cohort struct {
	mu      sync.Mutex
	cfg     CohortConfig
	groups  map[string][]cohortEntry
	nowFunc func() time.Time
}

// NewCohort creates a Cohort. Zero-value fields in cfg are replaced with defaults.
func NewCohort(cfg CohortConfig) *Cohort {
	def := DefaultCohortConfig()
	if cfg.MaxAge <= 0 {
		cfg.MaxAge = def.MaxAge
	}
	if cfg.MaxSize <= 0 {
		cfg.MaxSize = def.MaxSize
	}
	if cfg.GroupKey == "" {
		cfg.GroupKey = def.GroupKey
	}
	return &Cohort{
		cfg:     cfg,
		groups:  make(map[string][]cohortEntry),
		nowFunc: time.Now,
	}
}

// Add places tokenID into the group identified by groupValue.
func (c *Cohort) Add(groupValue, tokenID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.evict(groupValue)
	entries := c.groups[groupValue]
	if len(entries) >= c.cfg.MaxSize {
		entries = entries[1:]
	}
	c.groups[groupValue] = append(entries, cohortEntry{tokenID: tokenID, admitted: c.nowFunc()})
}

// Members returns the current (non-expired) token IDs in the given group.
func (c *Cohort) Members(groupValue string) []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.evict(groupValue)
	entries := c.groups[groupValue]
	out := make([]string, len(entries))
	for i, e := range entries {
		out[i] = e.tokenID
	}
	return out
}

// Groups returns all group keys that have at least one live member.
func (c *Cohort) Groups() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	keys := make([]string, 0, len(c.groups))
	for k, entries := range c.groups {
		if len(entries) > 0 {
			keys = append(keys, k)
		}
	}
	return keys
}

// evict removes expired entries from a single group. Caller must hold mu.
func (c *Cohort) evict(groupValue string) {
	now := c.nowFunc()
	entries := c.groups[groupValue]
	var live []cohortEntry
	for _, e := range entries {
		if now.Sub(e.admitted) < c.cfg.MaxAge {
			live = append(live, e)
		}
	}
	if len(live) == 0 {
		delete(c.groups, groupValue)
	} else {
		c.groups[groupValue] = live
	}
}
