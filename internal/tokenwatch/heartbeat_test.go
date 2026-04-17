package tokenwatch

import (
	"testing"
	"time"
)

func TestDefaultHeartbeatConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultHeartbeatConfig()
	if cfg.StaleAfter <= 0 {
		t.Fatal("expected positive StaleAfter")
	}
	if cfg.CriticalAfter <= cfg.StaleAfter {
		t.Fatal("expected CriticalAfter > StaleAfter")
	}
}

func TestNewHeartbeatDetector_ZeroValues_UsesDefaults(t *testing.T) {
	d := NewHeartbeatDetector(HeartbeatConfig{})
	def := DefaultHeartbeatConfig()
	if d.cfg.StaleAfter != def.StaleAfter {
		t.Errorf("got StaleAfter %v, want %v", d.cfg.StaleAfter, def.StaleAfter)
	}
}

func TestHeartbeatDetector_Check_NeverSeen_ReturnsCritical(t *testing.T) {
	d := NewHeartbeatDetector(DefaultHeartbeatConfig())
	alert := d.Check("tok-1", time.Now())
	if alert == nil {
		t.Fatal("expected alert for unseen token")
	}
	if alert.Level != LevelCritical {
		t.Errorf("got level %v, want critical", alert.Level)
	}
}

func TestHeartbeatDetector_Check_RecentBeat_ReturnsNil(t *testing.T) {
	d := NewHeartbeatDetector(DefaultHeartbeatConfig())
	d.Beat("tok-2")
	alert := d.Check("tok-2", time.Now())
	if alert != nil {
		t.Fatalf("expected nil alert for fresh token, got %+v", alert)
	}
}

func TestHeartbeatDetector_Check_StaleReturnsWarning(t *testing.T) {
	cfg := HeartbeatConfig{StaleAfter: 1 * time.Minute, CriticalAfter: 10 * time.Minute}
	d := NewHeartbeatDetector(cfg)
	d.mu.Lock()
	d.seen["tok-3"] = time.Now().Add(-2 * time.Minute)
	d.mu.Unlock()
	alert := d.Check("tok-3", time.Now())
	if alert == nil {
		t.Fatal("expected warning alert")
	}
	if alert.Level != LevelWarning {
		t.Errorf("got level %v, want warning", alert.Level)
	}
}

func TestHeartbeatDetector_Check_CriticalAfterLongSilence(t *testing.T) {
	cfg := HeartbeatConfig{StaleAfter: 1 * time.Minute, CriticalAfter: 5 * time.Minute}
	d := NewHeartbeatDetector(cfg)
	d.mu.Lock()
	d.seen["tok-4"] = time.Now().Add(-10 * time.Minute)
	d.mu.Unlock()
	alert := d.Check("tok-4", time.Now())
	if alert == nil || alert.Level != LevelCritical {
		t.Errorf("expected critical alert, got %+v", alert)
	}
}

func TestHeartbeatDetector_Beat_UpdatesTimestamp(t *testing.T) {
	cfg := HeartbeatConfig{StaleAfter: 1 * time.Minute, CriticalAfter: 5 * time.Minute}
	d := NewHeartbeatDetector(cfg)
	d.mu.Lock()
	d.seen["tok-5"] = time.Now().Add(-10 * time.Minute)
	d.mu.Unlock()
	d.Beat("tok-5")
	alert := d.Check("tok-5", time.Now())
	if alert != nil {
		t.Errorf("expected nil after fresh beat, got %+v", alert)
	}
}
