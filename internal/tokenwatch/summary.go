package tokenwatch

import (
	"fmt"
	"io"
	"os"
	"sort"
	"text/tabwriter"
	"time"

	"github.com/yourusername/vaultwatch/internal/alert"
)

// TokenSummary holds aggregated state for a set of watched tokens.
type TokenSummary struct {
	Total    int
	Healthy  int
	Warning  int
	Critical int
	Alerts   []alert.Alert
	At       time.Time
}

// Summarize builds a TokenSummary from a slice of alerts and a total token count.
func Summarize(alerts []alert.Alert, total int, at time.Time) TokenSummary {
	s := TokenSummary{
		Total:  total,
		Alerts: alerts,
		At:     at,
	}
	for _, a := range alerts {
		switch a.Level {
		case alert.Warning:
			s.Warning++
		case alert.Critical:
			s.Critical++
		}
	}
	s.Healthy = total - s.Warning - s.Critical
	if s.Healthy < 0 {
		s.Healthy = 0
	}
	return s
}

// Print writes a human-readable summary table to w.
// If w is nil, os.Stdout is used.
func (s TokenSummary) Print(w io.Writer) {
	if w == nil {
		w = os.Stdout
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintf(tw, "Token Watch Summary\t[%s]\n", s.At.UTC().Format(time.RFC3339))
	fmt.Fprintf(tw, "Total:\t%d\n", s.Total)
	fmt.Fprintf(tw, "Healthy:\t%d\n", s.Healthy)
	fmt.Fprintf(tw, "Warning:\t%d\n", s.Warning)
	fmt.Fprintf(tw, "Critical:\t%d\n", s.Critical)
	if len(s.Alerts) > 0 {
		fmt.Fprintln(tw, "\nToken\tLevel\tMessage")
		sorted := make([]alert.Alert, len(s.Alerts))
		copy(sorted, s.Alerts)
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].LeaseID < sorted[j].LeaseID
		})
		for _, a := range sorted {
			fmt.Fprintf(tw, "%s\t%s\t%s\n", a.LeaseID, a.Level, a.Message)
		}
	}
	tw.Flush()
}
