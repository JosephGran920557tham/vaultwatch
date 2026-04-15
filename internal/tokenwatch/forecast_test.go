package tokenwatch

import (
	"errors"
	"testing"
	"time"

	"github.com/vaultwatch/internal/alert"
)

func TestDefaultForecastConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultForecastConfig()
	if cfg.SampleWindow <= 0 {
		t.Error("expected positive SampleWindow")
	}
	if cfg.ProjectionHorizon <= 0 {
		t.Error("expected positive ProjectionHorizon")
	}
	if cfg.CriticalThreshold >= cfg.WarningThreshold {
		t.Error("expected CriticalThreshold < WarningThreshold")
	}
}

func TestNewForecastDetector_ZeroValues_UsesDefaults(t *testing.T) {
	d := NewForecastDetector(ForecastConfig{})
	def := DefaultForecastConfig()
	if d.cfg.SampleWindow != def.SampleWindow {
		t.Errorf("want SampleWindow %d, got %d", def.SampleWindow, d.cfg.SampleWindow)
	}
}

func TestForecastDetector_InsufficientSamples_ReturnsNil(t *testing.T) {
	d := NewForecastDetector(DefaultForecastConfig())
	d.Record(5 * time.Minute)
	if got := d.Check("tok-1"); got != nil {
		t.Errorf("expected nil with one sample, got %+v", got)
	}
}

func TestForecastDetector_ProjectedTTL_Warning(t *testing.T) {
	cfg := ForecastConfig{
		SampleWindow:      4,
		ProjectionHorizon: 10 * time.Minute,
		CriticalThreshold: 5 * time.Minute,
		WarningThreshold:  15 * time.Minute,
	}
	d := NewForecastDetector(cfg)
	// Simulate TTL decaying by 2 min per sample
	for _, ttl := range []time.Duration{30, 28, 26, 24} {
		d.Record(ttl * time.Minute)
	}
	a := d.Check("tok-warn")
	if a == nil {
		t.Fatal("expected a warning alert")
	}
	if a.Level != alert.LevelWarning {
		t.Errorf("expected Warning, got %v", a.Level)
	}
}

func TestForecastDetector_ProjectedTTL_Critical(t *testing.T) {
	cfg := ForecastConfig{
		SampleWindow:      3,
		ProjectionHorizon: 10 * time.Minute,
		CriticalThreshold: 20 * time.Minute,eshold:  30 * time.Minute,
	}
	d := NewForecastDetector(cfg)
	for _, ttl := range []time.Duration{25, 22, 19} {
		d.Record(ttl * time.Minute)
	}
	a := d.Check("tok-crit")
	if a == nil {
		t.Fatal("expected a critical alert")
	}
	if a.Level != alert.LevelCritical {
		t.Errorf("expected Critical, got %v", a.Level)
	}
}

func TestForecastScanner_Scan_NoAnomalies(t *testing.T) {
	reg := NewRegistry()
	_ = reg.Add("tok-stable")
	scanner := NewForecastScanner(reg, DefaultForecastConfig(), nil)
	lookup := func(id string) (TokenInfo, error) {
		return TokenInfo{ID: id, TTL: 60 * time.Minute}, nil
	}
	// Run twice to populate samples — stable TTL should not alert
	scanner.Scan(lookup)
	alerts := scanner.Scan(lookup)
	if len(alerts) != 0 {
		t.Errorf("expected no alerts for stable TTL, got %d", len(alerts))
	}
}

func TestForecastScanner_Scan_LookupError_Skipped(t *testing.T) {
	reg := NewRegistry()
	_ = reg.Add("tok-err")
	scanner := NewForecastScanner(reg, DefaultForecastConfig(), nil)
	lookup := func(id string) (TokenInfo, error) {
		return TokenInfo{}, errors.New("vault unavailable")
	}
	alerts := scanner.Scan(lookup)
	if len(alerts) != 0 {
		t.Errorf("expected no alerts on lookup error, got %d", len(alerts))
	}
}

func TestNewForecastScanner_NilRegistry_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil registry")
		}
	}()
	NewForecastScanner(nil, DefaultForecastConfig(), nil)
}
