package tokenwatch

import (
	"context"
	"fmt"
	"time"
)

// PipelineConfig holds configuration for the token watch pipeline.
type PipelineConfig struct {
	// PollInterval controls how often the pipeline checks tokens.
	PollInterval time.Duration
	// MaxConcurrency limits parallel token checks.
	MaxConcurrency int
}

// DefaultPipelineConfig returns a PipelineConfig with sensible defaults.
func DefaultPipelineConfig() PipelineConfig {
	return PipelineConfig{
		PollInterval:   30 * time.Second,
		MaxConcurrency: 8,
	}
}

// Pipeline orchestrates the full token-watch cycle: check, classify,
// deduplicate, throttle, and dispatch alerts.
type Pipeline struct {
	alerter     *Alerter
	dedup       *Deduplicator
	throttle    *Throttle
	cfg         PipelineConfig
}

// NewPipeline constructs a Pipeline from its dependencies.
func NewPipeline(alerter *Alerter, dedup *Deduplicator, throttle *Throttle, cfg PipelineConfig) (*Pipeline, error) {
	if alerter == nil {
		return nil, fmt.Errorf("tokenwatch: pipeline requires a non-nil Alerter")
	}
	if dedup == nil {
		return nil, fmt.Errorf("tokenwatch: pipeline requires a non-nil Deduplicator")
	}
	if throttle == nil {
		return nil, fmt.Errorf("tokenwatch: pipeline requires a non-nil Throttle")
	}
	if cfg.MaxConcurrency <= 0 {
		cfg.MaxConcurrency = DefaultPipelineConfig().MaxConcurrency
	}
	if cfg.PollInterval <= 0 {
		cfg.PollInterval = DefaultPipelineConfig().PollInterval
	}
	return &Pipeline{
		alerter:  alerter,
		dedup:    dedup,
		throttle: throttle,
		cfg:      cfg,
	}, nil
}

// Run executes the pipeline once and returns filtered alerts.
func (p *Pipeline) Run(ctx context.Context) ([]Alert, error) {
	all, err := p.alerter.CheckAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("tokenwatch pipeline: check failed: %w", err)
	}

	var out []Alert
	for _, a := range all {
		if !p.dedup.Allow(a.LeaseID) {
			continue
		}
		if !p.throttle.Allow(a.LeaseID) {
			continue
		}
		out = append(out, a)
	}
	return out, nil
}

// PollInterval returns the configured polling interval.
func (p *Pipeline) PollInterval() time.Duration {
	return p.cfg.PollInterval
}
