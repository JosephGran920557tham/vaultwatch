package report

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// JSONReporter serialises a Summary to JSON.
type JSONReporter struct {
	out io.Writer
}

// NewJSONReporter returns a JSONReporter writing to w.
// If w is nil, os.Stdout is used.
func NewJSONReporter(w io.Writer) *JSONReporter {
	if w == nil {
		w = os.Stdout
	}
	return &JSONReporter{out: w}
}

// jsonSummary is the serialisable form of Summary.
type jsonSummary struct {
	GeneratedAt string      `json:"generated_at"`
	Total       int         `json:"total"`
	Critical    int         `json:"critical"`
	Warning     int         `json:"warning"`
	Info        int         `json:"info"`
	Alerts      []jsonAlert `json:"alerts"`
}

type jsonAlert struct {
	LeaseID   string `json:"lease_id"`
	Severity  string `json:"severity"`
	Message   string `json:"message"`
	ExpiresAt string `json:"expires_at"`
}

// Write encodes the Summary as indented JSON and writes it to the reporter's writer.
func (r *JSONReporter) Write(s Summary) error {
	js := jsonSummary{
		GeneratedAt: s.GeneratedAt.Format("2006-01-02T15:04:05Z"),
		Total:       s.Total,
		Critical:    s.Critical,
		Warning:     s.Warning,
		Info:        s.Info,
	}
	for _, a := range s.Alerts {
		js.Alerts = append(js.Alerts, jsonAlert{
			LeaseID:   a.LeaseID,
			Severity:  string(a.Severity),
			Message:   a.Message,
			ExpiresAt: a.ExpiresAt.Format("2006-01-02T15:04:05Z"),
		})
	}
	enc, err := json.MarshalIndent(js, "", "  ")
	if err != nil {
		return fmt.Errorf("report: json marshal: %w", err)
	}
	_, err = fmt.Fprintln(r.out, string(enc))
	return err
}
