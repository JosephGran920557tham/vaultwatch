package tokenwatch

import (
	"errors"
	"testing"
	"time"

	"github.com/vaultwatch/internal/alert"
)

func makeEnvelopeAlert() alert.Alert {
	return alert.Alert{
		LeaseID: "token/test-123",
		Level:   alert.Warning,
		Message: "token expiring soon",
	}
}

func TestNewEnvelope_DefaultsMaxAttempts(t *testing.T) {
	env := NewEnvelope("tok-abc", makeEnvelopeAlert(), 0)
	if env.MaxAttempts != 3 {
		t.Errorf("expected MaxAttempts=3, got %d", env.MaxAttempts)
	}
}

func TestNewEnvelope_SetsFields(t *testing.T) {
	a := makeEnvelopeAlert()
	env := NewEnvelope("tok-xyz", a, 5)
	if env.Token != "tok-xyz" {
		t.Errorf("expected Token=tok-xyz, got %s", env.Token)
	}
	if env.Alert.LeaseID != a.LeaseID {
		t.Errorf("unexpected alert lease ID")
	}
	if env.Attempt != 0 {
		t.Errorf("expected Attempt=0, got %d", env.Attempt)
	}
	if env.MaxAttempts != 5 {
		t.Errorf("expected MaxAttempts=5, got %d", env.MaxAttempts)
	}
}

func TestEnvelope_Exhausted(t *testing.T) {
	env := NewEnvelope("tok", makeEnvelopeAlert(), 2)
	if env.Exhausted() {
		t.Error("should not be exhausted before any attempts")
	}
	env.Increment(errors.New("fail"))
	env.Increment(errors.New("fail again"))
	if !env.Exhausted() {
		t.Error("should be exhausted after max attempts")
	}
}

func TestEnvelope_Increment_RecordsError(t *testing.T) {
	env := NewEnvelope("tok", makeEnvelopeAlert(), 3)
	err := errors.New("delivery failed")
	env.Increment(err)
	if env.Attempt != 1 {
		t.Errorf("expected Attempt=1, got %d", env.Attempt)
	}
	if env.LastError != err {
		t.Errorf("expected LastError to be set")
	}
}

func TestEnvelope_Age_IsPositive(t *testing.T) {
	env := NewEnvelope("tok", makeEnvelopeAlert(), 3)
	time.Sleep(2 * time.Millisecond)
	if env.Age() <= 0 {
		t.Error("expected positive age")
	}
}
