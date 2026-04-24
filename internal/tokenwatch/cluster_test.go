package tokenwatch

import (
	"testing"
	"time"

	"github.com/vaultwatch/internal/alert"
)

func TestDefaultClusterConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultClusterConfig()
	if cfg.MinMembers <= 0 {
		t.Errorf("expected positive MinMembers, got %d", cfg.MinMembers)
	}
	if cfg.WarnThreshold <= 0 || cfg.WarnThreshold >= 1 {
		t.Errorf("unexpected WarnThreshold: %f", cfg.WarnThreshold)
	}
	if cfg.CritThreshold <= 0 || cfg.CritThreshold >= cfg.WarnThreshold {
		t.Errorf("unexpected CritThreshold: %f", cfg.CritThreshold)
	}
	if cfg.MemberTTL <= 0 {
		t.Errorf("expected positive MemberTTL, got %v", cfg.MemberTTL)
	}
}

func TestNewClusterDetector_ZeroValues_UsesDefaults(t *testing.T) {
	d := NewClusterDetector(ClusterConfig{})
	def := DefaultClusterConfig()
	if d.cfg.MinMembers != def.MinMembers {
		t.Errorf("expected MinMembers %d, got %d", def.MinMembers, d.cfg.MinMembers)
	}
}

func TestClusterDetector_Check_AllHealthy_ReturnsNil(t *testing.T) {
	d := NewClusterDetector(ClusterConfig{MinMembers: 2, WarnThreshold: 0.5, CritThreshold: 0.25, MemberTTL: time.Minute})
	d.Observe("node-1")
	d.Observe("node-2")
	if a := d.Check("tok"); a != nil {
		t.Errorf("expected nil alert for healthy cluster, got %+v", a)
	}
}

func TestClusterDetector_Check_Warning(t *testing.T) {
	d := NewClusterDetector(ClusterConfig{MinMembers: 4, WarnThreshold: 0.6, CritThreshold: 0.25, MemberTTL: time.Minute})
	// 2 of 4 = 0.5, below warn threshold 0.6
	d.Observe("node-1")
	d.Observe("node-2")
	a := d.Check("tok")
	if a == nil {
		t.Fatal("expected warning alert, got nil")
	}
	if a.Level != alert.LevelWarning {
		t.Errorf("expected Warning, got %v", a.Level)
	}
}

func TestClusterDetector_Check_Critical(t *testing.T) {
	d := NewClusterDetector(ClusterConfig{MinMembers: 8, WarnThreshold: 0.5, CritThreshold: 0.25, MemberTTL: time.Minute})
	// 1 of 8 = 0.125, below crit threshold 0.25
	d.Observe("node-1")
	a := d.Check("tok")
	if a == nil {
		t.Fatal("expected critical alert, got nil")
	}
	if a.Level != alert.LevelCritical {
		t.Errorf("expected Critical, got %v", a.Level)
	}
}

func TestClusterDetector_Check_ExpiredMembers_CountedOut(t *testing.T) {
	d := NewClusterDetector(ClusterConfig{MinMembers: 2, WarnThreshold: 0.9, CritThreshold: 0.4, MemberTTL: time.Millisecond})
	d.Observe("node-1")
	d.Observe("node-2")
	time.Sleep(5 * time.Millisecond)
	a := d.Check("tok")
	if a == nil {
		t.Fatal("expected alert after members expired, got nil")
	}
}

func TestClusterDetector_Check_LeaseIDPropagated(t *testing.T) {
	d := NewClusterDetector(ClusterConfig{MinMembers: 10, WarnThreshold: 0.9, CritThreshold: 0.5, MemberTTL: time.Minute})
	a := d.Check("my-token")
	if a == nil {
		t.Fatal("expected alert")
	}
	if a.LeaseID != "my-token" {
		t.Errorf("expected LeaseID 'my-token', got %q", a.LeaseID)
	}
}
