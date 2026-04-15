package tokenwatch

import (
	"testing"
	"time"

	"github.com/vaultwatch/internal/alert"
)

func newTestBurstScanner(t *testing.T, maxEvents int) (*BurstScanner, *Registry) {
	t.Helper()
	reg := NewRegistry()
	det, err := NewBurstDetector(BurstConfig{Window: time.Minute, MaxEvents: maxEvents})
	if err != nil {
		t.Fatalf("NewBurstDetector: %v", err)
	}
	return NewBurstScanner(reg, det), reg
}

func TestNewBurstScanner_NilRegistry_Panics(t *testing.T) {
	det, _ := NewBurstDetector(BurstConfig{})
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil registry")
		}
	}()
	NewBurstScanner(nil, det)
}

func TestNewBurstScanner_NilDetector_Panics(t *testing.T) {
	reg := NewRegistry()
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil detector")
		}
	}()
	NewBurstScanner(reg, nil)
}

func TestBurstScanner_Scan_NoAlerts_WhenUnderLimit(t *testing.T) {
	sc, reg := newTestBurstScanner(t, 5)
	_ = reg.Add("tok-1")
	for i := 0; i < 3; i++ {
		sc.Record("tok-1")
	}
	if got := sc.Scan(); len(got) != 0 {
		t.Errorf("expected no alerts, got %d", len(got))
	}
}

func TestBurstScanner_Scan_AlertWhenBursting(t *testing.T) {
	sc, reg := newTestBurstScanner(t, 2)
	_ = reg.Add("tok-burst")
	for i := 0; i < 5; i++ {
		sc.Record("tok-burst")
	}
	alerts := sc.Scan()
	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(alerts))
	}
	if alerts[0].LeaseID != "tok-burst" {
		t.Errorf("expected LeaseID tok-burst, got %s", alerts[0].LeaseID)
	}
	if alerts[0].Level != alert.LevelWarning {
		t.Errorf("expected Warning level, got %s", alerts[0].Level)
	}
}

func TestBurstScanner_Scan_OnlyRegisteredTokens(t *testing.T) {
	sc, reg := newTestBurstScanner(t, 2)
	_ = reg.Add("tok-reg")
	// Record events for an unregistered token too
	for i := 0; i < 5; i++ {
		sc.Record("tok-unreg")
		sc.Record("tok-reg")
	}
	alerts := sc.Scan()
	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert for registered token, got %d", len(alerts))
	}
	if alerts[0].LeaseID != "tok-reg" {
		t.Errorf("unexpected LeaseID: %s", alerts[0].LeaseID)
	}
}
