// Package version exposes build-time metadata for the vaultwatch binary.
//
// Variables such as Version, Commit, BuildDate, and GoVersion are intended
// to be overridden at link time using -ldflags, for example:
//
//	-ldflags "-X github.com/your-org/vaultwatch/internal/version.Version=1.0.0"
//	         "-X github.com/your-org/vaultwatch/internal/version.Commit=$(git rev-parse --short HEAD)"
//	         "-X github.com/your-org/vaultwatch/internal/version.BuildDate=$(date -u +%Y-%m-%d)"
//
// Use version.Get() to retrieve a structured Info value, or Info.String()
// for a human-readable summary suitable for CLI output.
package version
