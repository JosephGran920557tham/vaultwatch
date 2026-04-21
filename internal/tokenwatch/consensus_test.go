package tokenwatch

import (
	"testing"
	"time"

	"github.com/vaultwatch/internal/alert"
)

func makeConsensusAlert(id, level string) alert.Alert {
	return alert.Alert{LeaseID: id, Level: level, Message: "test"}
}

func TestDefaultConsensusConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultConsensusConfig()
	if cfg.Quorum < 2 {
		t.Errorf("expected Quorum >= 2, got %d", cfg.Quorum)
	}
	if cfg.Window <= 0 {
		t.Error("expected positive Window")
	}
	if cfg.MaxKeys <= 0 {
		t.Error("expected positive MaxKeys")
	}
}

func TestNewConsensus_ZeroValues_UsesDefaults(t *testing.T) {
	c := NewConsensus(ConsensusConfig{})
	if c.cfg.Quorum != DefaultConsensusConfig().Quorum {
		t.Errorf("expected default Quorum, got %d", c.cfg.Quorum)
	}
}

func TestConsensus_Vote_BelowQuorum_ReturnsFalse(t *testing.T) {
	c := NewConsensus(ConsensusConfig{Quorum: 3, Window: time.Minute, MaxKeys: 100})
	a := makeConsensusAlert("lease-1", "warning")
	if c.Vote("src-a", a) {
		t.Error("expected false before quorum reached")
	}
	if c.Vote("src-b", a) {
		t.Error("expected false at 2 of 3")
	}
}

func TestConsensus_Vote_ReachesQuorum_ReturnsTrue(t *testing.T) {
	c := NewConsensus(ConsensusConfig{Quorum: 2, Window: time.Minute, MaxKeys: 100})
	a := makeConsensusAlert("lease-2", "critical")
	c.Vote("src-a", a)
	if !c.Vote("src-b", a) {
		t.Error("expected true when quorum reached")
	}
}

func TestConsensus_Vote_SameSource_DoesNotDoubleCount(t *testing.T) {
	c := NewConsensus(ConsensusConfig{Quorum: 2, Window: time.Minute, MaxKeys: 100})
	a := makeConsensusAlert("lease-3", "warning")
	c.Vote("src-a", a)
	if c.Vote("src-a", a) {
		t.Error("same source should not count twice")
	}
}

func TestConsensus_Vote_ExpiredVotes_Evicted(t *testing.T) {
	now := time.Now()
	c := NewConsensus(ConsensusConfig{Quorum: 2, Window: 10 * time.Millisecond, MaxKeys: 100})
	c.now = func() time.Time { return now }

	a := makeConsensusAlert("lease-4", "info")
	c.Vote("src-a", a)

	// Advance time past window.
	c.now = func() time.Time { return now.Add(20 * time.Millisecond) }
	c.Vote("src-b", a) // triggers evict; src-a gone

	// src-a vote expired, so only src-b remains — below quorum.
	if c.Vote("src-c", a) {
		t.Error("expected false: src-a vote should have been evicted")
	}
}

func TestConsensus_Vote_MaxKeys_Enforced(t *testing.T) {
	c := NewConsensus(ConsensusConfig{Quorum: 2, Window: time.Minute, MaxKeys: 1})
	a1 := makeConsensusAlert("lease-10", "warning")
	a2 := makeConsensusAlert("lease-11", "warning")
	c.Vote("src-a", a1) // fills the single slot
	if c.Vote("src-b", a2) {
		t.Error("expected false: MaxKeys exceeded")
	}
}
