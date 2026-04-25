package tokenwatch

import (
	"sync"
	"time"
)

// DefaultPledgeConfig returns a PledgeConfig with sensible defaults.
func DefaultPledgeConfig() PledgeConfig {
	return PledgeConfig{
		MaxAge: 24 * time.Hour,
	}
}

// PledgeConfig controls how long a token pledge is considered valid.
type PledgeConfig struct {
	MaxAge time.Duration
}

func (c *PledgeConfig) applyDefaults() {
	if c.MaxAge <= 0 {
		c.MaxAge = DefaultPledgeConfig().MaxAge
	}
}

// PledgeEntry records when a token was first pledged and its optional note.
type PledgeEntry struct {
	Token    string
	Note     string
	PledgeAt time.Time
}

// Pledge tracks tokens that have been explicitly registered for monitoring
// with an optional human-readable note. Entries expire after MaxAge.
type Pledge struct {
	mu      sync.Mutex
	cfg     PledgeConfig
	entries map[string]PledgeEntry
}

// NewPledge creates a new Pledge with the given config.
func NewPledge(cfg PledgeConfig) *Pledge {
	cfg.applyDefaults()
	return &Pledge{
		cfg:     cfg,
		entries: make(map[string]PledgeEntry),
	}
}

// Register adds or refreshes a pledge for the given token.
func (p *Pledge) Register(token, note string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.entries[token] = PledgeEntry{
		Token:    token,
		Note:     note,
		PledgeAt: time.Now(),
	}
}

// Get returns the pledge entry for a token, if it exists and has not expired.
func (p *Pledge) Get(token string) (PledgeEntry, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	e, ok := p.entries[token]
	if !ok {
		return PledgeEntry{}, false
	}
	if time.Since(e.PledgeAt) > p.cfg.MaxAge {
		delete(p.entries, token)
		return PledgeEntry{}, false
	}
	return e, true
}

// Revoke removes a pledge entry for the given token.
func (p *Pledge) Revoke(token string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.entries, token)
}

// Len returns the number of active (non-expired) pledge entries.
func (p *Pledge) Len() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	now := time.Now()
	count := 0
	for k, e := range p.entries {
		if now.Sub(e.PledgeAt) > p.cfg.MaxAge {
			delete(p.entries, k)
			continue
		}
		count++
	}
	return count
}
