package tokenwatch_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/vaultwatch/internal/alert"
	"github.com/vaultwatch/internal/tokenwatch"
)

func makeInfo(accessor string, ttl time.Duration) *tokenwatch.TokenInfo {
	return &tokenwatch.TokenInfo{
		Accessor:    accessor,
		DisplayName: "test-token",
		TTL:         ttl,
	}
}

func TestNew_NilLookup(t *testing.T) {
	_, err := tokenwatch.New(nil, nil, time.Hour, time.Minute)
	if err == nil {
		t.Fatal("expected error for nil lookup")
	}
}

func TestNew_InvalidThresholds(t *testing.T) {
	lookup := func(_ context.Context, _ string) (*tokenwatch.TokenInfo, error) { return nil, nil }
	_, err := tokenwatch.New(lookup, nil, time.Minute, time.Hour)
	if err == nil {
		t.Fatal("expected error when critical >= warn")
	}
}

func TestCheck_NoAlerts_WhenTTLHigh(t *testing.T) {
	lookup := func(_ context.Context, acc string) (*tokenwatch.TokenInfo, error) {
		return makeInfo(acc, 24*time.Hour), nil
	}
	w, _ := tokenwatch.New(lookup, []string{"acc1"}, time.Hour, 10*time.Minute)
	alerts, err := w.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(alerts) != 0 {
		t.Fatalf("expected no alerts, got %d", len(alerts))
	}
}

func TestCheck_WarningAlert(t *testing.T) {
	lookup := func(_ context.Context, acc string) (*tokenwatch.TokenInfo, error) {
		return makeInfo(acc, 30*time.Minute), nil
	}
	w, _ := tokenwatch.New(lookup, []string{"acc1"}, time.Hour, 10*time.Minute)
	alerts, err := w.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(alerts))
	}
	if alerts[0].Level != alert.Warning {
		t.Errorf("expected Warning, got %v", alerts[0].Level)
	}
}

func TestCheck_CriticalAlert(t *testing.T) {
	lookup := func(_ context.Context, acc string) (*tokenwatch.TokenInfo, error) {
		return makeInfo(acc, 5*time.Minute), nil
	}
	w, _ := tokenwatch.New(lookup, []string{"acc1"}, time.Hour, 10*time.Minute)
	alerts, err := w.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(alerts))
	}
	if alerts[0].Level != alert.Critical {
		t.Errorf("expected Critical, got %v", alerts[0].Level)
	}
}

func TestCheck_LookupError_Propagates(t *testing.T) {
	lookupErr := errors.New("vault unavailable")
	lookup := func(_ context.Context, _ string) (*tokenwatch.TokenInfo, error) {
		return nil, lookupErr
	}
	w, _ := tokenwatch.New(lookup, []string{"acc1"}, time.Hour, 10*time.Minute)
	_, err := w.Check(context.Background())
	if err == nil {
		t.Fatal("expected error from lookup")
	}
	if !errors.Is(err, lookupErr) {
		t.Errorf("expected wrapped lookupErr, got %v", err)
	}
}
