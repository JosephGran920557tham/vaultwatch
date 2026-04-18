package tokenwatch

import (
	"sync"
	"time"
)

// DefaultLedgerConfig returns sensible defaults for the Ledger.
func DefaultLedgerConfig() LedgerConfig {
	return LedgerConfig{
		MaxEntries: 1000,
		TTL:        30 * time.Minute,
	}
}

// LedgerConfig controls retention behaviour of the Ledger.
type LedgerConfig struct {
	MaxEntries int
	TTL        time.Duration
}

// LedgerEntry records a single observation for a token.
type LedgerEntry struct {
	TokenID   string
	Timestamp time.Time
	TTL       time.Duration
	Level     string
}

// Ledger stores a bounded, time-limited history of token observations.
type Ledger struct {
	mu      sync.Mutex
	cfg     LedgerConfig
	entries []LedgerEntry
}

// NewLedger creates a Ledger with the supplied config.
// Zero-value fields are replaced with defaults.
func NewLedger(cfg LedgerConfig) *Ledger {
	def := DefaultLedgerConfig()
	if cfg.MaxEntries <= 0 {
		cfg.MaxEntries = def.MaxEntries
	}
	if cfg.TTL <= 0 {
		cfg.TTL = def.TTL
	}
	return &Ledger{cfg: cfg}
}

// Record appends an entry, evicting expired and excess entries first.
func (l *Ledger) Record(e LedgerEntry) {
	if e.Timestamp.IsZero() {
		e.Timestamp = time.Now()
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.evict()
	if len(l.entries) >= l.cfg.MaxEntries {
		l.entries = l.entries[1:]
	}
	l.entries = append(l.entries, e)
}

// List returns all non-expired entries for the given token ID.
func (l *Ledger) List(tokenID string) []LedgerEntry {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.evict()
	var out []LedgerEntry
	for _, e := range l.entries {
		if e.TokenID == tokenID {
			out = append(out, e)
		}
	}
	return out
}

// Len returns the total number of live entries across all tokens.
func (l *Ledger) Len() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.evict()
	return len(l.entries)
}

func (l *Ledger) evict() {
	cutoff := time.Now().Add(-l.cfg.TTL)
	keep := l.entries[:0]
	for _, e := range l.entries {
		if e.Timestamp.After(cutoff) {
			keep = append(keep, e)
		}
	}
	l.entries = keep
}
