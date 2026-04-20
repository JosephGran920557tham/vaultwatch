package tokenwatch

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/your-org/vaultwatch/internal/alert"
)

func TestDefaultIsolationConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultIsolationConfig()
	if cfg.MinPeers <= 0 {
		t.Errorf("expected positive MinPeers, got %d", cfg.MinPeers)
	}
	if cfg.Window <= 0 {
		t.Errorf("expected positive Window, got %v", cfg.Window)
	}
	if cfg.CriticalFactor <= 0 || cfg.CriticalFactor > 1 {
		t.Errorf("expected CriticalFactor in (0,1], got %v", cfg.CriticalFactor)
	}
}

func TestNewIsolationDetector_ZeroValues_UsesDefaults(t *testing.T) {
	det := NewIsolationDetector(IsolationConfig{})
	def := DefaultIsolationConfig()
	if det.cfg.MinPeers != def.MinPeers {
		t.Errorf("MinPeers: want %d, got %d", def.MinPeers, det.cfg.MinPeers)
	}
}

func TestIsolationDetector_Check_InsufficientPeers_ReturnsNil(t *testing.T) {
	det := NewIsolationDetector(IsolationConfig{MinPeers: 3, Window: time.Minute, CriticalFactor: 0.5})
	det.RecordPeer(30 * time.Second) // only 1 peer
	if a := det.Check("tok-1", 5*time.Second); a != nil {
		t.Errorf("expected nil alert with insufficient peers, got %+v", a)
	}
}

func TestIsolationDetector_Check_HealthyToken_ReturnsNil(t *testing.T) {
	det := NewIsolationDetector(IsolationConfig{MinPeers: 2, Window: time.Minute, CriticalFactor: 0.5})
	det.RecordPeer(60 * time.Second)
	det.RecordPeer(60 * time.Second)
	// token TTL equals median — not isolated
	if a := det.Check("tok-1", 60*time.Second); a != nil {
		t.Errorf("expected nil, got %+v", a)
	}
}

func TestIsolationDetector_Check_IsolatedToken_ReturnsCritical(t *testing.T) {
	det := NewIsolationDetector(IsolationConfig{MinPeers: 2, Window: time.Minute, CriticalFactor: 0.5})
	det.RecordPeer(100 * time.Second)
	det.RecordPeer(100 * time.Second)
	// token TTL is 10s, median is 100s, threshold is 50s → isolated
	a := det.Check("tok-isolated", 10*time.Second)
	if a == nil {
		t.Fatal("expected Critical alert, got nil")
	}
	if a.Level != alert.Critical {
		t.Errorf("expected Critical, got %v", a.Level)
	}
	if a.LeaseID != "tok-isolated" {
		t.Errorf("unexpected LeaseID %q", a.LeaseID)
	}
}

func TestNewIsolationScanner_NilRegistry_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil registry")
		}
	}()
	NewIsolationScanner(nil, NewIsolationDetector(IsolationConfig{}), func(string) (time.Duration, error) { return 0, nil })
}

func TestNewIsolationScanner_NilDetector_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil detector")
		}
	}()
	NewIsolationScanner(NewRegistry(), nil, func(string) (time.Duration, error) { return 0, nil })
}

func TestIsolationScanner_Scan_NoTokens_ReturnsEmpty(t *testing.T) {
	reg := NewRegistry()
	det := NewIsolationDetector(IsolationConfig{MinPeers: 2, Window: time.Minute, CriticalFactor: 0.5})
	scanner := NewIsolationScanner(reg, det, func(string) (time.Duration, error) { return 0, nil })
	alerts, err := scanner.Scan(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(alerts) != 0 {
		t.Errorf("expected empty alerts, got %d", len(alerts))
	}
}

func TestIsolationScanner_Scan_LookupError_Skipped(t *testing.T) {
	reg := NewRegistry()
	_ = reg.Add("tok-err")
	det := NewIsolationDetector(IsolationConfig{MinPeers: 2, Window: time.Minute, CriticalFactor: 0.5})
	scanner := NewIsolationScanner(reg, det, func(string) (time.Duration, error) {
		return 0, errors.New("vault unavailable")
	})
	alerts, err := scanner.Scan(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(alerts) != 0 {
		t.Errorf("expected no alerts when lookup fails, got %d", len(alerts))
	}
}
