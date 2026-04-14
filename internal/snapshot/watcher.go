package snapshot

import (
	"context"
	"fmt"

	"github.com/yourusername/vaultwatch/internal/alert"
)

// Source is any type that can produce a slice of current alerts.
type Source interface {
	Alerts(ctx context.Context) ([]alert.Alert, error)
}

// ChangeHandler is called with newly appeared alerts after each poll cycle.
type ChangeHandler func(newAlerts []alert.Alert)

// Watcher polls a Source, captures snapshots, and invokes a ChangeHandler
// whenever new lease IDs appear relative to the previous snapshot.
type Watcher struct {
	source  Source
	store   *Store
	onChange ChangeHandler
}

// NewWatcher constructs a Watcher. handler may be nil (changes are silently dropped).
func NewWatcher(src Source, store *Store, handler ChangeHandler) (*Watcher, error) {
	if src == nil {
		return nil, fmt.Errorf("snapshot: source must not be nil")
	}
	if store == nil {
		return nil, fmt.Errorf("snapshot: store must not be nil")
	}
	h := handler
	if h == nil {
		h = func([]alert.Alert) {}
	}
	return &Watcher{source: src, store: store, onChange: h}, nil
}

// Poll fetches the current alerts, saves a new snapshot, and fires onChange
// with any lease IDs that did not exist in the previous snapshot.
func (w *Watcher) Poll(ctx context.Context) error {
	alerts, err := w.source.Alerts(ctx)
	if err != nil {
		return fmt.Errorf("snapshot watcher poll: %w", err)
	}
	next := Capture(alerts)
	prev := w.store.Latest()
	newAlerts := Diff(prev, next)
	w.store.Save(next)
	if len(newAlerts) > 0 {
		w.onChange(newAlerts)
	}
	return nil
}
