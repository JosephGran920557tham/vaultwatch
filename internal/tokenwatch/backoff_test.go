package tokenwatch

import (
	"testing"
	"time"
)

func TestDefaultBackoffConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultBackoffConfig()
	if cfg.InitialInterval <= 0 {
		t.Fatal("expected positive InitialInterval")
	}
	if cfg.MaxInterval < cfg.InitialInterval {
		t.Fatal("expected MaxInterval >= InitialInterval")
	}
	if cfg.Multiplier <= 1.0 {
		t.Fatal("expected Multiplier > 1.0")
	}
	if cfg.MaxRetries <= 0 {
		t.Fatal("expected positive MaxRetries")
	}
}

func TestNewBackoff_InvalidInitialInterval(t *testing.T) {
	cfg := DefaultBackoffConfig()
	cfg.InitialInterval = 0
	_, err := NewBackoff(cfg)
	if err == nil {
		t.Fatal("expected error for zero InitialInterval")
	}
}

func TestNewBackoff_MaxIntervalBelowInitial(t *testing.T) {
	cfg := DefaultBackoffConfig()
	cfg.MaxInterval = cfg.InitialInterval - time.Millisecond
	_, err := NewBackoff(cfg)
	if err == nil {
		t.Fatal("expected error when MaxInterval < InitialInterval")
	}
}

func TestNewBackoff_MultiplierTooLow(t *testing.T) {
	cfg := DefaultBackoffConfig()
	cfg.Multiplier = 1.0
	_, err := NewBackoff(cfg)
	if err == nil {
		t.Fatal("expected error for Multiplier <= 1.0")
	}
}

func TestBackoff_Next_IncreasesDelay(t *testing.T) {
	cfg := DefaultBackoffConfig()
	b, err := NewBackoff(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	d1, ok1 := b.Next()
	d2, ok2 := b.Next()

	if !ok1 || !ok2 {
		t.Fatal("expected ok=true for first two attempts")
	}
	if d2 <= d1 {
		t.Errorf("expected d2 > d1, got d1=%v d2=%v", d1, d2)
	}
}

func TestBackoff_Next_CapsAtMaxInterval(t *testing.T) {
	cfg := BackoffConfig{
		InitialInterval: 100 * time.Millisecond,
		MaxInterval:     200 * time.Millisecond,
		Multiplier:      10.0,
		MaxRetries:      0,
	}
	b, err := NewBackoff(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i := 0; i < 5; i++ {
		d, ok := b.Next()
		if !ok {
			t.Fatal("expected ok=true with unlimited retries")
		}
		if d > cfg.MaxInterval {
			t.Errorf("delay %v exceeded MaxInterval %v", d, cfg.MaxInterval)
		}
	}
}

func TestBackoff_Next_StopsAfterMaxRetries(t *testing.T) {
	cfg := DefaultBackoffConfig()
	cfg.MaxRetries = 3
	b, err := NewBackoff(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i := 0; i < cfg.MaxRetries; i++ {
		_, ok := b.Next()
		if !ok {
			t.Fatalf("expected ok=true on attempt %d", i)
		}
	}
	_, ok := b.Next()
	if ok {
		t.Fatal("expected ok=false after MaxRetries exceeded")
	}
}

func TestBackoff_Reset_RestoresAttempt(t *testing.T) {
	cfg := DefaultBackoffConfig()
	b, _ := NewBackoff(cfg)

	b.Next()
	b.Next()
	if b.Attempt() != 2 {
		t.Fatalf("expected attempt=2, got %d", b.Attempt())
	}

	b.Reset()
	if b.Attempt() != 0 {
		t.Fatalf("expected attempt=0 after reset, got %d", b.Attempt())
	}
}
