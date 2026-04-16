package tokenwatch

import (
	"bytes"
	"strings"
	"testing"

	"github.com/vaultwatch/internal/alert"
)

func makeScoreAlert(id string, level alert.Level, msg string) alert.Alert {
	return alert.Alert{LeaseID: id, Level: level, Message: msg}
}

func TestScoreboard_Record_AddsPoints(t *testing.T) {
	sb := NewScoreboard()
	sb.Record(makeScoreAlert("tok-1", alert.LevelCritical, "expired"))
	top := sb.Top(1)
	if len(top) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(top))
	}
	if top[0].Score != 10 {
		t.Errorf("expected score 10, got %d", top[0].Score)
	}
}

func TestScoreboard_Record_Accumulates(t *testing.T) {
	sb := NewScoreboard()
	sb.Record(makeScoreAlert("tok-1", alert.LevelWarning, "warn"))
	sb.Record(makeScoreAlert("tok-1", alert.LevelWarning, "warn2"))
	top := sb.Top(1)
	if top[0].Score != 6 {
		t.Errorf("expected 6, got %d", top[0].Score)
	}
}

func TestScoreboard_Top_SortedDescending(t *testing.T) {
	sb := NewScoreboard()
	sb.Record(makeScoreAlert("low", alert.LevelInfo, "info"))
	sb.Record(makeScoreAlert("high", alert.LevelCritical, "crit"))
	sb.Record(makeScoreAlert("mid", alert.LevelWarning, "warn"))
	top := sb.Top(3)
	if top[0].TokenID != "high" {
		t.Errorf("expected high first, got %s", top[0].TokenID)
	}
	if top[2].TokenID != "low" {
		t.Errorf("expected low last, got %s", top[2].TokenID)
	}
}

func TestScoreboard_Top_LimitsResults(t *testing.T) {
	sb := NewScoreboard()
	for i := 0; i < 5; i++ {
		sb.Record(makeScoreAlert(string(rune('a'+i)), alert.LevelWarning, "w"))
	}
	if got := len(sb.Top(2)); got != 2 {
		t.Errorf("expected 2, got %d", got)
	}
}

func TestScoreboard_Reset_ClearsEntries(t *testing.T) {
	sb := NewScoreboard()
	sb.Record(makeScoreAlert("tok", alert.LevelCritical, "c"))
	sb.Reset()
	if got := len(sb.Top(10)); got != 0 {
		t.Errorf("expected 0 after reset, got %d", got)
	}
}

func TestScoreboard_Print_ContainsTokenID(t *testing.T) {
	sb := NewScoreboard()
	sb.Record(makeScoreAlert("my-token-id", alert.LevelCritical, "boom"))
	var buf bytes.Buffer
	sb.Print(&buf, 5)
	if !strings.Contains(buf.String(), "my-token-id") {
		t.Errorf("expected token id in output, got: %s", buf.String())
	}
}

func TestScoreboard_Print_NoEntries(t *testing.T) {
	sb := NewScoreboard()
	var buf bytes.Buffer
	sb.Print(&buf, 5)
	if !strings.Contains(buf.String(), "no entries") {
		t.Errorf("expected 'no entries', got: %s", buf.String())
	}
}
