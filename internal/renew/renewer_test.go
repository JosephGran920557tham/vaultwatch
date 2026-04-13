package renew

import (
	"context"
	"errors"
	"testing"
	"time"
)

// mockRenewer implements LeaseRenewer for testing.
type mockRenewer struct {
	err     error
	called  []string
}

func (m *mockRenewer) RenewLease(_ context.Context, leaseID string, _ time.Duration) error {
	m.called = append(m.called, leaseID)
	return m.err
}

func TestNewManager_NilRenewer(t *testing.T) {
	_, err := NewManager(nil)
	if err == nil {
		t.Fatal("expected error for nil renewer")
	}
}

func TestRenew_Success(t *testing.T) {
	mock := &mockRenewer{}
	mgr, _ := NewManager(mock)

	req := RenewRequest{LeaseID: "lease/abc", Increment: 30 * time.Minute}
	res := mgr.Renew(context.Background(), req)

	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if res.LeaseID != "lease/abc" {
		t.Errorf("expected lease/abc, got %s", res.LeaseID)
	}
	if res.RenewedAt.IsZero() {
		t.Error("RenewedAt should not be zero")
	}
}

func TestRenew_Error(t *testing.T) {
	mock := &mockRenewer{err: errors.New("vault error")}
	mgr, _ := NewManager(mock)

	res := mgr.Renew(context.Background(), RenewRequest{LeaseID: "lease/xyz", Increment: time.Hour})
	if res.Err == nil {
		t.Fatal("expected error but got nil")
	}
}

func TestRenewAll_SendsAll(t *testing.T) {
	mock := &mockRenewer{}
	mgr, _ := NewManager(mock)

	reqs := []RenewRequest{
		{LeaseID: "lease/1", Increment: time.Hour},
		{LeaseID: "lease/2", Increment: time.Hour},
		{LeaseID: "lease/3", Increment: time.Hour},
	}
	results := mgr.RenewAll(context.Background(), reqs)

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	if len(mock.called) != 3 {
		t.Errorf("expected 3 calls, got %d", len(mock.called))
	}
	for _, r := range results {
		if r.Err != nil {
			t.Errorf("unexpected error for %s: %v", r.LeaseID, r.Err)
		}
	}
}

func TestRenewAll_PartialErrors(t *testing.T) {
	callCount := 0
	mock := &mockRenewer{}
	_ = mock

	failRenewer := &conditionalRenewer{failOn: "lease/2"}
	mgr, _ := NewManager(failRenewer)

	reqs := []RenewRequest{
		{LeaseID: "lease/1", Increment: time.Hour},
		{LeaseID: "lease/2", Increment: time.Hour},
	}
	results := mgr.RenewAll(context.Background(), reqs)
	_ = callCount

	if results[0].Err != nil {
		t.Errorf("lease/1 should succeed")
	}
	if results[1].Err == nil {
		t.Errorf("lease/2 should fail")
	}
}

type conditionalRenewer struct{ failOn string }

func (c *conditionalRenewer) RenewLease(_ context.Context, leaseID string, _ time.Duration) error {
	if leaseID == c.failOn {
		return errors.New("forced failure")
	}
	return nil
}
