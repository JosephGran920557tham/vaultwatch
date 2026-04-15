package tokenwatch

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/vaultwatch/internal/alert"
)

func newTestAnomalyScanner(source TTLSource) *AnomalyScanner {
	reg := NewRegistry()
	_ = reg.Add("tok-1")
	_ = reg.Add("tok-2")
	det := NewAnomalyDetector(AnomalyConfig{
		MinTTL: 1 * time.Minute,
		MaxTTL: 24 * time.Hour,
	})
	return NewAnomalyScanner(reg, det, source)
}

func TestNewAnomalyScanner_NilRegistry_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil registry")
		}
	}()
	NewAnomalyScanner(nil, NewAnomalyDetector(AnomalyConfig{}), func(_ context.Context, _ string) (time.Duration, error) {
		return 0, nil
	})
}

func TestNewAnomalyScanner_NilDetector_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil detector")
		}
	}()
	NewAnomalyScanner(NewRegistry(), nil, func(_ context.Context, _ string) (time.Duration, error) {
		return 0, nil
	})
}

func TestAnomalyScanner_Scan_NoAnomalies(t *testing.T) {
	scanner := newTestAnomalyScanner(func(_ context.Context, _ string) (time.Duration, error) {
		return 6 * time.Hour, nil
	})
	alerts, err := scanner.Scan(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(alerts) != 0 {
		t.Fatalf("expected no alerts, got %d", len(alerts))
	}
}

func TestAnomalyScanner_Scan_DetectsLowTTL(t *testing.T) {
	scanner := newTestAnomalyScanner(func(_ context.Context, _ string) (time.Duration, error) {
		return 5 * time.Second, nil
	})
	alerts, err := scanner.Scan(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(alerts) != 2 {
		t.Fatalf("expected 2 critical alerts, got %d", len(alerts))
	}
	for _, a := range alerts {
		if a.Level != alert.LevelCritical {
			t.Errorf("expected Critical, got %s", a.Level)
		}
	}
}

func TestAnomalyScanner_Scan_SkipsLookupErrors(t *testing.T) {
	calls := 0
	scanner := newTestAnomalyScanner(func(_ context.Context, _ string) (time.Duration, error) {
		calls++
		if calls == 1 {
			return 0, errors.New("vault unavailable")
		}
		return 6 * time.Hour, nil
	})
	alerts, err := scanner.Scan(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// First token skipped due to error, second is healthy — no alerts.
	if len(alerts) != 0 {
		t.Fatalf("expected 0 alerts, got %d", len(alerts))
	}
}
