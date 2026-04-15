// Package tokenwatch provides monitoring of HashiCorp Vault token TTLs.
//
// It tracks a configurable set of token accessors and emits alert.Alert
// values when a token's remaining TTL falls below warning or critical
// thresholds. Accessors are managed through a thread-safe Registry that
// supports dynamic add/remove at runtime.
//
// # Architecture
//
// The Watcher is stateless between calls: each call to Check performs a
// fresh lookup of all registered accessors and compares the returned TTL
// against the configured warn and critical thresholds. The Registry holds
// the canonical list of accessors to monitor and may be updated concurrently
// while the Watcher is running.
//
// # Thresholds
//
// Two thresholds are supported:
//
//   - warn: alert is emitted when TTL is at or below this duration.
//   - critical: alert is emitted (at higher severity) when TTL is at or
//     below this duration. The critical threshold must be less than warn.
//
// Typical usage:
//
//	reg := tokenwatch.NewRegistry()
//	_ = reg.Add("s.abc123")
//
//	w, err := tokenwatch.New(vaultClient.LookupToken, reg.List(), time.Hour, 10*time.Minute)
//	if err != nil { /* handle */ }
//
//	alerts, err := w.Check(ctx)
//	if err != nil { /* handle */ }
//	for _, a := range alerts {
//		fmt.Println(a)
//	}
package tokenwatch
