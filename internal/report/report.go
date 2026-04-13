// Package report provides functionality for generating summary reports
// of Vault secret lease statuses collected during a monitor run.
package report

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/vaultwatch/internal/alert"
)

// Summary holds aggregated counts and details from a monitor run.
type Summary struct {
	GeneratedAt time.Time
	Total       int
	Critical    int
	Warning     int
	Info        int
	Alerts      []alert.Alert
}

// Builder constructs a Summary from a slice of alerts.
type Builder struct {
	out io.Writer
}

// NewBuilder returns a Builder that writes to w.
// If w is nil, os.Stdout is used.
func NewBuilder(w io.Writer) *Builder {
	if w == nil {
		w = os.Stdout
	}
	return &Builder{out: w}
}

// Build aggregates alerts into a Summary.
func (b *Builder) Build(alerts []alert.Alert) Summary {
	s := Summary{
		GeneratedAt: time.Now().UTC(),
		Total:       len(alerts),
		Alerts:      alerts,
	}
	for _, a := range alerts {
		switch a.Severity {
		case alert.SeverityCritical:
			s.Critical++
		case alert.SeverityWarning:
			s.Warning++
		default:
			s.Info++
		}
	}
	return s
}

// Print writes a human-readable summary to the Builder's writer.
func (b *Builder) Print(s Summary) {
	fmt.Fprintf(b.out, "=== VaultWatch Report [%s] ===\n", s.GeneratedAt.Format(time.RFC3339))
	fmt.Fprintf(b.out, "Total leases checked : %d\n", s.Total)
	fmt.Fprintf(b.out, "Critical             : %d\n", s.Critical)
	fmt.Fprintf(b.out, "Warning              : %d\n", s.Warning)
	fmt.Fprintf(b.out, "Info                 : %d\n", s.Info)
	if len(s.Alerts) == 0 {
		fmt.Fprintln(b.out, "No alerts to display.")
		return
	}
	fmt.Fprintln(b.out, "\nAlerts:")
	for _, a := range s.Alerts {
		fmt.Fprintf(b.out, "  [%s] %s — %s\n", a.Severity, a.LeaseID, a.Message)
	}
}
