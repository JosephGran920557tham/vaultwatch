package alert

import (
	"errors"
	"testing"
	"time"
)

// mockNotifier records sent alerts and optionally returns an error.
type mockNotifier struct {
	Sent  []Alert
	Err   error
}

func (m *mockNotifier) Send(a Alert) error {
	m.Sent = append(m.Sent, a)
	return m.Err
}

func TestDispatcher_Dispatch_FiltersInfo(t *testing.T) {
	n := &mockNotifier{}
	d := NewDispatcher(30, 10, n)

	leases := []LeaseInfo{
		{LeaseID: "secret/ok", ExpiresIn: 60 * time.Minute},   // INFO — skipped
		{LeaseID: "secret/warn", ExpiresIn: 20 * time.Minute}, // WARNING
		{LeaseID: "secret/crit", ExpiresIn: 5 * time.Minute},  // CRITICAL
	}

	if err := d.Dispatch(leases); err != nil {
		t.Fatalf("Dispatch() unexpected error: %v", err)
	}
	if len(n.Sent) != 2 {
		t.Errorf("expected 2 alerts sent, got %d", len(n.Sent))
	}
}

func TestDispatcher_DispatchAll_SendsAll(t *testing.T) {
	n := &mockNotifier{}
	d := NewDispatcher(30, 10, n)

	leases := []LeaseInfo{
		{LeaseID: "a", ExpiresIn: 60 * time.Minute},
		{LeaseID: "b", ExpiresIn: 20 * time.Minute},
	}

	if err := d.DispatchAll(leases); err != nil {
		t.Fatalf("DispatchAll() unexpected error: %v", err)
	}
	if len(n.Sent) != 2 {
		t.Errorf("expected 2 alerts, got %d", len(n.Sent))
	}
}

func TestDispatcher_Dispatch_NotifierError(t *testing.T) {
	n := &mockNotifier{Err: errors.New("connection refused")}
	d := NewDispatcher(30, 10, n)

	leases := []LeaseInfo{
		{LeaseID: "secret/crit", ExpiresIn: 5 * time.Minute},
	}

	err := d.Dispatch(leases)
	if err == nil {
		t.Error("expected error from failing notifier, got nil")
	}
}

func TestDispatcher_MultipleNotifiers(t *testing.T) {
	n1 := &mockNotifier{}
	n2 := &mockNotifier{}
	d := NewDispatcher(30, 10, n1, n2)

	leases := []LeaseInfo{
		{LeaseID: "secret/warn", ExpiresIn: 20 * time.Minute},
	}

	if err := d.Dispatch(leases); err != nil {
		t.Fatalf("Dispatch() unexpected error: %v", err)
	}
	if len(n1.Sent) != 1 || len(n2.Sent) != 1 {
		t.Errorf("expected both notifiers to receive 1 alert each")
	}
}
