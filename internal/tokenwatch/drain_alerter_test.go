package tokenwatch

import (
	"context"
	"testing"
	"time"
)

func newFastDrainDetector(t *testing.T, tokenID string) *DrainDetector {
	t.Helper()
	cfg := DrainConfig{SampleWindow: time.Minute, DrainThreshold: 5.0, MinSamples: 2}
	d, err := NewDrainDetector(cfg)
	if err != nil {
		t.Fatalf("NewDrainDetector: %v", err)
	}
	now := time.Now()
	d.mu.Lock()
	d.samples[tokenID] = []sample{
		{at: now.Add(-10 * time.Second), ttl: 400 * time.Second},
		{at: now, ttl: 300 * time.Second},
	}
	d.mu.Unlock()
	return d
}

func TestNewDrainAlerter_NilDetector(t *testing.T) {
	_, err := NewDrainAlerter(nil)
	if err == nil {
		t.Fatal("expected error for nil detector")
	}
}

func TestNewDrainAlerter_ValidDetector(t *testing.T) {
	d, _ := NewDrainDetector(DefaultDrainConfig())
	_, err := NewDrainAlerter(d)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDrainAlerter_Check_NoDrain_ReturnsNil(t *testing.T) {
	d, _ := NewDrainDetector(DefaultDrainConfig())
	a, _ := NewDrainAlerter(d)
	alrt, err := a.Check(context.Background(), "tok-quiet")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if alrt != nil {
		t.Fatalf("expected nil alert, got %+v", alrt)
	}
}

func TestDrainAlerter_Check_Draining_ReturnsAlert(t *testing.T) {
	const id = "tok-fast"
	det := newFastDrainDetector(t, id)
	a, err := NewDrainAlerter(det)
	if err != nil {
		t.Fatalf("NewDrainAlerter: %v", err)
	}
	alrt, err := a.Check(context.Background(), id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if alrt == nil {
		t.Fatal("expected alert for draining token")
	}
	if alrt.LeaseID != id {
		t.Errorf("want LeaseID %q got %q", id, alrt.LeaseID)
	}
}

func TestDrainAlerter_Check_CooldownSuppressesRepeat(t *testing.T) {
	const id = "tok-repeat"
	det := newFastDrainDetector(t, id)
	a, _ := NewDrainAlerter(det)
	ctx := context.Background()
	first, err := a.Check(ctx, id)
	if err != nil {
		t.Fatalf("first check error: %v", err)
	}
	if first == nil {
		t.Fatal("expected first alert")
	}
	second, err := a.Check(ctx, id)
	if err != nil {
		t.Fatalf("second check error: %v", err)
	}
	if second != nil {
		t.Fatal("expected second alert suppressed by cooldown")
	}
}
