package tokenwatch

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/your-org/vaultwatch/internal/alert"
)

func makeFanoutAlert() alert.Alert {
	return alert.Alert{LeaseID: "token-fanout", Level: alert.LevelWarning}
}

func TestDefaultFanoutConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultFanoutConfig()
	if cfg.Workers <= 0 {
		t.Errorf("expected positive Workers, got %d", cfg.Workers)
	}
}

func TestNewFanout_NilHandlers_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic with no handlers")
		}
	}()
	NewFanout(DefaultFanoutConfig())
}

func TestFanout_Dispatch_CallsAllHandlers(t *testing.T) {
	var count atomic.Int32
	h := func(_ context.Context, _ alert.Alert) error {
		count.Add(1)
		return nil
	}
	f := NewFanout(DefaultFanoutConfig(), h, h, h)
	errs := f.Dispatch(context.Background(), makeFanoutAlert())
	if len(errs) != 0 {
		t.Errorf("unexpected errors: %v", errs)
	}
	if count.Load() != 3 {
		t.Errorf("expected 3 handler calls, got %d", count.Load())
	}
}

func TestFanout_Dispatch_CollectsErrors(t *testing.T) {
	ok := func(_ context.Context, _ alert.Alert) error { return nil }
	fail := func(_ context.Context, _ alert.Alert) error { return errors.New("boom") }
	f := NewFanout(DefaultFanoutConfig(), ok, fail, fail)
	errs := f.Dispatch(context.Background(), makeFanoutAlert())
	if len(errs) != 2 {
		t.Errorf("expected 2 errors, got %d", len(errs))
	}
}

func TestNewFanout_ZeroWorkers_UsesDefault(t *testing.T) {
	h := func(_ context.Context, _ alert.Alert) error { return nil }
	f := NewFanout(FanoutConfig{Workers: 0}, h)
	if f.cfg.Workers != DefaultFanoutConfig().Workers {
		t.Errorf("expected default workers, got %d", f.cfg.Workers)
	}
}
