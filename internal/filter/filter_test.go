package filter_test

import (
	"testing"
	"time"

	"github.com/your-org/vaultwatch/internal/alert"
	"github.com/your-org/vaultwatch/internal/filter"
)

func makeAlerts() []alert.Alert {
	return []alert.Alert{
		{
			LeaseID:   "secret/db/prod",
			Level:     alert.LevelCritical,
			ExpiresAt: time.Now().Add(1 * time.Hour),
			Labels:    map[string]string{"env": "prod"},
		},
		{
			LeaseID:   "secret/db/staging",
			Level:     alert.LevelWarning,
			ExpiresAt: time.Now().Add(6 * time.Hour),
			Labels:    map[string]string{"env": "staging"},
		},
		{
			LeaseID:   "auth/token/xyz",
			Level:     alert.LevelInfo,
			ExpiresAt: time.Now().Add(24 * time.Hour),
			Labels:    map[string]string{"env": "prod"},
		},
	}
}

func TestFilter_NoOptions_ReturnsAll(t *testing.T) {
	alerts := makeAlerts()
	got := filter.Filter(alerts, filter.Options{})
	if len(got) != len(alerts) {
		t.Fatalf("expected %d alerts, got %d", len(alerts), len(got))
	}
}

func TestFilter_PathPrefix(t *testing.T) {
	got := filter.Filter(makeAlerts(), filter.Options{PathPrefix: "secret/db"})
	if len(got) != 2 {
		t.Fatalf("expected 2, got %d", len(got))
	}
}

func TestFilter_MinLevel_Warning(t *testing.T) {
	got := filter.Filter(makeAlerts(), filter.Options{MinLevel: alert.LevelWarning})
	if len(got) != 2 {
		t.Fatalf("expected 2, got %d", len(got))
	}
	for _, a := range got {
		if a.Level < alert.LevelWarning {
			t.Errorf("unexpected level %v", a.Level)
		}
	}
}

func TestFilter_Labels(t *testing.T) {
	got := filter.Filter(makeAlerts(), filter.Options{Labels: map[string]string{"env": "prod"}})
	if len(got) != 2 {
		t.Fatalf("expected 2, got %d", len(got))
	}
}

func TestFilter_CombinedCriteria(t *testing.T) {
	opts := filter.Options{
		PathPrefix: "secret/",
		MinLevel:   alert.LevelCritical,
		Labels:     map[string]string{"env": "prod"},
	}
	got := filter.Filter(makeAlerts(), opts)
	if len(got) != 1 {
		t.Fatalf("expected 1, got %d", len(got))
	}
	if got[0].LeaseID != "secret/db/prod" {
		t.Errorf("unexpected lease %s", got[0].LeaseID)
	}
}
