package tokenwatch

import (
	"testing"
	"time"
)

func TestDefaultWatermarkConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultWatermarkConfig()
	if cfg.LowWatermark <= 0 {
		t.Fatal("expected positive LowWatermark")
	}
	if cfg.HighWatermark <= cfg.LowWatermark {
		t.Fatal("expected HighWatermark > LowWatermark")
	}
}

func TestNewWatermarkDetector_ZeroValues_UsesDefaults(t *testing.T) {
	d := NewWatermarkDetector(WatermarkConfig{})
	def := DefaultWatermarkConfig()
	if d.cfg.LowWatermark != def.LowWatermark {
		t.Fatalf("expected %v got %v", def.LowWatermark, d.cfg.LowWatermark)
	}
}

func TestWatermarkDetector_Check_NoPeak_ReturnsNil(t *testing.T) {
	d := NewWatermarkDetector(WatermarkConfig{})
	// No Record call, so peak is zero — below HighWatermark, no alert.
	if a := d.Check("tok1", 10*time.Second); a != nil {
		t.Fatal("expected nil alert when peak not established")
	}
}

func TestWatermarkDetector_Check_AboveLow_ReturnsNil(t *testing.T) {
	d := NewWatermarkDetector(WatermarkConfig{
		LowWatermark:  5 * time.Minute,
		HighWatermark: 30 * time.Minute,
	})
	d.Record("tok1", 60*time.Minute)
	if a := d.Check("tok1", 10*time.Minute); a != nil {
		t.Fatal("expected nil when TTL above low watermark")
	}
}

func TestWatermarkDetector_Check_BelowLow_ReturnsWarning(t *testing.T) {
	d := NewWatermarkDetector(WatermarkConfig{
		LowWatermark:  5 * time.Minute,
		HighWatermark: 30 * time.Minute,
	})
	d.Record("tok1", 60*time.Minute)
	a := d.Check("tok1", 2*time.Minute)
	if a == nil {
		t.Fatal("expected warning alert")
	}
	if a.Level != LevelWarning {
		t.Fatalf("expected Warning got %v", a.Level)
	}
	if a.LeaseID != "tok1" {
		t.Fatalf("unexpected lease id %q", a.LeaseID)
	}
}

func TestWatermarkDetector_Record_UpdatesPeak(t *testing.T) {
	d := NewWatermarkDetector(WatermarkConfig{})
	d.Record("tok1", 10*time.Minute)
	d.Record("tok1", 5*time.Minute) // lower — should not replace
	d.mu.Lock()
	peak := d.peaks["tok1"]
	d.mu.Unlock()
	if peak != 10*time.Minute {
		t.Fatalf("expected peak 10m got %v", peak)
	}
}
