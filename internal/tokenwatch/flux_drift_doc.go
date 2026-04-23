// Package tokenwatch provides flux and drift detection for Vault token TTLs.
//
// # Flux Detection
//
// FluxDetector tracks the absolute delta between consecutive TTL samples for
// a given token. High flux indicates that a token's TTL is oscillating
// unpredictably — for example due to repeated renewals with inconsistent
// policies or competing renewal agents.
//
// FluxScanner wires a FluxDetector to the token Registry, running a full
// scan on demand and returning structured Alert values.
//
// # Drift Detection
//
// DriftDetector compares the current TTL of a token against a baseline
// captured on the first observation. Sustained deviation beyond configurable
// warning and critical thresholds surfaces as an alert.
//
// DriftScanner wires a DriftDetector to the token Registry, running a full
// scan on demand and returning structured Alert values.
//
// Both scanners skip tokens whose lookup fails, ensuring a single
// unavailable token does not block the rest of the scan.
package tokenwatch
