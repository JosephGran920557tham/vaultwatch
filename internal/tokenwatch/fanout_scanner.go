package tokenwatch

import (
	"context"
	"log"

	"github.com/your-org/vaultwatch/internal/alert"
)

// AlertSource is anything that can produce a slice of alerts.
type AlertSource interface {
	Scan(ctx context.Context) ([]alert.Alert, error)
}

// FanoutScanner runs multiple AlertSources and fans their results out
// through a Fanout dispatcher.
type FanoutScanner struct {
	sources []AlertSource
	fanout  *Fanout
	logger  *log.Logger
}

// NewFanoutScanner creates a FanoutScanner.
// Panics if sources or fanout are nil/empty.
func NewFanoutScanner(fanout *Fanout, logger *log.Logger, sources ...AlertSource) *FanoutScanner {
	if fanout == nil {
		panic("fanout_scanner: fanout must not be nil")
	}
	if len(sources) == 0 {
		panic("fanout_scanner: at least one source is required")
	}
	if logger == nil {
		logger = log.Default()
	}
	return &FanoutScanner{sources: sources, fanout: fanout, logger: logger}
}

// Run scans all sources and dispatches each alert through the fanout.
func (fs *FanoutScanner) Run(ctx context.Context) error {
	for _, src := range fs.sources {
		alerts, err := src.Scan(ctx)
		if err != nil {
			fs.logger.Printf("fanout_scanner: source error: %v", err)
			continue
		}
		for _, a := range alerts {
			if errs := fs.fanout.Dispatch(ctx, a); len(errs) > 0 {
				fs.logger.Printf("fanout_scanner: dispatch errors for %s: %v", a.LeaseID, errs)
			}
		}
	}
	return nil
}
