package filter_test

import (
	"testing"

	"github.com/your-org/vaultwatch/internal/alert"
	"github.com/your-org/vaultwatch/internal/filter"
)

func TestPresetOptions_CriticalOnly(t *testing.T) {
	opts := filter.PresetOptions(filter.PresetCriticalOnly)
	if opts.MinLevel != alert.LevelCritical {
		t.Errorf("expected LevelCritical, got %v", opts.MinLevel)
	}
}

func TestPresetOptions_WarningAndAbove(t *testing.T) {
	opts := filter.PresetOptions(filter.PresetWarningAndAbove)
	if opts.MinLevel != alert.LevelWarning {
		t.Errorf("expected LevelWarning, got %v", opts.MinLevel)
	}
}

func TestPresetOptions_Unknown_ReturnsEmpty(t *testing.T) {
	opts := filter.PresetOptions("unknown-preset")
	if opts.MinLevel != alert.LevelInfo {
		t.Errorf("expected LevelInfo (zero), got %v", opts.MinLevel)
	}
	if opts.PathPrefix != "" {
		t.Errorf("expected empty prefix, got %q", opts.PathPrefix)
	}
}

func TestMerge_OverrideWins(t *testing.T) {
	base := filter.Options{PathPrefix: "secret/", MinLevel: alert.LevelInfo}
	override := filter.Options{PathPrefix: "secret/db", MinLevel: alert.LevelCritical}
	result := filter.Merge(base, override)
	if result.PathPrefix != "secret/db" {
		t.Errorf("expected secret/db, got %s", result.PathPrefix)
	}
	if result.MinLevel != alert.LevelCritical {
		t.Errorf("expected LevelCritical, got %v", result.MinLevel)
	}
}

func TestMerge_Labels_AreCombined(t *testing.T) {
	base := filter.Options{Labels: map[string]string{"env": "prod"}}
	override := filter.Options{Labels: map[string]string{"team": "platform"}}
	result := filter.Merge(base, override)
	if result.Labels["env"] != "prod" {
		t.Error("base label 'env' should be preserved")
	}
	if result.Labels["team"] != "platform" {
		t.Error("override label 'team' should be present")
	}
}
