package tokenwatch

import (
	"testing"
	"time"
)

func TestDefaultJitterConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultJitterConfig()
	if cfg.Factor != 0.2 {
		t.Errorf("expected Factor 0.2, got %v", cfg.Factor)
	}
}

func TestNewJitter_ClampsNegativeFactor(t *testing.T) {
	j := NewJitter(JitterConfig{Factor: -0.5})
	if j.Factor() != 0 {
		t.Errorf("expected factor clamped to 0, got %v", j.Factor())
	}
}

func TestNewJitter_ClampsFactorAboveOne(t *testing.T) {
	j := NewJitter(JitterConfig{Factor: 1.5})
	if j.Factor() != 1 {
		t.Errorf("expected factor clamped to 1, got %v", j.Factor())
	}
}

func TestJitter_Apply_ZeroFactor_ReturnsBase(t *testing.T) {
	j := NewJitter(JitterConfig{Factor: 0})
	base := 10 * time.Second
	result := j.Apply(base)
	if result != base {
		t.Errorf("expected %v, got %v", base, result)
	}
}

func TestJitter_Apply_NegativeBase_ReturnsBase(t *testing.T) {
	j := NewJitter(DefaultJitterConfig())
	base := -5 * time.Second
	result := j.Apply(base)
	if result != base {
		t.Errorf("expected unchanged negative base %v, got %v", base, result)
	}
}

func TestJitter_Apply_IncreasesOrEqualBase(t *testing.T) {
	j := NewJitter(JitterConfig{Factor: 0.3})
	base := 10 * time.Second
	for i := 0; i < 50; i++ {
		result := j.Apply(base)
		if result < base {
			t.Errorf("jitter should not reduce base: got %v < %v", result, base)
		}
		max := base + time.Duration(float64(base)*0.3)
		if result > max {
			t.Errorf("jitter exceeded max %v: got %v", max, result)
		}
	}
}

func TestJitter_Apply_FullFactor_BoundedByBase(t *testing.T) {
	j := NewJitter(JitterConfig{Factor: 1.0})
	base := 5 * time.Second
	for i := 0; i < 50; i++ {
		result := j.Apply(base)
		if result < base || result > 2*base {
			t.Errorf("result %v outside expected range [%v, %v]", result, base, 2*base)
		}
	}
}
