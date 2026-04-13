package monitor_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/vaultwatch/internal/alert"
	"github.com/vaultwatch/internal/config"
	"github.com/vaultwatch/internal/monitor"
	"github.com/vaultwatch/internal/vault"
)

// stubChecker satisfies vault.LeaseChecker's interface for testing.
type stubChecker struct {
	leases []vault.LeaseInfo
	err    error
}

func (s *stubChecker) ListExpiring(_ context.Context, _ time.Duration) ([]vault.LeaseInfo, error) {
	return s.leases, s.err
}

// recordingNotifier captures dispatched alerts.
type recordingNotifier struct {
	sent []alert.Alert
}

func (r *recordingNotifier) Send(_ context.Context, a alert.Alert) error {
	r.sent = append(r.sent, a)
	return nil
}

func (r *recordingNotifier) Name() string { return "recorder" }

func defaultCfg() *config.Config {
	return &config.Config{
		PollInterval:      50 * time.Millisecond,
		WarnThreshold:     72 * time.Hour,
		CriticalThreshold: 24 * time.Hour,
	}
}

func TestMonitor_RunOnce_DispatchesAlerts(t *testing.T) {
	leases := []vault.LeaseInfo{
		{LeaseID: "secret/data/db#abc", TTL: 10 * time.Hour},
	}
	checker := &stubChecker{leases: leases}
	notifier := &recordingNotifier{}
	dispatcher := alert.NewDispatcher([]alert.Notifier{notifier})
	cfg := defaultCfg()

	m := monitor.New(checker, dispatcher, cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Millisecond)
	defer cancel()

	_ = m.Run(ctx) // exits when ctx is cancelled

	if len(notifier.sent) == 0 {
		t.Fatal("expected at least one alert to be dispatched")
	}
	if notifier.sent[0].LeaseID != "secret/data/db#abc" {
		t.Errorf("unexpected lease id: %s", notifier.sent[0].LeaseID)
	}
}

func TestMonitor_RunOnce_CheckerError_NoDispatch(t *testing.T) {
	checker := &stubChecker{err: errors.New("vault unavailable")}
	notifier := &recordingNotifier{}
	dispatcher := alert.NewDispatcher([]alert.Notifier{notifier})

	m := monitor.New(checker, dispatcher, defaultCfg())

	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()

	_ = m.Run(ctx)

	if len(notifier.sent) != 0 {
		t.Errorf("expected no alerts on checker error, got %d", len(notifier.sent))
	}
}

func TestMonitor_RunOnce_NoLeases_NoDispatch(t *testing.T) {
	checker := &stubChecker{leases: nil}
	notifier := &recordingNotifier{}
	dispatcher := alert.NewDispatcher([]alert.Notifier{notifier})

	m := monitor.New(checker, dispatcher, defaultCfg())

	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()

	_ = m.Run(ctx)

	if len(notifier.sent) != 0 {
		t.Errorf("expected no alerts when no leases, got %d", len(notifier.sent))
	}
}
