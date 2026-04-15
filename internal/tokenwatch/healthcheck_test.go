package tokenwatch

import (
	"context"
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/alert"
)

func makeHealthAlert(level, tokenID string) alert.Alert {
	return alert.Alert{
		LeaseID:   tokenID,
		Level:     level,
		Message:   "test alert",
		ExpiresAt: time.Now().Add(time.Hour),
	}
}

func TestNewHealthChecker_NilAlerter(t *testing.T) {
	c := DefaultExpiryClassifier()
	_, err := NewHealthChecker(nil, c)
	if err == nil {
		t.Fatal("expected error for nil alerter")
	}
}

func TestNewHealthChecker_NilClassifier(t *testing.T) {
	reg := NewRegistry()
	w, _ := New(func(_ context.Context, _ string) (TokenInfo, error) {
		return TokenInfo{TTL: time.Hour}, nil
	}, 30*time.Second, time.Hour)
	a, _ := NewAlerter(reg, w)
	_, err := NewHealthChecker(a, nil)
	if err == nil {
		t.Fatal("expected error for nil classifier")
	}
}

func TestHealthStatus_IsHealthy_NoCritical(t *testing.T) {
	s := HealthStatus{Healthy: 3, Warning: 1, Critical: 0, Total: 4}
	if !s.IsHealthy() {
		t.Error("expected healthy when no critical tokens")
	}
}

func TestHealthStatus_IsHealthy_WithCritical(t *testing.T) {
	s := HealthStatus{Healthy: 1, Warning: 0, Critical: 2, Total: 3}
	if s.IsHealthy() {
		t.Error("expected unhealthy when critical tokens exist")
	}
}

func TestHealthStatus_String_ContainsCounts(t *testing.T) {
	s := HealthStatus{Healthy: 2, Warning: 1, Critical: 1, Total: 4}
	str := s.String()
	for _, want := range []string{"total=4", "healthy=2", "warning=1", "critical=1"} {
		if !contains(str, want) {
			t.Errorf("String() missing %q, got: %s", want, str)
		}
	}
}

func TestHealthChecker_Last_NilBeforeCheck(t *testing.T) {
	reg := NewRegistry()
	w, _ := New(func(_ context.Context, _ string) (TokenInfo, error) {
		return TokenInfo{TTL: time.Hour}, nil
	}, 30*time.Second, time.Hour)
	a, _ := NewAlerter(reg, w)
	c := DefaultExpiryClassifier()
	hc, err := NewHealthChecker(a, c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hc.Last() != nil {
		t.Error("expected nil before first Check call")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsRune(s, sub))
}

func containsRune(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
