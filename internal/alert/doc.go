// Package alert provides alert construction, classification, and dispatching
// for VaultWatch lease expiration events.
//
// # Overview
//
// Leases retrieved from HashiCorp Vault are evaluated against configurable
// warning and critical thresholds (in minutes). Based on the remaining TTL,
// each lease is classified as INFO, WARNING, or CRITICAL.
//
// # Components
//
//   - [Alert]       — value type carrying lease ID, expiry, level, and message.
//   - [Notifier]    — interface for sending alerts (console, webhook, etc.).
//   - [ConsoleNotifier] — built-in notifier that writes to stdout.
//   - [Dispatcher]  — evaluates a batch of leases and fans out to notifiers.
//
// # Usage
//
//	console := alert.NewConsoleNotifier()
//	d := alert.NewDispatcher(warnMins, critMins, console)
//	if err := d.Dispatch(leases); err != nil {
//		log.Printf("alert dispatch error: %v", err)
//	}
package alert
