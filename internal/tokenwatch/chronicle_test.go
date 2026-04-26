package tokenwatch

import (
	"context"
	"testing"
	"time"

	"github.com/vaultwatch/internal/alert"
)

func makeChronicleAlert(leaseID string) alert.Alert {
	return alert.Alert{LeaseID: leaseID, Level: alert.LevelWarning}
}

func TestDefaultChronicleConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultChronicleConfig()
	if cfg.MaxEntries <= 0 {
		t.Error("expected positive MaxEntries")
	}
	if cfg.MaxAge <= 0 {
		t.Error("expected positive MaxAge")
	}
}

func TestNewChronicle_ZeroValues_UsesDefaults(t *testing.T) {
	c := NewChronicle(ChronicleConfig{})
	if c.cfg.MaxEntries <= 0 || c.cfg.MaxAge <= 0 {
		t.Error("expected defaults to be applied for zero config")
	}
}

func TestChronicle_Record_And_List(t *testing.T) {
	c := NewChronicle(DefaultChronicleConfig())
	a := makeChronicleAlert("lease-1")
	c.Record("tok-1", a)

	entries := c.List("tok-1")
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Alert.LeaseID != "lease-1" {
		t.Errorf("unexpected lease ID: %s", entries[0].Alert.LeaseID)
	}
}

func TestChronicle_List_Empty_ReturnsNil(t *testing.T) {
	c := NewChronicle(DefaultChronicleConfig())
	if got := c.List("unknown"); got != nil {
		t.Errorf("expected nil for unknown token, got %v", got)
	}
}

func TestChronicle_PrunesExpiredEntries(t *testing.T) {
	cfg := ChronicleConfig{MaxEntries: 100, MaxAge: 10 * time.Millisecond}
	c := NewChronicle(cfg)
	c.Record("tok-2", makeChronicleAlert("lease-old"))
	time.Sleep(20 * time.Millisecond)
	if got := c.List("tok-2"); len(got) != 0 {
		t.Errorf("expected expired entries to be pruned, got %d", len(got))
	}
}

func TestChronicle_CapsAtMaxEntries(t *testing.T) {
	cfg := ChronicleConfig{MaxEntries: 3, MaxAge: time.Hour}
	c := NewChronicle(cfg)
	for i := 0; i < 5; i++ {
		c.Record("tok-3", makeChronicleAlert("lease"))
	}
	if n := c.Len("tok-3"); n != 3 {
		t.Errorf("expected 3 entries after cap, got %d", n)
	}
}

func TestChronicleScanner_Scan_RecordsAndReturnsAlerts(t *testing.T) {
	reg := NewRegistry()
	_ = reg.Add("tok-a")
	chronicle := NewChronicle(DefaultChronicleConfig())

	source := func(_ context.Context, tokenID string) ([]alert.Alert, error) {
		return []alert.Alert{makeChronicleAlert(tokenID)}, nil
	}

	scanner := NewChronicleScanner(reg, chronicle, source)
	alerts, err := scanner.Scan(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(alerts))
	}
	if chronicle.Len("tok-a") != 1 {
		t.Errorf("expected chronicle to have 1 entry for tok-a")
	}
}
