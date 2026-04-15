package tokenwatch

import (
	"context"
	"fmt"
	"log"
	"time"
)

// DispatchFunc is called with each batch of alerts produced by the pipeline.
type DispatchFunc func(ctx context.Context, alerts []Alert) error

// PipelineRunner runs a Pipeline on a ticker and dispatches results.
type PipelineRunner struct {
	pipeline *Pipeline
	dispatch DispatchFunc
	logger   *log.Logger
}

// NewPipelineRunner constructs a PipelineRunner.
func NewPipelineRunner(pipeline *Pipeline, dispatch DispatchFunc, logger *log.Logger) (*PipelineRunner, error) {
	if pipeline == nil {
		return nil, fmt.Errorf("tokenwatch: PipelineRunner requires a non-nil Pipeline")
	}
	if dispatch == nil {
		return nil, fmt.Errorf("tokenwatch: PipelineRunner requires a non-nil DispatchFunc")
	}
	if logger == nil {
		logger = log.Default()
	}
	return &PipelineRunner{
		pipeline: pipeline,
		dispatch: dispatch,
		logger:   logger,
	}, nil
}

// Run starts the polling loop. It blocks until ctx is cancelled.
func (r *PipelineRunner) Run(ctx context.Context) error {
	if err := r.tick(ctx); err != nil {
		r.logger.Printf("tokenwatch pipeline: initial tick error: %v", err)
	}

	ticker := time.NewTicker(r.pipeline.PollInterval())
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := r.tick(ctx); err != nil {
				r.logger.Printf("tokenwatch pipeline: tick error: %v", err)
			}
		}
	}
}

func (r *PipelineRunner) tick(ctx context.Context) error {
	alerts, err := r.pipeline.Run(ctx)
	if err != nil {
		return fmt.Errorf("pipeline run: %w", err)
	}
	if len(alerts) == 0 {
		return nil
	}
	if err := r.dispatch(ctx, alerts); err != nil {
		return fmt.Errorf("dispatch: %w", err)
	}
	return nil
}
