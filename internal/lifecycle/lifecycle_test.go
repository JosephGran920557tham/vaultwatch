package lifecycle_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/vaultwatch/internal/lifecycle"
)

func TestNew_DefaultDrainTimeout(t *testing.T) {
	m := lifecycle.New(0)
	if m == nil {
		t.Fatal("expected non-nil manager")
	}
}

func TestStart_RunsHooksInOrder(t *testing.T) {
	m := lifecycle.New(5 * time.Second)
	var order []string

	m.OnStart("first", func(_ context.Context) error {
		order = append(order, "first")
		return nil
	})
	m.OnStart("second", func(_ context.Context) error {
		order = append(order, "second")
		return nil
	})

	if err := m.Start(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(order) != 2 || order[0] != "first" || order[1] != "second" {
		t.Errorf("unexpected order: %v", order)
	}
}

func TestStart_HookError_Aborts(t *testing.T) {
	m := lifecycle.New(5 * time.Second)
	boom := errors.New("boom")

	m.OnStart("bad", func(_ context.Context) error { return boom })

	err := m.Start(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, boom) {
		t.Errorf("expected boom, got %v", err)
	}
}

func TestStop_RunsHooksInReverseOrder(t *testing.T) {
	m := lifecycle.New(5 * time.Second)
	var order []string

	m.OnStop("first", func(_ context.Context) error {
		order = append(order, "first")
		return nil
	})
	m.OnStop("second", func(_ context.Context) error {
		order = append(order, "second")
		return nil
	})

	if err := m.Stop(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(order) != 2 || order[0] != "second" || order[1] != "first" {
		t.Errorf("unexpected reverse order: %v", order)
	}
}

func TestStop_HookError_ReturnsFirst(t *testing.T) {
	m := lifecycle.New(5 * time.Second)
	boom := errors.New("stop-boom")

	m.OnStop("bad", func(_ context.Context) error { return boom })

	err := m.Stop()
	if !errors.Is(err, boom) {
		t.Errorf("expected stop-boom, got %v", err)
	}
}
