// Package tokenwatch monitors Vault token TTLs and emits alerts
// when tokens are approaching expiration.
package tokenwatch

import (
	"context"
	"fmt"
	"time"

	"github.com/vaultwatch/internal/alert"
)

// TokenInfo holds metadata about a Vault token.
type TokenInfo struct {
	Accessor   string
	DisplayName string
	TTL        time.Duration
	Policies   []string
	Meta       map[string]string
}

// Lookup is a function that retrieves token info by accessor.
type Lookup func(ctx context.Context, accessor string) (*TokenInfo, error)

// Watcher monitors a set of token accessors and produces alerts.
type Watcher struct {
	lookup    Lookup
	accessors []string
	warn      time.Duration
	critical  time.Duration
}

// New creates a Watcher with the given lookup function and thresholds.
func New(lookup Lookup, accessors []string, warn, critical time.Duration) (*Watcher, error) {
	if lookup == nil {
		return nil, fmt.Errorf("tokenwatch: lookup function must not be nil")
	}
	if critical >= warn {
		return nil, fmt.Errorf("tokenwatch: critical threshold must be less than warn threshold")
	}
	return &Watcher{
		lookup:    lookup,
		accessors: accessors,
		warn:      warn,
		critical:  critical,
	}, nil
}

// Check evaluates all tracked token accessors and returns any alerts.
func (w *Watcher) Check(ctx context.Context) ([]alert.Alert, error) {
	var alerts []alert.Alert
	for _, acc := range w.accessors {
		info, err := w.lookup(ctx, acc)
		if err != nil {
			return nil, fmt.Errorf("tokenwatch: lookup accessor %q: %w", acc, err)
		}
		if a, ok := w.evaluate(info); ok {
			alerts = append(alerts, a)
		}
	}
	return alerts, nil
}

func (w *Watcher) evaluate(info *TokenInfo) (alert.Alert, bool) {
	switch {
	case info.TTL <= w.critical:
		return alert.Build(info.Accessor, info.TTL, alert.Critical), true
	case info.TTL <= w.warn:
		return alert.Build(info.Accessor, info.TTL, alert.Warning), true
	default:
		return alert.Alert{}, false
	}
}
