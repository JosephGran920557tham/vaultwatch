// Package tokenwatch — entropy sub-module.
//
// # Entropy Detection
//
// The EntropyDetector measures the diversity of TTL values observed for a
// single token over time.  A token whose TTL samples cluster tightly in the
// same narrow range exhibits low entropy, which can indicate:
//
//   - An external process renewing the token at a fixed interval without
//     Vault's built-in lease renewal jitter.
//   - A misconfigured client that always requests the same TTL, bypassing
//     server-side TTL policies.
//   - A replay or replay-adjacent attack where a stale token is being
//     re-presented.
//
// # Thresholds
//
// Entropy is normalised to the range [0, 1] using Shannon entropy divided by
// the maximum possible entropy for the observed number of distinct TTL buckets.
// Two thresholds are configurable:
//
//   - WarningThreshold  – entropy below this value emits a Warning alert.
//   - CriticalThreshold – entropy below this value emits a Critical alert
//     (takes precedence over Warning).
//
// # Usage
//
//	detector := tokenwatch.NewEntropyDetector(tokenwatch.DefaultEntropyConfig())
//	detector.Record(tokenID, ttl)
//	if a := detector.Check(tokenID); a != nil {
//		// handle alert
//	}
package tokenwatch
