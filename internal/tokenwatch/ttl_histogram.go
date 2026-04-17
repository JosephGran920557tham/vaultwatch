package tokenwatch

import (
	"fmt"
	"io"
	"sort"
	"sync"
	"time"
)

// TTLBucket represents a histogram bucket for TTL distribution.
type TTLBucket struct {
	Label string
	Count int
}

// TTLHistogram tracks the distribution of token TTLs across configurable buckets.
type TTLHistogram struct {
	mu      sync.Mutex
	buckets []time.Duration
	counts  []int
}

// NewTTLHistogram creates a histogram with the provided bucket boundaries (sorted ascending).
// Each bucket represents TTL <= boundary. A final bucket captures everything above the last boundary.
func NewTTLHistogram(boundaries []time.Duration) *TTLHistogram {
	sorted := make([]time.Duration, len(boundaries))
	copy(sorted, boundaries)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	return &TTLHistogram{
		buckets: sorted,
		counts:  make([]int, len(sorted)+1),
	}
}

// DefaultTTLHistogram returns a histogram with sensible default boundaries.
func DefaultTTLHistogram() *TTLHistogram {
	return NewTTLHistogram([]time.Duration{
		5 * time.Minute,
		15 * time.Minute,
		1 * time.Hour,
		6 * time.Hour,
		24 * time.Hour,
	})
}

// Record places the given TTL into the appropriate bucket.
func (h *TTLHistogram) Record(ttl time.Duration) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for i, b := range h.buckets {
		if ttl <= b {
			h.counts[i]++
			return
		}
	}
	h.counts[len(h.buckets)]++
}

// Buckets returns a snapshot of all bucket labels and counts.
func (h *TTLHistogram) Buckets() []TTLBucket {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make([]TTLBucket, len(h.buckets)+1)
	for i, b := range h.buckets {
		out[i] = TTLBucket{Label: fmt.Sprintf("<=%s", b), Count: h.counts[i]}
	}
	out[len(h.buckets)] = TTLBucket{Label: fmt.Sprintf(">%s", h.buckets[len(h.buckets)-1]), Count: h.counts[len(h.buckets)]}
	return out
}

// Print writes a simple text representation of the histogram to w.
func (h *TTLHistogram) Print(w io.Writer) {
	for _, b := range h.Buckets() {
		fmt.Fprintf(w, "  %-20s %d\n", b.Label, b.Count)
	}
}
