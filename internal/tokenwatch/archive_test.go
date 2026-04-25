package tokenwatch

import (
	"errors"
	"testing"
	"time"

	"github.com/vaultwatch/internal/alert"
)

func makeArchiveAlert(id string) alert.Alert {
	return alert.Alert{LeaseID: id, Level: alert.Warning}
}

func TestDefaultArchiveConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultArchiveConfig()
	if cfg.MaxEntries <= 0 {
		t.Fatalf("expected positive MaxEntries, got %d", cfg.MaxEntries)
	}
	if cfg.MaxAge <= 0 {
		t.Fatalf("expected positive MaxAge, got %v", cfg.MaxAge)
	}
}

func TestNewArchive_ZeroValues_UsesDefaults(t *testing.T) {
	a := NewArchive(ArchiveConfig{})
	def := DefaultArchiveConfig()
	if a.cfg.MaxEntries != def.MaxEntries {
		t.Fatalf("expected MaxEntries %d, got %d", def.MaxEntries, a.cfg.MaxEntries)
	}
	if a.cfg.MaxAge != def.MaxAge {
		t.Fatalf("expected MaxAge %v, got %v", def.MaxAge, a.cfg.MaxAge)
	}
}

func TestArchive_Record_And_List(t *testing.T) {
	a := NewArchive(DefaultArchiveConfig())
	a.Record(makeArchiveAlert("lease-1"))
	a.Record(makeArchiveAlert("lease-2"))

	list := a.List()
	if len(list) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(list))
	}
}

func TestArchive_Len_ReturnsCount(t *testing.T) {
	a := NewArchive(DefaultArchiveConfig())
	if a.Len() != 0 {
		t.Fatalf("expected 0, got %d", a.Len())
	}
	a.Record(makeArchiveAlert("x"))
	if a.Len() != 1 {
		t.Fatalf("expected 1, got %d", a.Len())
	}
}

func TestArchive_EvictsStaleEntries(t *testing.T) {
	a := NewArchive(ArchiveConfig{MaxEntries: 100, MaxAge: 10 * time.Millisecond})
	a.Record(makeArchiveAlert("old"))
	time.Sleep(20 * time.Millisecond)
	if a.Len() != 0 {
		t.Fatalf("expected stale entry to be evicted, got %d entries", a.Len())
	}
}

func TestArchive_CapsBeyondMaxEntries(t *testing.T) {
	a := NewArchive(ArchiveConfig{MaxEntries: 3, MaxAge: time.Hour})
	for i := 0; i < 5; i++ {
		a.Record(makeArchiveAlert("id"))
	}
	if a.Len() > 3 {
		t.Fatalf("expected at most 3 entries, got %d", a.Len())
	}
}

func TestNewArchiveScanner_NilArchive_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil archive")
		}
	}()
	NewArchiveScanner(nil, func() ([]alert.Alert, error) { return nil, nil }, nil, nil)
}

func TestNewArchiveScanner_NilSource_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil source")
		}
	}()
	NewArchiveScanner(NewArchive(DefaultArchiveConfig()), nil, nil, nil)
}

func TestArchiveScanner_Scan_RecordsAlerts(t *testing.T) {
	arch := NewArchive(DefaultArchiveConfig())
	source := func() ([]alert.Alert, error) {
		return []alert.Alert{makeArchiveAlert("a"), makeArchiveAlert("b")}, nil
	}
	s := NewArchiveScanner(arch, source, nil, nil)
	got := s.Scan()
	if len(got) != 2 {
		t.Fatalf("expected 2 alerts, got %d", len(got))
	}
	if arch.Len() != 2 {
		t.Fatalf("expected archive to hold 2 entries, got %d", arch.Len())
	}
}

func TestArchiveScanner_Scan_SourceError_ReturnsNil(t *testing.T) {
	arch := NewArchive(DefaultArchiveConfig())
	source := func() ([]alert.Alert, error) { return nil, errors.New("boom") }
	s := NewArchiveScanner(arch, source, nil, nil)
	if got := s.Scan(); got != nil {
		t.Fatalf("expected nil on source error, got %v", got)
	}
	if arch.Len() != 0 {
		t.Fatalf("expected empty archive on error, got %d", arch.Len())
	}
}

func TestArchiveScanner_Scan_ForwardsToDispatch(t *testing.T) {
	arch := NewArchive(DefaultArchiveConfig())
	var dispatched []alert.Alert
	source := func() ([]alert.Alert, error) {
		return []alert.Alert{makeArchiveAlert("fwd")}, nil
	}
	s := NewArchiveScanner(arch, source, func(al alert.Alert) {
		dispatched = append(dispatched, al)
	}, nil)
	s.Scan()
	if len(dispatched) != 1 {
		t.Fatalf("expected 1 dispatched alert, got %d", len(dispatched))
	}
}
