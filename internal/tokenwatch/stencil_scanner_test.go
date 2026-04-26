package tokenwatch

import (
	"errors"
	"testing"

	"github.com/vaultwatch/internal/alert"
)

func newTestStencilScanner(tokens []string, stencil *Stencil) (*StencilScanner, *Registry) {
	reg := NewRegistry()
	for _, id := range tokens {
		_ = reg.Add(id)
	}
	lookup := func(id string) (TokenInfo, error) {
		return TokenInfo{ID: id, TTL: 3600}, nil
	}
	return NewStencilScanner(reg, stencil, lookup), reg
}

func TestNewStencilScanner_NilRegistry_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil registry")
		}
	}()
	NewStencilScanner(nil, NewStencil(DefaultStencilConfig()), func(string) (TokenInfo, error) { return TokenInfo{}, nil })
}

func TestNewStencilScanner_NilStencil_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil stencil")
		}
	}()
	NewStencilScanner(NewRegistry(), nil, func(string) (TokenInfo, error) { return TokenInfo{}, nil })
}

func TestNewStencilScanner_NilLookup_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil lookup")
		}
	}()
	NewStencilScanner(NewRegistry(), NewStencil(DefaultStencilConfig()), nil)
}

func TestStencilScanner_Scan_NoTokens_ReturnsEmpty(t *testing.T) {
	sc, _ := newTestStencilScanner(nil, NewStencil(DefaultStencilConfig()))
	if got := sc.Scan(); len(got) != 0 {
		t.Errorf("expected empty slice, got %d alerts", len(got))
	}
}

func TestStencilScanner_Scan_NoTemplate_ReturnsEmpty(t *testing.T) {
	st := NewStencil(DefaultStencilConfig())
	sc, _ := newTestStencilScanner([]string{"tok-1"}, st)
	if got := sc.Scan(); len(got) != 0 {
		t.Errorf("expected empty when no template set, got %d", len(got))
	}
}

func TestStencilScanner_Scan_EmitsAlertForTemplate(t *testing.T) {
	st := NewStencil(DefaultStencilConfig())
	st.Set("tok-1", "token %s has a custom alert")
	sc, _ := newTestStencilScanner([]string{"tok-1", "tok-2"}, st)

	alerts := sc.Scan()
	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(alerts))
	}
	if alerts[0].LeaseID != "tok-1" {
		t.Errorf("unexpected LeaseID: %q", alerts[0].LeaseID)
	}
	if alerts[0].Level != alert.LevelInfo {
		t.Errorf("expected Info level, got %v", alerts[0].Level)
	}
}

func TestStencilScanner_Scan_LookupError_Skipped(t *testing.T) {
	st := NewStencil(DefaultStencilConfig())
	st.Set("tok-err", "template %s")
	reg := NewRegistry()
	_ = reg.Add("tok-err")
	lookup := func(string) (TokenInfo, error) { return TokenInfo{}, errors.New("vault unavailable") }
	sc := NewStencilScanner(reg, st, lookup)

	if got := sc.Scan(); len(got) != 0 {
		t.Errorf("expected empty on lookup error, got %d", len(got))
	}
}
