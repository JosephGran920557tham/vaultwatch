package tokenwatch

import (
	"testing"
	"time"
)

func TestNewThrottle_DefaultsOnZeroInterval(t *testing.T) {
	th := NewThrottle(ThrottleConfig{MinInterval: 0})
	if th.cfg.MinInterval != DefaultThrottleConfig().MinInterval {
		t.Errorf("expected default interval %v, got %v", DefaultThrottleConfig().MinInterval, th.cfg.MinInterval)
	}
}

func TestThrottle_Allow_FirstCallPermitted(t *testing.T) {
	th := NewThrottle(ThrottleConfig{MinInterval: time.Minute})
	if !th.Allow("tok-1") {
		t.Error("expected first call to be allowed")
	}
}

func TestThrottle_Allow_SecondCallSuppressed(t *testing.T) {
	th := NewThrottle(ThrottleConfig{MinInterval: time.Minute})
	th.Allow("tok-1")
	if th.Allow("tok-1") {
		t.Error("expected second call within interval to be suppressed")
	}
}

func TestThrottle_Allow_AfterInterval_Permitted(t *testing.T) {
	th := NewThrottle(ThrottleConfig{MinInterval: 10 * time.Millisecond})
	th.Allow("tok-2")
	time.Sleep(20 * time.Millisecond)
	if !th.Allow("tok-2") {
		t.Error("expected call after interval to be allowed")
	}
}

func TestThrottle_Allow_DifferentTokensIndependent(t *testing.T) {
	th := NewThrottle(ThrottleConfig{MinInterval: time.Minute})
	th.Allow("tok-a")
	if !th.Allow("tok-b") {
		t.Error("expected different token to be allowed independently")
	}
}

func TestThrottle_Reset_ClearsState(t *testing.T) {
	th := NewThrottle(ThrottleConfig{MinInterval: time.Minute})
	th.Allow("tok-3")
	th.Reset("tok-3")
	if !th.Allow("tok-3") {
		t.Error("expected allow after reset")
	}
}

func TestThrottle_ResetAll_ClearsAllState(t *testing.T) {
	th := NewThrottle(ThrottleConfig{MinInterval: time.Minute})
	th.Allow("tok-x")
	th.Allow("tok-y")
	th.ResetAll()
	if !th.Allow("tok-x") || !th.Allow("tok-y") {
		t.Error("expected all tokens allowed after ResetAll")
	}
}
