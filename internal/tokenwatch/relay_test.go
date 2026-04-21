package tokenwatch

import (
	"errors"
	"testing"
	"time"

	"github.com/vaultwatch/internal/alert"
)

func makeRelayAlert(id string) alert.Alert {
	return alert.Alert{LeaseID: id, Level: alert.Warning}
}

func TestDefaultRelayConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultRelayConfig()
	if cfg.BufferSize <= 0 {
		t.Errorf("expected positive BufferSize, got %d", cfg.BufferSize)
	}
	if cfg.FlushInterval <= 0 {
		t.Errorf("expected positive FlushInterval, got %v", cfg.FlushInterval)
	}
}

func TestNewRelay_NilDispatch_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil dispatch")
		}
	}()
	NewRelay(DefaultRelayConfig(), nil)
}

func TestNewRelay_ZeroValues_UsesDefaults(t *testing.T) {
	r := NewRelay(RelayConfig{}, func([]alert.Alert) error { return nil })
	if r.cfg.BufferSize <= 0 {
		t.Errorf("expected default BufferSize, got %d", r.cfg.BufferSize)
	}
	if r.cfg.FlushInterval <= 0 {
		t.Errorf("expected default FlushInterval, got %v", r.cfg.FlushInterval)
	}
}

func TestRelay_Enqueue_And_Flush(t *testing.T) {
	var received []alert.Alert
	r := NewRelay(DefaultRelayConfig(), func(batch []alert.Alert) error {
		received = append(received, batch...)
		return nil
	})
	r.Enqueue(makeRelayAlert("a"))
	r.Enqueue(makeRelayAlert("b"))
	if r.Len() != 2 {
		t.Fatalf("expected 2 buffered, got %d", r.Len())
	}
	if err := r.Flush(); err != nil {
		t.Fatalf("unexpected flush error: %v", err)
	}
	if r.Len() != 0 {
		t.Errorf("expected empty buffer after flush, got %d", r.Len())
	}
	if len(received) != 2 {
		t.Errorf("expected 2 dispatched, got %d", len(received))
	}
}

func TestRelay_Flush_EmptyBuffer_NoDispatch(t *testing.T) {
	called := false
	r := NewRelay(DefaultRelayConfig(), func([]alert.Alert) error {
		called = true
		return nil
	})
	if err := r.Flush(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Error("dispatch should not be called for empty buffer")
	}
}

func TestRelay_Enqueue_DropsOldestWhenFull(t *testing.T) {
	cfg := RelayConfig{BufferSize: 2, FlushInterval: time.Second}
	r := NewRelay(cfg, func([]alert.Alert) error { return nil })
	r.Enqueue(makeRelayAlert("first"))
	r.Enqueue(makeRelayAlert("second"))
	r.Enqueue(makeRelayAlert("third"))
	if r.Len() != 2 {
		t.Fatalf("expected buffer capped at 2, got %d", r.Len())
	}
	var received []alert.Alert
	r.dispatch = func(batch []alert.Alert) error {
		received = batch
		return nil
	}
	_ = r.Flush()
	if received[0].LeaseID != "second" {
		t.Errorf("expected oldest dropped; got %s", received[0].LeaseID)
	}
}

func TestRelay_Flush_PropagatesDispatchError(t *testing.T) {
	sentinel := errors.New("dispatch failed")
	r := NewRelay(DefaultRelayConfig(), func([]alert.Alert) error {
		return sentinel
	})
	r.Enqueue(makeRelayAlert("x"))
	if err := r.Flush(); !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}
