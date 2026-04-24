package tokenwatch

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/vaultwatch/internal/alert"
)

func TestDefaultBudgetConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultBudgetConfig()
	if cfg.MaxRenewals <= 0 {
		t.Errorf("expected positive MaxRenewals, got %d", cfg.MaxRenewals)
	}
	if cfg.Window <= 0 {
		t.Errorf("expected positive Window, got %v", cfg.Window)
	}
	if cfg.WarningRatio <= 0 || cfg.WarningRatio > 1 {
		t.Errorf("expected WarningRatio in (0,1], got %v", cfg.WarningRatio)
	}
}

func TestNewBudgetDetector_ZeroValues_UsesDefaults(t *testing.T) {
	d := NewBudgetDetector(BudgetConfig{})
	def := DefaultBudgetConfig()
	if d.cfg.MaxRenewals != def.MaxRenewals {
		t.Errorf("expected default MaxRenewals %d, got %d", def.MaxRenewals, d.cfg.MaxRenewals)
	}
}

func TestBudgetDetector_Check_NoRecord_ReturnsNil(t *testing.T) {
	d := NewBudgetDetector(DefaultBudgetConfig())
	if a := d.Check("tok-1"); a != nil {
		t.Errorf("expected nil alert for unseen token, got %v", a)
	}
}

func TestBudgetDetector_Check_Warning(t *testing.T) {
	cfg := BudgetConfig{MaxRenewals: 4, Window: time.Minute, WarningRatio: 0.75}
	d := NewBudgetDetector(cfg)
	// 3 renewals = 75% of 4 → warning
	for i := 0; i < 3; i++ {
		d.Record("tok-2")
	}
	a := d.Check("tok-2")
	if a == nil {
		t.Fatal("expected warning alert, got nil")
	}
	if a.Level != alert.LevelWarning {
		t.Errorf("expected warning level, got %v", a.Level)
	}
}

func TestBudgetDetector_Check_Critical(t *testing.T) {
	cfg := BudgetConfig{MaxRenewals: 3, Window: time.Minute, WarningRatio: 0.5}
	d := NewBudgetDetector(cfg)
	for i := 0; i < 3; i++ {
		d.Record("tok-3")
	}
	a := d.Check("tok-3")
	if a == nil {
		t.Fatal("expected critical alert, got nil")
	}
	if a.Level != alert.LevelCritical {
		t.Errorf("expected critical level, got %v", a.Level)
	}
}

func TestNewBudgetScanner_NilRegistry_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil registry")
		}
	}()
	NewBudgetScanner(nil, NewBudgetDetector(DefaultBudgetConfig()), func(_ context.Context, _ string) (TokenInfo, error) { return TokenInfo{}, nil })
}

func TestNewBudgetScanner_NilDetector_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil detector")
		}
	}()
	NewBudgetScanner(NewRegistry(), nil, func(_ context.Context, _ string) (TokenInfo, error) { return TokenInfo{}, nil })
}

func TestBudgetScanner_Scan_NoTokens_ReturnsEmpty(t *testing.T) {
	reg := NewRegistry()
	det := NewBudgetDetector(DefaultBudgetConfig())
	lookup := func(_ context.Context, _ string) (TokenInfo, error) { return TokenInfo{}, nil }
	s := NewBudgetScanner(reg, det, lookup)
	alerts := s.Scan(context.Background())
	if len(alerts) != 0 {
		t.Errorf("expected no alerts for empty registry, got %d", len(alerts))
	}
}

func TestBudgetScanner_Scan_LookupError_Skips(t *testing.T) {
	reg := NewRegistry()
	_ = reg.Add("tok-err")
	det := NewBudgetDetector(DefaultBudgetConfig())
	lookup := func(_ context.Context, _ string) (TokenInfo, error) {
		return TokenInfo{}, errors.New("vault unavailable")
	}
	s := NewBudgetScanner(reg, det, lookup)
	alerts := s.Scan(context.Background())
	if len(alerts) != 0 {
		t.Errorf("expected no alerts when lookup errors, got %d", len(alerts))
	}
}
