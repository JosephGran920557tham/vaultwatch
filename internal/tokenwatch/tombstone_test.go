package tokenwatch

import (
	"errors"
	"testing"
	"time"

	"github.com/vaultwatch/internal/alert"
)

func TestDefaultTombstoneConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultTombstoneConfig()
	if cfg.RetentionWindow <= 0 {
		t.Fatalf("expected positive retention window, got %v", cfg.RetentionWindow)
	}
}

func TestNewTombstone_ZeroWindow_UsesDefault(t *testing.T) {
	tb := NewTombstone(TombstoneConfig{})
	if tb.cfg.RetentionWindow != DefaultTombstoneConfig().RetentionWindow {
		t.Fatalf("expected default retention window")
	}
}

func TestTombstone_Revoke_And_IsRevoked(t *testing.T) {
	tb := NewTombstone(TombstoneConfig{RetentionWindow: time.Minute})
	tb.Revoke("tok-1")
	if !tb.IsRevoked("tok-1") {
		t.Fatal("expected tok-1 to be revoked")
	}
}

func TestTombstone_IsRevoked_Unknown_ReturnsFalse(t *testing.T) {
	tb := NewTombstone(TombstoneConfig{RetentionWindow: time.Minute})
	if tb.IsRevoked("unknown") {
		t.Fatal("expected false for unknown token")
	}
}

func TestTombstone_IsRevoked_Expired_ReturnsFalse(t *testing.T) {
	tb := NewTombstone(TombstoneConfig{RetentionWindow: time.Millisecond})
	tb.Revoke("tok-exp")
	tb.now = func() time.Time { return time.Now().Add(time.Hour) }
	if tb.IsRevoked("tok-exp") {
		t.Fatal("expected expired tombstone to return false")
	}
}

func TestTombstone_Purge_RemovesExpired(t *testing.T) {
	tb := NewTombstone(TombstoneConfig{RetentionWindow: time.Millisecond})
	tb.Revoke("tok-a")
	tb.Revoke("tok-b")
	tb.now = func() time.Time { return time.Now().Add(time.Hour) }
	n := tb.Purge()
	if n != 2 {
		t.Fatalf("expected 2 purged, got %d", n)
	}
	if tb.Len() != 0 {
		t.Fatalf("expected empty tombstone after purge")
	}
}

// stubScanner is a minimal Scanner for testing.
type stubScanner struct {
	alerts []alert.Alert
	err    error
}

func (s *stubScanner) Scan() ([]alert.Alert, error) { return s.alerts, s.err }

func TestNewTombstoneScanner_NilInner_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil inner")
		}
	}()
	NewTombstoneScanner(nil, NewTombstone(TombstoneConfig{}))
}

func TestNewTombstoneScanner_NilTombstone_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil tombstone")
		}
	}()
	NewTombstoneScanner(&stubScanner{}, nil)
}

func TestTombstoneScanner_Scan_FiltersRevokedTokens(t *testing.T) {
	tb := NewTombstone(TombstoneConfig{RetentionWindow: time.Minute})
	tb.Revoke("revoked-token")

	inner := &stubScanner{
		alerts: []alert.Alert{
			{LeaseID: "revoked-token"},
			{LeaseID: "active-token"},
		},
	}
	sc := NewTombstoneScanner(inner, tb)
	results, err := sc.Scan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 || results[0].LeaseID != "active-token" {
		t.Fatalf("expected only active-token, got %v", results)
	}
}

func TestTombstoneScanner_Scan_PropagatesError(t *testing.T) {
	tb := NewTombstone(TombstoneConfig{RetentionWindow: time.Minute})
	inner := &stubScanner{err: errors.New("vault unavailable")}
	sc := NewTombstoneScanner(inner, tb)
	_, err := sc.Scan()
	if err == nil {
		t.Fatal("expected error to be propagated")
	}
}
