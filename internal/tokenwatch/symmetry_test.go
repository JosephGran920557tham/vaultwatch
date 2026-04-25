package tokenwatch

import (
	"testing"
	"time"
)

func TestDefaultSymmetryConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultSymmetryConfig()
	if cfg.MaxSkew <= 0 {
		t.Fatal("expected positive MaxSkew")
	}
	if cfg.MinPeers < 2 {
		t.Fatal("expected MinPeers >= 2")
	}
}

func TestNewSymmetryDetector_ZeroValues_UsesDefaults(t *testing.T) {
	d := NewSymmetryDetector(SymmetryConfig{})
	def := DefaultSymmetryConfig()
	if d.cfg.MaxSkew != def.MaxSkew {
		t.Errorf("expected MaxSkew %v, got %v", def.MaxSkew, d.cfg.MaxSkew)
	}
	if d.cfg.MinPeers != def.MinPeers {
		t.Errorf("expected MinPeers %d, got %d", def.MinPeers, d.cfg.MinPeers)
	}
}

func TestSymmetryDetector_Check_InsufficientPeers_ReturnsNil(t *testing.T) {
	d := NewSymmetryDetector(SymmetryConfig{MaxSkew: 10 * time.Second, MinPeers: 3})
	d.Observe("tok-a", 100*time.Second)
	d.Observe("tok-b", 200*time.Second)
	// only 2 peers, MinPeers=3
	if a := d.Check("tok-a"); a != nil {
		t.Fatalf("expected nil alert with insufficient peers, got %+v", a)
	}
}

func TestSymmetryDetector_Check_WithinSkew_ReturnsNil(t *testing.T) {
	d := NewSymmetryDetector(SymmetryConfig{MaxSkew: 30 * time.Second, MinPeers: 2})
	d.Observe("tok-a", 100*time.Second)
	d.Observe("tok-b", 110*time.Second)
	if a := d.Check("tok-a"); a != nil {
		t.Fatalf("expected nil alert within skew, got %+v", a)
	}
}

func TestSymmetryDetector_Check_ExceedsSkew_ReturnsWarning(t *testing.T) {
	d := NewSymmetryDetector(SymmetryConfig{MaxSkew: 10 * time.Second, MinPeers: 2})
	d.Observe("tok-a", 20*time.Second)
	d.Observe("tok-b", 200*time.Second)
	a := d.Check("tok-a")
	if a == nil {
		t.Fatal("expected warning alert for skewed token")
	}
	if a.LeaseID != "tok-a" {
		t.Errorf("expected LeaseID tok-a, got %s", a.LeaseID)
	}
}

func TestNewSymmetryScanner_NilRegistry_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil registry")
		}
	}()
	NewSymmetryScanner(nil, NewSymmetryDetector(SymmetryConfig{}), func(string) (TokenInfo, error) { return TokenInfo{}, nil })
}

func TestNewSymmetryScanner_NilDetector_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil detector")
		}
	}()
	NewSymmetryScanner(NewRegistry(), nil, func(string) (TokenInfo, error) { return TokenInfo{}, nil })
}

func TestNewSymmetryScanner_NilLookup_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil lookup")
		}
	}()
	NewSymmetryScanner(NewRegistry(), NewSymmetryDetector(SymmetryConfig{}), nil)
}

func TestSymmetryScanner_Scan_NoTokens_ReturnsEmpty(t *testing.T) {
	s := NewSymmetryScanner(
		NewRegistry(),
		NewSymmetryDetector(SymmetryConfig{MaxSkew: 10 * time.Second, MinPeers: 2}),
		func(string) (TokenInfo, error) { return TokenInfo{}, nil },
	)
	alerts, err := s.Scan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(alerts) != 0 {
		t.Fatalf("expected no alerts, got %d", len(alerts))
	}
}
