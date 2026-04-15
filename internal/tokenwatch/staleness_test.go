package tokenwatch

import (
	"context"
	"testing"
	"time"

	"github.com/vaultwatch/internal/alert"
)

func TestDefaultStalenessConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultStalenessConfig()
	if cfg.WarnAfter <= 0 {
		t.Fatal("expected positive WarnAfter")
	}
	if cfg.CriticalAfter <= cfg.WarnAfter {
		t.Fatal("expected CriticalAfter > WarnAfter")
	}
}

func TestStalenessDetector_Fresh_ReturnsNil(t *testing.T) {
	d := NewStalenessDetector(DefaultStalenessConfig())
	a := d.Check("tok-1", time.Now())
	if a != nil {
		t.Fatalf("expected nil alert for fresh token, got %+v", a)
	}
}

func TestStalenessDetector_Warning(t *testing.T) {
	cfg := StalenessConfig{WarnAfter: 5 * time.Minute, CriticalAfter: 20 * time.Minute}
	d := NewStalenessDetector(cfg)
	d.now = func() time.Time { return time.Now().Add(10 * time.Minute) }
	a := d.Check("tok-2", time.Now())
	if a == nil {
		t.Fatal("expected warning alert")
	}
	if a.Level != alert.LevelWarning {
		t.Fatalf("expected LevelWarning, got %v", a.Level)
	}
}

func TestStalenessDetector_Critical(t *testing.T) {
	cfg := StalenessConfig{WarnAfter: 5 * time.Minute, CriticalAfter: 20 * time.Minute}
	d := NewStalenessDetector(cfg)
	d.now = func() time.Time { return time.Now().Add(25 * time.Minute) }
	a := d.Check("tok-3", time.Now())
	if a == nil {
		t.Fatal("expected critical alert")
	}
	if a.Level != alert.LevelCritical {
		t.Fatalf("expected LevelCritical, got %v", a.Level)
	}
	if a.LeaseID != "tok-3" {
		t.Fatalf("unexpected lease ID %s", a.LeaseID)
	}
}

func TestStalenessDetector_ZeroDurations_UsesDefaults(t *testing.T) {
	d := NewStalenessDetector(StalenessConfig{})
	if d.cfg.WarnAfter != DefaultStalenessConfig().WarnAfter {
		t.Fatal("expected default WarnAfter")
	}
}

// --- StalenessScanner ---

type mockLastSeen struct{ entries map[string]time.Time }

func (m *mockLastSeen) LastSeen(id string) (time.Time, bool) {
	v, ok := m.entries[id]
	return v, ok
}

func TestStalenessScanner_NilRegistry_Panics(t *testing.T) {
	defer func() { recover() }()
	NewStalenessScanner(nil, &mockLastSeen{}, nil, nil, nil)
	t.Fatal("expected panic")
}

func TestStalenessScanner_Scan_DispatchesAlerts(t *testing.T) {
	reg := NewRegistry()
	_ = reg.Add("tok-a")
	src := &mockLastSeen{entries: map[string]time.Time{
		"tok-a": time.Now().Add(-1 * time.Hour),
	}}
	cfg := StalenessConfig{WarnAfter: 5 * time.Minute, CriticalAfter: 30 * time.Minute}
	det := NewStalenessDetector(cfg)
	var dispatched []interface{}
	s := NewStalenessScanner(reg, src, det, func(a interface{}) { dispatched = append(dispatched, a) }, nil)
	if err := s.Scan(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dispatched) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(dispatched))
	}
}

func TestStalenessScanner_Scan_SkipsMissingLastSeen(t *testing.T) {
	reg := NewRegistry()
	_ = reg.Add("tok-b")
	src := &mockLastSeen{entries: map[string]time.Time{}}
	s := NewStalenessScanner(reg, src, nil, func(a interface{}) { panic("should not dispatch") }, nil)
	if err := s.Scan(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
