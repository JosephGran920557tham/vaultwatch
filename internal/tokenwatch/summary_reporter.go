package tokenwatch

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// SummaryReporter writes TokenSummary data to an io.Writer in a chosen format.
type SummaryReporter struct {
	w      io.Writer
	format string // "text" or "json"
}

// NewSummaryReporter creates a SummaryReporter. format must be "text" or "json".
// An unknown format defaults to "text".
func NewSummaryReporter(w io.Writer, format string) *SummaryReporter {
	if w == nil {
		panic("tokenwatch: SummaryReporter requires a non-nil writer")
	}
	if format != "json" {
		format = "text"
	}
	return &SummaryReporter{w: w, format: format}
}

// Report writes the summary in the configured format.
func (r *SummaryReporter) Report(s TokenSummary) error {
	switch r.format {
	case "json":
		return r.writeJSON(s)
	default:
		s.Print(r.w)
		return nil
	}
}

type summaryJSON struct {
	At       time.Time `json:"at"`
	Total    int       `json:"total"`
	Healthy  int       `json:"healthy"`
	Warning  int       `json:"warning"`
	Critical int       `json:"critical"`
	Alerts   []alertJSON `json:"alerts"`
}

type alertJSON struct {
	Token   string `json:"token"`
	Level   string `json:"level"`
	Message string `json:"message"`
}

func (r *SummaryReporter) writeJSON(s TokenSummary) error {
	payload := summaryJSON{
		At:       s.At.UTC(),
		Total:    s.Total,
		Healthy:  s.Healthy,
		Warning:  s.Warning,
		Critical: s.Critical,
	}
	for _, a := range s.Alerts {
		payload.Alerts = append(payload.Alerts, alertJSON{
			Token:   a.LeaseID,
			Level:   fmt.Sprintf("%s", a.Level),
			Message: a.Message,
		})
	}
	enc := json.NewEncoder(r.w)
	enc.SetIndent("", "  ")
	return enc.Encode(payload)
}
