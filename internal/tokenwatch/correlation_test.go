package tokenwatch

import (
	"testing"
	"time"

	"github.com/vaultwatch/internal/alert"
)

func TestDefaultCorrelationConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultCorrelationConfig()
	if cfg.Window <= 0 {
		t.Fatal("expected positive Window")
	}
	if cfg.MinEvents <= 0 {
		t.Fatal("expected positive MinEvents")
	}
	if cfg.ScoreThreshold <= 0 {
		t.Fatal("expected positive ScoreThreshold")
	}
}

func TestNewCorrelationDetector_ZeroValues_UsesDefaults(t *testing.T) {
	d := NewCorrelationDetector(CorrelationConfig{})
	def := DefaultCorrelationConfig()
	if d.cfg.Window != def.Window {
		t.Errorf("window: got %v want %v", d.cfg.Window, def.Window)
	}
}

func TestCorrelationDetector_Check_InsufficientEvents_ReturnsNil(t *testing.T) {
	d := NewCorrelationDetector(CorrelationConfig{
		Window:         time.Minute,
		MinEvents:      5,
		ScoreThreshold: 0.75,
	})
	d.Record("tok-1")
	d.Record("tok-1")
	got := d.Check("tok-1", alert.LevelWarning)
	if got != nil {
		t.Fatalf("expected nil alert, got %v", got)
	}
}

func TestCorrelationDetector_Check_SufficientEvents_ReturnsAlert(t *testing.T) {
	d := NewCorrelationDetector(CorrelationConfig{
		Window:         time.Minute,
		MinEvents:      3,
		ScoreThreshold: 0.75,
	})
	for i := 0; i < 4; i++ {
		d.Record("tok-2")
	}
	got := d.Check("tok-2", alert.LevelWarning)
	if got == nil {
		t.Fatal("expected alert, got nil")
	}
	if got.LeaseID != "tok-2" {
		t.Errorf("LeaseID: got %q want %q", got.LeaseID, "tok-2")
	}
}

func TestCorrelationDetector_Check_PrunesOldEvents(t *testing.T) {
	now := time.Now()
	d := NewCorrelationDetector(CorrelationConfig{
		Window:         10 * time.Second,
		MinEvents:      2,
		ScoreThreshold: 0.75,
	})
	// inject stale timestamps directly
	d.mu.Lock()
	d.bucket["tok-3"] = []time.Time{
		now.Add(-30 * time.Second),
		now.Add(-20 * time.Second),
	}
	d.mu.Unlock()
	got := d.Check("tok-3", alert.LevelInfo)
	if got != nil {
		t.Fatalf("expected nil after pruning stale events, got %v", got)
	}
}

func TestNewCorrelationScanner_NilRegistry_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil registry")
		}
	}()
	NewCorrelationScanner(nil, NewCorrelationDetector(CorrelationConfig{}), func(string) (TokenInfo, error) { return TokenInfo{}, nil })
}

func TestNewCorrelationScanner_NilDetector_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil detector")
		}
	}()
	NewCorrelationScanner(NewRegistry(), nil, func(string) (TokenInfo, error) { return TokenInfo{}, nil })
}

func TestCorrelationScanner_Scan_NoTokens_ReturnsEmpty(t *testing.T) {
	s := NewCorrelationScanner(
		NewRegistry(),
		NewCorrelationDetector(CorrelationConfig{}),
		func(string) (TokenInfo, error) { return TokenInfo{}, nil },
	)
	if got := s.Scan(); len(got) != 0 {
		t.Fatalf("expected empty, got %d alerts", len(got))
	}
}
