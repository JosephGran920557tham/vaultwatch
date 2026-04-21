package tokenwatch

import (
	"testing"
	"time"

	"github.com/your-org/vaultwatch/internal/alert"
)

func makeTriageAlert(level alert.Level, firedAt time.Time) alert.Alert {
	return alert.Alert{
		LeaseID: "tok-triage",
		Level:   level,
		FiredAt: firedAt,
		Message: "triage test",
	}
}

func TestDefaultTriageConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultTriageConfig()
	if cfg.CriticalWeight <= 0 {
		t.Error("expected positive CriticalWeight")
	}
	if cfg.WarningWeight <= 0 {
		t.Error("expected positive WarningWeight")
	}
	if cfg.RecencyHalfLife <= 0 {
		t.Error("expected positive RecencyHalfLife")
	}
}

func TestNewTriage_ZeroValues_UsesDefaults(t *testing.T) {
	tr := NewTriage(TriageConfig{})
	def := DefaultTriageConfig()
	if tr.cfg.CriticalWeight != def.CriticalWeight {
		t.Errorf("expected CriticalWeight %v, got %v", def.CriticalWeight, tr.cfg.CriticalWeight)
	}
}

func TestTriage_Score_CriticalHigherThanWarning(t *testing.T) {
	now := time.Now()
	tr := NewTriage(DefaultTriageConfig())
	crit := makeTriageAlert(alert.Critical, now)
	warn := makeTriageAlert(alert.Warning, now)
	if tr.Score(crit, now) <= tr.Score(warn, now) {
		t.Error("critical score should exceed warning score")
	}
}

func TestTriage_Score_DecaysWithAge(t *testing.T) {
	now := time.Now()
	tr := NewTriage(DefaultTriageConfig())
	fresh := makeTriageAlert(alert.Critical, now)
	old := makeTriageAlert(alert.Critical, now.Add(-30*time.Minute))
	if tr.Score(fresh, now) <= tr.Score(old, now) {
		t.Error("fresh alert should score higher than old alert")
	}
}

func TestTriage_Rank_SortedDescending(t *testing.T) {
	now := time.Now()
	tr := NewTriage(DefaultTriageConfig())
	alerts := []alert.Alert{
		makeTriageAlert(alert.Info, now),
		makeTriageAlert(alert.Critical, now),
		makeTriageAlert(alert.Warning, now),
	}
	entries := tr.Rank(alerts, now)
	for i := 1; i < len(entries); i++ {
		if entries[i].Score > entries[i-1].Score {
			t.Errorf("entry %d score %v > entry %d score %v; not sorted", i, entries[i].Score, i-1, entries[i-1].Score)
		}
	}
}

func TestTriage_Rank_Empty_ReturnsEmpty(t *testing.T) {
	tr := NewTriage(DefaultTriageConfig())
	result := tr.Rank(nil, time.Now())
	if len(result) != 0 {
		t.Errorf("expected empty result, got %d entries", len(result))
	}
}
