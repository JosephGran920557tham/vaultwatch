package tokenwatch

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// FingerprintConfig holds configuration for the fingerprint tracker.
type FingerprintConfig struct {
	// TTL is how long a fingerprint entry is retained.
	TTL time.Duration
}

// DefaultFingerprintConfig returns sensible defaults.
func DefaultFingerprintConfig() FingerprintConfig {
	return FingerprintConfig{
		TTL: 24 * time.Hour,
	}
}

// fingerprintEntry stores a computed fingerprint and its expiry.
type fingerprintEntry struct {
	hash    string
	expires time.Time
}

// Fingerprint tracks a stable hash of token metadata to detect
// unexpected identity changes across polling cycles.
type Fingerprint struct {
	cfg   FingerprintConfig
	mu    sync.Mutex
	store map[string]fingerprintEntry
}

// NewFingerprint creates a Fingerprint tracker with the given config.
// Zero-value fields are replaced with defaults.
func NewFingerprint(cfg FingerprintConfig) *Fingerprint {
	if cfg.TTL <= 0 {
		cfg.TTL = DefaultFingerprintConfig().TTL
	}
	return &Fingerprint{
		cfg:   cfg,
		store: make(map[string]fingerprintEntry),
	}
}

// Compute derives a deterministic hash from a set of key/value labels.
func Compute(labels map[string]string) string {
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var sb strings.Builder
	for _, k := range keys {
		sb.WriteString(k)
		sb.WriteByte('=')
		sb.WriteString(labels[k])
		sb.WriteByte(';')
	}
	sum := sha256.Sum256([]byte(sb.String()))
	return fmt.Sprintf("%x", sum)
}

// Track records the fingerprint for a token. It returns true if the
// fingerprint changed since the last call (or is new), false otherwise.
func (f *Fingerprint) Track(tokenID, hash string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	now := time.Now()
	entry, ok := f.store[tokenID]
	changed := !ok || entry.hash != hash
	f.store[tokenID] = fingerprintEntry{hash: hash, expires: now.Add(f.cfg.TTL)}
	return changed
}

// Evict removes entries whose TTL has elapsed.
func (f *Fingerprint) Evict() {
	f.mu.Lock()
	defer f.mu.Unlock()
	now := time.Now()
	for id, e := range f.store {
		if now.After(e.expires) {
			delete(f.store, id)
		}
	}
}
