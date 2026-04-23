package tokenwatch

import (
	"testing"
	"time"
)

func TestDefaultAffinityConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultAffinityConfig()
	if cfg.MaxAge <= 0 {
		t.Error("expected positive MaxAge")
	}
	if cfg.MinHits <= 0 {
		t.Error("expected positive MinHits")
	}
	if cfg.DecayRate <= 0 || cfg.DecayRate > 1 {
		t.Errorf("expected DecayRate in (0,1], got %v", cfg.DecayRate)
	}
}

func TestNewAffinityDetector_ZeroValues_UsesDefaults(t *testing.T) {
	d := NewAffinityDetector(AffinityConfig{})
	if d.cfg.MaxAge <= 0 {
		t.Error("expected default MaxAge")
	}
	if d.cfg.MinHits <= 0 {
		t.Error("expected default MinHits")
	}
}

func TestAffinityDetector_Observe_ScoreZeroBeforeMinHits(t *testing.T) {
	d := NewAffinityDetector(AffinityConfig{MaxAge: time.Minute, MinHits: 3, DecayRate: 0.1})
	score := d.Observe("tok-1")
	if score != 0 {
		t.Errorf("expected score 0 before MinHits, got %v", score)
	}
}

func TestAffinityDetector_Observe_ScoreRisesAfterMinHits(t *testing.T) {
	d := NewAffinityDetector(AffinityConfig{MaxAge: time.Minute, MinHits: 3, DecayRate: 0.1})
	var last float64
	for i := 0; i < 3; i++ {
		last = d.Observe("tok-2")
	}
	if last <= 0 {
		t.Errorf("expected positive score after MinHits, got %v", last)
	}
}

func TestAffinityDetector_Score_ReturnsZeroForUnknown(t *testing.T) {
	d := NewAffinityDetector(AffinityConfig{})
	if s := d.Score("unknown"); s != 0 {
		t.Errorf("expected 0 for unknown token, got %v", s)
	}
}

func TestAffinityDetector_Evict_RemovesStaleEntries(t *testing.T) {
	d := NewAffinityDetector(AffinityConfig{MaxAge: time.Millisecond, MinHits: 1, DecayRate: 0.1})
	d.Observe("tok-evict")
	time.Sleep(5 * time.Millisecond)
	d.Evict()
	if s := d.Score("tok-evict"); s != 0 {
		t.Errorf("expected evicted token to have score 0, got %v", s)
	}
}

func TestNewAffinityScanner_NilRegistry_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil registry")
		}
	}()
	NewAffinityScanner(nil, NewAffinityDetector(AffinityConfig{}), func(string) (TokenInfo, error) { return TokenInfo{}, nil }, 0.2)
}

func TestNewAffinityScanner_NilDetector_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil detector")
		}
	}()
	NewAffinityScanner(NewRegistry(), nil, func(string) (TokenInfo, error) { return TokenInfo{}, nil }, 0.2)
}

func TestAffinityScanner_Scan_NoTokens_ReturnsEmpty(t *testing.T) {
	reg := NewRegistry()
	det := NewAffinityDetector(AffinityConfig{MaxAge: time.Minute, MinHits: 1, DecayRate: 0.1})
	s := NewAffinityScanner(reg, det, func(string) (TokenInfo, error) { return TokenInfo{TTL: time.Hour}, nil }, 0.2)
	alerts, err := s.Scan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(alerts) != 0 {
		t.Errorf("expected no alerts for empty registry, got %d", len(alerts))
	}
}
