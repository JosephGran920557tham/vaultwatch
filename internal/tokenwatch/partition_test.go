package tokenwatch

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/your-org/vaultwatch/internal/alert"
)

func TestDefaultPartitionConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultPartitionConfig()
	if cfg.MaxBuckets <= 0 {
		t.Fatalf("expected positive MaxBuckets, got %d", cfg.MaxBuckets)
	}
	if cfg.BucketTTL <= 0 {
		t.Fatalf("expected positive BucketTTL, got %v", cfg.BucketTTL)
	}
}

func TestNewPartition_ZeroValues_UsesDefaults(t *testing.T) {
	p := NewPartition(PartitionConfig{})
	def := DefaultPartitionConfig()
	if p.cfg.MaxBuckets != def.MaxBuckets {
		t.Errorf("expected MaxBuckets=%d, got %d", def.MaxBuckets, p.cfg.MaxBuckets)
	}
	if p.cfg.BucketTTL != def.BucketTTL {
		t.Errorf("expected BucketTTL=%v, got %v", def.BucketTTL, p.cfg.BucketTTL)
	}
}

func TestPartition_Assign_And_Tokens(t *testing.T) {
	p := NewPartition(PartitionConfig{MaxBuckets: 4, BucketTTL: time.Minute})
	bucket := p.Assign("token-abc")
	tokens := p.Tokens(bucket)
	if len(tokens) != 1 || tokens[0] != "token-abc" {
		t.Errorf("expected [token-abc], got %v", tokens)
	}
}

func TestPartition_Tokens_EmptyBucket(t *testing.T) {
	p := NewPartition(PartitionConfig{MaxBuckets: 4, BucketTTL: time.Minute})
	if got := p.Tokens(3); got != nil {
		t.Errorf("expected nil for empty bucket, got %v", got)
	}
}

func TestPartition_Evict_RemovesStale(t *testing.T) {
	now := time.Now()
	p := NewPartition(PartitionConfig{MaxBuckets: 8, BucketTTL: time.Minute})
	p.now = func() time.Time { return now }
	p.Assign("old-token")
	p.now = func() time.Time { return now.Add(2 * time.Minute) }
	removed := p.Evict()
	if removed == 0 {
		t.Error("expected at least one stale bucket to be evicted")
	}
}

func TestNewPartitionScanner_NilRegistry_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil registry")
		}
	}()
	NewPartitionScanner(nil, NewPartition(PartitionConfig{}), nil, nil, nil)
}

func TestNewPartitionScanner_NilPartition_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil partition")
		}
	}()
	NewPartitionScanner(NewRegistry(), nil, nil, nil, nil)
}

func TestPartitionScanner_Scan_NoTokens_ReturnsEmpty(t *testing.T) {
	reg := NewRegistry()
	p := NewPartition(PartitionConfig{MaxBuckets: 4, BucketTTL: time.Minute})
	lookup := func(_ context.Context, _ string) (TokenInfo, error) { return TokenInfo{}, nil }
	detect := func(_ context.Context, _ TokenInfo) (*alert.Alert, error) { return nil, nil }
	scanner := NewPartitionScanner(reg, p, lookup, detect, nil)
	alerts, err := scanner.Scan(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(alerts) != 0 {
		t.Errorf("expected 0 alerts, got %d", len(alerts))
	}
}

func TestPartitionScanner_Scan_LookupError_Skipped(t *testing.T) {
	reg := NewRegistry()
	_ = reg.Add("tok-1")
	p := NewPartition(PartitionConfig{MaxBuckets: 4, BucketTTL: time.Minute})
	lookup := func(_ context.Context, _ string) (TokenInfo, error) {
		return TokenInfo{}, errors.New("lookup failed")
	}
	detect := func(_ context.Context, _ TokenInfo) (*alert.Alert, error) { return nil, nil }
	scanner := NewPartitionScanner(reg, p, lookup, detect, nil)
	alerts, err := scanner.Scan(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(alerts) != 0 {
		t.Errorf("expected 0 alerts after skipped lookup error, got %d", len(alerts))
	}
}
