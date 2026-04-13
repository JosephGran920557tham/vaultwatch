package cache_test

import (
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/cache"
)

func makeEntry(leaseID string) cache.Entry {
	return cache.Entry{
		LeaseID:   leaseID,
		ExpiresAt: time.Now().Add(time.Hour),
		Meta:      map[string]string{"path": "secret/data/foo"},
	}
}

func TestStore_SetAndGet(t *testing.T) {
	s := cache.New(5 * time.Minute)
	e := makeEntry("lease-1")
	s.Set("lease-1", e)

	got, ok := s.Get("lease-1")
	if !ok {
		t.Fatal("expected entry to be present")
	}
	if got.LeaseID != "lease-1" {
		t.Errorf("got LeaseID %q, want %q", got.LeaseID, "lease-1")
	}
}

func TestStore_Get_Missing(t *testing.T) {
	s := cache.New(5 * time.Minute)
	_, ok := s.Get("nonexistent")
	if ok {
		t.Fatal("expected miss for unknown lease")
	}
}

func TestStore_Get_Stale(t *testing.T) {
	s := cache.New(1 * time.Millisecond)
	s.Set("lease-stale", makeEntry("lease-stale"))
	time.Sleep(5 * time.Millisecond)

	_, ok := s.Get("lease-stale")
	if ok {
		t.Fatal("expected stale entry to be a miss")
	}
}

func TestStore_Delete(t *testing.T) {
	s := cache.New(5 * time.Minute)
	s.Set("lease-del", makeEntry("lease-del"))
	s.Delete("lease-del")

	_, ok := s.Get("lease-del")
	if ok {
		t.Fatal("expected entry to be deleted")
	}
}

func TestStore_Purge(t *testing.T) {
	s := cache.New(1 * time.Millisecond)
	s.Set("a", makeEntry("a"))
	s.Set("b", makeEntry("b"))
	time.Sleep(5 * time.Millisecond)

	removed := s.Purge()
	if removed != 2 {
		t.Errorf("Purge removed %d entries, want 2", removed)
	}
	if s.Len() != 0 {
		t.Errorf("Len() = %d after purge, want 0", s.Len())
	}
}

func TestStore_Len(t *testing.T) {
	s := cache.New(5 * time.Minute)
	if s.Len() != 0 {
		t.Fatalf("expected empty store, got %d", s.Len())
	}
	s.Set("x", makeEntry("x"))
	s.Set("y", makeEntry("y"))
	if s.Len() != 2 {
		t.Errorf("Len() = %d, want 2", s.Len())
	}
}
