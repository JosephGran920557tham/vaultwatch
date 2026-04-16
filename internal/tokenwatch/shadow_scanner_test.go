package tokenwatch

import (
	"errors"
	"testing"
	"time"
)

func newTestShadowScanner(lookup func(string) (time.Duration, error)) *ShadowScanner {
	reg := NewRegistry()
	_ = reg.Add("tok1")
	shadow := NewShadowRegistry(DefaultShadowConfig())
	return NewShadowScanner(reg, shadow, lookup, 0.5)
}

func TestNewShadowScanner_NilRegistry_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()
	NewShadowScanner(nil, NewShadowRegistry(DefaultShadowConfig()), func(string) (time.Duration, error) { return 0, nil }, 0.5)
}

func TestNewShadowScanner_NilShadow_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()
	NewShadowScanner(NewRegistry(), nil, func(string) (time.Duration, error) { return 0, nil }, 0.5)
}

func TestNewShadowScanner_NilLookup_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()
	NewShadowScanner(NewRegistry(), NewShadowRegistry(DefaultShadowConfig()), nil, 0.5)
}

func TestShadowScanner_Scan_NoShadow_SetsBaseline(t *testing.T) {
	s := newTestShadowScanner(func(string) (time.Duration, error) {
		return 60 * time.Second, nil
	})
	alerts := s.Scan()
	if len(alerts) != 0 {
		t.Errorf("expected no alerts on first scan, got %d", len(alerts))
	}
}

func TestShadowScanner_Scan_NoAlert_WhenTTLStable(t *testing.T) {
	s := newTestShadowScanner(func(string) (time.Duration, error) {
		return 60 * time.Second, nil
	})
	s.Scan()
	alerts := s.Scan()
	if len(alerts) != 0 {
		t.Errorf("expected no alerts for stable TTL, got %d", len(alerts))
	}
}

func TestShadowScanner_Scan_AlertOnDrop(t *testing.T) {
	call := 0
	s := newTestShadowScanner(func(string) (time.Duration, error) {
		call++
		if call == 1 {
			return 60 * time.Second, nil
		}
		return 10 * time.Second, nil
	})
	s.Scan()
	alerts := s.Scan()
	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(alerts))
	}
	if alerts[0].LeaseID != "tok1" {
		t.Errorf("unexpected lease ID: %s", alerts[0].LeaseID)
	}
}

func TestShadowScanner_Scan_LookupError_Skipped(t *testing.T) {
	s := newTestShadowScanner(func(string) (time.Duration, error) {
		return 0, errors.New("lookup failed")
	})
	alerts := s.Scan()
	if len(alerts) != 0 {
		t.Errorf("expected no alerts on lookup error, got %d", len(alerts))
	}
}
