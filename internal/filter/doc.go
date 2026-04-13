// Package filter provides composable filtering for Vault lease alerts.
//
// Filters can be chained by building an Options struct that specifies
// one or more of the following criteria:
//
//   - PathPrefix – retain only leases whose ID begins with a given path.
//   - MinLevel   – drop alerts below a minimum severity (Info/Warning/Critical).
//   - Labels     – retain only alerts whose metadata labels match all
//     supplied key/value pairs.
//
// Example usage:
//
//	filtered := filter.Filter(alerts, filter.Options{
//	    PathPrefix: "secret/db",
//	    MinLevel:   alert.LevelWarning,
//	})
package filter
