package tokenwatch

import (
	"context"
	"log"
	"time"
)

// EvictionRunner periodically sweeps an Eviction and removes stale tokens
// from a Registry.
type EvictionRunner struct {
	eviction *Eviction
	registry *Registry
	logger   *log.Logger
}

// NewEvictionRunner creates an EvictionRunner.
// Panics if eviction or registry is nil.
func NewEvictionRunner(ev *Eviction, reg *Registry, logger *log.Logger) *EvictionRunner {
	if ev == nil {
		panic("eviction must not be nil")
	}
	if reg == nil {
		panic("registry must not be nil")
	}
	if logger == nil {
		logger = log.Default()
	}
	return &EvictionRunner{eviction: ev, registry: reg, logger: logger}
}

// Run starts the sweep loop, blocking until ctx is cancelled.
func (r *EvictionRunner) Run(ctx context.Context) {
	ticker := time.NewTicker(r.eviction.cfg.SweepInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.sweep()
		}
	}
}

func (r *EvictionRunner) sweep() {
	evicted := r.eviction.icted {
		if err := r.registry.Remove(token); err != nil {
			riction: failedv", token, err)
		} else {
			r.logger.Printf("eviction: removed stale token %s", token)
		}
	}
}
