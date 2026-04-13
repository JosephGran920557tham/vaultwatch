package report_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/vaultwatch/internal/report"
)

func TestJSONReporter_Write_ValidJSON(t *testing.T) {
	var buf bytes.Buffer
	r := report.NewJSONReporter(&buf)

	b := report.NewBuilder(&buf)
	s := b.Build(makeAlerts())

	if err := r.Write(s); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var out map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
}

func TestJSONReporter_Write_Fields(t *testing.T) {
	var buf bytes.Buffer
	r := report.NewJSONReporter(&buf)

	b := report.NewBuilder(nil)
	s := b.Build(makeAlerts())

	_ = r.Write(s)

	var out map[string]interface{}
	_ = json.Unmarshal(buf.Bytes(), &out)

	if out["total"].(float64) != 3 {
		t.Errorf("expected total=3, got %v", out["total"])
	}
	if out["critical"].(float64) != 1 {
		t.Errorf("expected critical=1, got %v", out["critical"])
	}
	alerts, ok := out["alerts"].([]interface{})
	if !ok || len(alerts) != 3 {
		t.Errorf("expected 3 alerts in JSON, got %v", out["alerts"])
	}
}

func TestJSONReporter_Write_EmptySummary(t *testing.T) {
	var buf bytes.Buffer
	r := report.NewJSONReporter(&buf)

	b := report.NewBuilder(nil)
	s := b.Build(nil)

	if err := r.Write(s); err != nil {
		t.Fatalf("unexpected error on empty summary: %v", err)
	}

	var out map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if out["total"].(float64) != 0 {
		t.Errorf("expected total=0 for empty summary")
	}
}
