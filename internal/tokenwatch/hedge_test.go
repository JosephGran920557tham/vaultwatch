package tokenwatch

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestDefaultHedgeConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultHedgeConfig()
	if cfg.Delay <= 0 {
		t.Fatal("expected positive delay")
	}
	if cfg.MaxHedges <= 0 {
		t.Fatal("expected positive MaxHedges")
	}
}

func TestHedge_SucceedsOnFirstAttempt(t *testing.T) {
	cfg := HedgeConfig{Delay: 5 * time.Millisecond, MaxHedges: 2}
	v, err := Hedge(cfg, func() (interface{}, error) {
		return "ok", nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != "ok" {
		t.Fatalf("unexpected value: %v", v)
	}
}

func TestHedge_ReturnsFirstSuccess(t *testing.T) {
	var calls int64
	cfg := HedgeConfig{Delay: 5 * time.Millisecond, MaxHedges: 2}
	v, err := Hedge(cfg, func() (interface{}, error) {
		n := atomic.AddInt64(&calls, 1)
		if n == 1 {
			time.Sleep(20 * time.Millisecond)
			return nil, errors.New("slow")
		}
		return "fast", nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != "fast" {
		t.Fatalf("expected fast, got %v", v)
	}
}

func TestHedge_AllFail_ReturnsError(t *testing.T) {
	cfg := HedgeConfig{Delay: 2 * time.Millisecond, MaxHedges: 1}
	_, err := Hedge(cfg, func() (interface{}, error) {
		return nil, errors.New("boom")
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNewHedgeScanner_NilRegistry_Panics(t *testing.T) {
	defer func() { recover() }()
	NewHedgeScanner(nil, func(string) (TokenInfo, error) { return TokenInfo{}, nil }, DefaultHedgeConfig())
	t.Fatal("expected panic")
}

func TestNewHedgeScanner_NilLookup_Panics(t *testing.T) {
	defer func() { recover() }()
	NewHedgeScanner(NewRegistry(), nil, DefaultHedgeConfig())
	t.Fatal("expected panic")
}

func TestHedgeScanner_Scan_NoTokens_ReturnsEmpty(t *testing.T) {
	s := NewHedgeScanner(NewRegistry(), func(string) (TokenInfo, error) {
		return TokenInfo{}, nil
	}, DefaultHedgeConfig())
	if got := s.Scan(); len(got) != 0 {
		t.Fatalf("expected empty, got %d", len(got))
	}
}

func TestHedgeScanner_Scan_ErrorProducesAlert(t *testing.T) {
	reg := NewRegistry()
	_ = reg.Add("tok-1")
	cfg := HedgeConfig{Delay: 2 * time.Millisecond, MaxHedges: 1}
	s := NewHedgeScanner(reg, func(string) (TokenInfo, error) {
		return TokenInfo{}, errors.New("unreachable")
	}, cfg)
	alerts := s.Scan()
	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(alerts))
	}
}
