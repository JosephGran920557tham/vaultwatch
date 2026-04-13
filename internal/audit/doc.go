// Package audit provides structured, append-only audit logging for
// vaultwatch lease expiration events.
//
// Each alert processed by the monitor can be recorded as a newline-delimited
// JSON entry containing the lease ID, severity level, message, and remaining
// TTL at the time of observation.
//
// # Basic usage
//
//	// Write to a file
//	fw, err := audit.NewFileWriter("/var/log/vaultwatch/audit.log")
//	if err != nil { ... }
//	defer fw.Close()
//
//	logger := audit.NewLogger(fw)
//	logger.RecordAll(alerts)
//
// # Entry format
//
// Each entry is a single JSON object terminated by a newline:
//
//	{"timestamp":"2024-01-15T10:30:00Z","lease_id":"secret/db/creds",
//	 "level":"critical","message":"lease expires in 5m","ttl_seconds":300}
package audit
