package filter

import "github.com/your-org/vaultwatch/internal/alert"

// Preset names for common filter configurations.
const (
	PresetCriticalOnly = "critical-only"
	PresetWarningAndAbove = "warning-and-above"
	PresetAll = "all"
)

// PresetOptions returns a pre-built Options struct for well-known filter
// presets. Unknown preset names fall back to an empty Options (no filtering).
func PresetOptions(name string) Options {
	switch name {
	case PresetCriticalOnly:
		return Options{MinLevel: alert.LevelCritical}
	case PresetWarningAndAbove:
		return Options{MinLevel: alert.LevelWarning}
	default:
		return Options{}
	}
}

// Merge combines two Options structs. Fields from override take precedence
// over base when they are non-zero.
func Merge(base, override Options) Options {
	result := base
	if override.PathPrefix != "" {
		result.PathPrefix = override.PathPrefix
	}
	if override.MinLevel > result.MinLevel {
		result.MinLevel = override.MinLevel
	}
	if len(override.Labels) > 0 {
		merged := make(map[string]string, len(result.Labels)+len(override.Labels))
		for k, v := range result.Labels {
			merged[k] = v
		}
		for k, v := range override.Labels {
			merged[k] = v
		}
		result.Labels = merged
	}
	return result
}
