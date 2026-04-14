package snapshot_test

import (
	"context"
	"errors"
	"testing"

	"github.com/yourusername/vaultwatch/internal/alert"
	"github.com/yourusername/vaultwatch/internal/snapshot"
)

type stubSource struct {
	alerts []alert.Alert
	err    error
}

func (s *stubSource) Alerts(_ context.Context) ([]alert.Alert, error) {
	return s.alerts, s.err
}

func TestNewWatcher_NilSource(t *testing.T) {
	_, err := snapshot.NewWatcher(nil, snapshot.NewStore(), nil)
	if err == nil {
		t.Fatal("expected error for nil source")
	}
}

func TestNewWatcher_NilStore(t *testing.T) {
	_, err := snapshot.NewWatcher(&stubSource{}, nil, nil)
	if err == nil {
		t.Fatal("expected error for nil store")
	}
}

func TestWatcher_Poll_StoresSnapshot(t *testing.T) {
	src := &stubSource{alerts: []alert.Alert{makeAlert("lease-1")}}
	store := snapshot.NewStore()
	w, err := snapshot.NewWatcher(src, store, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := w.Poll(context.Background()); err != nil {
		t.Fatalf("poll error: %v", err)
	}
	if store.Latest() == nil {
		t.Fatal("expected snapshot to be stored")
	}
	if len(store.Latest().Entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(store.Latest().Entries))
	}
}

func TestWatcher_Poll_FiresOnChange(t *testing.T) {
	src := &stubSource{alerts: []alert.Alert{makeAlert("new-lease")}}
	store := snapshot.NewStore()
	var received []alert.Alert
	w, _ := snapshot.NewWatcher(src, store, func(a []alert.Alert) {
		received = append(received, a...)
	})
	_ = w.Poll(context.Background())
	if len(received) != 1 || received[0].LeaseID != "new-lease" {
		t.Errorf("onChange not fired correctly, got %v", received)
	}
}

func TestWatcher_Poll_SourceError(t *testing.T) {
	src := &stubSource{err: errors.New("vault unavailable")}
	w, _ := snapshot.NewWatcher(src, snapshot.NewStore(), nil)
	if err := w.Poll(context.Background()); err == nil {
		t.Fatal("expected error from source")
	}
}

func TestWatcher_Poll_SecondPoll_OnlyNewAlerts(t *testing.T) {
	src := &stubSource{alerts: []alert.Alert{makeAlert("a")}}
	store := snapshot.NewStore()
	var calls int
	w, _ := snapshot.NewWatcher(src, store, func(_ []alert.Alert) { calls++ })
	_ = w.Poll(context.Background()) // first: 1 new
	_ = w.Poll(context.Background()) // second: same lease, no new
	if calls != 1 {
		t.Errorf("expected onChange called once, got %d", calls)
	}
}
