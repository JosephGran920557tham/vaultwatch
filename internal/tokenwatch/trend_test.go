package tokenwatch

import (
	"errors"
	"testing"
	"time"

	"github.com/vaultwatch/internal/alert"
)

func TestDefaultTrendConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultTrendConfig()
	if cfg.SampleWindow <= 0 {
		t.Error("expected positive SampleWindow")
	}
	if cfg.MinSamples <= 0 {
		t.Error("expected positive MinSamples")
	}
	if cfg.DropThreshold <= 0 {
		t.Error("expected positive DropThreshold")
	}
}

func TestNewTrendDetector_ZeroValues_UsesDefaults(t *testing.T) {
	d := NewTrendDetector(TrendConfig{})
	def := DefaultTrendConfig()
	if d.cfg.SampleWindow != def.SampleWindow {
		t.Errorf("got SampleWindow %v, want %v", d.cfg.SampleWindow, def.SampleWindow)
	}
}

func TestTrendDetector_Check_InsufficientSamples(t *testing.T) {
	d := NewTrendDetector(TrendConfig{SampleWindow: time.Minute, MinSamples: 3, DropThreshold: 0.2})
	d.Record("tok1", 10*time.Minute)
	d.Record("tok1", 5*time.Minute)
	if a := d.Check("tok1"); a != nil {
		t.Errorf("expected nil alert with insufficient samples, got %+v", a)
	}
}

func TestTrendDetector_Check_NoDrop_ReturnsNil(t *testing.T) {
	d := NewTrendDetector(TrendConfig{SampleWindow: time.Minute, MinSamples: 2, DropThreshold: 0.5})
	d.Record("tok2", 10*time.Minute)
	d.Record("tok2", 9*time.Minute) // only 10% drop
	if a := d.Check("tok2"); a != nil {
		t.Errorf("expected nil alert for small drop, got %+v", a)
	}
}

func TestTrendDetector_Check_LargeDrop_ReturnsWarning(t *testing.T) {
	d := NewTrendDetector(TrendConfig{SampleWindow: time.Minute, MinSamples: 2, DropThreshold: 0.2})
	d.Record("tok3", 10*time.Minute)
	d.Record("tok3", 2*time.Minute) // 80% drop
	a := d.Check("tok3")
	if a == nil {
		t.Fatal("expected warning alert for large TTL drop")
	}
	if a.Level != alert.Warning {
		t.Errorf("got level %v, want Warning", a.Level)
	}
	if a.LeaseID != "tok3" {
		t.Errorf("got LeaseID %q, want tok3", a.LeaseID)
	}
}

func TestNewTrendScanner_NilRegistry_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil registry")
		}
	}()
	NewTrendScanner(nil, NewTrendDetector(TrendConfig{}), func(string) (TokenInfo, error) { return TokenInfo{}, nil }, nil)
}

func TestNewTrendScanner_NilDetector_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil detector")
		}
	}()
	NewTrendScanner(NewRegistry(), nil, func(string) (TokenInfo, error) { return TokenInfo{}, nil }, nil)
}

func TestTrendScanner_Scan_LookupError_Skipped(t *testing.T) {
	reg := NewRegistry()
	_ = reg.Add("badtok")
	det := NewTrendDetector(TrendConfig{SampleWindow: time.Minute, MinSamples: 2, DropThreshold: 0.2})
	scanner := NewTrendScanner(reg, det, func(id string) (TokenInfo, error) {
		return TokenInfo{}, errors.New("lookup failed")
	}, nil)
	alerts := scanner.Scan()
	if len(alerts) != 0 {
		t.Errorf("expected no alerts on lookup error, got %d", len(alerts))
	}
}

func TestTrendScanner_Scan_DetectsDrop(t *testing.T) {
	reg := NewRegistry()
	_ = reg.Add("droptok")
	det := NewTrendDetector(TrendConfig{SampleWindow: time.Minute, MinSamples: 2, DropThreshold: 0.2})
	call := 0
	ttls := []time.Duration{10 * time.Minute, 1 * time.Minute}
	scanner := NewTrendScanner(reg, det, func(id string) (TokenInfo, error) {
		ttl := ttls[call%len(ttls)]
		call++
		return TokenInfo{ID: id, TTL: ttl}, nil
	}, nil)
	scanner.Scan() // first sample
	alerts := scanner.Scan() // second sample — large drop
	if len(alerts) == 0 {
		t.Error("expected at least one trend alert")
	}
}
