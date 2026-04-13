// Package monitor provides the core polling loop for vaultwatch.
//
// A Monitor ties together a vault.LeaseChecker and an alert.Dispatcher,
// running on a configurable interval. On each tick it queries Vault for
// leases that will expire within the warn threshold, classifies each one,
// and routes the resulting alerts to every registered Notifier.
//
// Typical usage:
//
//	cfg, _ := config.Load("vaultwatch.yaml")
//	client, _ := vault.NewClient(cfg)
//	checker := vault.NewLeaseChecker(client)
//	notifier := alert.NewConsoleNotifier(os.Stdout)
//	dispatcher := alert.NewDispatcher([]alert.Notifier{notifier})
//	m := monitor.New(checker, dispatcher, cfg)
//	if err := m.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
//		log.Fatal(err)
//	}
package monitor
