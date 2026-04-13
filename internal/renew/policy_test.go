package renew

import (
	"testing"
	"time"
)

func TestDefaultPolicy_IsValid(t *testing.T) {
	p := DefaultPolicy()
	if err := p.Validate(); err != nil {
		t.Fatalf("default policy should be valid, got: %v", err)
	}
}

func TestValidate_NegativeRenewBefore(t *testing.T) {
	p := DefaultPolicy()
	p.RenewBefore = -1 * time.Minute
	if err := p.Validate(); err == nil {
		t.Error("expected error for negative RenewBefore")
	}
}

func TestValidate_ZeroIncrement(t *testing.T) {
	p := DefaultPolicy()
	p.Increment = 0
	if err := p.Validate(); err == nil {
		t.Error("expected error for zero Increment")
	}
}

func TestValidate_NegativeMaxRetries(t *testing.T) {
	p := DefaultPolicy()
	p.MaxRetries = -1
	if err := p.Validate(); err == nil {
		t.Error("expected error for negative MaxRetries")
	}
}

func TestShouldRenew_WithinWindow(t *testing.T) {
	p := Policy{RenewBefore: 10 * time.Minute, Increment: time.Hour, MaxRetries: 3}
	expiry := time.Now().Add(5 * time.Minute)
	if !p.ShouldRenew(expiry) {
		t.Error("expected ShouldRenew to return true when within window")
	}
}

func TestShouldRenew_OutsideWindow(t *testing.T) {
	p := Policy{RenewBefore: 10 * time.Minute, Increment: time.Hour, MaxRetries: 3}
	expiry := time.Now().Add(30 * time.Minute)
	if p.ShouldRenew(expiry) {
		t.Error("expected ShouldRenew to return false when outside window")
	}
}

func TestShouldRenew_AlreadyExpired(t *testing.T) {
	p := DefaultPolicy()
	expiry := time.Now().Add(-1 * time.Minute)
	if !p.ShouldRenew(expiry) {
		t.Error("expected ShouldRenew to return true for already-expired lease")
	}
}
