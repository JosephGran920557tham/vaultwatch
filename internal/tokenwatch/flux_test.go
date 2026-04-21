package tokenwatch

import (
	"testing"
	"time"
)

func TestDefaultFluxConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultFluxConfig()
	if cfg.Window <= 0 {
		t.Fatalf("expected positive Window, got %v", cfg.Window)
	}
	if cfg.MaxSamples <= 0 {
		t.Fatalf("expected positive MaxSamples, got %d", cfg.MaxSamples)
	}
}

func TestNewFluxDetector_ZeroValues_UsesDefaults(t *testing.T) {
	d := NewFluxDetector(FluxConfig{})
	def := DefaultFluxConfig()
	if d.cfg.Window != def.Window {
		t.Errorf("expected Window %v, got %v", def.Window, d.cfg.Window)
	}
	if d.cfg.MaxSamples != def.MaxSamples {
		t.Errorf("expected MaxSamples %d, got %d", def.MaxSamples, d.cfg.MaxSamples)
	}
}

func TestFluxDetector_Flux_InsufficientSamples_ReturnsZero(t *testing.T) {
	d := NewFluxDetector(DefaultFluxConfig())
	if got := d.Flux("tok1"); got != 0 {
		t.Errorf("expected 0, got %v", got)
	}
	d.Record("tok1", 10*time.Minute)
	if got := d.Flux("tok1"); got != 0 {
		t.Errorf("expected 0 with single sample, got %v", got)
	}
}

func TestFluxDetector_Flux_ReturnsAbsoluteDelta(t *testing.T) {
	d := NewFluxDetector(DefaultFluxConfig())
	d.Record("tok2", 10*time.Minute)
	d.Record("tok2", 6*time.Minute)
	got := d.Flux("tok2")
	if got != 4*time.Minute {
		t.Errorf("expected 4m flux, got %v", got)
	}
}

func TestFluxDetector_Flux_DifferentTokensIndependent(t *testing.T) {
	d := NewFluxDetector(DefaultFluxConfig())
	d.Record("a", 20*time.Minute)
	d.Record("a", 10*time.Minute)
	d.Record("b", 5*time.Minute)
	d.Record("b", 5*time.Minute)

	if got := d.Flux("a"); got != 10*time.Minute {
		t.Errorf("token a: expected 10m, got %v", got)
	}
	if got := d.Flux("b"); got != 0 {
		t.Errorf("token b: expected 0, got %v", got)
	}
}

func TestFluxDetector_Record_CapsAtMaxSamples(t *testing.T) {
	cfg := FluxConfig{Window: time.Hour, MaxSamples: 3}
	d := NewFluxDetector(cfg)
	for i := 0; i < 10; i++ {
		d.Record("tok", time.Duration(i)*time.Minute)
	}
	d.mu.Lock()
	count := len(d.samples["tok"])
	d.mu.Unlock()
	if count > 3 {
		t.Errorf("expected at most 3 samples, got %d", count)
	}
}
