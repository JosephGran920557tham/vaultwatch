package tokenwatch

import (
	"context"
	"testing"
	"time"

	"github.com/vaultwatch/internal/alert"
)

func makeQueueEnvelope(id string) Envelope {
	return NewEnvelope(id, alert.Alert{LeaseID: id, Level: alert.Warning}, 3)
}

func TestNewQueue_DefaultCap(t *testing.T) {
	q := NewQueue(0)
	if q.cap != 64 {
		t.Errorf("expected cap=64, got %d", q.cap)
	}
}

func TestQueue_PushAndPop(t *testing.T) {
	q := NewQueue(4)
	env := makeQueueEnvelope("tok-1")
	if !q.Push(env) {
		t.Fatal("expected push to succeed")
	}
	got, ok := q.Pop()
	if !ok {
		t.Fatal("expected pop to succeed")
	}
	if got.Token != "tok-1" {
		t.Errorf("unexpected token: %s", got.Token)
	}
}

func TestQueue_Pop_EmptyReturnsFalse(t *testing.T) {
	q := NewQueue(4)
	_, ok := q.Pop()
	if ok {
		t.Error("expected false on empty pop")
	}
}

func TestQueue_Push_FullReturnsFalse(t *testing.T) {
	q := NewQueue(2)
	q.Push(makeQueueEnvelope("a"))
	q.Push(makeQueueEnvelope("b"))
	if q.Push(makeQueueEnvelope("c")) {
		t.Error("expected push to fail on full queue")
	}
}

func TestQueue_Len(t *testing.T) {
	q := NewQueue(8)
	if q.Len() != 0 {
		t.Error("expected empty queue")
	}
	q.Push(makeQueueEnvelope("x"))
	q.Push(makeQueueEnvelope("y"))
	if q.Len() != 2 {
		t.Errorf("expected Len=2, got %d", q.Len())
	}
}

func TestQueue_Wait_ReturnsWhenItemAdded(t *testing.T) {
	q := NewQueue(4)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	go func() {
		time.Sleep(20 * time.Millisecond)
		q.Push(makeQueueEnvelope("async"))
	}()
	if err := q.Wait(ctx); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestQueue_Wait_CancelledContext(t *testing.T) {
	q := NewQueue(4)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := q.Wait(ctx); err == nil {
		t.Error("expected error on cancelled context")
	}
}
