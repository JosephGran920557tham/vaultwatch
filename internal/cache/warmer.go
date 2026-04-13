package cache

import (
	"context"
	"time"
)

// LeaseSource is anything that can supply a list of active lease IDs.
type LeaseSource interface {
	ListLeaseIDs(ctx context.Context) ([]string, error)
}

// LeaseLookup resolves a single lease ID into an Entry.
type LeaseLookup interface {
	LookupLease(ctx context.Context, leaseID string) (Entry, error)
}

// Warmer pre-populates a Store by fetching all active leases from Vault.
type Warmer struct {
	store  *Store
	src    LeaseSource
	lookup LeaseLookup
}

// NewWarmer constructs a Warmer.
func NewWarmer(store *Store, src LeaseSource, lookup LeaseLookup) *Warmer {
	return &Warmer{store: store, src: src, lookup: lookup}
}

// Warm fetches all lease IDs and populates the cache. It stops early if ctx
// is cancelled. Errors from individual lookups are skipped so a single bad
// lease does not abort the whole warm-up.
func (w *Warmer) Warm(ctx context.Context) error {
	ids, err := w.src.ListLeaseIDs(ctx)
	if err != nil {
		return err
	}
	for _, id := range ids {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		entry, err := w.lookup.LookupLease(ctx, id)
		if err != nil {
			continue
		}
		w.store.Set(id, entry)
	}
	return nil
}

// WarmWithTimeout is a convenience wrapper that applies a deadline.
func (w *Warmer) WarmWithTimeout(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return w.Warm(ctx)
}
