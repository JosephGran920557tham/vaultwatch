package tokenwatch

import (
	"testing"
	"time"
)

func TestDefaultFingerprintConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultFingerprintConfig()
	if cfg.TTL <= 0 {
		t.Fatalf("expected positive TTL, got %v", cfg.TTL)
	}
}

func TestNewFingerprint_ZeroTTL_UsesDefault(t *testing.T) {
	fp := NewFingerprint(FingerprintConfig{})
	if fp.cfg.TTL != DefaultFingerprintConfig().TTL {
		t.Fatalf("expected default TTL, got %v", fp.cfg.TTL)
	}
}

func TestCompute_DeterministicAndOrderIndependent(t *testing.T) {
	a := Compute(map[string]string{"role": "admin", "env": "prod"})
	b := Compute(map[string]string{"env": "prod", "role": "admin"})
	if a != b {
		t.Fatalf("expected same hash regardless of map order, got %q vs %q", a, b)
	}
}

func TestCompute_DifferentLabels_DifferentHash(t *testing.T) {
	a := Compute(map[string]string{"role": "admin"})
	b := Compute(map[string]string{"role": "viewer"})
	if a == b {
		t.Fatal("expected different hashes for different labels")
	}
}

func TestFingerprint_Track_NewToken_ReturnsTrue(t *testing.T) {
	fp := NewFingerprint(DefaultFingerprintConfig())
	changed := fp.Track("tok-1", "abc123")
	if !changed {
		t.Fatal("expected changed=true for new token")
	}
}

func TestFingerprint_Track_SameHash_ReturnsFalse(t *testing.T) {
	fp := NewFingerprint(DefaultFingerprintConfig())
	fp.Track("tok-1", "abc123")
	changed := fp.Track("tok-1", "abc123")
	if changed {
		t.Fatal("expected changed=false for same hash")
	}
}

func TestFingerprint_Track_DifferentHash_ReturnsTrue(t *testing.T) {
	fp := NewFingerprint(DefaultFingerprintConfig())
	fp.Track("tok-1", "abc123")
	changed := fp.Track("tok-1", "xyz789")
	if !changed {
		t.Fatal("expected changed=true when hash differs")
	}
}

func TestFingerprint_Evict_RemovesExpiredEntries(t *testing.T) {
	fp := NewFingerprint(FingerprintConfig{TTL: time.Millisecond})
	fp.Track("tok-1", "abc123")
	time.Sleep(5 * time.Millisecond)
	fp.Evict()
	// After eviction, the token should be treated as new.
	changed := fp.Track("tok-1", "abc123")
	if !changed {
		t.Fatal("expected changed=true after eviction")
	}
}

func TestFingerprint_Evict_KeepsFreshEntries(t *testing.T) {
	fp := NewFingerprint(DefaultFingerprintConfig())
	fp.Track("tok-fresh", "hash1")
	fp.Evict()
	changed := fp.Track("tok-fresh", "hash1")
	if changed {
		t.Fatal("expected changed=false — fresh entry should survive eviction")
	}
}
