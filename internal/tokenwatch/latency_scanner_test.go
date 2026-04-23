package tokenwatch

import (
	"context"
	"errors"
	"testing"
	"time"
)

func newTestLatencyScanner(t *testing.T) (*LatencyScanner, *Registry, *LatencyDetector) {
	t.Helper()
	reg := NewRegistry()
	det := NewLatencyDetector(DefaultLatencyConfig())
	lookup := func(_ context.Context, _ string) (TokenInfo, error) {
		return TokenInfo{TTL: 3600 * time.Second}, nil
	}
	return NewLatencyScanner(reg, det, lookup), reg, det
}

func TestNewLatencyScanner_NilRegistry_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil registry")
		}
	}()
	NewLatencyScanner(nil, NewLatencyDetector(DefaultLatencyConfig()), func(_ context.Context, _ string) (TokenInfo, error) { return TokenInfo{}, nil })
}

func TestNewLatencyScanner_NilDetector_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil detector")
		}
	}()
	NewLatencyScanner(NewRegistry(), nil, func(_ context.Context, _ string) (TokenInfo, error) { return TokenInfo{}, nil })
}

func TestNewLatencyScanner_NilLookup_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil lookup")
		}
	}()
	NewLatencyScanner(NewRegistry(), NewLatencyDetector(DefaultLatencyConfig()), nil)
}

func TestLatencyScanner_Scan_NoTokens_ReturnsEmpty(t *testing.T) {
	scanner, _, _ := newTestLatencyScanner(t)
	alerts := scanner.Scan(context.Background())
	if len(alerts) != 0 {
		t.Fatalf("expected no alerts, got %d", len(alerts))
	}
}

func TestLatencyScanner_Scan_LookupError_Skipped(t *testing.T) {
	reg := NewRegistry()
	_ = reg.Add("tok-err")
	det := NewLatencyDetector(DefaultLatencyConfig())
	lookup := func(_ context.Context, _ string) (TokenInfo, error) {
		return TokenInfo{}, errors.New("vault unavailable")
	}
	scanner := NewLatencyScanner(reg, det, lookup)
	alerts := scanner.Scan(context.Background())
	if len(alerts) != 0 {
		t.Fatalf("expected no alerts on lookup error, got %d", len(alerts))
	}
}

func TestLatencyScanner_Scan_RecordsAndReturnsAlert(t *testing.T) {
	reg := NewRegistry()
	_ = reg.Add("tok-slow")

	cfg := DefaultLatencyConfig()
	cfg.WarnThreshold = 1 * time.Nanosecond
	cfg.CritThreshold = 1 * time.Hour
	cfg.MinSamples = 1
	det := NewLatencyDetector(cfg)

	lookup := func(_ context.Context, _ string) (TokenInfo, error) {
		time.Sleep(2 * time.Millisecond)
		return TokenInfo{TTL: 3600 * time.Second}, nil
	}
	scanner := NewLatencyScanner(reg, det, lookup)
	alerts := scanner.Scan(context.Background())
	if len(alerts) == 0 {
		t.Fatal("expected at least one latency alert")
	}
}
