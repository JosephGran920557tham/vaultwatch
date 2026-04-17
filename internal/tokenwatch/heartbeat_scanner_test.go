package tokenwatch

import (
	"errors"
	"testing"
	"time"
)

func newTestHeartbeatScanner(lookup func(string) (*TokenInfo, error)) (*HeartbeatScanner, *Registry) {
	r := NewRegistry()
	cfg := HeartbeatConfig{StaleAfter: 1 * time.Minute, CriticalAfter: 5 * time.Minute}
	d := NewHeartbeatDetector(cfg)
	return NewHeartbeatScanner(r, d, lookup), r
}

func TestNewHeartbeatScanner_NilRegistry_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil registry")
		}
	}()
	d := NewHeartbeatDetector(DefaultHeartbeatConfig())
	NewHeartbeatScanner(nil, d, func(string) (*TokenInfo, error) { return nil, nil })
}

func TestNewHeartbeatScanner_NilDetector_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil detector")
		}
	}()
	NewHeartbeatScanner(NewRegistry(), nil, func(string) (*TokenInfo, error) { return nil, nil })
}

func TestNewHeartbeatScanner_NilLookup_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil lookup")
		}
	}()
	d := NewHeartbeatDetector(DefaultHeartbeatConfig())
	NewHeartbeatScanner(NewRegistry(), d, nil)
}

func TestHeartbeatScanner_Scan_NoTokens_ReturnsEmpty(t *testing.T) {
	s, _ := newTestHeartbeatScanner(func(string) (*TokenInfo, error) { return &TokenInfo{}, nil })
	results := s.Scan()
	if len(results) != 0 {
		t.Fatalf("expected empty results, got %d", len(results))
	}
}

func TestHeartbeatScanner_Scan_LookupSuccess_NoAlert(t *testing.T) {
	s, r := newTestHeartbeatScanner(func(string) (*TokenInfo, error) {
		return &TokenInfo{}, nil
	})
	_ = r.Add("tok-ok")
	alerts := s.Scan()
	if len(alerts) != 0 {
		t.Fatalf("expected no alerts after successful lookup, got %d", len(alerts))
	}
}

func TestHeartbeatScanner_Scan_LookupError_RaisesAlert(t *testing.T) {
	s, r := newTestHeartbeatScanner(func(string) (*TokenInfo, error) {
		return nil, errors.New("vault unavailable")
	})
	_ = r.Add("tok-fail")
	alerts := s.Scan()
	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(alerts))
	}
	if alerts[0].LeaseID != "tok-fail" {
		t.Errorf("unexpected lease id %q", alerts[0].LeaseID)
	}
}
