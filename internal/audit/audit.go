// Package audit provides a structured audit log for lease expiration events
// observed by vaultwatch, enabling traceability of alerts over time.
package audit

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/vaultwatch/internal/alert"
)

// Entry represents a single audit log record.
type Entry struct {
	Timestamp time.Time  `json:"timestamp"`
	LeaseID   string     `json:"lease_id"`
	Level     string     `json:"level"`
	Message   string     `json:"message"`
	TTL       int        `json:"ttl_seconds"`
}

// Logger writes audit entries to an io.Writer as newline-delimited JSON.
type Logger struct {
	writer io.Writer
	now    func() time.Time
}

// NewLogger returns a Logger that writes to w.
// Pass nil to write to os.Stdout.
func NewLogger(w io.Writer) *Logger {
	if w == nil {
		w = os.Stdout
	}
	return &Logger{writer: w, now: time.Now}
}

// Record writes an audit entry derived from the given alert.Alert.
func (l *Logger) Record(a alert.Alert) error {
	entry := Entry{
		Timestamp: l.now().UTC(),
		LeaseID:   a.LeaseID,
		Level:     string(a.Level),
		Message:   a.Message,
		TTL:       int(a.TTL.Seconds()),
	}
	b, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("audit: marshal entry: %w", err)
	}
	_, err = fmt.Fprintf(l.writer, "%s\n", b)
	return err
}

// RecordAll writes audit entries for each alert in the slice.
func (l *Logger) RecordAll(alerts []alert.Alert) error {
	for _, a := range alerts {
		if err := l.Record(a); err != nil {
			return err
		}
	}
	return nil
}
