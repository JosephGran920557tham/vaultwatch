package tokenwatch

import (
	"testing"
	"time"

	"github.com/vaultwatch/internal/alert"
)

func TestDefaultTopologyConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultTopologyConfig()
	if cfg.MaxNeighbors <= 0 {
		t.Fatal("expected positive MaxNeighbors")
	}
	if cfg.StaleAfter <= 0 {
		t.Fatal("expected positive StaleAfter")
	}
	if cfg.MinLinks <= 0 {
		t.Fatal("expected positive MinLinks")
	}
}

func TestNewTopologyDetector_ZeroValues_UsesDefaults(t *testing.T) {
	d := NewTopologyDetector(TopologyConfig{})
	def := DefaultTopologyConfig()
	if d.cfg.MinLinks != def.MinLinks {
		t.Fatalf("MinLinks: got %d, want %d", d.cfg.MinLinks, def.MinLinks)
	}
}

func TestTopologyDetector_Check_NoNode_ReturnsNil(t *testing.T) {
	d := NewTopologyDetector(DefaultTopologyConfig())
	if a := d.Check("unknown-token"); a != nil {
		t.Fatalf("expected nil, got %v", a)
	}
}

func TestTopologyDetector_Check_SufficientLinks_ReturnsNil(t *testing.T) {
	cfg := DefaultTopologyConfig()
	cfg.MinLinks = 2
	d := NewTopologyDetector(cfg)
	d.Link("tok-a", "tok-b")
	d.Link("tok-a", "tok-c")
	if a := d.Check("tok-a"); a != nil {
		t.Fatalf("expected nil for healthy token, got %v", a)
	}
}

func TestTopologyDetector_Check_InsufficientLinks_ReturnsWarning(t *testing.T) {
	cfg := DefaultTopologyConfig()
	cfg.MinLinks = 3
	d := NewTopologyDetector(cfg)
	d.Link("tok-a", "tok-b")
	a := d.Check("tok-a")
	if a == nil {
		t.Fatal("expected warning alert, got nil")
	}
	if a.Level != alert.LevelWarning {
		t.Fatalf("expected Warning, got %v", a.Level)
	}
}

func TestTopologyDetector_Check_StaleLinks_NotCounted(t *testing.T) {
	cfg := DefaultTopologyConfig()
	cfg.MinLinks = 2
	cfg.StaleAfter = 10 * time.Millisecond
	d := NewTopologyDetector(cfg)
	now := time.Now()
	d.now = func() time.Time { return now }
	d.Link("tok-a", "tok-b")
	d.Link("tok-a", "tok-c")
	// advance clock past stale threshold
	d.now = func() time.Time { return now.Add(20 * time.Millisecond) }
	a := d.Check("tok-a")
	if a == nil {
		t.Fatal("expected warning after links became stale")
	}
}

func TestTopologyScanner_Scan_NoTokens_ReturnsEmpty(t *testing.T) {
	reg := NewRegistry()
	det := NewTopologyDetector(DefaultTopologyConfig())
	lookup := func(id string) (TokenInfo, error) { return TokenInfo{}, nil }
	scanner := NewTopologyScanner(reg, det, lookup)
	alerts, err := scanner.Scan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(alerts) != 0 {
		t.Fatalf("expected no alerts, got %d", len(alerts))
	}
}

func TestTopologyScanner_Scan_LinksSharedLabelPeers(t *testing.T) {
	reg := NewRegistry()
	_ = reg.Add("tok-1")
	_ = reg.Add("tok-2")
	_ = reg.Add("tok-3")
	infos := map[string]TokenInfo{
		"tok-1": {Labels: map[string]string{"env": "prod"}},
		"tok-2": {Labels: map[string]string{"env": "prod"}},
		"tok-3": {Labels: map[string]string{"env": "prod"}},
	}
	lookup := func(id string) (TokenInfo, error) { return infos[id], nil }
	cfg := DefaultTopologyConfig()
	cfg.MinLinks = 2
	det := NewTopologyDetector(cfg)
	scanner := NewTopologyScanner(reg, det, lookup)
	alerts, err := scanner.Scan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// each token should have 2 links (the other two peers) — no alerts expected
	if len(alerts) != 0 {
		t.Fatalf("expected no alerts for fully-linked topology, got %d", len(alerts))
	}
}
