// Package tokenwatch provides facilities for monitoring HashiCorp Vault
// token lifecycles, including expiry classification, alert generation,
// deduplication, and throttling of repeated notifications.
//
// The ExpiryClassifier type maps a token's remaining TTL to an alert level
// (Info, Warning, or Critical) based on configurable thresholds. It is
// intended to be used alongside the Watcher and Alerter types to form a
// complete token monitoring pipeline.
//
// Alert Levels:
//
//   - Info: The token is healthy; its TTL exceeds the warning threshold.
//   - Warning: The token's TTL has fallen below the warning threshold but
//     remains above the critical threshold.
//   - Critical: The token's TTL has fallen below the critical threshold
//     and requires immediate attention.
//
// Typical usage:
//
//	ec := tokenwatch.DefaultExpiryClassifier()
//	if err := ec.Validate(); err != nil {
//	    log.Fatal(err)
//	}
//	level := ec.Classify(remainingTTL)
//	fmt.Println(ec.Summary(tokenID, remainingTTL))
package tokenwatch
