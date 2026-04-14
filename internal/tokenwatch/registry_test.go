package tokenwatch_test

import (
	"testing"

	"github.com/vaultwatch/internal/tokenwatch"
)

func TestRegistry_AddAndList(t *testing.T) {
	r := tokenwatch.NewRegistry()
	if err := r.Add("acc-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := r.Add("acc-2"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Len() != 2 {
		t.Errorf("expected 2 accessors, got %d", r.Len())
	}
	list := r.List()
	if len(list) != 2 {
		t.Errorf("expected list length 2, got %d", len(list))
	}
}

func TestRegistry_Add_Duplicate(t *testing.T) {
	r := tokenwatch.NewRegistry()
	_ = r.Add("acc-1")
	err := r.Add("acc-1")
	if err == nil {
		t.Fatal("expected error for duplicate accessor")
	}
}

func TestRegistry_Add_Empty(t *testing.T) {
	r := tokenwatch.NewRegistry()
	err := r.Add("")
	if err == nil {
		t.Fatal("expected error for empty accessor")
	}
}

func TestRegistry_Remove(t *testing.T) {
	r := tokenwatch.NewRegistry()
	_ = r.Add("acc-1")
	if err := r.Remove("acc-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Len() != 0 {
		t.Errorf("expected 0 accessors after remove, got %d", r.Len())
	}
}

func TestRegistry_Remove_NotFound(t *testing.T) {
	r := tokenwatch.NewRegistry()
	err := r.Remove("nonexistent")
	if err == nil {
		t.Fatal("expected error for missing accessor")
	}
}

func TestRegistry_List_Empty(t *testing.T) {
	r := tokenwatch.NewRegistry()
	list := r.List()
	if list == nil {
		t.Fatal("expected non-nil slice for empty registry")
	}
	if len(list) != 0 {
		t.Errorf("expected empty list, got %v", list)
	}
}
