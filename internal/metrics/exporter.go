package metrics

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"text/tabwriter"
)

// Exporter serialises a Registry snapshot to various output formats.
type Exporter struct {
	reg *Registry
}

// NewExporter wraps reg in an Exporter.
func NewExporter(reg *Registry) *Exporter {
	if reg == nil {
		reg = NewRegistry()
	}
	return &Exporter{reg: reg}
}

// WriteJSON encodes the current snapshot as a JSON object to w.
func (e *Exporter) WriteJSON(w io.Writer) error {
	snap := e.reg.Snapshot()
	return json.NewEncoder(w).Encode(snap)
}

// WriteText writes a human-readable table of metrics to w.
func (e *Exporter) WriteText(w io.Writer) error {
	snap := e.reg.Snapshot()

	// Stable ordering.
	keys := make([]string, 0, len(snap))
	for k := range snap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "METRIC\tVALUE")
	for _, k := range keys {
		fmt.Fprintf(tw, "%s\t%.4g\n", k, snap[k])
	}
	return tw.Flush()
}
