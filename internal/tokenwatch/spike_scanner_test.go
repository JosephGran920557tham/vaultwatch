package tokenwatch

import (
	"errors"
	"testing"
	"time"
)

func newTestSpikeScanner(lookup func(string) (TokenInfo, error)) (*SpikeScanner, *Registry) {
	reg := NewRegistry()
	det := NewSpikeDetector(SpikeConfig{Window: time.Minute, MinSamples: 2, MaxIncrease: 0.5})
	return NewSpikeScanner(reg, det, lookup, nil), reg
}

func TestNewSpikeScanner_NilRegistry_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil registry")
		}
	}()
	NewSpikeScanner(nil, NewSpikeDetector(SpikeConfig{}), func(string) (TokenInfo, error) { return TokenInfo{}, nil }, nil)
}

func TestNewSpikeScanner_NilDetector_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil detector")
		}
	}()
	NewSpikeScanner(NewRegistry(), nil, func(string) (TokenInfo, error) { return TokenInfo{}, nil }, nil)
}

func TestNewSpikeScanner_NilLookup_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil lookup")
		}
	}()
	NewSpikeScanner(NewRegistry(), NewSpikeDetector(SpikeConfig{}), nil, nil)
}

func TestSpikeScanner_Scan_NoTokens_ReturnsEmpty(t *testing.T) {
	sc, _ := newTestSpikeScanner(func(string) (TokenInfo, error) { return TokenInfo{TTL: 10 * time.Minute}, nil })
	if got := sc.Scan(); len(got) != 0 {
		t.Errorf("expected empty, got %d alerts", len(got))
	}
}

func TestSpikeScanner_Scan_LookupError_Skipped(t *testing.T) {
	sc, reg := newTestSpikeScanner(func(string) (TokenInfo, error) {
		return TokenInfo{}, errors.New("vault unavailable")
	})
	_ = reg.Add("tok-err")
	if got := sc.Scan(); len(got) != 0 {
		t.Errorf("expected no alerts on error, got %d", len(got))
	}
}

func TestSpikeScanner_Scan_DetectsSpike(t *testing.T) {
	calls := 0
	ttls := []time.Duration{10 * time.Minute, 10 * time.Minute, 30 * time.Minute}
	lookup := func(string) (TokenInfo, error) {
		ttl := ttls[calls%len(ttls)]
		calls++
		return TokenInfo{TTL: ttl}, nil
	}
	sc, reg := newTestSpikeScanner(lookup)
	_ = reg.Add("tok-spike")
	// first two scans: stable
	sc.Scan()
	sc.Scan()
	// third scan: spike
	alerts := sc.Scan()
	if len(alerts) == 0 {
		t.Error("expected spike alert on third scan")
	}
}
