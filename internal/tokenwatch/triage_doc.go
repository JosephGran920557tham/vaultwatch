// Package tokenwatch provides token lifecycle monitoring for HashiCorp Vault.
//
// # Triage
//
// The triage subsystem scores and ranks alerts by urgency so that operators
// can focus on the most critical issues first.
//
// Scoring is based on two factors:
//
//  1. Severity weight — Critical alerts receive the highest base score,
//     followed by Warning, then Info.
//
//  2. Recency decay — Older alerts are down-weighted using an exponential
//     half-life, so a fresh warning may outrank a stale critical.
//
// Usage:
//
//	tr := tokenwatch.NewTriage(tokenwatch.DefaultTriageConfig())
//	entries := tr.Rank(alerts, time.Now())
//	for _, e := range entries {
//		fmt.Printf("[%.2f] %s\n", e.Score, e.Alert.Message)
//	}
//
// TriageScanner wraps a Triage together with an alert source and a dispatch
// function, providing a single Scan method suitable for use inside a
// schedule.Scheduler loop.
package tokenwatch
