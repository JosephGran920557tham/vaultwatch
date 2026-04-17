package tokenwatch

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/your-org/vaultwatch/internal/alert"
)

func newTestGraceScanner(lookup GraceLookup) *GraceScanner {
	reg := NewRegistry()
	det := NewGraceDetector(GraceConfig{
		WarningBefore:  10 * time.Minute,
		CriticalBefore: 2 * time.Minute,
	})
	return NewGraceScanner(reg, det, lookup)
}

func TestNewGraceScanner_NilRegistry_Panics(t *testing.T) {
	defer func() { recover() }()
	NewGraceScanner(nil, NewGraceDetector(DefaultGraceConfig()), func(_ context.Context, _ string) (time.Duration, error) { return 0, nil })
	t.Fatal("expected panic")
}

func TestNewGraceScanner_NilDetector_Panics(t *testing.T) {
	defer func() { recover() }()
	NewGraceScanner(NewRegistry(), nil, func(_ context.Context, _ string) (time.Duration, error) { return 0, nil })
	t.Fatal("expected panic")
}

func TestNewGraceScanner_NilLookup_Panics(t *testing.T) {
	defer func() { recover() }()
	NewGraceScanner(NewRegistry(), NewGraceDetector(DefaultGraceConfig()), nil)
	t.Fatal("expected panic")
}

func TestGraceScanner_Scan_NoTokens_ReturnsEmpty(t *testing.T) {
	s := newTestGraceScanner(func(_ context.Context, _ string) (time.Duration, error) {
		return time.Hour, nil
	})
	results := s.Scan(context.Background())
	if len(results) != 0 {
		t.Errorf("expected empty, got %d", len(results))
	}
}

func TestGraceScanner_Scan_DetectsGrace(t *testing.T) {
	s := newTestGraceScanner(func(_ context.Context, _ string) (time.Duration, error) {
		return 5 * time.Minute, nil
	})
	_ = s.registry.Add("tok-grace")
	results := s.Scan(context.Background())
	if len(results) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(results))
	}
	if results[0].Level != alert.Warning {
		t.Errorf("expected Warning, got %v", results[0].Level)
	}
}

func TestGraceScanner_Scan_LookupError_Skipped(t *testing.T) {
	s := newTestGraceScanner(func(_ context.Context, _ string) (time.Duration, error) {
		return 0, errors.New("vault unavailable")
	})
	_ = s.registry.Add("tok-err")
	results := s.Scan(context.Background())
	if len(results) != 0 {
		t.Errorf("expected empty on error, got %d", len(results))
	}
}
