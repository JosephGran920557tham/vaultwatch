package tokenwatch

import (
	"context"
	"errors"
	"testing"
	"time"
)

func newTestWatermarkScanner() (*WatermarkScanner, *Registry) {
	reg := NewRegistry()
	det := NewWatermarkDetector(WatermarkConfig{
		LowWatermark:  5 * time.Minute,
		HighWatermark: 30 * time.Minute,
	})
	lookup := func(_ context.Context, id string) (*TokenInfo, error) {
		return &TokenInfo{ID: id, TTL: 2 * time.Minute}, nil
	}
	return NewWatermarkScanner(reg, det, lookup), reg
}

func TestNewWatermarkScanner_NilRegistry_Panics(t *testing.T) {
	defer func() { recover() }()
	NewWatermarkScanner(nil, NewWatermarkDetector(WatermarkConfig{}), func(_ context.Context, _ string) (*TokenInfo, error) { return nil, nil })
	t.Fatal("expected panic")
}

func TestNewWatermarkScanner_NilDetector_Panics(t *testing.T) {
	defer func() { recover() }()
	NewWatermarkScanner(NewRegistry(), nil, func(_ context.Context, _ string) (*TokenInfo, error) { return nil, nil })
	t.Fatal("expected panic")
}

func TestNewWatermarkScanner_NilLookup_Panics(t *testing.T) {
	defer func() { recover() }()
	NewWatermarkScanner(NewRegistry(), NewWatermarkDetector(WatermarkConfig{}), nil)
	t.Fatal("expected panic")
}

func TestWatermarkScanner_Scan_NoTokens_ReturnsEmpty(t *testing.T) {
	s, _ := newTestWatermarkScanner()
	results := s.Scan(context.Background())
	if len(results) != 0 {
		t.Fatalf("expected empty, got %d", len(results))
	}
}

func TestWatermarkScanner_Scan_AlertWhenBelowLow(t *testing.T) {
	reg := NewRegistry()
	_ = reg.Add("tok-low")
	det := NewWatermarkDetector(WatermarkConfig{
		LowWatermark:  5 * time.Minute,
		HighWatermark: 30 * time.Minute,
	})
	// Pre-establish a high peak so detector fires.
	det.Record("tok-low", 60*time.Minute)
	lookup := func(_ context.Context, id string) (*TokenInfo, error) {
		return &TokenInfo{ID: id, TTL: 1 * time.Minute}, nil
	}
	s := NewWatermarkScanner(reg, det, lookup)
	alerts := s.Scan(context.Background())
	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert got %d", len(alerts))
	}
}

func TestWatermarkScanner_Scan_LookupError_Skipped(t *testing.T) {
	reg := NewRegistry()
	_ = reg.Add("tok-err")
	det := NewWatermarkDetector(WatermarkConfig{})
	lookup := func(_ context.Context, _ string) (*TokenInfo, error) {
		return nil, errors.New("vault unavailable")
	}
	s := NewWatermarkScanner(reg, det, lookup)
	if alerts := s.Scan(context.Background()); len(alerts) != 0 {
		t.Fatalf("expected 0 alerts got %d", len(alerts))
	}
}
