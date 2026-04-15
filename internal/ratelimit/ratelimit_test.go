package ratelimit

import (
	"context"
	"testing"
	"time"
)

func TestNew_InvalidRate(t *testing.T) {
	_, err := New(Config{Rate: 0, Burst: 5})
	if err == nil {
		t.Fatal("expected error for zero rate")
	}
}

func TestNew_InvalidBurst(t *testing.T) {
	_, err := New(Config{Rate: 1, Burst: -1})
	if err == nil {
		t.Fatal("expected error for negative burst")
	}
}

func TestAllow_ConsumesToken(t *testing.T) {
	l, err := New(Config{Rate: 1, Burst: 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !l.Allow() {
		t.Fatal("expected first Allow() to return true")
	}
	if !l.Allow() {
		t.Fatal("expected second Allow() to return true")
	}
	if l.Allow() {
		t.Fatal("expected third Allow() to return false (burst exhausted)")
	}
}

func TestAllow_RefillsOverTime(t *testing.T) {
	l, err := New(Config{Rate: 10, Burst: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// drain the bucket
	l.Allow()

	// advance the fake clock by 200ms — should add 2 tokens, capped at burst=1
	now := time.Now()
	l.clock = func() time.Time { return now.Add(200 * time.Millisecond) }

	if !l.Allow() {
		t.Fatal("expected Allow() to succeed after clock advance")
	}
}

func TestWait_SucceedsWithinContext(t *testing.T) {
	l, err := New(Config{Rate: 100, Burst: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := l.Wait(ctx); err != nil {
		t.Fatalf("unexpected error from Wait: %v", err)
	}
}

func TestWait_CancelledContext(t *testing.T) {
	l, err := New(Config{Rate: 0.001, Burst: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// drain the bucket so Wait must block
	l.Allow()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	if err := l.Wait(ctx); err == nil {
		t.Fatal("expected context cancellation error")
	}
}

func TestNew_ValidConfig(t *testing.T) {
	tests := []struct {
		name  string
		cfg   Config
	}{
		{"minimal burst", Config{Rate: 1, Burst: 0}},
		{"high rate", Config{Rate: 1000, Burst: 500}},
		{"fractional rate", Config{Rate: 0.5, Burst: 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.cfg)
			if err != nil {
				t.Fatalf("unexpected error for config %+v: %v", tt.cfg, err)
			}
		})
	}
}
