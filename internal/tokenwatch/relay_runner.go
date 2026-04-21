package tokenwatch

import (
	"context"
	"log"
	"time"
)

// RelayRunner periodically flushes a Relay on a fixed interval and
// also performs a final flush when the context is cancelled.
type RelayRunner struct {
	relay  *Relay
	logger *log.Logger
}

// NewRelayRunner constructs a RelayRunner. relay must not be nil.
func NewRelayRunner(relay *Relay, logger *log.Logger) *RelayRunner {
	if relay == nil {
		panic("relay_runner: relay must not be nil")
	}
	if logger == nil {
		logger = log.Default()
	}
	return &RelayRunner{relay: relay, logger: logger}
}

// Run starts the flush loop. It blocks until ctx is done, then
// performs one final flush before returning.
func (rr *RelayRunner) Run(ctx context.Context) error {
	ticker := time.NewTicker(rr.relay.cfg.FlushInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := rr.relay.Flush(); err != nil {
				rr.logger.Printf("relay_runner: flush error: %v", err)
			}
		case <-ctx.Done():
			if err := rr.relay.Flush(); err != nil {
				rr.logger.Printf("relay_runner: final flush error: %v", err)
			}
			return ctx.Err()
		}
	}
}
