package tokenwatch

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/vaultwatch/internal/alert"
)

// stubLookup returns a fixed TokenInfo or error.
type stubLookup struct {
	info TokenInfo
	err  error
}

func (s *stubLookup) lookup(_ context.Context, _ string) (TokenInfo, error) {
	return s.info, s.err
}

func newAlerterFixture(t *testing.T, info TokenInfo, lookupErr error) (*Alerter, *Registry) {
	t.Helper()
	reg := NewRegistry()
	thresholds := Thresholds{Warning: 30 * time.Minute, Critical: 10 * time.Minute}
	stub := &stubLookup{info: info, err: lookupErr}
	watcher, err := New(stub.lookup, thresholds)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	throttle := NewThrottle(ThrottleConfig{MinInterval: time.Hour})
	alerter, err := NewAlerter(reg, watcher, throttle)
	if err != nil {
		t.Fatalf("NewAlerter: %v", err)
	}
	return alerter, reg
}

func TestNewAlerter_NilRegistry(t *testing.T) {
	_, err := NewAlerter(nil, &Watcher{}, NewThrottle(DefaultThrottleConfig()))
	if err == nil {
		t.Error("expected error for nil registry")
	}
}

func TestNewAlerter_NilWatcher(t *testing.T) {
	_, err := NewAlerter(NewRegistry(), nil, NewThrottle(DefaultThrottleConfig()))
	if err == nil {
		t.Error("expected error for nil watcher")
	}
}

func TestAlerter_CheckAll_NoTokens(t *testing.T) {
	alerter, _ := newAlerterFixture(t, TokenInfo{TTL: time.Hour}, nil)
	alerts, err := alerter.CheckAll(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(alerts) != 0 {
		t.Errorf("expected 0 alerts, got %d", len(alerts))
	}
}

func TestAlerter_CheckAll_ProducesAlert(t *testing.T) {
	alerter, reg := newAlerterFixture(t, TokenInfo{TTL: 5 * time.Minute}, nil)
	reg.Add("tok-critical")
	alerts, err := alerter.CheckAll(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(alerts) == 0 {
		t.Error("expected at least one alert for critical TTL")
	}
}

func TestAlerter_CheckAll_ThrottlesSameToken(t *testing.T) {
	alerter, reg := newAlerterFixture(t, TokenInfo{TTL: 5 * time.Minute}, nil)
	reg.Add("tok-dup")
	alerter.CheckAll(context.Background())
	alerts, _ := alerter.CheckAll(context.Background())
	if len(alerts) != 0 {
		t.Errorf("expected throttled alerts to be suppressed, got %d", len(alerts))
	}
}

func TestAlerter_CheckAll_SkipsLookupError(t *testing.T) {
	alerter, reg := newAlerterFixture(t, TokenInfo{}, errors.New("vault unavailable"))
	reg.Add("tok-err")
	alerts, err := alerter.CheckAll(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(alerts) != 0 {
		t.Errorf("expected 0 alerts on lookup error, got %d", len(alerts))
	}
}

func TestBuildTokenAlert_ContainsID(t *testing.T) {
	al := buildTokenAlert("my-token", 2*time.Minute, alert.LevelCritical)
	if al.LeaseID != "my-token" {
		t.Errorf("expected lease ID 'my-token', got %q", al.LeaseID)
	}
	if al.Level != alert.LevelCritical {
		t.Errorf("expected critical level, got %v", al.Level)
	}
}
