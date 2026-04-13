// Package schedule provides a simple interval-based scheduler for
// vaultwatch monitoring passes.
//
// Usage:
//
//	runner := monitor.New(cfg, vaultClient, dispatcher)
//	sched := schedule.New(runner, 5*time.Minute, nil)
//	if err := sched.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
//		log.Fatal(err)
//	}
//
// The scheduler fires RunOnce immediately on Start, then repeats at the
// configured interval. Transient errors from RunOnce are logged but do not
// halt the scheduler; only context cancellation stops it.
package schedule
