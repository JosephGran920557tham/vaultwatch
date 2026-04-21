package tokenwatch

import (
	"context"
	"sync"
	"testing"
	"time"
)

// fakeReaperStore is an in-memory ReaperStore for testing.
type fakeReaperStore struct {
	mu      sync.Mutex
	entries map[string]time.Time
}

func newFakeReaperStore(entries map[string]time.Time) *fakeReaperStore {
	if entries == nil {
		entries = make(map[string]time.Time)
	}
	return &fakeReaperStore{entries: entries}
}

func (s *fakeReaperStore) List() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]string, 0, len(s.entries))
	for id := range s.entries {
		out = append(out, id)
	}
	return out
}

func (s *fakeReaperStore) LastSeen(id string) (time.Time, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.entries[id]
	return t, ok
}

func (s *fakeReaperStore) Evict(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, id)
}

func TestDefaultReaperConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultReaperConfig()
	if cfg.Interval <= 0 {
		t.Errorf("expected positive Interval, got %v", cfg.Interval)
	}
	if cfg.MaxAge <= 0 {
		t.Errorf("expected positive MaxAge, got %v", cfg.MaxAge)
	}
	if cfg.BatchSize <= 0 {
		t.Errorf("expected positive BatchSize, got %d", cfg.BatchSize)
	}
}

func TestNewReaper_NilStore_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil store")
		}
	}()
	NewReaper(DefaultReaperConfig(), nil, nil)
}

func TestNewReaper_ZeroValues_UsesDefaults(t *testing.T) {
	store := newFakeReaperStore(nil)
	r := NewReaper(ReaperConfig{}, store, nil)
	if r.cfg.Interval != DefaultReaperConfig().Interval {
		t.Errorf("expected default interval")
	}
}

func TestReaper_Reap_EvictsStaleTokens(t *testing.T) {
	old := time.Now().Add(-2 * time.Hour)
	recent := time.Now()
	store := newFakeReaperStore(map[string]time.Time{
		"stale":  old,
		"active": recent,
	})
	cfg := ReaperConfig{Interval: time.Minute, MaxAge: time.Hour, BatchSize: 10}
	r := NewReaper(cfg, store, nil)
	evicted := r.Reap()
	if len(evicted) != 1 || evicted[0] != "stale" {
		t.Errorf("expected [stale] evicted, got %v", evicted)
	}
	if r.TotalReaped() != 1 {
		t.Errorf("expected TotalReaped=1, got %d", r.TotalReaped())
	}
}

func TestReaper_Reap_RespectsMaxAge_KeepsRecent(t *testing.T) {
	store := newFakeReaperStore(map[string]time.Time{
		"fresh": time.Now(),
	})
	cfg := ReaperConfig{Interval: time.Minute, MaxAge: time.Hour, BatchSize: 10}
	r := NewReaper(cfg, store, nil)
	evicted := r.Reap()
	if len(evicted) != 0 {
		t.Errorf("expected no evictions, got %v", evicted)
	}
}

func TestReaper_Run_StopsOnContextCancel(t *testing.T) {
	store := newFakeReaperStore(nil)
	cfg := ReaperConfig{Interval: 10 * time.Millisecond, MaxAge: time.Hour, BatchSize: 10}
	r := NewReaper(cfg, store, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel()
	done := make(chan struct{})
	go func() {
		r.Run(ctx)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Error("Run did not stop after context cancellation")
	}
}
