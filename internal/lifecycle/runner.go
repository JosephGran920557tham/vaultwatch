package lifecycle

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// Runner ties a Manager to OS signal handling, starting services and
// blocking until SIGINT or SIGTERM is received before draining.
type Runner struct {
	mgr     *Manager
	signals []os.Signal
}

// NewRunner returns a Runner backed by the given Manager.
// If no signals are provided, SIGINT and SIGTERM are used.
func NewRunner(mgr *Manager, sigs ...os.Signal) *Runner {
	if len(sigs) == 0 {
		sigs = []os.Signal{syscall.SIGINT, syscall.SIGTERM}
	}
	return &Runner{mgr: mgr, signals: sigs}
}

// Run starts the manager, blocks on signals, then stops the manager.
// The provided context is passed to start hooks.
func (r *Runner) Run(ctx context.Context) error {
	if err := r.mgr.Start(ctx); err != nil {
		return err
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, r.signals...)
	defer signal.Stop(quit)

	select {
	case <-quit:
	case <-ctx.Done():
	}

	return r.mgr.Stop()
}
