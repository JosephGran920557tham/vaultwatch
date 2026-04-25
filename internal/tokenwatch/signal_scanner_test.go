package tokenwatch

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/vaultwatch/internal/alert"
)

// fakeSignalSource is a stub aggregatorSource.
type fakeSignalSource struct {
	alerts []alert.Alert
	err    error
}

func (f *fakeSignalSource) Scan(_ context.Context) ([]alert.Alert, error) {
	return f.alerts, f.err
}

func newTestSignalScanner(alerts []alert.Alert, cfg SignalConfig) *SignalScanner {
	src := &fakeSignalSource{alerts: alerts}
	agg := NewSignalAggregator(cfg)
	return NewSignalScanner(src, agg, nil)
}

func TestNewSignalScanner_NilSource_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil source")
		}
	}()
	NewSignalScanner(nil, NewSignalAggregator(DefaultSignalConfig()), nil)
}

func TestNewSignalScanner_NilAggregator_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil aggregator")
		}
	}()
	NewSignalScanner(&fakeSignalSource{}, nil, nil)
}

func TestSignalScanner_Scan_BelowThreshold_ReturnsEmpty(t *testing.T) {
	a := alert.Alert{LeaseID: "tok-1", Level: alert.LevelWarning}
	sc := newTestSignalScanner([]alert.Alert{a}, SignalConfig{MinStrength: 3, DecayWindow: time.Minute})
	out, err := sc.Scan(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 0 {
		t.Errorf("expected 0 alerts, got %d", len(out))
	}
}

func TestSignalScanner_Scan_ReachesThreshold_ReturnsAlert(t *testing.T) {
	a := alert.Alert{LeaseID: "tok-1", Level: alert.LevelCritical}
	src := &fakeSignalSource{alerts: []alert.Alert{a}}
	agg := NewSignalAggregator(SignalConfig{MinStrength: 2, DecayWindow: time.Minute})
	sc := NewSignalScanner(src, agg, nil)

	// First scan — below threshold.
	out, _ := sc.Scan(context.Background())
	if len(out) != 0 {
		t.Fatalf("expected 0 on first scan, got %d", len(out))
	}
	// Second scan — meets threshold.
	out, err := sc.Scan(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 1 {
		t.Errorf("expected 1 alert, got %d", len(out))
	}
}

func TestSignalScanner_Scan_SourceError_Propagates(t *testing.T) {
	src := &fakeSignalSource{err: errors.New("vault unavailable")}
	agg := NewSignalAggregator(DefaultSignalConfig())
	sc := NewSignalScanner(src, agg, nil)
	_, err := sc.Scan(context.Background())
	if err == nil {
		t.Error("expected error from source")
	}
}
