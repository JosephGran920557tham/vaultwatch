// Package tokenwatch provides token lifecycle monitoring for HashiCorp Vault.
//
// # Escalation
//
// The Escalation type tracks how often a given token fires alerts within a
// sliding time window. When the number of occurrences reaches or exceeds the
// configured Threshold, the alert level is promoted to EscalateLevel
// (typically Critical), ensuring that persistently misbehaving tokens receive
// higher-priority attention even if they were originally classified as Warning.
//
// Usage:
//
//	esc := tokenwatch.NewEscalation(tokenwatch.EscalationConfig{
//		Window:        5 * time.Minute,
//		Threshold:     3,
//		EscalateLevel: alert.LevelCritical,
//	})
//
//	effectiveLevel := esc.Check(tokenID, originalLevel)
//
// Call Reset to clear state for a token that has been remediated.
package tokenwatch
