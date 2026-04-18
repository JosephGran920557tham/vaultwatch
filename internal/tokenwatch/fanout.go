package tokenwatch

import (
	"context"
	"sync"

	"github.com/your-org/vaultwatch/internal/alert"
)

// FanoutConfig holds configuration for the Fanout dispatcher.
type FanoutConfig struct {
	// Workers controls the number of concurrent dispatch goroutines.
	Workers int
}

// DefaultFanoutConfig returns a FanoutConfig with sensible defaults.
func DefaultFanoutConfig() FanoutConfig {
	return FanoutConfig{Workers: 4}
}

// Fanout dispatches alerts concurrently to multiple handler functions.
type Fanout struct {
	cfg      FanoutConfig
	handlers []func(context.Context, alert.Alert) error
}

// NewFanout creates a Fanout with the given config and handlers.
// Panics if no handlers are provided.
func NewFanout(cfg FanoutConfig, handlers ...func(context.Context, alert.Alert) error) *Fanout {
	if len(handlers) == 0 {
		panic("fanout: at least one handler is required")
	}
	if cfg.Workers <= 0 {
		cfg.Workers = DefaultFanoutConfig().Workers
	}
	return &Fanout{cfg: cfg, handlers: handlers}
}

// Dispatch sends the alert to all handlers concurrently and collects errors.
func (f *Fanout) Dispatch(ctx context.Context, a alert.Alert) []error {
	type result struct {
		err error
	}

	sem := make(chan struct{}, f.cfg.Workers)
	results := make([]result, len(f.handlers))
	var wg sync.WaitGroup

	for i, h := range f.handlers {
		wg.Add(1)
		go func(idx int, fn func(context.Context, alert.Alert) error) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			results[idx] = result{err: fn(ctx, a)}
		}(i, h)
	}
	wg.Wait()

	var errs []error
	for _, r := range results {
		if r.err != nil {
			errs = append(errs, r.err)
		}
	}
	return errs
}
