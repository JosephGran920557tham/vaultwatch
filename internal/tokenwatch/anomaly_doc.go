// Package tokenwatch provides TTL anomaly detection for Vault tokens.
//
// # Anomaly Detection
//
// The AnomalyDetector checks whether a token's remaining TTL falls outside
// a configured range [MinTTL, MaxTTL]:
//
//   - A TTL below MinTTL produces a Critical alert, indicating the token is
//     dangerously close to expiry or was issued with an unexpectedly short
//     lifetime.
//
//   - A TTL above MaxTTL produces a Warning alert, indicating the token was
//     issued with an unusually long lifetime which may violate policy.
//
//   - A TTL within [MinTTL, MaxTTL] produces no alert.
//
// # Scanning
//
// AnomalyScanner integrates with the Registry to iterate over all tracked
// tokens, fetch their current TTLs via a pluggable TTLSource, and collect
// any anomaly alerts. Lookup errors for individual tokens are silently skipped
// so that a single unreachable token does not prevent the rest from being
// checked.
//
// # Usage
//
//	det := tokenwatch.NewAnomalyDetector(tokenwatch.DefaultAnomalyConfig())
//	scanner := tokenwatch.NewAnomalyScanner(registry, det, myTTLSource)
//	alerts, err := scanner.Scan(ctx)
package tokenwatch
