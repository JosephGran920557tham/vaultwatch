// Package tokenwatch provides facilities for monitoring HashiCorp Vault
// token leases and detecting anomalous conditions.
//
// # Fingerprint
//
// The Fingerprint subsystem tracks a stable SHA-256 hash of a token's
// label metadata across polling cycles. When the hash changes between
// scans, it indicates that the token's identity or associated metadata
// has been mutated — which may signal an unexpected rotation, privilege
// escalation, or misconfiguration.
//
// Usage:
//
//	fp := tokenwatch.NewFingerprint(tokenwatch.DefaultFingerprintConfig())
//	hash := tokenwatch.Compute(labels)
//	if fp.Track(tokenID, hash) {
//		// fingerprint changed — emit alert
//	}
//
// FingerprintScanner integrates with the Registry to perform a full sweep
// across all registered tokens on each call to Scan, returning a slice of
// alert.Alert values for any token whose fingerprint has changed.
//
// Expired entries are removed via Evict, which should be called
// periodically (e.g. from an EvictionRunner) to bound memory usage.
package tokenwatch
