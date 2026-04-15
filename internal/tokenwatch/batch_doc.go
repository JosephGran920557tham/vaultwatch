// Package tokenwatch provides facilities for monitoring HashiCorp Vault
// token leases and alerting on upcoming expirations.
//
// # Batch Processing
//
// BatchRunner executes token health checks across all tokens registered in a
// Registry concurrently, respecting a configurable concurrency limit to avoid
// overwhelming the Vault API.
//
//	runner, _ := tokenwatch.NewBatchRunner(registry, alerter, 8)
//	results := runner.Run(ctx)
//
// BatchReporter formats the aggregated BatchReport for human-readable text
// output or structured JSON suitable for downstream consumption.
//
//	rpt := tokenwatch.BuildBatchReport(results)
//	tokenwatch.NewBatchReporter(os.Stdout, "json").Write(rpt)
package tokenwatch
