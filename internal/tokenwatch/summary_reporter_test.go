package tokenwatch_test

import {
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/alert"
	"github.com/yourusername/vaultwatch/internal/tokenwatch"
}

func makeSummary() tokenwatch.TokenSummary {
	alerts := []alert.Alert{
		{LeaseID: "token/aaa", Level: alert.Warning, Message: "warn"},
		{LeaseID: "token/bbb", Level: alert.Critical, Message: "crit"},
	}
	return tokenwatch.Summarize(alerts, 5, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
}

func TestNewSummaryReporter_NilWriter_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil writer")
		}
	}()
	tokenwatch.NewSummaryReporter(nil, "text")
}

func TestNewSummaryReporter_UnknownFormat_DefaultsToText(t *testing.T) {
	var buf bytes.Buffer
	r := tokenwatch.NewSummaryReporter(&buf, "yaml")
	if err := r.Report(makeSummary()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "Total:") {
		t.Error("expected text output with 'Total:'")
	}
}

func TestSummaryReporter_TextFormat(t *testing.T) {
	var buf bytes.Buffer
	r := tokenwatch.NewSummaryReporter(&buf, "text")
	if err := r.Report(makeSummary()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Warning:") || !strings.Contains(out, "Critical:") {
		t.Errorf("text output missing expected fields: %s", out)
	}
}

func TestSummaryReporter_JSONFormat_ValidJSON(t *testing.T) {
	var buf bytes.Buffer
	r := tokenwatch.NewSummaryReporter(&buf, "json")
	if err := r.Report(makeSummary()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var out map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
}

func TestSummaryReporter_JSONFormat_Fields(t *testing.T) {
	var buf bytes.Buffer
	r := tokenwatch.NewSummaryReporter(&buf, "json")
	_ = r.Report(makeSummary())

	var out map[string]interface{}
	_ = json.Unmarshal(buf.Bytes(), &out)

	for _, key := range []string{"at", "total", "healthy", "warning", "critical", "alerts"} {
		if _, ok := out[key]; !ok {
			t.Errorf("missing JSON field %q", key)
		}
	}
	if v, _ := out["total"].(float64); int(v) != 5 {
		t.Errorf("expected total=5, got %v", out["total"])
	}
}
