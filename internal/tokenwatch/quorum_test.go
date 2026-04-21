package tokenwatch

import (
	"testing"
	"time"

	"github.com/vaultwatch/internal/alert"
)

func TestDefaultQuorumConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultQuorumConfig()
	if cfg.MinVoters <= 0 {
		t.Error("expected MinVoters > 0")
	}
	if cfg.Threshold <= 0 || cfg.Threshold > 1 {
		t.Error("expected Threshold in (0,1]")
	}
	if cfg.WindowSize <= 0 {
		t.Error("expected WindowSize > 0")
	}
}

func TestNewQuorumDetector_ZeroValues_UsesDefaults(t *testing.T) {
	d := NewQuorumDetector(QuorumConfig{})
	def := DefaultQuorumConfig()
	if d.cfg.MinVoters != def.MinVoters {
		t.Errorf("MinVoters: got %d, want %d", d.cfg.MinVoters, def.MinVoters)
	}
	if d.cfg.WindowSize != def.WindowSize {
		t.Errorf("WindowSize: got %v, want %v", d.cfg.WindowSize, def.WindowSize)
	}
}

func TestQuorumDetector_Check_InsufficientVoters_ReturnsNil(t *testing.T) {
	d := NewQuorumDetector(QuorumConfig{MinVoters: 3, Threshold: 0.6, WindowSize: time.Minute, CriticalMin: 0.9})
	d.Vote("tok1", true)
	d.Vote("tok1", false)
	if got := d.Check("tok1"); got != nil {
		t.Errorf("expected nil with insufficient voters, got %v", got)
	}
}

func TestQuorumDetector_Check_QuorumReached_ReturnsNil(t *testing.T) {
	d := NewQuorumDetector(QuorumConfig{MinVoters: 3, Threshold: 0.6, WindowSize: time.Minute, CriticalMin: 0.9})
	for i := 0; i < 4; i++ {
		d.Vote("tok2", true)
	}
	if got := d.Check("tok2"); got != nil {
		t.Errorf("expected nil when quorum reached, got %v", got)
	}
}

func TestQuorumDetector_Check_QuorumFailed_ReturnsWarning(t *testing.T) {
	d := NewQuorumDetector(QuorumConfig{MinVoters: 3, Threshold: 0.6, WindowSize: time.Minute, CriticalMin: 0.9})
	for i := 0; i < 3; i++ {
		d.Vote("tok3", false)
	}
	a := d.Check("tok3")
	if a == nil {
		t.Fatal("expected alert when quorum not reached")
	}
	if a.Level != alert.LevelWarning && a.Level != alert.LevelCritical {
		t.Errorf("unexpected level: %v", a.Level)
	}
	if a.LeaseID != "tok3" {
		t.Errorf("expected LeaseID tok3, got %s", a.LeaseID)
	}
}

func TestQuorumDetector_Vote_PrunesOldEntries(t *testing.T) {
	d := NewQuorumDetector(QuorumConfig{MinVoters: 2, Threshold: 0.6, WindowSize: 10 * time.Millisecond, CriticalMin: 0.9})
	d.Vote("tok4", true)
	d.Vote("tok4", true)
	time.Sleep(20 * time.Millisecond)
	d.Vote("tok4", false) // only this should remain after pruning
	if got := d.Check("tok4"); got != nil {
		// only 1 vote after pruning, below MinVoters=2
		t.Errorf("expected nil due to insufficient voters after pruning, got %v", got)
	}
}
