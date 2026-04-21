// Package tokenwatch provides monitoring primitives for HashiCorp Vault token
// leases.
//
// # Roster
//
// Roster is a bounded, time-limited registry of active token IDs. It is
// designed to complement the core Registry by providing a lightweight
// "who's alive" view that does not require a full token lookup.
//
// Typical usage:
//
//	roster := tokenwatch.NewRoster(tokenwatch.DefaultRosterConfig())
//
//	// Record activity whenever a token is observed.
//	if err := roster.Touch(tokenID, labels); err != nil {
//		log.Printf("roster full: %v", err)
//	}
//
//	// Enumerate tokens seen within the TTL window.
//	for _, id := range roster.Active() {
//		// process id …
//	}
//
//	// Periodically remove expired entries.
//	n := roster.Prune()
//	log.Printf("pruned %d stale roster entries", n)
//
// Roster is safe for concurrent use by multiple goroutines.
package tokenwatch
