// Package tokenwatch provides token lifecycle monitoring for HashiCorp Vault.
//
// # Scoreboard
//
// Scoreboard aggregates per-token risk scores by accumulating points each time
// an alert is recorded for a given token. Points are weighted by severity:
//
//   - Critical → 10 points
//   - Warning  →  3 points
//   - Info     →  1 point
//
// Tokens with the highest accumulated scores can be retrieved via Top(n),
// which returns entries sorted in descending order. This makes it easy to
// identify the most at-risk tokens at a glance.
//
// Example usage:
//
//	sb := tokenwatch.NewScoreboard()
//	sb.Record(alert)
//	for _, e := range sb.Top(5) {
//		fmt.Printf("%s: %d\n", e.TokenID, e.Score)
//	}
//
// Call Reset between monitoring cycles to prevent unbounded score growth.
package tokenwatch
