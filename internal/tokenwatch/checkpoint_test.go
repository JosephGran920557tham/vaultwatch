package tokenwatch

import (
	"testing"
	"time"
)

func TestDefaultCheckpointConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultCheckpointConfig()
	if cfg.MaxAge <= 0 {
		t.Fatal("expected positive MaxAge")
	}
}

func TestNewCheckpoint_ZeroMaxAge_UsesDefault(t *testing.T) {
	cp := NewCheckpoint(CheckpointConfig{})
	if cp.cfg.MaxAge != DefaultCheckpointConfig().MaxAge {
		t.Fatalf("expected default MaxAge, got %v", cp.cfg.MaxAge)
	}
}

func TestCheckpoint_RecordAndGet(t *testing.T) {
	cp := NewCheckpoint(DefaultCheckpointConfig())
	cp.Record("tok1", 5*time.Minute)

	ttl, ok := cp.Get("tok1")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if ttl != 5*time.Minute {
		t.Fatalf("expected 5m, got %v", ttl)
	}
}

func TestCheckpoint_Get_Missing(t *testing.T) {
	cp := NewCheckpoint(DefaultCheckpointConfig())
	_, ok := cp.Get("nonexistent")
	if ok {
		t.Fatal("expected missing entry")
	}
}

func TestCheckpoint_Get_Expired(t *testing.T) {
	cp := NewCheckpoint(CheckpointConfig{MaxAge: 1 * time.Millisecond})
	cp.Record("tok2", 3*time.Minute)
	time.Sleep(5 * time.Millisecond)

	_, ok := cp.Get("tok2")
	if ok {
		t.Fatal("expected expired entry to be evicted")
	}
}

func TestCheckpoint_Delete(t *testing.T) {
	cp := NewCheckpoint(DefaultCheckpointConfig())
	cp.Record("tok3", 2*time.Minute)
	cp.Delete("tok3")

	_, ok := cp.Get("tok3")
	if ok {
		t.Fatal("expected deleted entry to be gone")
	}
}

func TestCheckpoint_Len(t *testing.T) {
	cp := NewCheckpoint(DefaultCheckpointConfig())
	if cp.Len() != 0 {
		t.Fatal("expected empty checkpoint")
	}
	cp.Record("a", time.Minute)
	cp.Record("b", time.Minute)
	if cp.Len() != 2 {
		t.Fatalf("expected 2, got %d", cp.Len())
	}
}
