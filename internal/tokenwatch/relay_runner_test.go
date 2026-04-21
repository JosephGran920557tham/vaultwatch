package tokenwatch

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/vaultwatch/internal/alert"
)

func TestNewRelayRunner_NilRelay_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil relay")
		}
	}()
	NewRelayRunner(nil, nil)
}

func TestNewRelayRunner_NilLogger_UsesDefault(t *testing.T) {
	relay := NewRelay(DefaultRelayConfig(), func([]alert.Alert) error { return nil })
	rr := NewRelayRunner(relay, nil)
	if rr.logger == nil {
		t.Error("expected non-nil logger")
	}
}

func TestRelayRunner_Run_FlushesOnInterval(t *testing.T) {
	var flushCount int64
	cfg := RelayConfig{BufferSize: 16, FlushInterval: 20 * time.Millisecond}
	relay := NewRelay(cfg, func(batch []alert.Alert) error {
		atomic.AddInt64(&flushCount, 1)
		return nil
	})
	relay.Enqueue(makeRelayAlert("z"))

	rr := NewRelayRunner(relay, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 70*time.Millisecond)
	defer cancel()
	_ = rr.Run(ctx)

	if atomic.LoadInt64(&flushCount) < 2 {
		t.Errorf("expected at least 2 flushes, got %d", flushCount)
	}
}

func TestRelayRunner_Run_FinalFlushOnCancel(t *testing.T) {
	var received []alert.Alert
	cfg := RelayConfig{BufferSize: 16, FlushInterval: 10 * time.Second}
	relay := NewRelay(cfg, func(batch []alert.Alert) error {
		received = append(received, batch...)
		return nil
	})
	relay.Enqueue(makeRelayAlert("final"))

	rr := NewRelayRunner(relay, nil)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()
	_ = rr.Run(ctx)

	if len(received) != 1 || received[0].LeaseID != "final" {
		t.Errorf("expected final flush of 'final', got %v", received)
	}
}

func TestRelayRunner_Run_StopsOnContextCancel(t *testing.T) {
	cfg := RelayConfig{BufferSize: 8, FlushInterval: 5 * time.Second}
	relay := NewRelay(cfg, func([]alert.Alert) error { return nil })
	rr := NewRelayRunner(relay, nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	done := make(chan struct{})
	go func() {
		_ = rr.Run(ctx)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Error("Run did not stop after context cancel")
	}
}
