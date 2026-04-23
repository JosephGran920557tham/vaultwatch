// Package tokenwatch provides token lifecycle monitoring for HashiCorp Vault.
//
// # Latency Detection
//
// The LatencyDetector tracks per-token lookup latency samples and emits
// alerts when observed latency exceeds configurable warning or critical
// thresholds.
//
// Configuration:
//
//	type LatencyConfig struct {
//	    MinSamples    int           // minimum samples before alerting
//	    WarnThreshold time.Duration // latency above this emits a Warning
//	    CritThreshold time.Duration // latency above this emits a Critical
//	}
//
// Usage:
//
//	det := NewLatencyDetector(DefaultLatencyConfig())
//	if alert := det.Check(tokenID, measuredLatency); alert != nil {
//	    // handle alert
//	}
//
// The LatencyScanner wraps a Registry and a LatencyDetector, measuring
// round-trip lookup time for each registered token on every Scan call.
package tokenwatch
