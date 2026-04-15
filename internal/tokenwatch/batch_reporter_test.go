package tokenwatch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func makeBatchResults(n int, withErr bool) []BatchResult {
	results := make([]BatchResult, n)
	for i := range results {
		results[i] = BatchResult{TokenID: fmt.Sprintf("tok-%d", i)}
		if withErr && i == 0 {
			results[i].Err = fmt.Errorf("lookup failed")
		}
		if i == 1 {
			results[i].Alerts = []Alert{{LeaseID: "lease-1"}}
		}
	}
	return results
}

func TestBuildBatchReport_Counts(t *testing.T) {
	results := makeBatchResults(3, true)
	rpt := BuildBatchReport(results)
	if rpt.Total != 3 {
		t.Errorf("expected Total=3, got %d", rpt.Total)
	}
	if rpt.Errors != 1 {
		t.Errorf("expected Errors=1, got %d", rpt.Errors)
	}
	if rpt.WithAlerts != 1 {
		t.Errorf("expected WithAlerts=1, got %d", rpt.WithAlerts)
	}
}

func TestBatchReporter_TextFormat(t *testing.T) {
	var buf bytes.Buffer
	rpt := BuildBatchReport(makeBatchResults(2, false))
	NewBatchReporter(&buf, "text").Write(rpt)
	out := buf.String()
	if !strings.Contains(out, "total=2") {
		t.Errorf("expected total=2 in output, got: %s", out)
	}
}

func TestBatchReporter_JSONFormat_ValidJSON(t *testing.T) {
	var buf bytes.Buffer
	rpt := BuildBatchReport(makeBatchResults(2, true))
	if err := NewBatchReporter(&buf, "json").Write(rpt); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var out map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if out["total"].(float64) != 2 {
		t.Errorf("expected total 2, got %v", out["total"])
	}
}

func TestNewBatchReporter_UnknownFormat_DefaultsToText(t *testing.T) {
	br := NewBatchReporter(&bytes.Buffer{}, "xml")
	if br.format != "text" {
		t.Errorf("expected text format fallback, got %s", br.format)
	}
}
