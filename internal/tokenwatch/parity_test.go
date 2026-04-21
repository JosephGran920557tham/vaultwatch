package tokenwatch

import (
	"context"
	"testing"
	"time"

	"github.com/vaultwatch/internal/alert"
)

func TestDefaultParityConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultParityConfig()
	if cfg.MaxSkew <= 0 {
		t.Fatal("expected positive MaxSkew")
	}
	if cfg.WarningThreshold <= 0 || cfg.WarningThreshold >= 1 {
		t.Fatalf("unexpected WarningThreshold: %v", cfg.WarningThreshold)
	}
	if cfg.CriticalThreshold <= cfg.WarningThreshold {
		t.Fatal("CriticalThreshold should exceed WarningThreshold")
	}
}

func TestNewParityDetector_ZeroValues_UsesDefaults(t *testing.T) {
	d := NewParityDetector(ParityConfig{})
	def := DefaultParityConfig()
	if d.cfg.MaxSkew != def.MaxSkew {
		t.Errorf("MaxSkew: got %v, want %v", d.cfg.MaxSkew, def.MaxSkew)
	}
}

func TestParityDetector_Check_NoPair_ReturnsNil(t *testing.T) {
	d := NewParityDetector(DefaultParityConfig())
	result := d.Check("tok-1", 5*time.Minute, 5*time.Minute)
	if result != nil {
		t.Fatalf("expected nil, got %+v", result)
	}
}

func TestParityDetector_Check_WithinThreshold_ReturnsNil(t *testing.T) {
	d := NewParityDetector(DefaultParityConfig())
	result := d.Check("tok-1", 5*time.Minute, 5*time.Minute+10*time.Second)
	if result != nil {
		t.Fatalf("expected nil for small skew, got %+v", result)
	}
}

func TestParityDetector_Check_Warning(t *testing.T) {
	cfg := DefaultParityConfig()
	d := NewParityDetector(cfg)
	// skew = 40s > MaxSkew(30s), ratio ~13% > WarningThreshold(10%)
	a := d.Check("tok-1", 5*time.Minute, 5*time.Minute+40*time.Second)
	if a == nil {
		t.Fatal("expected warning alert")
	}
	if a.Level != alert.LevelWarning {
		t.Errorf("expected Warning, got %v", a.Level)
	}
}

func TestParityDetector_Check_Critical(t *testing.T) {
	cfg := DefaultParityConfig()
	d := NewParityDetector(cfg)
	// skew = 90s > MaxSkew(30s), ratio = 90/300 = 30% > CriticalThreshold(25%)
	a := d.Check("tok-1", 5*time.Minute, 5*time.Minute+90*time.Second)
	if a == nil {
		t.Fatal("expected critical alert")
	}
	if a.Level != alert.LevelCritical {
		t.Errorf("expected Critical, got %v", a.Level)
	}
}

func TestNewParityScanner_NilRegistry_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil registry")
		}
	}()
	NewParityScanner(nil, NewParityDetector(DefaultParityConfig()), func(_ context.Context, _ string) (TokenInfo, error) {
		return TokenInfo{}, nil
	})
}

func TestParityScanner_Scan_NoTokens_ReturnsEmpty(t *testing.T) {
	reg := NewRegistry()
	det := NewParityDetector(DefaultParityConfig())
	lookup := func(_ context.Context, id string) (TokenInfo, error) {
		return TokenInfo{TTL: 5 * time.Minute}, nil
	}
	s := NewParityScanner(reg, det, lookup)
	alerts := s.Scan(context.Background())
	if len(alerts) != 0 {
		t.Fatalf("expected empty, got %d alerts", len(alerts))
	}
}

func TestParityScanner_Scan_DetectsMismatch(t *testing.T) {
	reg := NewRegistry()
	_ = reg.Add("tok-a")
	det := NewParityDetector(DefaultParityConfig())
	det.Pair("tok-a", "tok-b")

	ttls := map[string]time.Duration{
		"tok-a": 5 * time.Minute,
		"tok-b": 5*time.Minute + 90*time.Second,
	}
	lookup := func(_ context.Context, id string) (TokenInfo, error) {
		return TokenInfo{TTL: ttls[id]}, nil
	}
	s := NewParityScanner(reg, det, lookup)
	alerts := s.Scan(context.Background())
	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(alerts))
	}
	if alerts[0].LeaseID != "tok-a" {
		t.Errorf("unexpected lease ID: %s", alerts[0].LeaseID)
	}
}
