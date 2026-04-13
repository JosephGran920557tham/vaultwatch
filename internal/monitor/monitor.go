package monitor

import (
	"context"
	"log"
	"time"

	"github.com/vaultwatch/internal/alert"
	"github.com/vaultwatch/internal/config"
	"github.com/vaultwatch/internal/vault"
)

// Monitor orchestrates periodic lease checks and dispatches alerts.
type Monitor struct {
	checker    *vault.LeaseChecker
	dispatcher *alert.Dispatcher
	cfg        *config.Config
}

// New creates a Monitor with the provided dependencies.
func New(checker *vault.LeaseChecker, dispatcher *alert.Dispatcher, cfg *config.Config) *Monitor {
	return &Monitor{
		checker:    checker,
		dispatcher: dispatcher,
		cfg:        cfg,
	}
}

// Run starts the monitoring loop, ticking at the configured interval.
// It blocks until ctx is cancelled.
func (m *Monitor) Run(ctx context.Context) error {
	ticker := time.NewTicker(m.cfg.PollInterval)
	defer ticker.Stop()

	log.Printf("[monitor] starting — poll interval %s, warn threshold %s, critical threshold %s",
		m.cfg.PollInterval, m.cfg.WarnThreshold, m.cfg.CriticalThreshold)

	// Run once immediately before waiting for the first tick.
	m.runOnce(ctx)

	for {
		select {
		case <-ticker.C:
			m.runOnce(ctx)
		case <-ctx.Done():
			log.Println("[monitor] shutting down")
			return ctx.Err()
		}
	}
}

// runOnce performs a single check-and-dispatch cycle.
func (m *Monitor) runOnce(ctx context.Context) {
	leases, err := m.checker.ListExpiring(ctx, m.cfg.WarnThreshold)
	if err != nil {
		log.Printf("[monitor] lease check error: %v", err)
		return
	}

	if len(leases) == 0 {
		log.Println("[monitor] no expiring leases found")
		return
	}

	log.Printf("[monitor] found %d expiring lease(s)", len(leases))

	for _, lease := range leases {
		a := alert.Build(lease, m.cfg.WarnThreshold, m.cfg.CriticalThreshold)
		if err := m.dispatcher.Dispatch(ctx, a); err != nil {
			log.Printf("[monitor] dispatch error for lease %s: %v", lease.LeaseID, err)
		}
	}
}
