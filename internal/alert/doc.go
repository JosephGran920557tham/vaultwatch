// Package alert provides alert construction, classification, and dispatching
// for VaultWatch lease expiration events.
//
// # Overview
//
// Leases retrieved from HashiCorp Vault are evaluated against configurable
// warning and critical thresholds (in minutes). Based on the remaining TTL,
// each lease is classified as INFO, WARNING, or CRITICAL.
//
//   - INFO     — lease TTL is above the warning threshold.
//   - WARNING  — lease TTL is at or below the warning threshold.
//   - CRITICAL — lease TTL is at or below the critical threshold.
//
// # Components
//
//   - [Alert]             — value type carrying lease ID, expiry, level, and message.
//   - [Notifier]          — interface for sending alerts (console, webhook, etc.).
//   - [ConsoleNotifier]   — built-in notifier that writes to stdout.
//   - [Dispatcher]        — evaluates a batch of leases and fans out to notifiers.
//
// # Usage
//
//	console := alert.NewConsoleNotifier()
//	d := alert.NewDispatcher(warnMins, critMins, console)
//	if err := d.Dispatch(leases); err != nil {
//		log.Printf("alert dispatch error: %v", err)
//	}
//
// Multiple notifiers can be composed by passing them all to [NewDispatcher]:
//
//	d := alert.NewDispatcher(warnMins, critMins, console, webhook, pager)
package alert
