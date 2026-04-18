package tokenwatch

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/your-org/vaultwatch/internal/alert"
)

type stubSource struct {
	alerts []alert.Alert
	err    error
}

func (s *stubSource) Scan(_ context.Context) ([]alert.Alert, error) {
	return s.alerts, s.err
}

func TestNewFanoutScanner_NilFanout_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic")
		}
	}()
	NewFanoutScanner(nil, nil, &stubSource{})
}

func TestNewFanoutScanner_NoSources_Panics(t *testing.T) {
	h := func(_ context.Context, _ alert.Alert) error { return nil }
	f := NewFanout(DefaultFanoutConfig(), h)
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic")
		}
	}()
	NewFanoutScanner(f, nil)
}

func TestFanoutScanner_Run_DispatchesAllAlerts(t *testing.T) {
	var dispatched atomic.Int32
	h := func(_ context.Context, _ alert.Alert) error {
		dispatched.Add(1)
		return nil
	}
	f := NewFanout(DefaultFanoutConfig(), h)
	src := &stubSource{
		alerts: []alert.Alert{
			{LeaseID: "t1", Level: alert.LevelWarning},
			{LeaseID: "t2", Level: alert.LevelCritical},
		},
	}
	fs := NewFanoutScanner(f, nil, src)
	if err := fs.Run(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dispatched.Load() != 2 {
		t.Errorf("expected 2 dispatches, got %d", dispatched.Load())
	}
}

func TestFanoutScanner_Run_SourceError_Continues(t *testing.T) {
	var dispatched atomic.Int32
	h := func(_ context.Context, _ alert.Alert) error {
		dispatched.Add(1)
		return nil
	}
	f := NewFanout(DefaultFanoutConfig(), h)
	bad := &stubSource{err: errors.New("source down")}
	good := &stubSource{alerts: []alert.Alert{{LeaseID: "ok", Level: alert.LevelInfo}}}
	fs := NewFanoutScanner(f, nil, bad, good)
	if err := fs.Run(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dispatched.Load() != 1 {
		t.Errorf("expected 1 dispatch, got %d", dispatched.Load())
	}
}
