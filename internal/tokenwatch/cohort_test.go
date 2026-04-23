package tokenwatch

import (
	"testing"
	"time"
)

func TestDefaultCohortConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultCohortConfig()
	if cfg.MaxAge <= 0 {
		t.Fatal("expected positive MaxAge")
	}
	if cfg.MaxSize <= 0 {
		t.Fatal("expected positive MaxSize")
	}
	if cfg.GroupKey == "" {
		t.Fatal("expected non-empty GroupKey")
	}
}

func TestNewCohort_ZeroValues_UsesDefaults(t *testing.T) {
	c := NewCohort(CohortConfig{})
	if c.cfg.MaxAge <= 0 {
		t.Fatal("expected default MaxAge")
	}
	if c.cfg.MaxSize <= 0 {
		t.Fatal("expected default MaxSize")
	}
}

func TestCohort_Add_And_Members(t *testing.T) {
	c := NewCohort(DefaultCohortConfig())
	c.Add("prod", "token-1")
	c.Add("prod", "token-2")
	c.Add("staging", "token-3")

	prod := c.Members("prod")
	if len(prod) != 2 {
		t.Fatalf("expected 2 prod members, got %d", len(prod))
	}
	staging := c.Members("staging")
	if len(staging) != 1 {
		t.Fatalf("expected 1 staging member, got %d", len(staging))
	}
}

func TestCohort_Members_Expired_Excluded(t *testing.T) {
	c := NewCohort(CohortConfig{MaxAge: 10 * time.Millisecond, MaxSize: 10, GroupKey: "env"})
	frozen := time.Now()
	c.nowFunc = func() time.Time { return frozen }
	c.Add("prod", "token-old")

	// advance time past MaxAge
	c.nowFunc = func() time.Time { return frozen.Add(20 * time.Millisecond) }
	members := c.Members("prod")
	if len(members) != 0 {
		t.Fatalf("expected 0 members after expiry, got %d", len(members))
	}
}

func TestCohort_Groups_ReturnsNonEmptyGroups(t *testing.T) {
	c := NewCohort(DefaultCohortConfig())
	c.Add("a", "t1")
	c.Add("b", "t2")
	groups := c.Groups()
	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}
}

func TestCohort_MaxSize_CapsEntries(t *testing.T) {
	c := NewCohort(CohortConfig{MaxAge: time.Hour, MaxSize: 2, GroupKey: "env"})
	c.Add("prod", "t1")
	c.Add("prod", "t2")
	c.Add("prod", "t3") // should evict t1
	members := c.Members("prod")
	if len(members) != 2 {
		t.Fatalf("expected 2 members after cap, got %d", len(members))
	}
	for _, m := range members {
		if m == "t1" {
			t.Fatal("t1 should have been evicted")
		}
	}
}
