package tokenwatch

import (
	"context"
	"fmt"
	"sync"
)

// BatchResult holds the outcome of a single token check within a batch run.
type BatchResult struct {
	TokenID string
	Alerts  []Alert
	Err     error
}

// BatchRunner executes token checks concurrently across all registered tokens.
type BatchRunner struct {
	registry  *Registry
	alerter   *Alerter
	concurrency int
}

// NewBatchRunner creates a BatchRunner. concurrency controls the maximum number
// of goroutines used; if <= 0 it defaults to 4.
func NewBatchRunner(registry *Registry, alerter *Alerter, concurrency int) (*BatchRunner, error) {
	if registry == nil {
		return nil, fmt.Errorf("tokenwatch: registry must not be nil")
	}
	if alerter == nil {
		return nil, fmt.Errorf("tokenwatch: alerter must not be nil")
	}
	if concurrency <= 0 {
		concurrency = 4
	}
	return &BatchRunner{registry: registry, alerter: alerter, concurrency: concurrency}, nil
}

// Run checks all registered tokens concurrently and returns one BatchResult per token.
func (b *BatchRunner) Run(ctx context.Context) []BatchResult {
	tokens := b.registry.List()
	results := make([]BatchResult, len(tokens))

	sem := make(chan struct{}, b.concurrency)
	var wg sync.WaitGroup

	for i, id := range tokens {
		wg.Add(1)
		go func(idx int, tokenID string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			alerts, err := b.alerter.CheckToken(ctx, tokenID)
			results[idx] = BatchResult{TokenID: tokenID, Alerts: alerts, Err: err}
		}(i, id)
	}

	wg.Wait()
	return results
}
