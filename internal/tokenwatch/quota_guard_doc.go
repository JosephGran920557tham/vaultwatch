// Package tokenwatch provides token lifecycle monitoring for HashiCorp Vault.
//
// # QuotaGuard
//
// QuotaGuard limits the number of alerts emitted for a given token within
// a configurable sliding time window. This prevents alert storms when a
// token is repeatedly checked and consistently found in a degraded state.
//
// Usage:
//
//	guard := tokenwatch.NewQuotaGuard(tokenwatch.QuotaGuardConfig{
//	    MaxAlerts: 5,
//	    Window:    time.Minute,
//	})
//
//	if guard.Allow(tokenID) {
//	    dispatcher.Dispatch(ctx, alert)
//	}
//
// Zero values in QuotaGuardConfig are replaced with sensible defaults
// via DefaultQuotaGuardConfig.
package tokenwatch
