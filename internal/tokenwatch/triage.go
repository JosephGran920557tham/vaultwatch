package tokenwatch

import (
	"sort"
	"time"

	"github.com/your-org/vaultwatch/internal/alert"
)

// DefaultTriageConfig returns sensible defaults for the triage scorer.
func DefaultTriageConfig() TriageConfig {
	return TriageConfig{
		CriticalWeight: 10,
		WarningWeight:  3,
		InfoWeight:     1,
		RecencyHalfLife: 5 * time.Minute,
	}
}

// TriageConfig controls how alerts are scored during triage.
type TriageConfig struct {
	CriticalWeight  float64
	WarningWeight   float64
	InfoWeight      float64
	RecencyHalfLife time.Duration
}

// TriageEntry holds a scored alert.
type TriageEntry struct {
	Alert alert.Alert
	Score float64
}

// Triage scores and sorts alerts by urgency, applying recency decay.
type Triage struct {
	cfg TriageConfig
}

// NewTriage constructs a Triage scorer with the given config.
// Zero-value fields are replaced with defaults.
func NewTriage(cfg TriageConfig) *Triage {
	def := DefaultTriageConfig()
	if cfg.CriticalWeight <= 0 {
		cfg.CriticalWeight = def.CriticalWeight
	}
	if cfg.WarningWeight <= 0 {
		cfg.WarningWeight = def.WarningWeight
	}
	if cfg.InfoWeight <= 0 {
		cfg.InfoWeight = def.InfoWeight
	}
	if cfg.RecencyHalfLife <= 0 {
		cfg.RecencyHalfLife = def.RecencyHalfLife
	}
	return &Triage{cfg: cfg}
}

// Score returns a numeric urgency score for a single alert.
func (t *Triage) Score(a alert.Alert, now time.Time) float64 {
	var base float64
	switch a.Level {
	case alert.Critical:
		base = t.cfg.CriticalWeight
	case alert.Warning:
		base = t.cfg.WarningWeight
	default:
		base = t.cfg.InfoWeight
	}
	age := now.Sub(a.FiredAt)
	if age < 0 {
		age = 0
	}
	// exponential decay: score halves every RecencyHalfLife
	decay := 1.0
	if t.cfg.RecencyHalfLife > 0 {
		halves := float64(age) / float64(t.cfg.RecencyHalfLife)
		decay = 1.0 / (1.0 + halves)
	}
	return base * decay
}

// Rank scores all alerts and returns them sorted descending by score.
func (t *Triage) Rank(alerts []alert.Alert, now time.Time) []TriageEntry {
	entries := make([]TriageEntry, len(alerts))
	for i, a := range alerts {
		entries[i] = TriageEntry{Alert: a, Score: t.Score(a, now)}
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Score > entries[j].Score
	})
	return entries
}
