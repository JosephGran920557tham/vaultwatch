package tokenwatch

import (
	"context"
	"testing"
	"time"
)

func newTestDriftScanner(t *testing.T) (*DriftScanner, *Registry, *DriftDetector) {
	t.Helper()
	reg := NewRegistry()
	det := NewDriftDetector(DefaultDriftConfig())
	lookup := func(_ context.Context, id string) (TokenInfo, error) {
		return TokenInfo{ID: id, TTL: 30 * time.Minute}, nil
	}
	return NewDriftScanner(reg, det, lookup), reg, det
}

func TestNewDriftScanner_NilRegistry_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil registry")
		}
	}()
	det := NewDriftDetector(DefaultDriftConfig())
	NewDriftScanner(nil, det, func(_ context.Context, id string) (TokenInfo, error) {
		return TokenInfo{}, nil
	})
}

func TestNewDriftScanner_NilDetector_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil detector")
		}
	}()
	reg := NewRegistry()
	NewDriftScanner(reg, nil, func(_ context.Context, id string) (TokenInfo, error) {
		return TokenInfo{}, nil
	})
}

func TestNewDriftScanner_NilLookup_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil lookup")
		}
	}()
	reg := NewRegistry()
	det := NewDriftDetector(DefaultDriftConfig())
	NewDriftScanner(reg, det, nil)
}

func TestDriftScanner_Scan_NoTokens_ReturnsEmpty(t *testing.T) {
	s, _, _ := newTestDriftScanner(t)
	alerts, err := s.Scan(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(alerts) != 0 {
		t.Fatalf("expected 0 alerts, got %d", len(alerts))
	}
}

func TestDriftScanner_Scan_NoBaseline_NoAlert(t *testing.T) {
	s, reg, _ := newTestDriftScanner(t)
	_ = reg.Add("tok-a")
	alerts, err := s.Scan(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(alerts) != 0 {
		t.Fatalf("expected 0 alerts before baseline established, got %d", len(alerts))
	}
}

func TestDriftScanner_Scan_DetectsDrift(t *testing.T) {
	reg := NewRegistry()
	_ = reg.Add("tok-drift")

	cfg := DefaultDriftConfig()
	cfg.WarningDelta = 1 * time.Minute
	cfg.CriticalDelta = 10 * time.Minute
	det := NewDriftDetector(cfg)

	// Establish a baseline of 30 minutes.
	det.Check("tok-drift", 30*time.Minute)

	// Lookup now returns a heavily drifted TTL.
	lookup := func(_ context.Context, id string) (TokenInfo, error) {
		return TokenInfo{ID: id, TTL: 5 * time.Minute}, nil
	}
	s := NewDriftScanner(reg, det, lookup)
	alerts, err := s.Scan(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(alerts) == 0 {
		t.Fatal("expected at least one drift alert")
	}
	if alerts[0].LeaseID != "tok-drift" {
		t.Errorf("expected lease id tok-drift, got %s", alerts[0].LeaseID)
	}
	if alerts[0].Level != LevelCritical {
		t.Errorf("expected critical level, got %s", alerts[0].Level)
	}
}
