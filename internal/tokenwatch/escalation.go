package tokenwatch

import (
	"sync"
	"time"

	"github.com/vaultwatch/internal/alert"
)

// DefaultEscalationConfig returns sensible escalation defaults.
func DefaultEscalationConfig() EscalationConfig {
	return EscalationConfig{
		Window:       5 * time.Minute,
		Threshold:    3,
		EscalateLevel: alert.LevelCritical,
	}
}

// EscalationConfig controls when repeated alerts are escalated.
type EscalationConfig struct {
	Window        time.Duration
	Threshold     int
	EscalateLevel alert.Level
}

type escalationEntry struct {
	count int
	first time.Time
}

// Escalation tracks repeated alerts per token and escalates level when
// the same token fires more than Threshold times within Window.
type Escalation struct {
	mu      sync.Mutex
	cfg     EscalationConfig
	entries map[string]*escalationEntry
	now     func() time.Time
}

// NewEscalation constructs an Escalation with the given config.
func NewEscalation(cfg EscalationConfig) *Escalation {
	if cfg.Window <= 0 {
		cfg.Window = DefaultEscalationConfig().Window
	}
	if cfg.Threshold <= 0 {
		cfg.Threshold = DefaultEscalationConfig().Threshold
	}
	return &Escalation{
		cfg:     cfg,
		entries: make(map[string]*escalationEntry),
		now:     time.Now,
	}
}

// Check records an alert occurrence and returns the (possibly escalated) level.
func (e *Escalation) Check(tokenID string, level alert.Level) alert.Level {
	e.mu.Lock()
	defer e.mu.Unlock()

	now := e.now()
	ent, ok := e.entries[tokenID]
	if !ok || now.Sub(ent.first) > e.cfg.Window {
		e.entries[tokenID] = &escalationEntry{count: 1, first: now}
		return level
	}
	ent.count++
	if ent.count >= e.cfg.Threshold && level < e.cfg.EscalateLevel {
		return e.cfg.EscalateLevel
	}
	return level
}

// Reset clears the escalation state for a token.
func (e *Escalation) Reset(tokenID string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.entries, tokenID)
}
