// Package tokenwatch provides facilities for monitoring HashiCorp Vault
// token leases, classifying their expiry status, and dispatching alerts.
//
// # Backoff
//
// The Backoff type implements exponential back-off for retry loops that
// interact with Vault (e.g. token renewal, health checks). It is
// intentionally stateful so callers can drive the retry loop themselves
// without spawning goroutines:
//
//	b, err := tokenwatch.NewBackoff(tokenwatch.DefaultBackoffConfig())
//	if err != nil {
//		log.Fatal(err)
//	}
//	for {
//		err := doWork()
//		if err == nil {
//			b.Reset()
//			break
//		}
//		delay, ok := b.Next()
//		if !ok {
//			log.Fatal("max retries exceeded")
//		}
//		time.Sleep(delay)
//	}
package tokenwatch
