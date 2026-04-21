// Package tokenwatch provides token lifecycle monitoring for HashiCorp Vault.
//
// # Consensus
//
// The Consensus type implements quorum-based alert suppression. Rather than
// forwarding every alert from every scanner, Consensus requires that a
// configurable number of independent sources (the quorum) agree that an
// alert condition exists before it is propagated downstream.
//
// This reduces noise caused by transient failures or single-source anomalies.
//
// Usage:
//
//	cfg := tokenwatch.ConsensusConfig{
//		Quorum: 2,
//		Window: 5 * time.Minute,
//		MaxKeys: 1000,
//	}
//	c := tokenwatch.NewConsensus(cfg)
//
//	// Wrap multiple scanners so only agreed-upon alerts pass through.
//	scanner := tokenwatch.NewConsensusScanner(c, srcA, srcB, srcC)
//	alerts, err := scanner.Scan(tokenID)
//
// Votes expire after Window, preventing stale agreement from persisting
// indefinitely. MaxKeys bounds memory usage when many distinct alerts are
// tracked concurrently.
package tokenwatch
