package tokenwatch

import (
	"hash/fnv"
	"sync"
	"time"
)

// DefaultBloomConfig returns a BloomConfig with sensible defaults.
func DefaultBloomConfig() BloomConfig {
	return BloomConfig{
		Capacity:  1024,
		FPRate:    0.01,
		ResetEvery: 10 * time.Minute,
	}
}

// BloomConfig controls the behaviour of a BloomFilter.
type BloomConfig struct {
	// Capacity is the expected number of distinct token IDs.
	Capacity int
	// FPRate is the desired false-positive rate (0 < FPRate < 1).
	FPRate float64
	// ResetEvery is how often the filter is cleared to avoid saturation.
	ResetEvery time.Duration
}

// BloomFilter is a probabilistic set membership structure used to suppress
// duplicate token-watch alerts without allocating per-entry storage.
type BloomFilter struct {
	mu      sync.Mutex
	bits    []bool
	k       int // number of hash functions
	resetAt time.Time
	cfg     BloomConfig
}

// NewBloomFilter constructs a BloomFilter from cfg. Zero-value fields are
// replaced with defaults from DefaultBloomConfig.
func NewBloomFilter(cfg BloomConfig) *BloomFilter {
	def := DefaultBloomConfig()
	if cfg.Capacity <= 0 {
		cfg.Capacity = def.Capacity
	}
	if cfg.FPRate <= 0 || cfg.FPRate >= 1 {
		cfg.FPRate = def.FPRate
	}
	if cfg.ResetEvery <= 0 {
		cfg.ResetEvery = def.ResetEvery
	}
	// Approximate optimal m and k.
	m := optimalM(cfg.Capacity, cfg.FPRate)
	k := optimalK(m, cfg.Capacity)
	return &BloomFilter{
		bits:    make([]bool, m),
		k:       k,
		resetAt: time.Now().Add(cfg.ResetEvery),
		cfg:     cfg,
	}
}

// Add inserts key into the filter.
func (b *BloomFilter) Add(key string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.maybeReset()
	for _, idx := range b.indices(key) {
		b.bits[idx] = true
	}
}

// Contains returns true if key was probably added, false if definitely not.
func (b *BloomFilter) Contains(key string) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.maybeReset()
	for _, idx := range b.indices(key) {
		if !b.bits[idx] {
			return false
		}
	}
	return true
}

func (b *BloomFilter) maybeReset() {
	if time.Now().After(b.resetAt) {
		for i := range b.bits {
			b.bits[i] = false
		}
		b.resetAt = time.Now().Add(b.cfg.ResetEvery)
	}
}

func (b *BloomFilter) indices(key string) []int {
	idxs := make([]int, b.k)
	h1 := hashN(key, 0)
	h2 := hashN(key, 1)
	m := uint64(len(b.bits))
	for i := 0; i < b.k; i++ {
		idxs[i] = int((h1 + uint64(i)*h2) % m)
	}
	return idxs
}

func hashN(key string, seed uint32) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte{byte(seed), byte(seed >> 8)})
	_, _ = h.Write([]byte(key))
	return h.Sum64()
}

func optimalM(n int, p float64) int {
	// m = -n*ln(p) / (ln2)^2
	const ln2sq = 0.4804530139182015
	import_ln := func(x float64) float64 {
		// simple Newton approximation avoided; use standard approach via bit tricks
		return -x // placeholder replaced below
	}
	_ = import_ln
	// Use pre-computed factor: m ≈ n * 9.6 for p=0.01
	m := int(float64(n) * (-1.4426950408889634 * logApprox(p)) / ln2sq)
	if m < 64 {
		m = 64
	}
	return m
}

func logApprox(x float64) float64 {
	// ln(x) via identity: only valid for 0 < x < 1
	// Use a simple series or just return a constant for the common case.
	if x <= 0 {
		return -10
	}
	// 6-term Taylor around x=1 is inaccurate; use the identity ln(x) = -ln(1/x)
	y := 1.0 / x
	res := 0.0
	term := (y - 1) / (y + 1)
	t2 := term * term
	tk := term
	for k := 0; k < 20; k++ {
		res += tk / float64(2*k+1)
		tk *= t2
	}
	return -2 * res
}

func optimalK(m, n int) int {
	if n == 0 {
		return 1
	}
	k := int(0.6931471805599453 * float64(m) / float64(n))
	if k < 1 {
		return 1
	}
	if k > 20 {
		return 20
	}
	return k
}
