// Package snapshot provides point-in-time capture, storage, and diffing of
// Vault lease alert states.
//
// # Overview
//
// A Snapshot is an immutable collection of lease alerts captured at a single
// moment. The Store holds the most recent snapshot and supports lock-safe
// atomic replacement, making it safe to use across goroutines.
//
// # Diff
//
// Diff compares two snapshots and returns only the alerts whose LeaseID did
// not appear in the previous snapshot. This allows callers to react only to
// newly detected expirations rather than re-alerting on every poll cycle.
//
// # Watcher
//
// Watcher combines a Source (anything that yields current alerts), a Store,
// and a ChangeHandler. Each call to Poll fetches fresh alerts, saves a new
// snapshot, and invokes the handler with any net-new lease IDs.
package snapshot
