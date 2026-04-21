package tokenwatch

import (
	"errors"
	"testing"
	"time"
)

func newTestQuorumScanner(lookup func(string) (TokenInfo, error)) *QuorumScanner {
	reg := NewRegistry()
	det := NewQuorumDetector(QuorumConfig{
		MinVoters:   2,
		Threshold:   0.6,
		WindowSize:  time.Minute,
		CriticalMin: 0.9,
	})
	return NewQuorumScanner(reg, det, lookup)
}

func TestNewQuorumScanner_NilRegistry_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil registry")
		}
	}()
	NewQuorumScanner(nil, NewQuorumDetector(QuorumConfig{}), func(string) (TokenInfo, error) { return TokenInfo{}, nil })
}

func TestNewQuorumScanner_NilDetector_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil detector")
		}
	}()
	NewQuorumScanner(NewRegistry(), nil, func(string) (TokenInfo, error) { return TokenInfo{}, nil })
}

func TestNewQuorumScanner_NilLookup_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil lookup")
		}
	}()
	NewQuorumScanner(NewRegistry(), NewQuorumDetector(QuorumConfig{}), nil)
}

func TestQuorumScanner_Scan_NoTokens_ReturnsEmpty(t *testing.T) {
	s := newTestQuorumScanner(func(string) (TokenInfo, error) { return TokenInfo{TTL: time.Hour}, nil })
	got := s.Scan()
	if len(got) != 0 {
		t.Errorf("expected empty, got %d alerts", len(got))
	}
}

func TestQuorumScanner_Scan_LookupError_Skipped(t *testing.T) {
	reg := NewRegistry()
	_ = reg.Add("tok-err")
	det := NewQuorumDetector(QuorumConfig{MinVoters: 2, Threshold: 0.6, WindowSize: time.Minute, CriticalMin: 0.9})
	s := NewQuorumScanner(reg, det, func(string) (TokenInfo, error) { return TokenInfo{}, errors.New("lookup failed") })
	got := s.Scan()
	if len(got) != 0 {
		t.Errorf("expected empty when lookup errors, got %d", len(got))
	}
}

func TestQuorumScanner_Scan_AlertWhenQuorumFails(t *testing.T) {
	reg := NewRegistry()
	_ = reg.Add("tok-q")
	det := NewQuorumDetector(QuorumConfig{MinVoters: 2, Threshold: 0.6, WindowSize: time.Minute, CriticalMin: 0.9})
	// Pre-seed failing votes so quorum is already broken before Scan
	det.Vote("tok-q", false)
	s := NewQuorumScanner(reg, det, func(string) (TokenInfo, error) {
		return TokenInfo{TTL: 0}, nil // TTL=0 → unhealthy vote
	})
	got := s.Scan()
	if len(got) == 0 {
		t.Error("expected at least one quorum alert")
	}
	if got[0].LeaseID != "tok-q" {
		t.Errorf("expected LeaseID tok-q, got %s", got[0].LeaseID)
	}
}
