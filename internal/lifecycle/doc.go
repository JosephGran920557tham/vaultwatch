// Package lifecycle provides ordered startup and graceful shutdown
// coordination for vaultwatch background services.
//
// Usage:
//
//	mgr := lifecycle.New(10 * time.Second)
//
//	mgr.OnStart("vault-client", func(ctx context.Context) error {
//		return client.Connect(ctx)
//	})
//	mgr.OnStop("vault-client", func(ctx context.Context) error {
//		return client.Close()
//	})
//
//	runner := lifecycle.NewRunner(mgr)
//	if err := runner.Run(context.Background()); err != nil {
//		log.Fatal(err)
//	}
//
// Start hooks run in registration order; stop hooks run in reverse order
// so that dependencies are torn down cleanly.
package lifecycle
