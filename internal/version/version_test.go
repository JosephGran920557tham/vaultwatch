package version_test

import (
	"strings"
	"testing"

	"github.com/your-org/vaultwatch/internal/version"
)

func TestGet_ReturnsDefaults(t *testing.T) {
	info := version.Get()

	if info.Version == "" {
		t.Error("expected non-empty Version")
	}
	if info.Commit == "" {
		t.Error("expected non-empty Commit")
	}
	if info.BuildDate == "" {
		t.Error("expected non-empty BuildDate")
	}
	if info.GoVersion == "" {
		t.Error("expected non-empty GoVersion")
	}
}

func TestInfo_String_ContainsVersion(t *testing.T) {
	version.Version = "1.2.3"
	version.Commit = "abc1234"
	version.BuildDate = "2024-01-01"
	version.GoVersion = "go1.22.0"

	info := version.Get()
	s := info.String()

	for _, want := range []string{"vaultwatch", "1.2.3", "abc1234", "2024-01-01", "go1.22.0"} {
		if !strings.Contains(s, want) {
			t.Errorf("String() missing %q, got: %s", want, s)
		}
	}
}

func TestInfo_Short_ReturnsVersion(t *testing.T) {
	version.Version = "2.0.0"

	info := version.Get()
	if got := info.Short(); got != "2.0.0" {
		t.Errorf("Short() = %q, want %q", got, "2.0.0")
	}
}

func TestInfo_String_StartsWithVaultwatch(t *testing.T) {
	info := version.Get()
	if !strings.HasPrefix(info.String(), "vaultwatch ") {
		t.Errorf("String() should start with 'vaultwatch ', got: %s", info.String())
	}
}
