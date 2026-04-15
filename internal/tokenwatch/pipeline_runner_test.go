package tokenwatch

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewPipelineRunner_NilPipeline(t *testing.T) {
	_, err := NewPipelineRunner(nil, func(_ context.Context, _ []Alert) error { return nil }, nil)
	if err == nil {
		t.Fatal("expected error for nil pipeline")
	}
}

func TestNewPipelineRunner_NilDispatch(t *testing.T) {
	p := newTestPipeline(t, nil, nil)
	_, err := NewPipelineRunner(p, nil, nil)
	if err == nil {
		t.Fatal("expected error for nil dispatch func")
	}
}

func TestNewPipelineRunner_NilLogger_UsesDefault(t *testing.T) {
	p := newTestPipeline(t, nil, nil)
	r, err := NewPipelineRunner(p, func(_ context.Context, _ []Alert) error { return nil }, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.logger == nil {
		t.Error("expected non-nil logger")
	}
}

func TestPipelineRunner_Run_StopsOnContextCancel(t *testing.T) {
	p := newTestPipeline(t, nil, nil)
	// Use a very short interval so the ticker fires quickly.
	p.cfg.PollInterval = 10 * time.Millisecond

	var dispatched int32
	dispatch := func(_ context.Context, alerts []Alert) error {
		atomic.AddInt32(&dispatched, int32(len(alerts)))
		return nil
	}

	r, err := NewPipelineRunner(p, dispatch, nil)
	if err != nil {
		t.Fatalf("NewPipelineRunner: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err = r.Run(ctx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected DeadlineExceeded, got %v", err)
	}
}

func TestPipelineRunner_DispatchError_DoesNotStop(t *testing.T) {
	p := newTestPipeline(t, nil, nil)
	p.cfg.PollInterval = 10 * time.Millisecond

	var calls int32
	dispatch := func(_ context.Context, _ []Alert) error {
		atomic.AddInt32(&calls, 1)
		return errors.New("dispatch failure")
	}

	r, err := NewPipelineRunner(p, dispatch, nil)
	if err != nil {
		t.Fatalf("NewPipelineRunner: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel()

	// Run should continue despite dispatch errors and only stop on ctx cancel.
	err = r.Run(ctx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected DeadlineExceeded, got %v", err)
	}
}
