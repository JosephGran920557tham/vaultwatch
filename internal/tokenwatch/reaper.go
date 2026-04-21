package tokenwatch

import (
	"context"
	"log"
	"sync"
	"time"
)

// DefaultReaperConfig returns a ReaperConfig with sensible defaults.
func DefaultReaperConfig() ReaperConfig {
	return ReaperConfig{
		Interval:   5 * time.Minute,
		MaxAge:     30 * time.Minute,
		BatchSize:  50,
	}
}

// ReaperConfig controls how the Reaper scans and removes expired token entries.
type ReaperConfig struct {
	Interval  time.Duration
	MaxAge    time.Duration
	BatchSize int
}

// ReaperEntry represents a tracked token with its last-seen timestamp.
type ReaperEntry struct {
	TokenID  string
	LastSeen time.Time
}

// ReaperStore is the interface the Reaper uses to list and evict tokens.
type ReaperStore interface {
	List() []string
	LastSeen(tokenID string) (time.Time, bool)
	Evict(tokenID string)
}

// Reaper periodically scans a ReaperStore and removes tokens that have not
// been seen within the configured MaxAge window.
type Reaper struct {
	cfg    ReaperConfig
	store  ReaperStore
	logger *log.Logger
	mu     sync.Mutex
	reaped []string
}

// NewReaper creates a Reaper. It panics if store is nil.
func NewReaper(cfg ReaperConfig, store ReaperStore, logger *log.Logger) *Reaper {
	if store == nil {
		panic("reaper: store must not be nil")
	}
	if cfg.Interval <= 0 {
		cfg.Interval = DefaultReaperConfig().Interval
	}
	if cfg.MaxAge <= 0 {
		cfg.MaxAge = DefaultReaperConfig().MaxAge
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = DefaultReaperConfig().BatchSize
	}
	if logger == nil {
		logger = log.Default()
	}
	return &Reaper{cfg: cfg, store: store, logger: logger}
}

// Run starts the reaper loop, ticking at cfg.Interval until ctx is cancelled.
func (r *Reaper) Run(ctx context.Context) {
	ticker := time.NewTicker(r.cfg.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.Reap()
		}
	}
}

// Reap performs a single reap pass, evicting tokens older than MaxAge.
func (r *Reaper) Reap() []string {
	tokens := r.store.List()
	cutoff := time.Now().Add(-r.cfg.MaxAge)
	evicted := make([]string, 0)
	for i := 0; i < len(tokens) && len(evicted) < r.cfg.BatchSize; i++ {
		id := tokens[i]
		if ts, ok := r.store.LastSeen(id); ok && ts.Before(cutoff) {
			r.store.Evict(id)
			evicted = append(evicted, id)
			r.logger.Printf("reaper: evicted token %s (last seen %s)", id, ts.Format(time.RFC3339))
		}
	}
	r.mu.Lock()
	r.reaped = append(r.reaped, evicted...)
	r.mu.Unlock()
	return evicted
}

// TotalReaped returns the cumulative count of evicted tokens since creation.
func (r *Reaper) TotalReaped() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.reaped)
}
