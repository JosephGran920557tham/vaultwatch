package tokenwatch_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/alert"
	"github.com/yourusername/vaultwatch/internal/tokenwatch"
)

func makeSummaryAlerts() []alert.Alert {
	return []alert.Alert{
		{LeaseID: "token/abc", Level: alert.Warning, Message: "expires soon"},
		{LeaseID: "token/xyz", Level: alert.Critical, Message: "expires very soon"},
		{LeaseID: "token/def", Level: alert.Warning, Message: "expires soon"},
	}
}

func TestSummarize_Counts(t *testing.T) {
	alerts := makeSummaryAlerts()
	s := tokenwatch.Summarize(alerts, 10, time.Now())

	if s.Total != 10 {
		t.Errorf("expected Total=10, got %d", s.Total)
	}
	if s.Warning != 2 {
		t.Errorf("expected Warning=2, got %d", s.Warning)
	}
	if s.Critical != 1 {
		t.Errorf("expected Critical=1, got %d", s.Critical)
	}
	if s.Healthy != 7 {
		t.Errorf("expected Healthy=7, got %d", s.Healthy)
	}
}

func TestSummarize_Empty(t *testing.T) {
	s := tokenwatch.Summarize(nil, 5, time.Now())
	if s.Total != 5 || s.Healthy != 5 || s.Warning != 0 || s.Critical != 0 {
		t.Errorf("unexpected counts: %+v", s)
	}
}

func TestSummarize_HealthyFloor(t *testing.T) {
	// More alerts than total should not produce negative healthy count.
	alerts := makeSummaryAlerts()
	s := tokenwatch.Summarize(alerts, 1, time.Now())
	if s.Healthy < 0 {
		t.Errorf("Healthy should not be negative, got %d", s.Healthy)
	}
}

func TestPrint_ContainsTokenIDs(t *testing.T) {
	alerts := makeSummaryAlerts()
	s := tokenwatch.Summarize(alerts, 10, time.Now())
	var buf bytes.Buffer
	s.Print(&buf)
	out := buf.String()

	for _, id := range []string{"token/abc", "token/xyz", "token/def"} {
		if !strings.Contains(out, id) {
			t.Errorf("expected output to contain %q", id)
		}
	}
}

func TestPrint_ContainsCounts(t *testing.T) {
	s := tokenwatch.Summarize(makeSummaryAlerts(), 10, time.Now())
	var buf bytes.Buffer
	s.Print(&buf)
	out := buf.String()

	for _, want := range []string{"Total:", "Healthy:", "Warning:", "Critical:"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected output to contain %q", want)
		}
	}
}

func TestPrint_NilWriter_DoesNotPanic(t *testing.T) {
	s := tokenwatch.Summarize(nil, 3, time.Now())
	// Should not panic when w is nil (falls back to stdout).
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Print panicked: %v", r)
		}
	}()
	s.Print(nil)
}
