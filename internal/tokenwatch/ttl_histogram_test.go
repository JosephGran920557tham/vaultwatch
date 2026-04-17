package tokenwatch

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestDefaultTTLHistogram_HasBuckets(t *testing.T) {
	h := DefaultTTLHistogram()
	if len(h.Buckets()) == 0 {
		t.Fatal("expected non-empty buckets")
	}
}

func TestTTLHistogram_Record_CorrectBucket(t *testing.T) {
	h := NewTTLHistogram([]time.Duration{
		5 * time.Minute,
		30 * time.Minute,
		2 * time.Hour,
	})
	h.Record(3 * time.Minute)
	h.Record(10 * time.Minute)
	h.Record(1 * time.Hour)
	h.Record(5 * time.Hour)

	buckets := h.Buckets()
	if buckets[0].Count != 1 {
		t.Errorf("bucket 0 want 1 got %d", buckets[0].Count)
	}
	if buckets[1].Count != 1 {
		t.Errorf("bucket 1 want 1 got %d", buckets[1].Count)
	}
	if buckets[2].Count != 1 {
		t.Errorf("bucket 2 want 1 got %d", buckets[2].Count)
	}
	if buckets[3].Count != 1 {
		t.Errorf("overflow bucket want 1 got %d", buckets[3].Count)
	}
}

func TestTTLHistogram_Record_ExactBoundary(t *testing.T) {
	h := NewTTLHistogram([]time.Duration{10 * time.Minute})
	h.Record(10 * time.Minute)
	buckets := h.Buckets()
	if buckets[0].Count != 1 {
		t.Errorf("exact boundary should fall in first bucket, got %d", buckets[0].Count)
	}
}

func TestTTLHistogram_Buckets_Labels(t *testing.T) {
	h := NewTTLHistogram([]time.Duration{1 * time.Hour})
	buckets := h.Buckets()
	if len(buckets) != 2 {
		t.Fatalf("expected 2 buckets, got %d", len(buckets))
	}
	if !strings.Contains(buckets[0].Label, "<=") {
		t.Errorf("first bucket label should contain '<=' got %q", buckets[0].Label)
	}
	if !strings.Contains(buckets[1].Label, ">") {
		t.Errorf("overflow bucket label should contain '>' got %q", buckets[1].Label)
	}
}

func TestTTLHistogram_Print_ContainsCounts(t *testing.T) {
	h := NewTTLHistogram([]time.Duration{5 * time.Minute})
	h.Record(1 * time.Minute)
	h.Record(1 * time.Minute)
	var buf bytes.Buffer
	h.Print(&buf)
	output := buf.String()
	if !strings.Contains(output, "2") {
		t.Errorf("expected count 2 in output, got: %s", output)
	}
}

func TestTTLHistogram_SortsBoundaries(t *testing.T) {
	h := NewTTLHistogram([]time.Duration{1 * time.Hour, 5 * time.Minute, 30 * time.Minute})
	h.Record(10 * time.Minute)
	buckets := h.Buckets()
	// 10m should land in the <=30m bucket (index 1 after sorting)
	if buckets[1].Count != 1 {
		t.Errorf("expected 10m in second bucket after sort, got %+v", buckets)
	}
}
