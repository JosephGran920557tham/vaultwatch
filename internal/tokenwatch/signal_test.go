package tokenwatch

import (
	"testing"
	"time"

	"github.com/vaultwatch/internal/alert"
)

func TestDefaultSignalConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultSignalConfig()
	if cfg.MinStrength <= 0 {
		t.Errorf("expected positive MinStrength, got %d", cfg.MinStrength)
	}
	if cfg.DecayWindow <= 0 {
		t.Errorf("expected positive DecayWindow, got %v", cfg.DecayWindow)
	}
}

func TestNewSignalAggregator_ZeroValues_UsesDefaults(t *testing.T) {
	agg := NewSignalAggregator(SignalConfig{})
	def := DefaultSignalConfig()
	if agg.cfg.MinStrength != def.MinStrength {
		t.Errorf("expected MinStrength %d, got %d", def.MinStrength, agg.cfg.MinStrength)
	}
	if agg.cfg.DecayWindow != def.DecayWindow {
		t.Errorf("expected DecayWindow %v, got %v", def.DecayWindow, agg.cfg.DecayWindow)
	}
}

func TestSignalAggregator_Observe_BelowThreshold_ReturnsFalse(t *testing.T) {
	agg := NewSignalAggregator(SignalConfig{MinStrength: 3, DecayWindow: time.Minute})
	result := agg.Observe("tok-1", alert.LevelWarning)
	if result {
		t.Error("expected false below threshold")
	}
}

func TestSignalAggregator_Observe_ReachesThreshold_ReturnsTrue(t *testing.T) {
	agg := NewSignalAggregator(SignalConfig{MinStrength: 2, DecayWindow: time.Minute})
	agg.Observe("tok-1", alert.LevelWarning)
	result := agg.Observe("tok-1", alert.LevelWarning)
	if !result {
		t.Error("expected true at threshold")
	}
}

func TestSignalAggregator_Reset_ClearsStrength(t *testing.T) {
	agg := NewSignalAggregator(SignalConfig{MinStrength: 2, DecayWindow: time.Minute})
	agg.Observe("tok-1", alert.LevelWarning)
	agg.Reset("tok-1")
	if s := agg.Strength("tok-1"); s != 0 {
		t.Errorf("expected 0 after reset, got %d", s)
	}
}

func TestSignalAggregator_Strength_DecaysAfterWindow(t *testing.T) {
	agg := NewSignalAggregator(SignalConfig{MinStrength: 2, DecayWindow: time.Millisecond})
	agg.Observe("tok-1", alert.LevelWarning)
	time.Sleep(5 * time.Millisecond)
	if s := agg.Strength("tok-1"); s != 0 {
		t.Errorf("expected 0 after decay, got %d", s)
	}
}

func TestSignalAggregator_DifferentTokensAreIndependent(t *testing.T) {
	agg := NewSignalAggregator(SignalConfig{MinStrength: 2, DecayWindow: time.Minute})
	agg.Observe("tok-a", alert.LevelCritical)
	if s := agg.Strength("tok-b"); s != 0 {
		t.Errorf("expected 0 for tok-b, got %d", s)
	}
}
