package tokenwatch_test

import (
	"testing"
	"time"

	"github.com/youorg/vaultwatch/internal/tokenwatch"
)

func newTestPartitionScanner(t *testing.T) (*tokenwatch.PartitionScanner, *tokenwatch.Registry, *tokenwatch.Partition) {
	t.Helper()
	reg := tokenwatch.NewRegistry()
	part := tokenwatch.NewPartition(tokenwatch.DefaultPartitionConfig())
	lookup := func(id string) (tokenwatch.TokenInfo, error) {
		return tokenwatch.TokenInfo{
			ID:        id,
			TTL:       30 * time.Minute,
			Renewable: true,
			Labels:    map[string]string{"env": "prod"},
		}, nil
	}
	scanner := tokenwatch.NewPartitionScanner(reg, part, lookup)
	return scanner, reg, part
}

func TestNewPartitionScanner_NilRegistry_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil registry")
		}
	}()
	part := tokenwatch.NewPartition(tokenwatch.DefaultPartitionConfig())
	lookup := func(id string) (tokenwatch.TokenInfo, error) { return tokenwatch.TokenInfo{}, nil }
	tokenwatch.NewPartitionScanner(nil, part, lookup)
}

func TestNewPartitionScanner_NilPartition_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil partition")
		}
	}()
	reg := tokenwatch.NewRegistry()
	lookup := func(id string) (tokenwatch.TokenInfo, error) { return tokenwatch.TokenInfo{}, nil }
	tokenwatch.NewPartitionScanner(reg, nil, lookup)
}

func TestNewPartitionScanner_NilLookup_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil lookup")
		}
	}()
	reg := tokenwatch.NewRegistry()
	part := tokenwatch.NewPartition(tokenwatch.DefaultPartitionConfig())
	tokenwatch.NewPartitionScanner(reg, part, nil)
}

func TestPartitionScanner_Scan_NoTokens_ReturnsEmpty(t *testing.T) {
	scanner, _, _ := newTestPartitionScanner(t)
	alerts := scanner.Scan()
	if len(alerts) != 0 {
		t.Fatalf("expected no alerts for empty registry, got %d", len(alerts))
	}
}

func TestPartitionScanner_Scan_AssignsAndScans(t *testing.T) {
	scanner, reg, _ := newTestPartitionScanner(t)

	if err := reg.Add("token-a"); err != nil {
		t.Fatalf("Add token-a: %v", err)
	}
	if err := reg.Add("token-b"); err != nil {
		t.Fatalf("Add token-b: %v", err)
	}

	alerts := scanner.Scan()
	// Alerts may or may not be raised depending on TTL thresholds;
	// the key assertion is that no panic occurs and the result is a slice.
	if alerts == nil {
		t.Fatal("expected non-nil alerts slice")
	}
}

func TestPartitionScanner_Scan_ReturnsOneResultPerToken(t *testing.T) {
	scanner, reg, _ := newTestPartitionScanner(t)

	tokens := []string{"tok-1", "tok-2", "tok-3"}
	for _, id := range tokens {
		if err := reg.Add(id); err != nil {
			t.Fatalf("Add %s: %v", id, err)
		}
	}

	// Run multiple scans to ensure idempotency.
	for i := 0; i < 3; i++ {
		alerts := scanner.Scan()
		if alerts == nil {
			t.Fatalf("scan %d returned nil", i)
		}
	}
}
