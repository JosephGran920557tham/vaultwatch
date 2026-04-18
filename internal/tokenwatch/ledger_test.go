package tokenwatch

import (
	"testing"
	"time"
)

func TestDefaultLedgerConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultLedgerConfig()
	if cfg.MaxEntries <= 0 {
		t.Fatal("expected positive MaxEntries")
	}
	if cfg.TTL <= 0 {
		t.Fatal("expected positive TTL")
	}
}

func TestNewLedger_ZeroValues_UsesDefaults(t *testing.T) {
	l := NewLedger(LedgerConfig{})
	def := DefaultLedgerConfig()
	if l.cfg.MaxEntries != def.MaxEntries {
		t.Errorf("got MaxEntries %d, want %d", l.cfg.MaxEntries, def.MaxEntries)
	}
}

func TestLedger_Record_And_List(t *testing.T) {
	l := NewLedger(DefaultLedgerConfig())
	l.Record(LedgerEntry{TokenID: "tok1", TTL: 5 * time.Minute})
	l.Record(LedgerEntry{TokenID: "tok2", TTL: 2 * time.Minute})

	entries := l.List("tok1")
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry for tok1, got %d", len(entries))
	}
	if entries[0].TokenID != "tok1" {
		t.Errorf("unexpected token id %q", entries[0].TokenID)
	}
}

func TestLedger_Len_CountsAllTokens(t *testing.T) {
	l := NewLedger(DefaultLedgerConfig())
	l.Record(LedgerEntry{TokenID: "a", TTL: time.Minute})
	l.Record(LedgerEntry{TokenID: "b", TTL: time.Minute})
	if l.Len() != 2 {
		t.Errorf("expected Len 2, got %d", l.Len())
	}
}

func TestLedger_EvictsExpiredEntries(t *testing.T) {
	l := NewLedger(LedgerConfig{MaxEntries: 100, TTL: 10 * time.Millisecond})
	l.Record(LedgerEntry{TokenID: "old", Timestamp: time.Now().Add(-time.Second), TTL: time.Minute})
	time.Sleep(20 * time.Millisecond)
	if l.Len() != 0 {
		t.Errorf("expected expired entry to be evicted, got Len %d", l.Len())
	}
}

func TestLedger_CapsAtMaxEntries(t *testing.T) {
	l := NewLedger(LedgerConfig{MaxEntries: 3, TTL: time.Minute})
	for i := 0; i < 5; i++ {
		l.Record(LedgerEntry{TokenID: "tok", TTL: time.Minute})
	}
	if l.Len() > 3 {
		t.Errorf("expected at most 3 entries, got %d", l.Len())
	}
}

func TestNewLedgerScanner_NilRegistry_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil Registry")
		}
	}()
	NewLedgerScanner(nil, func(id string) (TokenInfo, error) { return TokenInfo{}, nil }, NewLedger(DefaultLedgerConfig()))
}

func TestLedgerScanner_Scan_PopulatesLedger(t *testing.T) {
	reg := NewRegistry()
	_ = reg.Add("tok1")

	lookup := func(id string) (TokenInfo, error) {
		return TokenInfo{ID: id, TTL: 10 * time.Minute}, nil
	}
	l := NewLedger(DefaultLedgerConfig())
	s := NewLedgerScanner(reg, lookup, l)
	s.Scan()

	if l.Len() != 1 {
		t.Errorf("expected 1 entry after scan, got %d", l.Len())
	}
	entries := l.List("tok1")
	if len(entries) == 0 {
		t.Fatal("expected entry for tok1")
	}
	if entries[0].TTL != 10*time.Minute {
		t.Errorf("unexpected TTL %v", entries[0].TTL)
	}
}
