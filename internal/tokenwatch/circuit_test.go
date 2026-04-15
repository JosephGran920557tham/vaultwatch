package tokenwatch

import (
	"testing"
	"time"
)

func TestDefaultCircuitConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultCircuitConfig()
	if cfg.MaxFailures < 1 {
		t.Errorf("expected MaxFailures >= 1, got %d", cfg.MaxFailures)
	}
	if cfg.OpenDuration <= 0 {
		t.Errorf("expected positive OpenDuration, got %v", cfg.OpenDuration)
	}
}

func TestNewCircuit_InvalidConfig_UsesDefaults(t *testing.T) {
	c := NewCircuit(CircuitConfig{MaxFailures: 0, OpenDuration: 0})
	if c.cfg.MaxFailures < 1 {
		t.Errorf("expected default MaxFailures, got %d", c.cfg.MaxFailures)
	}
	if c.cfg.OpenDuration <= 0 {
		t.Errorf("expected default OpenDuration, got %v", c.cfg.OpenDuration)
	}
}

func TestCircuit_Allow_InitiallyClosed(t *testing.T) {
	c := NewCircuit(DefaultCircuitConfig())
	if err := c.Allow(); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestCircuit_Opens_AfterMaxFailures(t *testing.T) {
	cfg := CircuitConfig{MaxFailures: 3, OpenDuration: 10 * time.Second}
	c := NewCircuit(cfg)
	for i := 0; i < 3; i++ {
		c.RecordFailure()
	}
	if c.State() != CircuitOpen {
		t.Errorf("expected CircuitOpen, got %v", c.State())
	}
	if err := c.Allow(); err != ErrCircuitOpen {
		t.Errorf("expected ErrCircuitOpen, got %v", err)
	}
}

func TestCircuit_RecordSuccess_Resets(t *testing.T) {
	cfg := CircuitConfig{MaxFailures: 2, OpenDuration: 10 * time.Second}
	c := NewCircuit(cfg)
	c.RecordFailure()
	c.RecordFailure()
	if c.State() != CircuitOpen {
		t.Fatal("expected circuit to be open")
	}
	// Manually move to half-open by backdating openedAt.
	c.mu.Lock()
	c.openedAt = time.Now().Add(-20 * time.Second)
	c.mu.Unlock()
	// Allow should transition to half-open.
	if err := c.Allow(); err != nil {
		t.Fatalf("expected nil in half-open, got %v", err)
	}
	c.RecordSuccess()
	if c.State() != CircuitClosed {
		t.Errorf("expected CircuitClosed after success, got %v", c.State())
	}
}

func TestCircuit_HalfOpen_TransitionsOnTimeout(t *testing.T) {
	cfg := CircuitConfig{MaxFailures: 1, OpenDuration: 50 * time.Millisecond}
	c := NewCircuit(cfg)
	c.RecordFailure()
	if err := c.Allow(); err != ErrCircuitOpen {
		t.Fatalf("expected ErrCircuitOpen immediately, got %v", err)
	}
	time.Sleep(60 * time.Millisecond)
	if err := c.Allow(); err != nil {
		t.Errorf("expected nil after timeout, got %v", err)
	}
	if c.State() != CircuitHalfOpen {
		t.Errorf("expected CircuitHalfOpen, got %v", c.State())
	}
}

func TestCircuit_MultipleFailures_OnlyOpensOnce(t *testing.T) {
	cfg := CircuitConfig{MaxFailures: 2, OpenDuration: 10 * time.Second}
	c := NewCircuit(cfg)
	for i := 0; i < 5; i++ {
		c.RecordFailure()
	}
	if c.State() != CircuitOpen {
		t.Errorf("expected CircuitOpen, got %v", c.State())
	}
}
