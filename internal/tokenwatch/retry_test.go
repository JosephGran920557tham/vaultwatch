package tokenwatch

import (
	"errors"
	"testing"
	"time"
)

func TestDefaultRetryConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultRetryConfig()
	if cfg.MaxAttempts != 3 {
		t.Errorf("expected MaxAttempts=3, got %d", cfg.MaxAttempts)
	}
	if cfg.Multiplier != 2.0 {
		t.Errorf("expected Multiplier=2.0, got %f", cfg.Multiplier)
	}
	if cfg.InitialDelay <= 0 {
		t.Error("expected positive InitialDelay")
	}
	if cfg.MaxDelay <= cfg.InitialDelay {
		t.Error("expected MaxDelay > InitialDelay")
	}
}

func TestNewRetry_InvalidMaxAttempts(t *testing.T) {
	cfg := DefaultRetryConfig()
	cfg.MaxAttempts = 0
	_, err := NewRetry(cfg)
	if err == nil {
		t.Fatal("expected error for MaxAttempts=0")
	}
}

func TestNewRetry_InvalidMultiplier(t *testing.T) {
	cfg := DefaultRetryConfig()
	cfg.Multiplier = 0.5
	_, err := NewRetry(cfg)
	if err == nil {
		t.Fatal("expected error for Multiplier < 1")
	}
}

func TestRetry_Do_SucceedsOnFirstAttempt(t *testing.T) {
	r, err := NewRetry(DefaultRetryConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	calls := 0
	err = r.Do(func() error {
		calls++
		return nil
	})
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if calls != 1 {
		t.Errorf("expected 1 call, got %d", calls)
	}
}

func TestRetry_Do_RetriesOnError(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: time.Millisecond,
		MaxDelay:     10 * time.Millisecond,
		Multiplier:   2.0,
	}
	r, _ := NewRetry(cfg)
	calls := 0
	sentinel := errors.New("transient")
	err := r.Do(func() error {
		calls++
		if calls < 3 {
			return sentinel
		}
		return nil
	})
	if err != nil {
		t.Errorf("expected success on 3rd attempt, got %v", err)
	}
	if calls != 3 {
		t.Errorf("expected 3 calls, got %d", calls)
	}
}

func TestRetry_Do_ReturnsLastError(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts:  2,
		InitialDelay: time.Millisecond,
		MaxDelay:     10 * time.Millisecond,
		Multiplier:   1.5,
	}
	r, _ := NewRetry(cfg)
	sentinel := errors.New("always fails")
	err := r.Do(func() error { return sentinel })
	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}

func TestRetry_Attempts_ReturnsConfigured(t *testing.T) {
	cfg := DefaultRetryConfig()
	cfg.MaxAttempts = 5
	r, _ := NewRetry(cfg)
	if r.Attempts() != 5 {
		t.Errorf("expected 5, got %d", r.Attempts())
	}
}
