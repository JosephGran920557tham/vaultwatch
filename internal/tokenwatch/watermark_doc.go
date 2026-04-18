// Package tokenwatch — watermark detection.
//
// WatermarkDetector tracks the historical peak TTL observed for each token and
// raises a warning alert whenever the current TTL falls below the configured
// low-watermark threshold, provided a sufficiently high peak has previously
// been recorded (above the high-watermark).
//
// This allows vaultwatch to distinguish between tokens that have always had a
// short TTL (no alert) and tokens whose TTL has unexpectedly shrunk (alert).
//
// Usage:
//
//	det := tokenwatch.NewWatermarkDetector(tokenwatch.WatermarkConfig{
//		LowWatermark:  5 * time.Minute,
//		HighWatermark: 1 * time.Hour,
//	})
//	det.Record(tokenID, observedTTL)  // call each poll cycle
//	if alert := det.Check(tokenID, observedTTL); alert != nil {
//		// handle alert
//	}
package tokenwatch
