package tokenwatch

import (
	"context"
	"log"
	"time"
)

// LastSeenSource returns the last-seen timestamp for a given token ID.
type LastSeenSource interface {
	LastSeen(tokenID string) (time.Time, bool)
}

// StalenessScanner periodically scans all registered tokens for staleness.
type StalenessScanner struct {
	registry *Registry
	source   LastSeenSource
	detector *StalenessDetector
	dispatch func(a interface{})
	logger   *log.Logger
}

// NewStalenessScanner constructs a StalenessScanner.
// dispatch is called with each non-nil alert produced by the detector.
func NewStalenessScanner(
	reg *Registry,
	src LastSeenSource,
	det *StalenessDetector,
	dispatch func(a interface{}),
	logger *log.Logger,
) *StalenessScanner {
	if reg == nil {
		panic("tokenwatch: StalenessScanner requires a non-nil Registry")
	}
	if src == nil {
		panic("tokenwatch: StalenessScanner requires a non-nil LastSeenSource")
	}
	if det == nil {
		det = NewStalenessDetector(DefaultStalenessConfig())
	}
	if logger == nil {
		logger = log.Default()
	}
	return &StalenessScanner{
		registry: reg,
		source:   src,
		detector: det,
		dispatch: dispatch,
		logger:   logger,
	}
}

// Scan iterates over all registered tokens, checks staleness, and dispatches alerts.
// It respects context cancellation between tokens.
func (s *StalenessScanner) Scan(ctx context.Context) error {
	for _, id := range s.registry.List() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		ls, ok := s.source.LastSeen(id)
		if !ok {
			s.logger.Printf("staleness: no last-seen record for token %s, skipping", id)
			continue
		}
		if a := s.detector.Check(id, ls); a != nil {
			if s.dispatch != nil {
				s.dispatch(a)
			}
		}
	}
	return nil
}
