package tokenwatch

import (
	"testing"
	"time"
)

func TestDefaultQuotaGuardConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultQuotaGuardConfig()
	if cfg.MaxAlerts <= 0 {
		t.Errorf("expected positive MaxAlerts, got %d", cfg.MaxAlerts)
	}
	if cfg.Window <= 0 {
		t.Errorf("expected positive Window, got %s", cfg.Window)
	}
}

func TestNewQuotaGuard_ZeroValues_UsesDefaults(t *testing.T) {
	g := NewQuotaGuard(QuotaGuardConfig{})
	if g.cfg.MaxAlerts <= 0 {
		t.Error("expected defaults to be applied")
	}
}

func TestQuotaGuard_Allow_FirstCallPermitted(t *testing.T) {
	g := NewQuotaGuard(QuotaGuardConfig{MaxAlerts: 3, Window: time.Minute})
	if !g.Allow("tok1") {
		t.Error("expected first call to be allowed")
	}
}

func TestQuotaGuard_Allow_ExhaustsLimit(t *testing.T) {
	g := NewQuotaGuard(QuotaGuardConfig{MaxAlerts: 2, Window: time.Minute})
	g.Allow("tok1")
	g.Allow("tok1")
	if g.Allow("tok1") {
		t.Error("expected third call to be denied")
	}
}

func TestQuotaGuard_Allow_DifferentTokensIndependent(t *testing.T) {
	g := NewQuotaGuard(QuotaGuardConfig{MaxAlerts: 1, Window: time.Minute})
	g.Allow("tok1")
	if !g.Allow("tok2") {
		t.Error("expected tok2 to be allowed independently")
	}
}

func TestQuotaGuard_Reset_ClearsState(t *testing.T) {
	g := NewQuotaGuard(QuotaGuardConfig{MaxAlerts: 1, Window: time.Minute})
	g.Allow("tok1")
	if g.Allow("tok1") {
		t.Error("expected second call to be denied before reset")
	}
	g.Reset("tok1")
	if !g.Allow("tok1") {
		t.Error("expected call to be allowed after reset")
	}
}

func TestQuotaGuard_Allow_AfterWindowExpires_Permitted(t *testing.T) {
	g := NewQuotaGuard(QuotaGuardConfig{MaxAlerts: 1, Window: 10 * time.Millisecond})
	g.Allow("tok1")
	time.Sleep(20 * time.Millisecond)
	if !g.Allow("tok1") {
		t.Error("expected call to be allowed after window expires")
	}
}

func TestQuotaGuard_String_ContainsMax(t *testing.T) {
	g := NewQuotaGuard(QuotaGuardConfig{MaxAlerts: 5, Window: time.Minute})
	s := g.String()
	if s == "" {
		t.Error("expected non-empty string")
	}
}
