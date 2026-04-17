// Package tokenwatch — heartbeat subsystem.
//
// The heartbeat subsystem detects tokens that have stopped being observed by
// the VaultWatch polling loop. Each time a token is successfully looked up in
// Vault, HeartbeatDetector.Beat is called to record the current time.
//
// If a token has not been seen within StaleAfter a Warning alert is raised.
// If it has not been seen within CriticalAfter a Critical alert is raised.
// Tokens that have never been observed are immediately Critical.
//
// Usage:
//
//	detector := tokenwatch.NewHeartbeatDetector(tokenwatch.DefaultHeartbeatConfig())
//	scanner  := tokenwatch.NewHeartbeatScanner(registry, detector, vaultLookup)
//	alerts   := scanner.Scan()
package tokenwatch
