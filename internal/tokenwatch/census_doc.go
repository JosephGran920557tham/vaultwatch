// Package tokenwatch provides components for monitoring HashiCorp Vault token
// lifecycles, detecting anomalies, and dispatching alerts.
//
// # Census
//
// Census tracks the live population of tokens observed during a scan cycle.
// It is designed to give downstream detectors a consistent view of which
// tokens are currently active, without requiring each detector to query the
// registry independently.
//
// Basic usage:
//
//	census := tokenwatch.NewCensus(tokenwatch.DefaultCensusConfig())
//	census.Observe("s.abc123", map[string]string{"env": "prod"})
//	activeIDs := census.Active()
//
// CensusScanner automates population by iterating the Registry and calling
// the provided lookup function for each token:
//
//	scanner := tokenwatch.NewCensusScanner(registry, census, vaultClient.LookupToken)
//	scanner.Scan(ctx)
//
// Stale entries (older than MaxAge) are excluded from Active() and can be
// explicitly removed by calling Evict().
package tokenwatch
