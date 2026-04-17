// Package tokenwatch — capacity detection
//
// The CapacityDetector measures what fraction of the configured maximum token
// slots are currently occupied and classifies the result as ok, warning, or
// critical.
//
// Thresholds are expressed as fractions in [0, 1]:
//
//	WarnThreshold  – default 0.75  (75 % full → warning)
//	CritThreshold  – default 0.90  (90 % full → critical)
//
// CapacityScanner wraps a Registry and a CapacityDetector, providing a
// single Scan() call that returns an *alert.Alert (or nil when healthy).
// It is designed to be plugged into the existing pipeline alongside other
// scanners such as AnomalyScanner or BurstScanner.
//
// Example:
//
//	det := tokenwatch.NewCapacityDetector(tokenwatch.DefaultCapacityConfig())
//	scanner := tokenwatch.NewCapacityScanner(registry, det)
//	if a := scanner.Scan(); a != nil {
//		dispatcher.Dispatch(ctx, *a)
//	}
package tokenwatch
