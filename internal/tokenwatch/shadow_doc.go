// Package tokenwatch provides token lifecycle monitoring for HashiCorp Vault.
//
// # Shadow Registry
//
// The ShadowRegistry stores a lightweight copy of previously observed token
// TTL values. Entries are timestamped and expire after a configurable TTL,
// ensuring stale comparisons are not used.
//
// # Shadow Scanner
//
// ShadowScanner compares each registered token's current TTL against its
// shadow entry. When the TTL has dropped by more than the configured ratio
// (default 50 %), a Warning alert is emitted. After each scan the shadow
// entry is refreshed with the latest observed value.
//
// This is useful for detecting unexpected TTL reductions that may indicate
// external revocation or lease misconfiguration, complementing the existing
// expiry and trend detectors.
package tokenwatch
