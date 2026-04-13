// Package schedule provides periodic execution of monitor runs.
package schedule

import (
	"context"
	"log"
	"time"
)

// Runner is any type that can perform a single monitoring pass.
type Runner interface {
	RunOnce(ctx context.Context) error
}

// Scheduler triggers a Runner at a fixed interval until the context is cancelled.
type Scheduler struct {
	runner   Runner
	interval time.Duration
	log      *log.Logger
}

// New creates a Scheduler that calls runner.RunOnce every interval.
func New(runner Runner, interval time.Duration, logger *log.Logger) *Scheduler {
	if logger == nil {
		logger = log.Default()
	}
	return &Scheduler{
		runner:   runner,
		interval: interval,
		log:      logger,
	}
}

// Start blocks, running the runner immediately and then on every tick.
// It returns when ctx is cancelled.
func (s *Scheduler) Start(ctx context.Context) error {
	if err := s.tick(ctx); err != nil {
		return err
	}

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.log.Println("scheduler: stopped")
			return ctx.Err()
		case <-ticker.C:
			if err := s.tick(ctx); err != nil {
				s.log.Printf("scheduler: run error: %v", err)
			}
		}
	}
}

func (s *Scheduler) tick(ctx context.Context) error {
	s.log.Println("scheduler: running monitor pass")
	return s.runner.RunOnce(ctx)
}
