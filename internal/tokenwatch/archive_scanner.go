package tokenwatch

import (
	"log"

	"github.com/vaultwatch/internal/alert"
)

// ArchiveScanner feeds alerts produced by a source scanner into an Archive
// and optionally forwards them to a downstream dispatch function.
type ArchiveScanner struct {
	archive  *Archive
	source   func() ([]alert.Alert, error)
	dispatch func(alert.Alert)
	logger   *log.Logger
}

// NewArchiveScanner creates an ArchiveScanner. Panics if archive or source is nil.
func NewArchiveScanner(
	archive *Archive,
	source func() ([]alert.Alert, error),
	dispatch func(alert.Alert),
	logger *log.Logger,
) *ArchiveScanner {
	if archive == nil {
		panic("tokenwatch: NewArchiveScanner: archive must not be nil")
	}
	if source == nil {
		panic("tokenwatch: NewArchiveScanner: source must not be nil")
	}
	if logger == nil {
		logger = log.Default()
	}
	return &ArchiveScanner{
		archive:  archive,
		source:   source,
		dispatch: dispatch,
		logger:   logger,
	}
}

// Scan runs the source, records each alert in the archive, and optionally
// forwards it downstream. Errors from the source are logged and skipped.
func (s *ArchiveScanner) Scan() []alert.Alert {
	alerts, err := s.source()
	if err != nil {
		s.logger.Printf("archive_scanner: source error: %v", err)
		return nil
	}
	for _, al := range alerts {
		s.archive.Record(al)
		if s.dispatch != nil {
			s.dispatch(al)
		}
	}
	return alerts
}
