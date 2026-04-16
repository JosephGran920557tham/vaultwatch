package tokenwatch

import (
	"sync"
	"time"
)

// DefaultHedgeConfig returns a HedgeConfig with sensible defaults.
func DefaultHedgeConfig() HedgeConfig {
	return HedgeConfig{
		Delay:      50 * time.Millisecond,
		MaxHedges:  2,
	}
}

// HedgeConfig controls hedged-request behaviour.
type HedgeConfig struct {
	// Delay is how long to wait before issuing a hedge request.
	Delay time.Duration
	// MaxHedges is the maximum number of parallel hedge attempts.
	MaxHedges int
}

// HedgeResult carries the value or error returned by a hedged call.
type HedgeResult struct {
	Value interface{}
	Err   error
}

// Hedge runs fn up to cfg.MaxHedges+1 times, launching a new attempt after
// each cfg.Delay, and returns the first successful result. If all attempts
// fail, the last error is returned.
func Hedge(cfg HedgeConfig, fn func() (interface{}, error)) (interface{}, error) {
	if cfg.Delay <= 0 {
		cfg.Delay = DefaultHedgeConfig().Delay
	}
	if cfg.MaxHedges <= 0 {
		cfg.MaxHedges = DefaultHedgeConfig().MaxHedges
	}

	resultCh := make(chan HedgeResult, cfg.MaxHedges+1)
	var wg sync.WaitGroup

	launch := func() {
		wg.Add(1)
		go func() {
			defer wg.Done()
			v, err := fn()
			resultCh <- HedgeResult{Value: v, Err: err}
		}()
	}

	launch()
	for i := 0; i < cfg.MaxHedges; i++ {
		time.Sleep(cfg.Delay)
		launch()
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	var last error
	for r := range resultCh {
		if r.Err == nil {
			return r.Value, nil
		}
		last = r.Err
	}
	return nil, last
}
