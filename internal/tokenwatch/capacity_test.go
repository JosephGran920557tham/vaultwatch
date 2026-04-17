package tokenwatch

import (
	"testing"
)

func TestDefaultCapacityConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultCapacityConfig()
	if cfg.MaxTokens <= 0 {
		t.Fatal("expected positive MaxTokens")
	}
	if cfg.WarnThreshold <= 0 || cfg.WarnThreshold >= 1 {
		t.Fatal("WarnThreshold out of range")
	}
	if cfg.CritThreshold <= cfg.WarnThreshold {
		t.Fatal("CritThreshold must exceed WarnThreshold")
	}
}

func TestNewCapacityDetector_ZeroValues_UsesDefaults(t *testing.T) {
	d := NewCapacityDetector(CapacityConfig{})
	def := DefaultCapacityConfig()
	if d.cfg.MaxTokens != def.MaxTokens {
		t.Fatalf("expected %d, got %d", def.MaxTokens, d.cfg.MaxTokens)
	}
}

func TestCapacityDetector_Check_Ok(t *testing.T) {
	d := NewCapacityDetector(DefaultCapacityConfig())
	res := d.Check(10)
	if res.Level != "ok" {
		t.Fatalf("expected ok, got %s", res.Level)
	}
	if res.Count != 10 {
		t.Fatalf("expected count 10, got %d", res.Count)
	}
}

func TestCapacityDetector_Check_Warning(t *testing.T) {
	cfg := DefaultCapacityConfig()
	cfg.MaxTokens = 100
	d := NewCapacityDetector(cfg)
	res := d.Check(80) // 80% >= 75% warn threshold
	if res.Level != "warning" {
		t.Fatalf("expected warning, got %s", res.Level)
	}
}

func TestCapacityDetector_Check_Critical(t *testing.T) {
	cfg := DefaultCapacityConfig()
	cfg.MaxTokens = 100
	d := NewCapacityDetector(cfg)
	res := d.Check(95) // 95% >= 90% crit threshold
	if res.Level != "critical" {
		t.Fatalf("expected critical, got %s", res.Level)
	}
}

func TestCapacityDetector_LastChecked_UpdatedAfterCheck(t *testing.T) {
	d := NewCapacityDetector(DefaultCapacityConfig())
	if !d.LastChecked().IsZero() {
		t.Fatal("expected zero time before first check")
	}
	d.Check(1)
	if d.LastChecked().IsZero() {
		t.Fatal("expected non-zero time after check")
	}
}

func TestCapacityScanner_Scan_Ok_ReturnsNil(t *testing.T) {
	reg := NewRegistry()
	det := NewCapacityDetector(DefaultCapacityConfig())
	s := NewCapacityScanner(reg, det)
	if a := s.Scan(); a != nil {
		t.Fatalf("expected nil alert for empty registry, got %+v", a)
	}
}

func TestCapacityScanner_Scan_Warning(t *testing.T) {
	reg := NewRegistry()
	for i := 0; i < 80; i++ {
		_ = reg.Add(TokenID(fmt.Sprintf("tok-%d", i)))
	}
	cfg := DefaultCapacityConfig()
	cfg.MaxTokens = 100
	det := NewCapacityDetector(cfg)
	s := NewCapacityScanner(reg, det)
	a := s.Scan()
	if a == nil {
		t.Fatal("expected alert")
	}
	if a.Level != "warning" {
		t.Fatalf("expected warning, got %s", a.Level)
	}
}
