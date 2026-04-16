package tokenwatch

import (
	"testing"
	"time"

	"github.com/vaultwatch/internal/alert"
)

func TestDefaultAgingConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultAgingConfig()
	if cfg.WarnAfter <= 0 {
		t.Fatal("expected positive WarnAfter")
	}
	if cfg.CriticalAfter <= cfg.WarnAfter {
		t.Fatal("expected CriticalAfter > WarnAfter")
	}
}

func TestNewAgingDetector_ZeroValues_UsesDefaults(t *testing.T) {
	d := NewAgingDetector(AgingConfig{})
	def := DefaultAgingConfig()
	if d.cfg.WarnAfter != def.WarnAfter {
		t.Errorf("WarnAfter: got %v, want %v", d.cfg.WarnAfter, def.WarnAfter)
	}
	if d.cfg.CriticalAfter != def.CriticalAfter {
		t.Errorf("CriticalAfter: got %v, want %v", d.cfg.CriticalAfter, def.CriticalAfter)
	}
}

func TestAgingDetector_Check_YoungToken_ReturnsNil(t *testing.T) {
	d := NewAgingDetector(AgingConfig{WarnAfter: time.Hour, CriticalAfter: 2 * time.Hour})
	alrt := d.Check("tok-1", d.now().Add(-30*time.Minute))
	if alrt != nil {
		t.Fatalf("expected nil, got %+v", alrt)
	}
}

func TestAgingDetector_Check_Warning(t *testing.T) {
	d := NewAgingDetector(AgingConfig{WarnAfter: time.Hour, CriticalAfter: 2 * time.Hour})
	alrt := d.Check("tok-2", d.now().Add(-90*time.Minute))
	if alrt == nil {
		t.Fatal("expected alert, got nil")
	}
	if alrt.Level != alert.LevelWarning {
		t.Errorf("level: got %v, want Warning", alrt.Level)
	}
	if alrt.LeaseID != "tok-2" {
		t.Errorf("leaseID: got %v, want tok-2", alrt.LeaseID)
	}
}

func TestAgingDetector_Check_Critical(t *testing.T) {
	d := NewAgingDetector(AgingConfig{WarnAfter: time.Hour, CriticalAfter: 2 * time.Hour})
	alrt := d.Check("tok-3", d.now().Add(-3*time.Hour))
	if alrt == nil {
		t.Fatal("expected alert, got nil")
	}
	if alrt.Level != alert.LevelCritical {
		t.Errorf("level: got %v, want Critical", alrt.Level)
	}
}

func TestAgingDetector_Check_LabelSet(t *testing.T) {
	d := NewAgingDetector(AgingConfig{WarnAfter: time.Hour, CriticalAfter: 2 * time.Hour})
	alrt := d.Check("tok-4", d.now().Add(-90*time.Minute))
	if alrt == nil {
		t.Fatal("expected alert")
	}
	if alrt.Labels["detector"] != "aging" {
		t.Errorf("label detector: got %q, want aging", alrt.Labels["detector"])
	}
}
