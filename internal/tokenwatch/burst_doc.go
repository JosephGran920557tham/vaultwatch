// Package tokenwatch provides burst detection for Vault token activity.
//
// # Burst Detection
//
// BurstDetector tracks the number of events (e.g. renewals, checks) recorded
// for a given token key within a configurable sliding time window. When the
// event count exceeds the configured MaxEvents threshold the token is
// considered to be "bursting".
//
// BurstScanner integrates BurstDetector with the token Registry so that
// callers can call Scan() to obtain a slice of alert.Alert values for every
// currently-bursting token.
//
// # Usage
//
//	det, _ := tokenwatch.NewBurstDetector(tokenwatch.BurstConfig{
//		Window:    30 * time.Second,
//		MaxEvents: 10,
//	})
//	scanner := tokenwatch.NewBurstScanner(registry, det)
//
//	// Record events as they occur:
//	scanner.Record(tokenID)
//
//	// Periodically scan for bursting tokens:
//	alerts := scanner.Scan()
package tokenwatch
