// Package tokenwatch provides token lifecycle monitoring for HashiCorp Vault.
//
// # Circuit Breaker
//
// The Circuit type implements a simple three-state circuit breaker pattern
// (Closed → Open → Half-Open) to protect token check operations from
// cascading failures when Vault is unreachable or returning errors.
//
// Usage:
//
//	cfg := tokenwatch.DefaultCircuitConfig()
//	cb := tokenwatch.NewCircuit(cfg)
//
//	if err := cb.Allow(); err != nil {
//	    // circuit is open — skip the operation
//	    return err
//	}
//	if err := doVaultCall(); err != nil {
//	    cb.RecordFailure()
//	    return err
//	}
//	cb.RecordSuccess()
package tokenwatch
