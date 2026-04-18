package tokenwatch

import "time"

// LedgerScanner periodically snapshots token TTLs into a Ledger.
type LedgerScanner struct {
	registry *Registry
	lookup   func(id string) (TokenInfo, error)
	ledger   *Ledger
}

// NewLedgerScanner constructs a LedgerScanner.
// Panics if any argument is nil.
func NewLedgerScanner(r *Registry, lookup func(id string) (TokenInfo, error), l *Ledger) *LedgerScanner {
	if r == nil {
		panic("tokenwatch: LedgerScanner requires non-nil Registry")
	}
	if lookup == nil {
		panic("tokenwatch: LedgerScanner requires non-nil lookup")
	}
	if l == nil {
		panic("tokenwatch: LedgerScanner requires non-nil Ledger")
	}
	return &LedgerScanner{registry: r, lookup: lookup, ledger: l}
}

// Scan iterates all registered tokens, looks up their current TTL, and
// records an entry in the Ledger. Lookup errors are silently skipped.
func (s *LedgerScanner) Scan() {
	for _, id := range s.registry.List() {
		info, err := s.lookup(id)
		if err != nil {
			continue
		}
		s.ledger.Record(LedgerEntry{
			TokenID:   id,
			Timestamp: time.Now(),
			TTL:       info.TTL,
		})
	}
}
