package tokenwatch

import (
	"errors"
	"testing"
	"time"
)

func newTestSentinelScanner(lookup func(string) (TokenInfo, error)) (*SentinelScanner, *Registry) {
	reg := NewRegistry()
	det := NewSentinelDetector(SentinelConfig{MissWindow: time.Millisecond, MissThreshold: 1})
	return NewSentinelScanner(reg, det, lookup), reg
}

func TestNewSentinelScanner_NilRegistry_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil registry")
		}
	}()
	NewSentinelScanner(nil, NewSentinelDetector(SentinelConfig{}), func(string) (TokenInfo, error) { return TokenInfo{}, nil })
}

func TestNewSentinelScanner_NilDetector_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil detector")
		}
	}()
	NewSentinelScanner(NewRegistry(), nil, func(string) (TokenInfo, error) { return TokenInfo{}, nil })
}

func TestNewSentinelScanner_NilLookup_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil lookup")
		}
	}()
	NewSentinelScanner(NewRegistry(), NewSentinelDetector(SentinelConfig{}), nil)
}

func TestSentinelScanner_Scan_NoTokens_ReturnsEmpty(t *testing.T) {
	scanner, _ := newTestSentinelScanner(func(string) (TokenInfo, error) { return TokenInfo{}, nil })
	alerts := scanner.Scan()
	if len(alerts) != 0 {
		t.Errorf("expected 0 alerts, got %d", len(alerts))
	}
}

func TestSentinelScanner_Scan_PingsReachableTokens(t *testing.T) {
	scanner, reg := newTestSentinelScanner(func(string) (TokenInfo, error) {
		return TokenInfo{TTL: 300}, nil
	})
	_ = reg.Add("tok-ok")
	alerts := scanner.Scan()
	if len(alerts) != 0 {
		t.Errorf("expected 0 alerts for reachable token, got %d", len(alerts))
	}
}

func TestSentinelScanner_Scan_AlertsForMissingToken(t *testing.T) {
	lookupErr := errors.New("not found")
	scanner, reg := newTestSentinelScanner(func(string) (TokenInfo, error) {
		return TokenInfo{}, lookupErr
	})
	_ = reg.Add("tok-gone")
	// first scan: initialises entry, no miss yet
	scanner.Scan()
	// advance detector clock by sleeping past MissWindow
	time.Sleep(3 * time.Millisecond)
	alerts := scanner.Scan()
	if len(alerts) == 0 {
		t.Error("expected at least one alert for missing token")
	}
}
