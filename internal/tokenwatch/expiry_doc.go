// Package tokenwatch provides facilities for monitoring HashiCorp Vault
// token lifecycles, including expiry classification, alert generation,
// deduplication, and throttling of repeated notifications.
//
// The ExpiryClassifier type maps a token's remaining TTL to an alert level
// (Info, Warning, or Critical) based on configurable thresholds. It is
// intended to be used alongside the Watcher and Alerter types to form a
// complete token monitoring pipeline.
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
