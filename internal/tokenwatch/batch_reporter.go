package tokenwatch

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// BatchReport summarises the outcome of a BatchRunner.Run call.
type BatchReport struct {
	Timestamp   time.Time       `json:"timestamp"`
	Total       int             `json:"total"`
	WithAlerts  int             `json:"with_alerts"`
	Errors      int             `json:"errors"`
	Results     []BatchResult   `json:"-"` // not serialised directly
}

// BuildBatchReport constructs a BatchReport from raw results.
func BuildBatchReport(results []BatchResult) BatchReport {
	rpt := BatchReport{
		Timestamp: time.Now().UTC(),
		Total:     len(results),
	}
	for _, r := range results {
		if r.Err != nil {
			rpt.Errors++
		}
		if len(r.Alerts) > 0 {
			rpt.WithAlerts++
		}
	}
	rpt.Results = results
	return rpt
}

// BatchReporter writes BatchReports in text or JSON format.
type BatchReporter struct {
	w      io.Writer
	format string
}

// NewBatchReporter creates a BatchReporter. format must be "text" or "json".
func NewBatchReporter(w io.Writer, format string) *BatchReporter {
	if format != "json" {
		format = "text"
	}
	return &BatchReporter{w: w, format: format}
}

// Write outputs the report to the underlying writer.
func (r *BatchReporter) Write(rpt BatchReport) error {
	if r.format == "json" {
		return json.NewEncoder(r.w).Encode(rpt)
	}
	_, err := fmt.Fprintf(r.w,
		"[%s] batch: total=%d with_alerts=%d errors=%d\n",
		rpt.Timestamp.Format(time.RFC3339),
		rpt.Total, rpt.WithAlerts, rpt.Errors,
	)
	return err
}
