package tokenwatch

import (
	"errors"
	"testing"
	"time"
)

func newTestEntropyScanner(lookup func(string) (TokenInfo, error)) *EntropyScanner {
	r := NewRegistry()
	cfg := DefaultEntropyConfig()
	cfg.MinSamples = 2
	cfg.CriticalThreshold = 0.99 // force alert on identical samples
	d := NewEntropyDetector(cfg)
	return NewEntropyScanner(r, d, lookup)
}

func TestNewEntropyScanner_NilRegistry_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil registry")
		}
	}()
	NewEntropyScanner(nil, NewEntropyDetector(DefaultEntropyConfig()), func(string) (TokenInfo, error) { return TokenInfo{}, nil })
}

func TestNewEntropyScanner_NilDetector_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil detector")
		}
	}()
	NewEntropyScanner(NewRegistry(), nil, func(string) (TokenInfo, error) { return TokenInfo{}, nil })
}

func TestNewEntropyScanner_NilLookup_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil lookup")
		}
	}()
	NewEntropyScanner(NewRegistry(), NewEntropyDetector(DefaultEntropyConfig()), nil)
}

func TestEntropyScanner_Scan_NoTokens_ReturnsEmpty(t *testing.T) {
	s := newTestEntropyScanner(func(string) (TokenInfo, error) {
		return TokenInfo{TTL: 30 * time.Minute}, nil
	})
	alerts, err := s.Scan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(alerts) != 0 {
		t.Fatalf("expected no alerts, got %d", len(alerts))
	}
}

func TestEntropyScanner_Scan_LookupError_SkipsToken(t *testing.T) {
	r := NewRegistry()
	_ = r.Add("bad-token")
	cfg := DefaultEntropyConfig()
	cfg.MinSamples = 1
	d := NewEntropyDetector(cfg)
	s := NewEntropyScanner(r, d, func(string) (TokenInfo, error) {
		return TokenInfo{}, errors.New("vault unavailable")
	})
	alerts, err := s.Scan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(alerts) != 0 {
		t.Fatalf("expected no alerts when lookup fails, got %d", len(alerts))
	}
}

func TestEntropyScanner_Scan_DetectsLowEntropy(t *testing.T) {
	r := NewRegistry()
	_ = r.Add("tok-a")

	cfg := DefaultEntropyConfig()
	cfg.MinSamples = 2
	cfg.CriticalThreshold = 0.99
	d := NewEntropyDetector(cfg)

	lookup := func(string) (TokenInfo, error) {
		return TokenInfo{TTL: 30 * time.Minute}, nil
	}
	s := NewEntropyScanner(r, d, lookup)

	// First scan seeds the sample; second scan should trigger alert.
	_, _ = s.Scan()
	alerts, err := s.Scan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(alerts) == 0 {
		t.Fatal("expected at least one low-entropy alert")
	}
	if alerts[0].LeaseID != "tok-a" {
		t.Errorf("expected LeaseID 'tok-a', got %q", alerts[0].LeaseID)
	}
}
