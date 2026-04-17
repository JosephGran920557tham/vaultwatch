// Package tokenwatch — debounce.go
//
// Debounce suppresses repeated notifications for the same token within a
// configurable wait window. Unlike Cooldown (which resets on each call),
// Debounce gates on the time since the last *allowed* emission, making it
// suitable for edge-triggered alert pipelines where you want at most one
// alert per token per window regardless of how many checks fire.
//
// Usage:
//
//	d := tokenwatch.NewDebounce(tokenwatch.DebounceConfig{Wait: 30 * time.Second})
//	if d.Allow(tokenID) {
//		// send alert
//	}
package tokenwatch
