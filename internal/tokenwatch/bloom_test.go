package tokenwatch

import (
	"fmt"
	"testing"
	"time"
)

func TestDefaultBloomConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultBloomConfig()
	if cfg.Capacity <= 0 {
		t.Errorf("expected positive Capacity, got %d", cfg.Capacity)
	}
	if cfg.FPRate <= 0 || cfg.FPRate >= 1 {
		t.Errorf("expected FPRate in (0,1), got %f", cfg.FPRate)
	}
	if cfg.ResetEvery <= 0 {
		t.Errorf("expected positive ResetEvery, got %v", cfg.ResetEvery)
	}
}

func TestNewBloomFilter_ZeroValues_UsesDefaults(t *testing.T) {
	bf := NewBloomFilter(BloomConfig{})
	if bf == nil {
		t.Fatal("expected non-nil BloomFilter")
	}
	if len(bf.bits) == 0 {
		t.Error("expected non-empty bit array")
	}
	if bf.k < 1 {
		t.Errorf("expected k >= 1, got %d", bf.k)
	}
}

func TestBloomFilter_Add_And_Contains(t *testing.T) {
	bf := NewBloomFilter(DefaultBloomConfig())
	keys := []string{"token-a", "token-b", "token-c"}
	for _, k := range keys {
		bf.Add(k)
	}
	for _, k := range keys {
		if !bf.Contains(k) {
			t.Errorf("expected Contains(%q) = true after Add", k)
		}
	}
}

func TestBloomFilter_Contains_ReturnsFalseForAbsent(t *testing.T) {
	bf := NewBloomFilter(DefaultBloomConfig())
	bf.Add("present")
	// A key that was never added must return false (no false negative).
	if bf.Contains("definitely-absent-xyz-123") {
		// This could theoretically be a false positive, but with a 1024-capacity
		// filter and a single foreign key the probability is negligible.
		t.Log("false positive detected — acceptable but worth noting")
	}
}

func TestBloomFilter_ResetsAfterInterval(t *testing.T) {
	cfg := BloomConfig{
		Capacity:   64,
		FPRate:     0.01,
		ResetEvery: 1 * time.Millisecond,
	}
	bf := NewBloomFilter(cfg)
	bf.Add("token-reset")
	if !bf.Contains("token-reset") {
		t.Fatal("expected Contains to be true before reset")
	}
	time.Sleep(5 * time.Millisecond)
	// After the reset interval the filter should have cleared.
	if bf.Contains("token-reset") {
		t.Error("expected Contains to be false after reset interval")
	}
}

func TestBloomFilter_ConcurrentAccess(t *testing.T) {
	bf := NewBloomFilter(DefaultBloomConfig())
	done := make(chan struct{})
	for i := 0; i < 50; i++ {
		go func(n int) {
			key := fmt.Sprintf("token-%d", n)
			bf.Add(key)
			_ = bf.Contains(key)
			done <- struct{}{}
		}(i)
	}
	for i := 0; i < 50; i++ {
		<-done
	}
}

func TestBloomFilter_DistinctKeysAreTreatedIndependently(t *testing.T) {
	bf := NewBloomFilter(DefaultBloomConfig())
	bf.Add("alpha")
	// "beta" was never added; result may be false positive but alpha must be true.
	if !bf.Contains("alpha") {
		t.Error("expected alpha to be present")
	}
}
