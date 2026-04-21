package tokenwatch

import (
	"errors"
	"testing"
	"time"
)

func newTestMirrorScanner(lookup func(string) (TokenInfo, error)) *MirrorScanner {
	reg := NewRegistry()
	_ = reg.Add("tok-1")
	mirror := NewMirror(DefaultMirrorConfig())
	return NewMirrorScanner(reg, mirror, lookup, 30*time.Second)
}

func TestNewMirrorScanner_NilRegistry_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil Registry")
		}
	}()
	NewMirrorScanner(nil, NewMirror(DefaultMirrorConfig()), func(string) (TokenInfo, error) { return TokenInfo{}, nil }, 0)
}

func TestNewMirrorScanner_NilMirror_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil Mirror")
		}
	}()
	NewMirrorScanner(NewRegistry(), nil, func(string) (TokenInfo, error) { return TokenInfo{}, nil }, 0)
}

func TestNewMirrorScanner_NilLookup_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil lookup")
		}
	}()
	NewMirrorScanner(NewRegistry(), NewMirror(DefaultMirrorConfig()), nil, 0)
}

func TestMirrorScanner_Scan_NoTokens_ReturnsEmpty(t *testing.T) {
	reg := NewRegistry()
	mirror := NewMirror(DefaultMirrorConfig())
	s := NewMirrorScanner(reg, mirror, func(string) (TokenInfo, error) { return TokenInfo{}, nil }, 0)
	alerts, err := s.Scan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(alerts) != 0 {
		t.Fatalf("expected no alerts, got %d", len(alerts))
	}
}

func TestMirrorScanner_Scan_NoDrift_ReturnsEmpty(t *testing.T) {
	s := newTestMirrorScanner(func(string) (TokenInfo, error) {
		return TokenInfo{TTL: 10 * time.Minute}, nil
	})
	// first scan seeds the mirror
	_, _ = s.Scan()
	// second scan with same TTL — no drift
	alerts, err := s.Scan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(alerts) != 0 {
		t.Fatalf("expected no alerts, got %d", len(alerts))
	}
}

func TestMirrorScanner_Scan_DetectsDrift(t *testing.T) {
	calls := 0
	s := newTestMirrorScanner(func(string) (TokenInfo, error) {
		calls++
		if calls == 1 {
			return TokenInfo{TTL: 10 * time.Minute}, nil
		}
		return TokenInfo{TTL: 1 * time.Minute}, nil // big drop
	})
	_, _ = s.Scan() // seed
	alerts, err := s.Scan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(alerts) == 0 {
		t.Fatal("expected drift alert")
	}
	if alerts[0].Level != LevelWarning {
		t.Fatalf("expected Warning, got %s", alerts[0].Level)
	}
}

func TestMirrorScanner_Scan_LookupError_Skips(t *testing.T) {
	s := newTestMirrorScanner(func(string) (TokenInfo, error) {
		return TokenInfo{}, errors.New("vault unavailable")
	})
	alerts, err := s.Scan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(alerts) != 0 {
		t.Fatalf("expected no alerts on lookup error, got %d", len(alerts))
	}
}
