// Package tokenwatch — relay subsystem
//
// A Relay buffers alert.Alert values produced by multiple concurrent
// scanner goroutines and forwards them in batches to a downstream
// dispatch function. This decouples high-frequency producers from
// potentially slow consumers (e.g. HTTP notifiers or audit loggers).
//
// # Components
//
//   - Relay — thread-safe buffer with configurable capacity and a
//     Flush method that drains the buffer and calls the dispatch func.
//
//   - RelayRunner — runs a periodic flush loop driven by
//     Relay.cfg.FlushInterval and performs a final flush when the
//     supplied context is cancelled.
//
// # Usage
//
//	relay := tokenwatch.NewRelay(
//		tokenwatch.DefaultRelayConfig(),
//		func(batch []alert.Alert) error {
//			return dispatcher.DispatchAll(batch)
//		},
//	)
//	runner := tokenwatch.NewRelayRunner(relay, logger)
//	go runner.Run(ctx)
//
// Enqueue is safe to call from multiple goroutines simultaneously.
package tokenwatch
