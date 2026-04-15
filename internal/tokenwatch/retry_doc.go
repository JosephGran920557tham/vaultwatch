// Package tokenwatch provides utilities for monitoring and alerting on
// HashiCorp Vault token lifecycle events.
//
// # Retry
//
// The Retry type executes a function up to a configured number of times,
// applying exponential backoff between attempts. It is intended for wrapping
// transient Vault API calls that may fail intermittently due to network
// conditions or Vault leadership changes.
//
// Example usage:
//
//	r, err := tokenwatch.NewRetry(tokenwatch.DefaultRetryConfig())
//	if err != nil {
//		log.Fatal(err)
//	}
//	err = r.Do(func() error {
//		return vaultClient.LookupToken(tokenID)
//	})
//
The delay between attempts grows by Multiplier each time, capped at MaxDelay.
package tokenwatch
