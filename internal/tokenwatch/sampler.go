package tokenwatch

import (
	"sync"
	"time"
)

// DefaultSamplerConfig returns sensible defaults for the Sampler.
func DefaultSamplerConfig() SamplerConfig {
	return SamplerConfig{
		MaxSamples: 60,
		MaxAge:     10 * time.Minute,
	}
}

// SamplerConfig controls how many TTL samples are retained per token.
type SamplerConfig struct {
	MaxSamples int
	MaxAge     time.Duration
}

// sample holds a single TTL observation.
type sample struct {
	TTL       time.Duration
	RecordedAt time.Time
}

// Sampler records TTL observations per token and returns recent samples.
type Sampler struct {
	mu      sync.Mutex
	cfg     SamplerConfig
	buckets map[string][]sample
}

// NewSampler creates a Sampler with the given config, applying defaults for
// zero values.
func NewSampler(cfg SamplerConfig) *Sampler {
	def := DefaultSamplerConfig()
	if cfg.MaxSamples <= 0 {
		cfg.MaxSamples = def.MaxSamples
	}
	if cfg.MaxAge <= 0 {
		cfg.MaxAge = def.MaxAge
	}
	return &Sampler{
		cfg:     cfg,
		buckets: make(map[string][]sample),
	}
}

// Record adds a TTL observation for the given token ID.
func (s *Sampler) Record(tokenID string, ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	bucket := s.buckets[tokenID]

	// Evict stale entries.
	cutoff := now.Add(-s.cfg.MaxAge)
	filtered := bucket[:0]
	for _, sm := range bucket {
		if sm.RecordedAt.After(cutoff) {
			filtered = append(filtered, sm)
		}
	}

	filtered = append(filtered, sample{TTL: ttl, RecordedAt: now})
	if len(filtered) > s.cfg.MaxSamples {
		filtered = filtered[len(filtered)-s.cfg.MaxSamples:]
	}
	s.buckets[tokenID] = filtered
}

// Samples returns a copy of recent TTL observations for the given token ID.
func (s *Sampler) Samples(tokenID string) []time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()

	bucket := s.buckets[tokenID]
	out := make([]time.Duration, len(bucket))
	for i, sm := range bucket {
		out[i] = sm.TTL
	}
	return out
}

// Clear removes all samples for the given token ID.
func (s *Sampler) Clear(tokenID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.buckets, tokenID)
}
