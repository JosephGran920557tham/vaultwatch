package tokenwatch

import (
	"fmt"
	"sync"
	"time"
)

// DefaultPartitionConfig returns sensible defaults for the Partition.
func DefaultPartitionConfig() PartitionConfig {
	return PartitionConfig{
		MaxBuckets: 16,
		BucketTTL:  10 * time.Minute,
	}
}

// PartitionConfig controls bucketing behaviour.
type PartitionConfig struct {
	MaxBuckets int
	BucketTTL  time.Duration
}

// partitionEntry holds the token IDs assigned to a single bucket.
type partitionEntry struct {
	tokens    map[string]struct{}
	updatedAt time.Time
}

// Partition distributes token IDs across a fixed number of buckets,
// enabling sharded processing and stale-bucket eviction.
type Partition struct {
	mu      sync.Mutex
	cfg     PartitionConfig
	buckets map[int]*partitionEntry
	now     func() time.Time
}

// NewPartition creates a Partition with the given config.
// Zero-value fields are replaced with defaults.
func NewPartition(cfg PartitionConfig) *Partition {
	def := DefaultPartitionConfig()
	if cfg.MaxBuckets <= 0 {
		cfg.MaxBuckets = def.MaxBuckets
	}
	if cfg.BucketTTL <= 0 {
		cfg.BucketTTL = def.BucketTTL
	}
	return &Partition{
		cfg:     cfg,
		buckets: make(map[int]*partitionEntry),
		now:     time.Now,
	}
}

// Assign places tokenID into the bucket derived from its hash.
func (p *Partition) Assign(tokenID string) int {
	p.mu.Lock()
	defer p.mu.Unlock()

	bucket := p.bucketFor(tokenID)
	e, ok := p.buckets[bucket]
	if !ok {
		e = &partitionEntry{tokens: make(map[string]struct{})}
		p.buckets[bucket] = e
	}
	e.tokens[tokenID] = struct{}{}
	e.updatedAt = p.now()
	return bucket
}

// Tokens returns all token IDs in the given bucket.
func (p *Partition) Tokens(bucket int) []string {
	p.mu.Lock()
	defer p.mu.Unlock()

	e, ok := p.buckets[bucket]
	if !ok {
		return nil
	}
	out := make([]string, 0, len(e.tokens))
	for id := range e.tokens {
		out = append(out, id)
	}
	return out
}

// Evict removes buckets that have not been updated within BucketTTL.
func (p *Partition) Evict() int {
	p.mu.Lock()
	defer p.mu.Unlock()

	cutoff := p.now().Add(-p.cfg.BucketTTL)
	removed := 0
	for k, e := range p.buckets {
		if e.updatedAt.Before(cutoff) {
			delete(p.buckets, k)
			removed++
		}
	}
	return removed
}

// String returns a human-readable summary.
func (p *Partition) String() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return fmt.Sprintf("Partition{buckets=%d, maxBuckets=%d}", len(p.buckets), p.cfg.MaxBuckets)
}

func (p *Partition) bucketFor(tokenID string) int {
	h := 0
	for _, c := range tokenID {
		h = (h*31 + int(c)) & 0x7fffffff
	}
	return h % p.cfg.MaxBuckets
}
