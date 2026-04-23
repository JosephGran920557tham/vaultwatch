package tokenwatch

import (
	"context"
	"testing"
	"time"
)

func newTestFluxScanner(t *testing.T) (*FluxScanner, *Registry, *FluxDetector) {
	t.Helper()
	reg := NewRegistry()
	cfg := DefaultFluxConfig()
	det := NewFluxDetector(cfg)
	lookup := func(_ context.Context, id string) (TokenInfo, error) {
		return TokenInfo{ID: id, TTL: 10 * time.Minute}, nil
	}
	s := NewFluxScanner(reg, det, lookup)
	return s, reg, det
}

func TestNewFluxScanner_NilRegistry_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil registry")
		}
	}()
	cfg := DefaultFluxConfig()
	det := NewFluxDetector(cfg)
	NewFluxScanner(nil, det, func(_ context.Context, id string) (TokenInfo, error) {
		return TokenInfo{}, nil
	})
}

func TestNewFluxScanner_NilDetector_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil detector")
		}
	}()
	reg := NewRegistry()
	NewFluxScanner(reg, nil, func(_ context.Context, id string) (TokenInfo, error) {
		return TokenInfo{}, nil
	})
}

func TestNewFluxScanner_NilLookup_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil lookup")
		}
	}()
	reg := NewRegistry()
	cfg := DefaultFluxConfig()
	det := NewFluxDetector(cfg)
	NewFluxScanner(reg, det, nil)
}

func TestFluxScanner_Scan_NoTokens_ReturnsEmpty(t *testing.T) {
	s, _, _ := newTestFluxScanner(t)
	alerts, err := s.Scan(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(alerts) != 0 {
		t.Fatalf("expected 0 alerts, got %d", len(alerts))
	}
}

func TestFluxScanner_Scan_InsufficientSamples_NoAlert(t *testing.T) {
	s, reg, _ := newTestFluxScanner(t)
	_ = reg.Add("tok-1")
	alerts, err := s.Scan(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(alerts) != 0 {
		t.Fatalf("expected 0 alerts with only 1 sample, got %d", len(alerts))
	}
}

func TestFluxScanner_Scan_DetectsFlux(t *testing.T) {
	reg := NewRegistry()
	_ = reg.Add("tok-flux")

	cfg := DefaultFluxConfig()
	cfg.WarningFlux = 1 * time.Second
	cfg.CriticalFlux = 5 * time.Minute
	det := NewFluxDetector(cfg)

	// Pre-seed detector with oscillating samples.
	det.Record("tok-flux", 10*time.Minute)
	det.Record("tok-flux", 1*time.Second)

	call := 0
	lookup := func(_ context.Context, id string) (TokenInfo, error) {
		call++
		ttl := 10 * time.Minute
		if call%2 == 0 {
			ttl = 1 * time.Second
		}
		return TokenInfo{ID: id, TTL: ttl}, nil
	}

	s := NewFluxScanner(reg, det, lookup)
	alerts, err := s.Scan(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(alerts) == 0 {
		t.Fatal("expected at least one flux alert")
	}
	if alerts[0].LeaseID != "tok-flux" {
		t.Errorf("expected lease id tok-flux, got %s", alerts[0].LeaseID)
	}
}
