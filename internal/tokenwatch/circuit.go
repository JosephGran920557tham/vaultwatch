package tokenwatch

import (
	"errors"
	"sync"
	"time"
)

// ErrCircuitOpen is returned when the circuit breaker is in the open state.
var ErrCircuitOpen = errors.New("circuit breaker is open")

// CircuitState represents the state of the circuit breaker.
type CircuitState int

const (
	CircuitClosed CircuitState = iota
	CircuitOpen
	CircuitHalfOpen
)

// CircuitConfig holds configuration for the circuit breaker.
type CircuitConfig struct {
	// MaxFailures before the circuit opens.
	MaxFailures int
	// OpenDuration is how long the circuit stays open before moving to half-open.
	OpenDuration time.Duration
}

// DefaultCircuitConfig returns a CircuitConfig with sensible defaults.
func DefaultCircuitConfig() CircuitConfig {
	return CircuitConfig{
		MaxFailures:  5,
		OpenDuration: 30 * time.Second,
	}
}

// Circuit is a simple circuit breaker for token operations.
type Circuit struct {
	mu         sync.Mutex
	cfg        CircuitConfig
	failures   int
	state      CircuitState
	openedAt   time.Time
}

// NewCircuit creates a new Circuit with the given config.
// If MaxFailures < 1 or OpenDuration <= 0, defaults are applied.
func NewCircuit(cfg CircuitConfig) *Circuit {
	if cfg.MaxFailures < 1 {
		cfg.MaxFailures = DefaultCircuitConfig().MaxFailures
	}
	if cfg.OpenDuration <= 0 {
		cfg.OpenDuration = DefaultCircuitConfig().OpenDuration
	}
	return &Circuit{cfg: cfg, state: CircuitClosed}
}

// Allow returns nil if the operation is permitted, or ErrCircuitOpen if not.
func (c *Circuit) Allow() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	switch c.state {
	case CircuitOpen:
		if time.Since(c.openedAt) >= c.cfg.OpenDuration {
			c.state = CircuitHalfOpen
			return nil
		}
		return ErrCircuitOpen
	default:
		return nil
	}
}

// RecordSuccess resets the circuit to closed.
func (c *Circuit) RecordSuccess() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.failures = 0
	c.state = CircuitClosed
}

// RecordFailure increments the failure count and may open the circuit.
func (c *Circuit) RecordFailure() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.failures++
	if c.failures >= c.cfg.MaxFailures && c.state != CircuitOpen {
		c.state = CircuitOpen
		c.openedAt = time.Now()
	}
}

// State returns the current CircuitState.
func (c *Circuit) State() CircuitState {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.state
}
