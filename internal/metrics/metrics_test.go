package metrics_test

import (
	"sync"
	"testing"

	"github.com/yourusername/vaultwatch/internal/metrics"
)

func TestCounter_IncAndValue(t *testing.T) {
	reg := metrics.NewRegistry()
	c := reg.Counter("checks_total")
	c.Inc()
	c.Inc()
	c.Add(3)
	if got := c.Value(); got != 5 {
		t.Fatalf("expected 5, got %d", got)
	}
}

func TestCounter_Concurrent(t *testing.T) {
	reg := metrics.NewRegistry()
	c := reg.Counter("concurrent")
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); c.Inc() }()
	}
	wg.Wait()
	if got := c.Value(); got != 100 {
		t.Fatalf("expected 100, got %d", got)
	}
}

func TestGauge_SetAndValue(t *testing.T) {
	reg := metrics.NewRegistry()
	g := reg.Gauge("expiring_leases")
	g.Set(42.5)
	if got := g.Value(); got != 42.5 {
		t.Fatalf("expected 42.5, got %f", got)
	}
	g.Set(0)
	if got := g.Value(); got != 0 {
		t.Fatalf("expected 0 after reset, got %f", got)
	}
}

func TestRegistry_SameNameReturnsSameInstance(t *testing.T) {
	reg := metrics.NewRegistry()
	c1 := reg.Counter("foo")
	c2 := reg.Counter("foo")
	c1.Inc()
	if c2.Value() != 1 {
		t.Fatal("expected same counter instance for same name")
	}
}

func TestRegistry_Snapshot(t *testing.T) {
	reg := metrics.NewRegistry()
	reg.Counter("alerts_sent").Add(7)
	reg.Gauge("vault_health").Set(1.0)

	snap := reg.Snapshot()
	if snap["alerts_sent"] != 7 {
		t.Fatalf("expected alerts_sent=7, got %f", snap["alerts_sent"])
	}
	if snap["vault_health"] != 1.0 {
		t.Fatalf("expected vault_health=1.0, got %f", snap["vault_health"])
	}
}

func TestRegistry_Snapshot_Empty(t *testing.T) {
	reg := metrics.NewRegistry()
	snap := reg.Snapshot()
	if len(snap) != 0 {
		t.Fatalf("expected empty snapshot, got %d entries", len(snap))
	}
}
