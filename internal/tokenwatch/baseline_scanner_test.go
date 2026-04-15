package tokenwatch

import (
	"errors"
	"testing"
	"time"
)

func newTestBaselineScanner(lookup func(string) (*TokenInfo, error)) *BaselineScanner {
	reg := NewRegistry()
	det := NewBaselineDetector(BaselineConfig{SampleWindow: 5, DeviationPct: 25})
	return NewBaselineScanner(reg, det, lookup)
}

func TestNewBaselineScanner_NilRegistry_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil registry")
		}
	}()
	NewBaselineScanner(nil, NewBaselineDetector(DefaultBaselineConfig()), func(string) (*TokenInfo, error) { return nil, nil })
}

func TestNewBaselineScanner_NilDetector_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil detector")
		}
	}()
	NewBaselineScanner(NewRegistry(), nil, func(string) (*TokenInfo, error) { return nil, nil })
}

func TestNewBaselineScanner_NilLookup_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil lookup")
		}
	}()
	NewBaselineScanner(NewRegistry(), NewBaselineDetector(DefaultBaselineConfig()), nil)
}

func TestBaselineScanner_Scan_NoTokens_ReturnsEmpty(t *testing.T) {
	s := newTestBaselineScanner(func(id string) (*TokenInfo, error) {
		return &TokenInfo{ID: id, TTL: 60 * time.Minute}, nil
	})
	alerts := s.Scan()
	if len(alerts) != 0 {
		t.Errorf("expected no alerts for empty registry, got %d", len(alerts))
	}
}

func TestBaselineScanner_Scan_LookupError_Skips(t *testing.T) {
	reg := NewRegistry()
	_ = reg.Add("bad-token")
	det := NewBaselineDetector(DefaultBaselineConfig())
	s := NewBaselineScanner(reg, det, func(string) (*TokenInfo, error) {
		return nil, errors.New("vault unreachable")
	})
	alerts := s.Scan()
	if len(alerts) != 0 {
		t.Errorf("expected no alerts when lookup errors, got %d", len(alerts))
	}
}

func TestBaselineScanner_Scan_DetectsDeviation(t *testing.T) {
	reg := NewRegistry()
	_ = reg.Add("dev-token")
	det := NewBaselineDetector(BaselineConfig{SampleWindow: 5, DeviationPct: 25})

	// Seed detector with high-TTL samples
	for i := 0; i < 5; i++ {
		det.Record("dev-token", 60*time.Minute)
	}

	s := NewBaselineScanner(reg, det, func(id string) (*TokenInfo, error) {
		return &TokenInfo{ID: id, TTL: 10 * time.Minute}, nil
	})

	alerts := s.Scan()
	if len(alerts) == 0 {
		t.Error("expected at least one baseline-deviation alert")
	}
	if alerts[0].LeaseID != "dev-token" {
		t.Errorf("expected LeaseID dev-token, got %s", alerts[0].LeaseID)
	}
}
