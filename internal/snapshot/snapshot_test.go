package snapshot_test

import (
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/alert"
	"github.com/yourusername/vaultwatch/internal/snapshot"
)

func makeAlert(id string) alert.Alert {
	return alert.Alert{
		LeaseID:   id,
		Path:      "secret/data/" + id,
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}
}

func TestCapture_SetsTimestampAndEntries(t *testing.T) {
	alerts := []alert.Alert{makeAlert("a"), makeAlert("b")}
	before := time.Now().UTC()
	snap := snapshot.Capture(alerts)
	after := time.Now().UTC()

	if snap.TakenAt.Before(before) || snap.TakenAt.After(after) {
		t.Errorf("TakenAt out of expected range")
	}
	if len(snap.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(snap.Entries))
	}
	if snap.Entries[0].Alert.LeaseID != "a" {
		t.Errorf("unexpected lease id: %s", snap.Entries[0].Alert.LeaseID)
	}
}

func TestCapture_EmptyAlerts(t *testing.T) {
	snap := snapshot.Capture([]alert.Alert{})
	if len(snap.Entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(snap.Entries))
	}
	if snap.TakenAt.IsZero() {
		t.Error("expected TakenAt to be set even for empty alert list")
	}
}

func TestStore_SaveAndLatest(t *testing.T) {
	store := snapshot.NewStore()
	if store.Latest() != nil {
		t.Fatal("expected nil before first save")
	}
	snap := snapshot.Capture([]alert.Alert{makeAlert("x")})
	store.Save(snap)
	if store.Latest() != snap {
		t.Error("Latest did not return saved snapshot")
	}
}

func TestDiff_NilPrev_ReturnsAll(t *testing.T) {
	next := snapshot.Capture([]alert.Alert{makeAlert("1"), makeAlert("2")})
	diff := snapshot.Diff(nil, next)
	if len(diff) != 2 {
		t.Errorf("expected 2, got %d", len(diff))
	}
}

func TestDiff_ReturnsOnlyNew(t *testing.T) {
	prev := snapshot.Capture([]alert.Alert{makeAlert("a"), makeAlert("b")})
	next := snapshot.Capture([]alert.Alert{makeAlert("a"), makeAlert("b"), makeAlert("c")})
	diff := snapshot.Diff(prev, next)
	if len(diff) != 1 {
		t.Fatalf("expected 1 new alert, got %d", len(diff))
	}
	if diff[0].LeaseID != "c" {
		t.Errorf("expected lease c, got %s", diff[0].LeaseID)
	}
}

func TestDiff_NoNewAlerts(t *testing.T) {
	prev := snapshot.Capture([]alert.Alert{makeAlert("a")})
	next := snapshot.Capture([]alert.Alert{makeAlert("a")})
	diff := snapshot.Diff(prev, next)
	if len(diff) != 0 {
		t.Errorf("expected 0 new alerts, got %d", len(diff))
	}
}
