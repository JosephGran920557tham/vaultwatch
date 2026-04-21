package tokenwatch

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/your-org/vaultwatch/internal/alert"
)

func newTestTriageScanner(source func(ctx context.Context) ([]alert.Alert, error), dispatch func([]TriageEntry) error) *TriageScanner {
	return NewTriageScanner(NewTriage(DefaultTriageConfig()), source, dispatch, nil)
}

func TestNewTriageScanner_NilTriage_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil triage")
		}
	}()
	NewTriageScanner(nil, func(_ context.Context) ([]alert.Alert, error) { return nil, nil }, func(_ []TriageEntry) error { return nil }, nil)
}

func TestNewTriageScanner_NilSource_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil source")
		}
	}()
	NewTriageScanner(NewTriage(DefaultTriageConfig()), nil, func(_ []TriageEntry) error { return nil }, nil)
}

func TestNewTriageScanner_NilDispatch_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil dispatch")
		}
	}()
	NewTriageScanner(NewTriage(DefaultTriageConfig()), func(_ context.Context) ([]alert.Alert, error) { return nil, nil }, nil, nil)
}

func TestTriageScanner_Scan_DispatchesRankedAlerts(t *testing.T) {
	now := time.Now()
	source := func(_ context.Context) ([]alert.Alert, error) {
		return []alert.Alert{
			{LeaseID: "a", Level: alert.Info, FiredAt: now},
			{LeaseID: "b", Level: alert.Critical, FiredAt: now},
		}, nil
	}
	var got []TriageEntry
	dispatch := func(entries []TriageEntry) error {
		got = entries
		return nil
	}
	s := newTestTriageScanner(source, dispatch)
	if err := s.Scan(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(got))
	}
	if got[0].Alert.LeaseID != "b" {
		t.Errorf("expected critical alert first, got %q", got[0].Alert.LeaseID)
	}
}

func TestTriageScanner_Scan_SourceError_ReturnsError(t *testing.T) {
	source := func(_ context.Context) ([]alert.Alert, error) {
		return nil, errors.New("source failure")
	}
	s := newTestTriageScanner(source, func(_ []TriageEntry) error { return nil })
	if err := s.Scan(context.Background()); err == nil {
		t.Error("expected error from source failure")
	}
}

func TestTriageScanner_Scan_EmptySource_NoDispatch(t *testing.T) {
	called := false
	source := func(_ context.Context) ([]alert.Alert, error) { return nil, nil }
	dispatch := func(_ []TriageEntry) error { called = true; return nil }
	s := newTestTriageScanner(source, dispatch)
	if err := s.Scan(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Error("dispatch should not be called for empty source")
	}
}
