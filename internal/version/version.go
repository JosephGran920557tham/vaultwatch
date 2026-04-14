// Package version provides build-time version information for vaultwatch.
package version

import "fmt"

// Build-time variables injected via ldflags.
var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
	GoVersion = "unknown"
)

// Info holds structured version metadata.
type Info struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildDate string `json:"build_date"`
	GoVersion string `json:"go_version"`
}

// Get returns the current version Info populated from build-time variables.
func Get() Info {
	return Info{
		Version:   Version,
		Commit:    Commit,
		BuildDate: BuildDate,
		GoVersion: GoVersion,
	}
}

// String returns a human-readable one-line version string.
func (i Info) String() string {
	return fmt.Sprintf("vaultwatch %s (commit=%s, built=%s, go=%s)",
		i.Version, i.Commit, i.BuildDate, i.GoVersion)
}

// Short returns only the semver version string.
func (i Info) Short() string {
	return i.Version
}
