// Package tokenwatch provides monitoring of HashiCorp Vault token TTLs.
//
// It tracks a configurable set of token accessors and emits alert.Alert
// values when a token's remaining TTL falls below warning or critical
// thresholds. Accessors are managed through a thread-safe Registry that
// supports dynamic add/remove at runtime.
//
// Typical usage:
//
//	reg := tokenwatch.NewRegistry()
//	_ = reg.Add("s.abc123")
//
//	w, err := tokenwatch.New(vaultClient.LookupToken, reg.List(), time.Hour, 10*time.Minute)
//	if err != nil { /* handle */ }
//
//	alerts, err := w.Check(ctx)
//	if err != nil { /* handle */ }
//	for _, a := range alerts {
//		fmt.Println(a)
//	}
package tokenwatch
