package tokenwatch

import (
	"context"
	"testing"
)

func TestNewBatchRunner_NilRegistry(t *testing.T) {
	_, err := NewBatchRunner(nil, &Alerter{}, 2)
	if err == nil {
		t.Fatal("expected error for nil registry")
	}
}

func TestNewBatchRunner_NilAlerter(t *testing.T) {
	reg := NewRegistry()
	_, err := NewBatchRunner(reg, nil, 2)
	if err == nil {
		t.Fatal("expected error for nil alerter")
	}
}

func TestNewBatchRunner_DefaultConcurrency(t *testing.T) {
	reg := NewRegistry()
	br, err := NewBatchRunner(reg, &Alerter{}, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if br.concurrency != 4 {
		t.Errorf("expected concurrency 4, got %d", br.concurrency)
	}
}

func TestBatchRunner_Run_EmptyRegistry(t *testing.T) {
	reg := NewRegistry()
	br, _ := NewBatchRunner(reg, &Alerter{}, 2)
	results := br.Run(context.Background())
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestBatchRunner_Run_ReturnsOneResultPerToken(t *testing.T) {
	reg := NewRegistry()
	_ = reg.Add("tok-a")
	_ = reg.Add("tok-b")
	_ = reg.Add("tok-c")

	// Alerter with no-op internals — CheckToken will error; we just check result count.
	br, _ := NewBatchRunner(reg, &Alerter{}, 2)
	results := br.Run(context.Background())
	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}
	ids := map[string]bool{}
	for _, r := range results {
		ids[r.TokenID] = true
	}
	for _, id := range []string{"tok-a", "tok-b", "tok-c"} {
		if !ids[id] {
			t.Errorf("missing result for token %s", id)
		}
	}
}
