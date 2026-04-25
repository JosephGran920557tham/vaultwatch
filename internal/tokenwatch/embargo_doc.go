// Package tokenwatch provides token lifecycle monitoring for HashiCorp Vault.
//
// # Embargo
//
// The Embargo type provides a time-bounded suppression mechanism for token
// alerts. When a token is placed under embargo, all alert checks for that
// token are suppressed for a configurable window. This is useful for:
//
//   - Planned maintenance windows where known-bad tokens should not trigger
//     pages or notifications.
//   - Graceful rollover periods during token rotation where a brief gap in
//     validity is expected.
//   - Integration testing environments where transient expiry states are
//     intentional.
//
// Usage:
//
//	emb := tokenwatch.NewEmbargo(tokenwatch.DefaultEmbargoConfig())
//	emb.Place("s.mytoken123")          // suppress for the configured window
//	if emb.IsSuppressed("s.mytoken123") {
//	    // skip alert dispatch
//	}
//	emb.Lift("s.mytoken123")           // remove embargo early if needed
//
// Embargoes expire automatically; no background goroutine is required.
package tokenwatch
