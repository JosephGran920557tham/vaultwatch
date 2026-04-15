package tokenwatch

import (
	"bytes"
	"log"
	"testing"
	"time"
)

func TestDefaultWindowGuardConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultWindowGuardConfig()
	if cfg.WindowSize <= 0 {
		t.Errorf("expected positive WindowSize, got %v", cfg.WindowSize)
	}
	if cfg.MaxAlerts <= 0 {
		t.Errorf("expected positive MaxAlerts, got %d", cfg.MaxAlerts)
	}
}

func TestNewWindowGuard_InvalidWindowSize(t *testing.T) {
	_, err := NewWindowGuard(WindowGuardConfig{WindowSize: 0, MaxAlerts: 3}, nil)
	if err == nil {
		t.Fatal("expected error for zero WindowSize")
	}
}

func TestNewWindowGuard_InvalidMaxAlerts(t *testing.T) {
	_, err := NewWindowGuard(WindowGuardConfig{WindowSize: time.Minute, MaxAlerts: 0}, nil)
	if err == nil {
		t.Fatal("expected error for zero MaxAlerts")
	}
}

func TestWindowGuard_Allow_FirstCallPermitted(t *testing.T) {
	g, _ := NewWindowGuard(WindowGuardConfig{WindowSize: time.Minute, MaxAlerts: 2}, nil)
	if !g.Allow("tok-a") {
		t.Fatal("expected first allow")
	}
}

func TestWindowGuard_Allow_ExceedsLimit_Denied(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	g, _ := NewWindowGuard(WindowGuardConfig{WindowSize: time.Minute, MaxAlerts: 1}, logger)

	g.Allow("tok-b")
	if g.Allow("tok-b") {
		t.Fatal("expected second call to be denied")
	}
	if buf.Len() == 0 {
		t.Fatal("expected suppression log entry")
	}
}

func TestWindowGuard_Count_ReflectsAllows(t *testing.T) {
	g, _ := NewWindowGuard(WindowGuardConfig{WindowSize: time.Minute, MaxAlerts: 5}, nil)
	g.Allow("tok-c")
	g.Allow("tok-c")
	if got := g.Count("tok-c"); got != 2 {
		t.Errorf("expected count 2, got %d", got)
	}
}

func TestWindowGuard_Reset_ClearsState(t *testing.T) {
	g, _ := NewWindowGuard(WindowGuardConfig{WindowSize: time.Minute, MaxAlerts: 1}, nil)
	g.Allow("tok-d")
	g.Reset("tok-d")
	if !g.Allow("tok-d") {
		t.Fatal("expected allow after reset")
	}
}

func TestWindowGuard_NilLogger_UsesDefault(t *testing.T) {
	g, err := NewWindowGuard(DefaultWindowGuardConfig(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.logger == nil {
		t.Fatal("expected non-nil logger")
	}
}
