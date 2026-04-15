package tokenwatch

import (
	"testing"
	"time"
)

func TestNewQuota_InvalidMaxAlerts(t *testing.T) {
	_, err := NewQuota(QuotaConfig{MaxAlertsPerWindow: 0, Window: time.Minute})
	if err == nil {
		t.Fatal("expected error for zero MaxAlertsPerWindow")
	}
}

func TestNewQuota_InvalidWindow(t *testing.T) {
	_, err := NewQuota(QuotaConfig{MaxAlertsPerWindow: 5, Window: 0})
	if err == nil {
		t.Fatal("expected error for zero Window")
	}
}

func TestDefaultQuotaConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultQuotaConfig()
	if cfg.MaxAlertsPerWindow <= 0 {
		t.Errorf("expected positive MaxAlertsPerWindow, got %d", cfg.MaxAlertsPerWindow)
	}
	if cfg.Window <= 0 {
		t.Errorf("expected positive Window, got %v", cfg.Window)
	}
}

func TestQuota_Allow_FirstCallPermitted(t *testing.T) {
	q, _ := NewQuota(QuotaConfig{MaxAlertsPerWindow: 3, Window: time.Minute})
	if !q.Allow("tok-1") {
		t.Error("expected first Allow to return true")
	}
}

func TestQuota_Allow_ExhaustsLimit(t *testing.T) {
	q, _ := NewQuota(QuotaConfig{MaxAlertsPerWindow: 3, Window: time.Minute})
	for i := 0; i < 3; i++ {
		if !q.Allow("tok-x") {
			t.Fatalf("expected Allow to be true on call %d", i+1)
		}
	}
	if q.Allow("tok-x") {
		t.Error("expected Allow to be false after quota exhausted")
	}
}

func TestQuota_Allow_DifferentTokensIndependent(t *testing.T) {
	q, _ := NewQuota(QuotaConfig{MaxAlertsPerWindow: 1, Window: time.Minute})
	q.Allow("tok-a")
	if !q.Allow("tok-b") {
		t.Error("expected tok-b to be allowed independently of tok-a")
	}
}

func TestQuota_Remaining_DecreasesWithAllows(t *testing.T) {
	q, _ := NewQuota(QuotaConfig{MaxAlertsPerWindow: 5, Window: time.Minute})
	if q.Remaining("tok-1") != 5 {
		t.Errorf("expected 5 remaining before any allows")
	}
	q.Allow("tok-1")
	q.Allow("tok-1")
	if q.Remaining("tok-1") != 3 {
		t.Errorf("expected 3 remaining after 2 allows, got %d", q.Remaining("tok-1"))
	}
}

func TestQuota_Allow_ResetsAfterWindowExpires(t *testing.T) {
	q, _ := NewQuota(QuotaConfig{MaxAlertsPerWindow: 2, Window: 50 * time.Millisecond})
	q.Allow("tok-r")
	q.Allow("tok-r")
	if q.Allow("tok-r") {
		t.Error("expected quota to be exhausted")
	}
	time.Sleep(60 * time.Millisecond)
	if !q.Allow("tok-r") {
		t.Error("expected quota to reset after window expiry")
	}
}
