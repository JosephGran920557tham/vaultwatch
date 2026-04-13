package lifecycle_test

import (
	"context"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/vaultwatch/internal/lifecycle"
)

func TestRunner_StopsOnContextCancel(t *testing.T) {
	mgr := lifecycle.New(2 * time.Second)

	started := make(chan struct{})
	stopped := make(chan struct{})

	mgr.OnStart("svc", func(_ context.Context) error {
		close(started)
		return nil
	})
	mgr.OnStop("svc", func(_ context.Context) error {
		close(stopped)
		return nil
	})

	runner := lifecycle.NewRunner(mgr)
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error,() { done <- runner.Run(ctx) }()

	<-started
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("runner did not stop in time")
	}

	select {
	case <-stopped:
	default:
		t.Error("stop hook was not called")
	}
}

func TestRunner_StopsOnSignal(t *testing.T) {
	mgr := lifecycle.New(2 * time.Second)

	var stopCalled bool
	mgr.OnStop("svc", func(_ context.Context) error {
		stopCalled = true
		return nil
	})

	// Use SIGUSR1 so we don't kill the test process.
	runner := lifecycle.NewRunner(mgr, syscall.SIGUSR1)
	ctx := context.Background()

	done := make(chan error, 1)
	go func() { done <- runner.Run(ctx) }()

	time.Sleep(50 * time.Millisecond)
	p, _ := os.FindProcess(os.Getpid())
	_ = p.Signal(syscall.SIGUSR1)

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("runner did not stop after signal")
	}

	if !stopCalled {
		t.Error("stop hook was not called after signal")
	}
}
