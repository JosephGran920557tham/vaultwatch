// Package lifecycle manages the startup and graceful shutdown of vaultwatch
// background services, coordinating context cancellation and drain timeouts.
package lifecycle

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Hook is a function invoked during a lifecycle phase.
type Hook func(ctx context.Context) error

// Manager coordinates startup and shutdown hooks for registered services.
type Manager struct {
	mu           sync.Mutex
	startHooks   []namedHook
	stopHooks    []namedHook
	drainTimeout time.Duration
}

type namedHook struct {
	name string
	fn   Hook
}

// New returns a Manager with the given drain timeout applied during shutdown.
func New(drainTimeout time.Duration) *Manager {
	if drainTimeout <= 0 {
		drainTimeout = 10 * time.Second
	}
	return &Manager{drainTimeout: drainTimeout}
}

// OnStart registers a hook to be called during Start.
func (m *Manager) OnStart(name string, fn Hook) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.startHooks = append(m.startHooks, namedHook{name, fn})
}

// OnStop registers a hook to be called during Stop (LIFO order).
func (m *Manager) OnStop(name string, fn Hook) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stopHooks = append(m.stopHooks, namedHook{name, fn})
}

// Start runs all registered start hooks sequentially.
func (m *Manager) Start(ctx context.Context) error {
	m.mu.Lock()
	hooks := make([]namedHook, len(m.startHooks))
	copy(hooks, m.startHooks)
	m.mu.Unlock()

	for _, h := range hooks {
		if err := h.fn(ctx); err != nil {
			return fmt.Errorf("lifecycle: start hook %q failed: %w", h.name, err)
		}
	}
	return nil
}

// Stop runs all registered stop hooks in reverse order with the drain timeout.
func (m *Manager) Stop() error {
	m.mu.Lock()
	hooks := make([]namedHook, len(m.stopHooks))
	copy(hooks, m.stopHooks)
	m.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), m.drainTimeout)
	defer cancel()

	var errs []error
	for i := len(hooks) - 1; i >= 0; i-- {
		h := hooks[i]
		if err := h.fn(ctx); err != nil {
			errs = append(errs, fmt.Errorf("lifecycle: stop hook %q failed: %w", h.name, err))
		}
	}
	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}
