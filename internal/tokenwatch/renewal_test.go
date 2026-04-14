package tokenwatch

import (
	"errors"
	"testing"
	"time"
)

func TestNewRenewalManager_NilRenewer(t *testing.T) {
	_, err := NewRenewalManager(nil, DefaultExpiryClassifier())
	if err == nil {
		t.Fatal("expected error for nil renewer")
	}
}

func TestNewRenewalManager_NilClassifier(t *testing.T) {
	_, err := NewRenewalManager(func(string) (time.Duration, error) { return 0, nil }, nil)
	if err == nil {
		t.Fatal("expected error for nil classifier")
	}
}

func TestMaybeRenew_SkipsInfo(t *testing.T) {
	called := false
	renewer := func(string) (time.Duration, error) { called = true; return time.Hour, nil }
	m, _ := NewRenewalManager(renewer, DefaultExpiryClassifier())

	// TTL well above any threshold → Info level, no renewal.
	if err := m.MaybeRenew("tok", 24*time.Hour); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Error("renewer should not have been called for Info-level TTL")
	}
}

func TestMaybeRenew_RenewsOnWarning(t *testing.T) {
	called := false
	renewer := func(string) (time.Duration, error) { called = true; return time.Hour, nil }
	m, _ := NewRenewalManager(renewer, DefaultExpiryClassifier())

	// TTL within warning window.
	if err := m.MaybeRenew("tok", 20*time.Minute); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("renewer should have been called for Warning-level TTL")
	}
}

func TestMaybeRenew_SuppressesWithinCooldown(t *testing.T) {
	count := 0
	renewer := func(string) (time.Duration, error) { count++; return time.Hour, nil }
	m, _ := NewRenewalManager(renewer, DefaultExpiryClassifier())
	fixed := time.Now()
	m.now = func() time.Time { return fixed }

	_ = m.MaybeRenew("tok", 20*time.Minute)
	_ = m.MaybeRenew("tok", 20*time.Minute) // within cooldown

	if count != 1 {
		t.Errorf("expected 1 renewal, got %d", count)
	}
}

func TestMaybeRenew_AllowsAfterCooldown(t *testing.T) {
	count := 0
	renewer := func(string) (time.Duration, error) { count++; return time.Hour, nil }
	m, _ := NewRenewalManager(renewer, DefaultExpiryClassifier())
	now := time.Now()
	m.now = func() time.Time { return now }

	_ = m.MaybeRenew("tok", 20*time.Minute)
	now = now.Add(2 * time.Minute) // advance past cooldown
	_ = m.MaybeRenew("tok", 20*time.Minute)

	if count != 2 {
		t.Errorf("expected 2 renewals, got %d", count)
	}
}

func TestMaybeRenew_PropagatesError(t *testing.T) {
	expected := errors.New("vault unavailable")
	renewer := func(string) (time.Duration, error) { return 0, expected }
	m, _ := NewRenewalManager(renewer, DefaultExpiryClassifier())

	err := m.MaybeRenew("tok", 5*time.Minute)
	if err == nil {
		t.Fatal("expected error from renewer")
	}
	if !errors.Is(err, expected) {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestLastRecord_NotFound(t *testing.T) {
	m, _ := NewRenewalManager(func(string) (time.Duration, error) { return 0, nil }, DefaultExpiryClassifier())
	_, ok := m.LastRecord("missing")
	if ok {
		t.Error("expected no record for unknown token")
	}
}
