package tokenwatch

import (
	"errors"
	"testing"
	"time"

	"github.com/vaultwatch/internal/alert"
)

// stubSource is a test AlertSource.
type stubSource struct {
	alerts []alert.Alert
	err    error
}

func (s *stubSource) Scan(_ string) ([]alert.Alert, error) {
	return s.alerts, s.err
}

func newTestConsensusScanner(quorum int, sources ...AlertSource) *ConsensusScanner {
	cfg := ConsensusConfig{Quorum: quorum, Window: time.Minute, MaxKeys: 1000}
	return NewConsensusScanner(NewConsensus(cfg), sources...)
}

func TestNewConsensusScanner_NilConsensus_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil Consensus")
		}
	}()
	NewConsensusScanner(nil, &stubSource{})
}

func TestNewConsensusScanner_NoSources_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for empty sources")
		}
	}()
	cfg := DefaultConsensusConfig()
	NewConsensusScanner(NewConsensus(cfg))
}

func TestConsensusScanner_Scan_BelowQuorum_ReturnsEmpty(t *testing.T) {
	a := makeConsensusAlert("lease-20", "warning")
	src := &stubSource{alerts: []alert.Alert{a}}
	cs := newTestConsensusScanner(2, src) // only 1 source, quorum=2
	results, err := cs.Scan("tok-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results below quorum, got %d", len(results))
	}
}

func TestConsensusScanner_Scan_QuorumReached_ReturnsAlert(t *testing.T) {
	a := makeConsensusAlert("lease-21", "critical")
	src1 := &stubSource{alerts: []alert.Alert{a}}
	src2 := &stubSource{alerts: []alert.Alert{a}}
	cs := newTestConsensusScanner(2, src1, src2)
	results, err := cs.Scan("tok-2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result at quorum, got %d", len(results))
	}
}

func TestConsensusScanner_Scan_SourceError_Skipped(t *testing.T) {
	a := makeConsensusAlert("lease-22", "warning")
	src1 := &stubSource{err: errors.New("unavailable")}
	src2 := &stubSource{alerts: []alert.Alert{a}}
	src3 := &stubSource{alerts: []alert.Alert{a}}
	cs := newTestConsensusScanner(2, src1, src2, src3)
	results, _ := cs.Scan("tok-3")
	if len(results) != 1 {
		t.Errorf("expected 1 result when one source errors, got %d", len(results))
	}
}

func TestConsensusScanner_Scan_NoDuplicates(t *testing.T) {
	a := makeConsensusAlert("lease-23", "critical")
	src1 := &stubSource{alerts: []alert.Alert{a}}
	src2 := &stubSource{alerts: []alert.Alert{a}}
	src3 := &stubSource{alerts: []alert.Alert{a}}
	cs := newTestConsensusScanner(2, src1, src2, src3)
	results, _ := cs.Scan("tok-4")
	if len(results) != 1 {
		t.Errorf("expected exactly 1 deduplicated result, got %d", len(results))
	}
}
