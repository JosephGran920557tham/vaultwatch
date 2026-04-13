package report_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/vaultwatch/internal/alert"
	"github.com/vaultwatch/internal/report"
)

func makeAlerts() []alert.Alert {
	return []alert.Alert{
		{LeaseID: "lease/critical/1", Severity: alert.SeverityCritical, Message: "expires in 1h", ExpiresAt: time.Now().Add(1 * time.Hour)},
		{LeaseID: "lease/warning/2", Severity: alert.SeverityWarning, Message: "expires in 6h", ExpiresAt: time.Now().Add(6 * time.Hour)},
		{LeaseID: "lease/info/3", Severity: alert.SeverityInfo, Message: "expires in 24h", ExpiresAt: time.Now().Add(24 * time.Hour)},
	}
}

func TestBuild_Counts(t *testing.T) {
	b := report.NewBuilder(nil)
	s := b.Build(makeAlerts())

	if s.Total != 3 {
		t.Errorf("expected Total=3, got %d", s.Total)
	}
	if s.Critical != 1 {
		t.Errorf("expected Critical=1, got %d", s.Critical)
	}
	if s.Warning != 1 {
		t.Errorf("expected Warning=1, got %d", s.Warning)
	}
	if s.Info != 1 {
		t.Errorf("expected Info=1, got %d", s.Info)
	}
}

func TestBuild_Empty(t *testing.T) {
	b := report.NewBuilder(nil)
	s := b.Build(nil)

	if s.Total != 0 || s.Critical != 0 || s.Warning != 0 || s.Info != 0 {
		t.Error("expected all zero counts for empty alert slice")
	}
}

func TestPrint_ContainsLeaseID(t *testing.T) {
	var buf bytes.Buffer
	b := report.NewBuilder(&buf)
	s := b.Build(makeAlerts())
	b.Print(s)

	output := buf.String()
	if !strings.Contains(output, "lease/critical/1") {
		t.Error("expected output to contain lease/critical/1")
	}
	if !strings.Contains(output, "VaultWatch Report") {
		t.Error("expected output to contain report header")
	}
}

func TestPrint_NoAlerts(t *testing.T) {
	var buf bytes.Buffer
	b := report.NewBuilder(&buf)
	s := b.Build(nil)
	b.Print(s)

	if !strings.Contains(buf.String(), "No alerts to display") {
		t.Error("expected 'No alerts to display' message for empty summary")
	}
}
