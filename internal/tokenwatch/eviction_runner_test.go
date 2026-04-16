package tokenwatch

import (
	"context"
	"testing"
	"time"
)

func TestNewEvictionRunner_NilEviction_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil eviction")
		}
	}()
	NewEvictionRunner(nil, NewRegistry(), nil)
}

func TestNewEvictionRunner_NilRegistry_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil registry")
		}
	}()
	ev := NewEviction(DefaultEvictionConfig())
	NewEvictionRunner(ev, nil, nil)
}

func TestNewEvictionRunner_NilLogger_UsesDefault(t *testing.T) {
	ev := NewEviction(DefaultEvictionConfig())
	reg := NewRegistry()
	runner := NewEvictionRunner(ev, reg, nil)
	if runner.logger == nil {
		t.Fatal("expected non-nil logger")
	}
}

func TestEvictionRunner_Run_EvictsStaleTokens(t *testing.T) {
	cfg := EvictionConfig{
		MaxAge:        10 * time.Millisecond,
		SweepInterval: 20 * time.Millisecond,
	}
	ev := NewEviction(cfg)
	reg := NewRegistry()

	if err := reg.Add("tok-stale"); err != nil {
		t.Fatalf("unexpected error adding token: %v", err)
	}
	ev.Touch("tok-stale")

	time.Sleep(15 * time.Millisecond) // let MaxAge expire

	runner := NewEvictionRunner(ev, reg, nil)
	ctx, cancel := context.WithCancel(context.Background())
	go runner.Run(ctx)

	time.Sleep(50 * time.Millisecond)
	cancel()

	tokens := reg.List()
	for _, tok := range tokens {
		if tok == "tok-stale" {
			t.Error("expected tok-stale to be evicted from registry")
		}
	}
}

func TestEvictionRunner_Run_StopsOnContextCancel(t *testing.T) {
	ev := NewEviction(EvictionConfig{MaxAge: time.Hour, SweepInterval: time.Hour})
	reg := NewRegistry()
	runner := NewEvictionRunner(ev, reg, nil)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		runner.Run(ctx)
		close(done)
	}()
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Run did not stop after context cancel")
	}
}
