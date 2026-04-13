package ratelimit

import (
	"context"
	"errors"
	"testing"
)

func TestNewMiddleware_NilLimiter(t *testing.T) {
	_, err := NewMiddleware(nil, false)
	if err == nil {
		t.Fatal("expected error for nil limiter")
	}
}

func TestMiddleware_NoWait_AllowsWhenToken(t *testing.T) {
	l, _ := New(Config{Rate: 10, Burst: 2})
	mw, err := NewMiddleware(l, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	called := false
	fn := mw.Wrap(func(ctx context.Context) error {
		called = true
		return nil
	})

	if err := fn(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected wrapped function to be called")
	}
}

func TestMiddleware_NoWait_BlocksWhenExhausted(t *testing.T) {
	l, _ := New(Config{Rate: 0.001, Burst: 1})
	// drain the single token
	l.Allow()

	mw, _ := NewMiddleware(l, false)
	fn := mw.Wrap(func(ctx context.Context) error {
		return nil
	})

	err := fn(context.Background())
	if !errors.Is(err, ErrRateLimited) {
		t.Fatalf("expected ErrRateLimited, got %v", err)
	}
}

func TestMiddleware_Wait_PropagatesWrappedError(t *testing.T) {
	l, _ := New(Config{Rate: 100, Burst: 5})
	mw, _ := NewMiddleware(l, true)

	sentinel := errors.New("check failed")
	fn := mw.Wrap(func(ctx context.Context) error {
		return sentinel
	})

	err := fn(context.Background())
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}

func TestMiddleware_Wait_CancelledContext(t *testing.T) {
	l, _ := New(Config{Rate: 0.001, Burst: 1})
	// drain the token so Wait must block
	l.Allow()

	mw, _ := NewMiddleware(l, true)
	fn := mw.Wrap(func(ctx context.Context) error { return nil })

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := fn(ctx); err == nil {
		t.Fatal("expected error from cancelled context")
	}
}
