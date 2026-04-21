package tokenwatch

import (
	"context"
	"errors"
	"testing"
	"time"
)

func newTestCensusScanner(t *testing.T) (*Registry, *Census, *CensusScanner) {
	t.Helper()
	reg := NewRegistry()
	census := NewCensus(DefaultCensusConfig())
	lookup := func(_ context.Context, id string) (TokenInfo, error) {
		return TokenInfo{
			ID:     id,
			TTL:    30 * time.Minute,
			Labels: map[string]string{"env": "test"},
		}, nil
	}
	scanner := NewCensusScanner(reg, census, lookup)
	return reg, census, scanner
}

func TestNewCensusScanner_NilRegistry_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil registry")
		}
	}()
	NewCensusScanner(nil, NewCensus(DefaultCensusConfig()), func(_ context.Context, id string) (TokenInfo, error) {
		return TokenInfo{}, nil
	})
}

func TestNewCensusScanner_NilCensus_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil census")
		}
	}()
	NewCensusScanner(NewRegistry(), nil, func(_ context.Context, id string) (TokenInfo, error) {
		return TokenInfo{}, nil
	})
}

func TestNewCensusScanner_NilLookup_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil lookup")
		}
	}()
	NewCensusScanner(NewRegistry(), NewCensus(DefaultCensusConfig()), nil)
}

func TestCensusScanner_Scan_NoTokens_CensusEmpty(t *testing.T) {
	_, census, scanner := newTestCensusScanner(t)
	scanner.Scan(context.Background())
	if census.Len() != 0 {
		t.Fatalf("expected 0 active, got %d", census.Len())
	}
}

func TestCensusScanner_Scan_PopulatesCensus(t *testing.T) {
	reg, census, scanner := newTestCensusScanner(t)
	_ = reg.Add("tok-1")
	_ = reg.Add("tok-2")

	scanner.Scan(context.Background())

	if census.Len() != 2 {
		t.Fatalf("expected 2 active, got %d", census.Len())
	}
}

func TestCensusScanner_Scan_LookupError_Skips(t *testing.T) {
	reg := NewRegistry()
	census := NewCensus(DefaultCensusConfig())
	_ = reg.Add("tok-ok")
	_ = reg.Add("tok-fail")

	lookup := func(_ context.Context, id string) (TokenInfo, error) {
		if id == "tok-fail" {
			return TokenInfo{}, errors.New("lookup error")
		}
		return TokenInfo{ID: id, TTL: time.Minute}, nil
	}
	scanner := NewCensusScanner(reg, census, lookup)
	scanner.Scan(context.Background())

	if census.Len() != 1 {
		t.Fatalf("expected 1 active after skip, got %d", census.Len())
	}
}
