// Package filter provides lease filtering capabilities for vaultwatch.
// It allows narrowing down leases by path prefix, severity level, and
// custom label selectors before alerts are dispatched.
package filter

import (
	"strings"

	"github.com/your-org/vaultwatch/internal/alert"
)

// Options holds the criteria used to filter alerts.
type Options struct {
	// PathPrefix filters alerts whose LeaseID starts with this prefix.
	PathPrefix string
	// MinLevel drops alerts below this severity (Info < Warning < Critical).
	MinLevel alert.Level
	// Labels retains only alerts whose Labels map contains all key/value pairs.
	Labels map[string]string
}

// Filter returns the subset of alerts that satisfy all criteria in opts.
func Filter(alerts []alert.Alert, opts Options) []alert.Alert {
	out := make([]alert.Alert, 0, len(alerts))
	for _, a := range alerts {
		if !matchesPrefix(a, opts.PathPrefix) {
			continue
		}
		if a.Level < opts.MinLevel {
			continue
		}
		if !matchesLabels(a, opts.Labels) {
			continue
		}
		out = append(out, a)
	}
	return out
}

func matchesPrefix(a alert.Alert, prefix string) bool {
	if prefix == "" {
		return true
	}
	return strings.HasPrefix(a.LeaseID, prefix)
}

func matchesLabels(a alert.Alert, labels map[string]string) bool {
	for k, v := range labels {
		if a.Labels[k] != v {
			return false
		}
	}
	return true
}
