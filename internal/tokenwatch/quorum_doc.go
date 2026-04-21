// Package tokenwatch provides token lifecycle monitoring for HashiCorp Vault.
//
// # Quorum Detector
//
// The QuorumDetector tracks rolling health votes for each token over a
// configurable time window. When the fraction of healthy votes falls below the
// configured threshold, an alert is emitted.
//
// Votes are cast by calling Vote(tokenID, healthy) and evaluated by Check(tokenID).
// The detector prunes votes older than WindowSize on every Vote call, keeping
// memory usage bounded.
//
// Alert levels:
//   - Warning  – healthy ratio < Threshold
//   - Critical – healthy ratio < (1 - CriticalMin)
//
// QuorumScanner integrates the detector with the token Registry, iterating all
// registered tokens, casting a vote derived from the token's current TTL, and
// returning any resulting alerts in a single Scan call.
package tokenwatch
