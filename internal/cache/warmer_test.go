package cache_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/cache"
)

// stubSource implements LeaseSource.
type stubSource struct {
	ids []string
	err error
}

func (s *stubSource) ListLeaseIDs(_ context.Context) ([]string, error) {
	return s.ids, s.err
}

// stubLookup implements LeaseLookup.
type stubLookup struct {
	entries map[string]cache.Entry
	err     error
}

func (l *stubLookup) LookupLease(_ context.Context, id string) (cache.Entry, error) {
	if l.err != nil {
		return cache.Entry{}, l.err
	}
	return l.entries[id], nil
}

func TestWarmer_Warm_PopulatesStore(t *testing.T) {
	store := cache.New(5 * time.Minute)
	src := &stubSource{ids: []string{"lease-a", "lease-b"}}
	lookup := &stubLookup{entries: map[string]cache.Entry{
		"lease-a": {LeaseID: "lease-a", ExpiresAt: time.Now().Add(time.Hour)},
		"lease-b": {LeaseID: "lease-b", ExpiresAt: time.Now().Add(2 * time.Hour)},
	}}

	w := cache.NewWarmer(store, src, lookup)
	if err := w.Warm(context.Background()); err != nil {
		t.Fatalf("Warm() error: %v", err)
	}
	if store.Len() != 2 {
		t.Errorf("store.Len() = %d, want 2", store.Len())
	}
}

func TestWarmer_Warm_SourceError(t *testing.T) {
	store := cache.New(5 * time.Minute)
	src := &stubSource{err: errors.New("vault unavailable")}
	w := cache.NewWarmer(store, src, &stubLookup{})

	if err := w. == nil {
		t.Fatal("expected error from source, got nil")
	}
}

func TestWarmer_Warm_LookupErrorSkipped(t *testing.T) {
	store := cache.New(5 * time.Minute)
	src := &stubSource{ids: []string{"lease-x"}}
	lookup := &stubLookup{err: errors.New("not found")}
	w := cache.NewWarmer(store, src, lookup)

	if err := w.Warm(context.Background()); err != nil {
		t.Fatalf("Warm() should not propagate lookup errors, got: %v", err)
	}
	if store.Len() != 0 {
		t.Errorf("expected empty store after lookup errors, got %d", store.Len())
	}
}

func TestWarmer_WarmWithTimeout(t *testing.T) {
	store := cache.New(5 * time.Minute)
	src := &stubSource{ids: []string{"lease-t"}}
	lookup := &stubLookup{entries: map[string]cache.Entry{
		"lease-t": {LeaseID: "lease-t", ExpiresAt: time.Now().Add(time.Hour)},
	}}
	w := cache.NewWarmer(store, src, lookup)

	if err := w.WarmWithTimeout(5 * time.Second); err != nil {
		t.Fatalf("WarmWithTimeout() error: %v", err)
	}
	if store.Len() != 1 {
		t.Errorf("store.Len() = %d, want 1", store.Len())
	}
}

func TestWarmer_Warm_EmptySource(t *testing.T) {
	// Warming with no lease IDs should succeed and leave the store empty.
	store := cache.New(5 * time.Minute)
	src := &stubSource{ids: []string{}}
	w := cache.NewWarmer(store, src, &stubLookup{})

	if err := w.Warm(context.Background()); err != nil {
		t.Fatalf("Warm() error on empty source: %v", err)
	}
	if store.Len() != 0 {
		t.Errorf("store.Len() = %d, want 0", store.Len())
	}
}
