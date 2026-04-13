package audit_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/vaultwatch/internal/alert"
	"github.com/vaultwatch/internal/audit"
)

func makeAlert(id, msg string, ttl time.Duration, level alert.Level) alert.Alert {
	return alert.Alert{
		LeaseID: id,
		Message: msg,
		TTL:     ttl,
		Level:   level,
	}
}

func TestRecord_WritesValidJSON(t *testing.T) {
	var buf bytes.Buffer
	l := audit.NewLogger(&buf)
	a := makeAlert("lease/abc", "expiring soon", 30*time.Second, alert.LevelWarning)

	if err := l.Record(a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var entry audit.Entry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if entry.LeaseID != "lease/abc" {
		t.Errorf("expected lease_id lease/abc, got %s", entry.LeaseID)
	}
	if entry.Level != "warning" {
		t.Errorf("expected level warning, got %s", entry.Level)
	}
	if entry.TTL != 30 {
		t.Errorf("expected ttl 30, got %d", entry.TTL)
	}
}

func TestRecordAll_WritesMultipleLines(t *testing.T) {
	var buf bytes.Buffer
	l := audit.NewLogger(&buf)
	alerts := []alert.Alert{
		makeAlert("lease/1", "msg1", 10*time.Second, alert.LevelInfo),
		makeAlert("lease/2", "msg2", 20*time.Second, alert.LevelCritical),
	}

	if err := l.RecordAll(alerts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
}

func TestNewLogger_NilWriter_UsesStdout(t *testing.T) {
	// Should not panic when nil is passed.
	l := audit.NewLogger(nil)
	if l == nil {
		t.Fatal("expected non-nil logger")
	}
}
