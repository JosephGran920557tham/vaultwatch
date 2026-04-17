package tokenwatch

import (
	"context"
	"errors"
	"testing"
	"time"
)

func newTestAgingScanner(t *testing.T) (*AgingScanner, *Registry) {
	t.Helper()
	reg := NewRegistry()
	det := NewAgingDetector(DefaultAgingConfig())
	lookup := func(_ context.Context, id string) (TokenInfo, error) {
		return TokenInfo{CreatedAt: time.Now().Add(-48 * time.Hour), TTL: time.Hour}, nil
	}
	return NewAgingScanner(reg, det, lookup), reg
}

func TestNewAgingScanner_NilRegistry_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil registry")
		}
	}()
	NewAgingScanner(nil, NewAgingDetector(DefaultAgingConfig()), func(_ context.Context, _ string) (TokenInfo, error) { return TokenInfo{}, nil })
}

func TestNewAgingScanner_NilDetector_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil detector")
		}
	}()
	NewAgingScanner(NewRegistry(), nil, func(_ context.Context, _ string) (TokenInfo, error) { return TokenInfo{}, nil })
}

func TestNewAgingScanner_NilLookup_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil lookup")
		}
	}()
	NewAgingScanner(NewRegistry(), NewAgingDetector(DefaultAgingConfig()), nil)
}

func TestAgingScanner_Scan_NoTokens_ReturnsEmpty(t *testing.T) {
	scanner, _ := newTestAgingScanner(t)
	alerts := scanner.Scan(context.Background())
	if len(alerts) != 0 {
		t.Fatalf("expected 0 alerts, got %d", len(alerts))
	}
}

func TestAgingScanner_Scan_LookupError_Skipped(t *testing.T) {
	reg := NewRegistry()
	det := NewAgingDetector(DefaultAgingConfig())
	lookup := func(_ context.Context, _ string) (TokenInfo, error) {
		return TokenInfo{}, errors.New("vault unavailable")
	}
	scanner := NewAgingScanner(reg, det, lookup)
	_ = reg.Add("tok-err")
	alerts := scanner.Scan(context.Background())
	if len(alerts) != 0 {
		t.Fatalf("expected 0 alerts on lookup error, got %d", len(alerts))
	}
}

func TestAgingScanner_Scan_OldToken_ReturnsAlert(t *testing.T) {
	reg := NewRegistry()
	cfg := DefaultAgingConfig()
	cfg.WarningAge = time.Hour
	cfg.CriticalAge = 10 * time.Hour
	det := NewAgingDetector(cfg)
	lookup := func(_ context.Context, _ string) (TokenInfo, error) {
		return TokenInfo{CreatedAt: time.Now().Add(-5 * time.Hour), TTL: time.Hour}, nil
	}
	scanner := NewAgingScanner(reg, det, lookup)
	_ = reg.Add("tok-old")
	alerts := scanner.Scan(context.Background())
	if len(alerts) == 0 {
		t.Fatal("expected at least one aging alert")
	}
	if alerts[0].LeaseID != "tok-old" {
		t.Errorf("unexpected lease ID: %s", alerts[0].LeaseID)
	}
}
