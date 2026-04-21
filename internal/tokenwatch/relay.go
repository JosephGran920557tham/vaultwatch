package tokenwatch

import (
	"errors"
	"sync"
	"time"

	"github.com/vaultwatch/internal/alert"
)

// DefaultRelayConfig returns a RelayConfig with sensible defaults.
func DefaultRelayConfig() RelayConfig {
	return RelayConfig{
		BufferSize: 64,
		FlushInterval: 5 * time.Second,
	}
}

// RelayConfig controls the behaviour of a Relay.
type RelayConfig struct {
	BufferSize    int
	FlushInterval time.Duration
}

// Relay buffers alerts from multiple sources and forwards them in
// batches to a downstream dispatch function.
type Relay struct {
	cfg      RelayConfig
	mu       sync.Mutex
	buffer   []alert.Alert
	dispatch func([]alert.Alert) error
}

// NewRelay constructs a Relay. dispatch must not be nil.
func NewRelay(cfg RelayConfig, dispatch func([]alert.Alert) error) *Relay {
	if dispatch == nil {
		panic("relay: dispatch func must not be nil")
	}
	if cfg.BufferSize <= 0 {
		cfg.BufferSize = DefaultRelayConfig().BufferSize
	}
	if cfg.FlushInterval <= 0 {
		cfg.FlushInterval = DefaultRelayConfig().FlushInterval
	}
	return &Relay{cfg: cfg, dispatch: dispatch}
}

// Enqueue adds an alert to the internal buffer. If the buffer is
// full the oldest entry is dropped to make room.
func (r *Relay) Enqueue(a alert.Alert) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.buffer) >= r.cfg.BufferSize {
		r.buffer = r.buffer[1:]
	}
	r.buffer = append(r.buffer, a)
}

// Flush drains the buffer and forwards all buffered alerts.
// Returns an error if dispatch fails; the buffer is still cleared.
func (r *Relay) Flush() error {
	r.mu.Lock()
	if len(r.buffer) == 0 {
		r.mu.Unlock()
		return nil
	}
	batch := make([]alert.Alert, len(r.buffer))
	copy(batch, r.buffer)
	r.buffer = r.buffer[:0]
	r.mu.Unlock()
	return r.dispatch(batch)
}

// Len returns the current number of buffered alerts.
func (r *Relay) Len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.buffer)
}

// ErrRelayFull is returned when the relay buffer cannot accept more alerts.
var ErrRelayFull = errors.New("relay: buffer full")
