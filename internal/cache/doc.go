// Package cache implements a thread-safe, TTL-based in-memory store for
// Vault lease metadata used by vaultwatch.
//
// # Overview
//
// The Store type holds [Entry] values keyed by lease ID. Each entry is
// considered fresh for the duration of the configured TTL; reads after that
// window return a cache miss, prompting callers to re-fetch from Vault.
//
// # Warming
//
// The Warmer type pre-populates a Store at startup by listing all active
// lease IDs from a [LeaseSource] and resolving each one via a [LeaseLookup].
// Individual lookup failures are silently skipped so that a single bad lease
// does not prevent the rest of the cache from being populated.
//
// # Maintenance
//
// Call [Store.Purge] periodically (e.g., from the scheduler) to reclaim
// memory occupied by stale entries.
package cache
