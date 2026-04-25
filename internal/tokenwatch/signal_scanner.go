package tokenwatch

import (
	"context"
	"log"

	"github.com/vaultwatch/internal/alert"
)

// signalSource is any scanner that can produce alerts for a set of tokens.
type signalSource interface {
	Scan(ctx context.Context) ([]alert.Alert, error)
}

// SignalScanner wraps an underlying scanner and gates its output through
// a SignalAggregator so that transient single-sample alerts are suppressed.
type SignalScanner struct {
	source aggregatorSource
	agg    *SignalAggregator
	log    *log.Logger
}

type aggregatorSource interface {
	Scan(ctx context.Context) ([]alert.Alert, error)
}

// NewSignalScanner creates a SignalScanner.
// Panics if source or agg are nil.
func NewSignalScanner(source aggregatorSource, agg *SignalAggregator, logger *log.Logger) *SignalScanner {
	if source == nil {
		panic("tokenwatch: SignalScanner source must not be nil")
	}
	if agg == nil {
		panic("tokenwatch: SignalScanner aggregator must not be nil")
	}
	if logger == nil {
		logger = log.Default()
	}
	return &SignalScanner{source: source, agg: agg, log: logger}
}

// Scan delegates to the underlying source and filters alerts that have
// not yet reached the required signal strength.
func (s *SignalScanner) Scan(ctx context.Context) ([]alert.Alert, error) {
	raw, err := s.source.Scan(ctx)
	if err != nil {
		return nil, err
	}

	var out []alert.Alert
	for _, a := range raw {
		if s.agg.Observe(a.LeaseID, a.Level) {
			out = append(out, a)
		}
	}
	return out, nil
}
