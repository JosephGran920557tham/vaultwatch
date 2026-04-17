package tokenwatch

import (
	"testing"
	"time"

	"github.com/vaultwatch/internal/alert"
)

func TestDefaultEscalationConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultEscalationConfig()
	if cfg.Window <= 0 {
		t.Fatal("expected positive window")
	}
	if cfg.Threshold <= 0 {
		t.Fatal("expected positive threshold")
	}
}

func TestNewEscalation_ZeroValues_UsesDefaults(t *testing.T) {
	e := NewEscalation(EscalationConfig{})
	if e.cfg.Window <= 0 {
		t.Fatal("expected default window")
	}
	if e.cfg.Threshold <= 0 {
		t.Fatal("expected default threshold")
	}
}

func TestEscalation_Check_FirstCall_ReturnsOriginalLevel(t *testing.T) {
	e := NewEscalation(EscalationConfig{Window: time.Minute, Threshold: 3, EscalateLevel: alert.LevelCritical})
	level := e.Check("tok-1", alert.LevelWarning)
	if level != alert.LevelWarning {
		t.Fatalf("expected warning, got %v", level)
	}
}

func TestEscalation_Check_BelowThreshold_NoEscalation(t *testing.T) {
	e := NewEscalation(EscalationConfig{Window: time.Minute, Threshold: 3, EscalateLevel: alert.LevelCritical})
	e.Check("tok-1", alert.LevelWarning)
	level := e.Check("tok-1", alert.LevelWarning)
	if level != alert.LevelWarning {
		t.Fatalf("expected warning before threshold, got %v", level)
	}
}

func TestEscalation_Check_AtThreshold_Escalates(t *testing.T) {
	e := NewEscalation(EscalationConfig{Window: time.Minute, Threshold: 3, EscalateLevel: alert.LevelCritical})
	e.Check("tok-1", alert.LevelWarning)
	e.Check("tok-1", alert.LevelWarning)
	level := e.Check("tok-1", alert.LevelWarning)
	if level != alert.LevelCritical {
		t.Fatalf("expected critical after threshold, got %v", level)
	}
}

func TestEscalation_Check_AfterWindow_Resets(t *testing.T) {
	now := time.Now()
	e := NewEscalation(EscalationConfig{Window: time.Minute, Threshold: 2, EscalateLevel: alert.LevelCritical})
	e.now = func() time.Time { return now }
	e.Check("tok-1", alert.LevelWarning)
	e.Check("tok-1", alert.LevelWarning)
	e.now = func() time.Time { return now.Add(2 * time.Minute) }
	level := e.Check("tok-1", alert.LevelWarning)
	if level != alert.LevelWarning {
		t.Fatalf("expected reset after window, got %v", level)
	}
}

func TestEscalation_Reset_ClearsState(t *testing.T) {
	e := NewEscalation(EscalationConfig{Window: time.Minute, Threshold: 2, EscalateLevel: alert.LevelCritical})
	e.Check("tok-1", alert.LevelWarning)
	e.Check("tok-1", alert.LevelWarning)
	e.Reset("tok-1")
	level := e.Check("tok-1", alert.LevelWarning)
	if level != alert.LevelWarning {
		t.Fatalf("expected warning after reset, got %v", level)
	}
}

func TestEscalation_DifferentTokens_Independent(t *testing.T) {
	e := NewEscalation(EscalationConfig{Window: time.Minute, Threshold: 2, EscalateLevel: alert.LevelCritical})
	e.Check("tok-1", alert.LevelWarning)
	e.Check("tok-1", alert.LevelWarning)
	level := e.Check("tok-2", alert.LevelWarning)
	if level != alert.LevelWarning {
		t.Fatalf("expected tok-2 unaffected, got %v", level)
	}
}
