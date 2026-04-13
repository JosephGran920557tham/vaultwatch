package alert

import (
	"fmt"
	"log"
	"time"
)

// LeaseInfo represents a single lease returned from the checker.
type LeaseInfo struct {
	LeaseID   string
	ExpiresIn time.Duration
}

// Dispatcher evaluates leases and dispatches alerts via registered notifiers.
type Dispatcher struct {
	Notifiers []Notifier
	WarnMins  int
	CritMins  int
}

// NewDispatcher creates a Dispatcher with the given thresholds and notifiers.
func NewDispatcher(warnMins, critMins int, notifiers ...Notifier) *Dispatcher {
	return &Dispatcher{
		Notifiers: notifiers,
		WarnMins:  warnMins,
		CritMins:  critMins,
	}
}

// Dispatch evaluates a slice of LeaseInfo and sends alerts for non-INFO leases.
// All notifier errors are logged but do not stop processing.
func (d *Dispatcher) Dispatch(leases []LeaseInfo) error {
	var lastErr error
	for _, l := range leases {
		a := Build(l.LeaseID, l.ExpiresIn, d.WarnMins, d.CritMins)
		if a.Level == LevelInfo {
			continue
		}
		if err := d.sendToAll(l.LeaseID, a); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// DispatchAll sends alerts for every lease regardless of level.
func (d *Dispatcher) DispatchAll(leases []LeaseInfo) error {
	var lastErr error
	for _, l := range leases {
		a := Build(l.LeaseID, l.ExpiresIn, d.WarnMins, d.CritMins)
		if err := d.sendToAll(l.LeaseID, a); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// DispatchOne evaluates a single LeaseInfo and sends an alert if its level is
// not INFO. It returns an error if any notifier fails.
func (d *Dispatcher) DispatchOne(l LeaseInfo) error {
	a := Build(l.LeaseID, l.ExpiresIn, d.WarnMins, d.CritMins)
	if a.Level == LevelInfo {
		return nil
	}
	return d.sendToAll(l.LeaseID, a)
}

// sendToAll delivers an alert to every registered notifier, logging errors as
// they occur. It returns the last error encountered, or nil if all succeeded.
func (d *Dispatcher) sendToAll(leaseID string, a Alert) error {
	var lastErr error
	for _, n := range d.Notifiers {
		if err := n.Send(a); err != nil {
			log.Printf("alert notifier error for lease %s: %v", leaseID, err)
			lastErr = fmt.Errorf("notifier send failed: %w", err)
		}
	}
	return lastErr
}
