package tokenwatch

import (
	"context"
	"log"
	"time"

	"github.com/your-org/vaultwatch/internal/alert"
)

// TriageScanner runs a Triage scorer against a source of alerts and
// dispatches the ranked results to a handler.
type TriageScanner struct {
	triage   *Triage
	source   func(ctx context.Context) ([]alert.Alert, error)
	dispatch func([]TriageEntry) error
	logger   *log.Logger
}

// NewTriageScanner constructs a TriageScanner.
// Panics if triage, source, or dispatch are nil.
func NewTriageScanner(
	triage *Triage,
	source func(ctx context.Context) ([]alert.Alert, error),
	dispatch func([]TriageEntry) error,
	logger *log.Logger,
) *TriageScanner {
	if triage == nil {
		panic("tokenwatch: NewTriageScanner: triage is nil")
	}
	if source == nil {
		panic("tokenwatch: NewTriageScanner: source is nil")
	}
	if dispatch == nil {
		panic("tokenwatch: NewTriageScanner: dispatch is nil")
	}
	if logger == nil {
		logger = log.Default()
	}
	return &TriageScanner{
		triage:   triage,
		source:   source,
		dispatch: dispatch,
		logger:   logger,
	}
}

// Scan fetches alerts from the source, ranks them, and dispatches the result.
func (s *TriageScanner) Scan(ctx context.Context) error {
	alerts, err := s.source(ctx)
	if err != nil {
		s.logger.Printf("triage_scanner: source error: %v", err)
		return err
	}
	if len(alerts) == 0 {
		return nil
	}
	entries := s.triage.Rank(alerts, time.Now())
	if err := s.dispatch(entries); err != nil {
		s.logger.Printf("triage_scanner: dispatch error: %v", err)
		return err
	}
	return nil
}
