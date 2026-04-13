package schedule_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/youorg/vaultwatch/internal/schedule"
)

type mockRunner struct {
	calls atomic.Int32
	err   error
}

func (m *mockRunner) RunOnce(_ context.Context) error {
	m.calls.Add(1)
	return m.err
}

func TestScheduler_RunsImmediately(t *testing.T) {
	r := &mockRunner{}
	s := schedule.New(r, 10*time.Second, nil)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() { done <- s.Start(ctx) }()

	// Give the scheduler time to fire the first tick.
	time.Sleep(50 * time.Millisecond)
	cancel()
	<-done

	if r.calls.Load() < 1 {
		t.Errorf("expected at least 1 call, got %d", r.calls.Load())
	}
}

func TestScheduler_TicksAtInterval(t *testing.T) {
	r := &mockRunner{}
	interval := 60 * time.Millisecond
	s := schedule.New(r, interval, nil)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- s.Start(ctx) }()

	time.Sleep(200 * time.Millisecond)
	cancel()
	<-done

	if r.calls.Load() < 3 {
		t.Errorf("expected >=3 calls, got %d", r.calls.Load())
	}
}

func TestScheduler_StopsOnContextCancel(t *testing.T) {
	r := &mockRunner{}
	s := schedule.New(r, 1*time.Hour, nil)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	err := s.Start(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestScheduler_RunnerErrorDoesNotStop(t *testing.T) {
	r := &mockRunner{err: errors.New("vault unavailable")}
	interval := 50 * time.Millisecond
	s := schedule.New(r, interval, nil)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- s.Start(ctx) }()

	time.Sleep(160 * time.Millisecond)
	cancel()
	<-done

	// Should have retried despite errors.
	if r.calls.Load() < 2 {
		t.Errorf("expected >=2 calls even with errors, got %d", r.calls.Load())
	}
}
