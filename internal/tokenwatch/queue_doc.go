// Package tokenwatch provides token lifecycle monitoring for HashiCorp Vault.
//
// # Envelope
//
// An Envelope wraps an alert.Alert with delivery metadata including the
// originating token ID, creation timestamp, attempt counter, and the last
// error encountered during delivery. Envelopes are created via NewEnvelope
// and progress through the delivery pipeline until either successfully sent
// or exhausted (Attempt >= MaxAttempts).
//
// # Queue
//
// Queue is a bounded, thread-safe FIFO queue of Envelopes. It is the
// primary handoff point between alert producers (scanners, detectors) and
// alert consumers (dispatchers, notifiers).
//
// Typical usage:
//
//	q := tokenwatch.NewQueue(128)
//	q.Push(tokenwatch.NewEnvelope(tokenID, alert, 3))
//
//	if err := q.Wait(ctx); err == nil {
//		env, _ := q.Pop()
//		// deliver env.Alert ...
//	}
//
// Push returns false when the queue is at capacity; callers should handle
// back-pressure (e.g. drop, log, or apply rate limiting) accordingly.
package tokenwatch
