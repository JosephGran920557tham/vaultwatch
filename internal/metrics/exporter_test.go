package metrics_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/yourusername/vaultwatch/internal/metrics"
)

func TestExporter_WriteJSON_ValidOutput(t *testing.T) {
	reg := metrics.NewRegistry()
	reg.Counter("checks").Add(3)
	reg.Gauge("leases").Set(12)

	e := metrics.NewExporter(reg)
	var buf bytes.Buffer
	if err := e.WriteJSON(&buf); err != nil {
		t.Fatalf("WriteJSON error: %v", err)
	}

	var out map[string]float64
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if out["checks"] != 3 {
		t.Errorf("expected checks=3, got %f", out["checks"])
	}
	if out["leases"] != 12 {
		t.Errorf("expected leases=12, got %f", out["leases"])
	}
}

func TestExporter_WriteText_ContainsHeaders(t *testing.T) {
	reg := metrics.NewRegistry()
	reg.Counter("alerts_sent").Inc()

	e := metrics.NewExporter(reg)
	var buf bytes.Buffer
	if err := e.WriteText(&buf); err != nil {
		t.Fatalf("WriteText error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "METRIC") {
		t.Error("expected header 'METRIC' in text output")
	}
	if !strings.Contains(output, "alerts_sent") {
		t.Error("expected 'alerts_sent' in text output")
	}
}

func TestExporter_WriteText_SortedKeys(t *testing.T) {
	reg := metrics.NewRegistry()
	reg.Counter("z_metric").Inc()
	reg.Counter("a_metric").Inc()

	e := metrics.NewExporter(reg)
	var buf bytes.Buffer
	_ = e.WriteText(&buf)

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	// lines[0] is header, lines[1] should be a_metric
	if len(lines) < 3 {
		t.Fatalf("expected at least 3 lines, got %d", len(lines))
	}
	if !strings.HasPrefix(strings.TrimSpace(lines[1]), "a_metric") {
		t.Errorf("expected a_metric first, got: %s", lines[1])
	}
}

func TestNewExporter_NilRegistry(t *testing.T) {
	// Should not panic.
	e := metrics.NewExporter(nil)
	var buf bytes.Buffer
	if err := e.WriteJSON(&buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
