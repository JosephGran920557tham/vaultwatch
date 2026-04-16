package tokenwatch

import (
	"testing"
	"time"
)

func TestDefaultSamplerConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultSamplerConfig()
	if cfg.MaxSamples <= 0 {
		t.Errorf("expected positive MaxSamples, got %d", cfg.MaxSamples)
	}
	if cfg.MaxAge <= 0 {
		t.Errorf("expected positive MaxAge, got %v", cfg.MaxAge)
	}
}

func TestNewSampler_ZeroValues_UsesDefaults(t *testing.T) {
	s := NewSampler(SamplerConfig{})
	def := DefaultSamplerConfig()
	if s.cfg.MaxSamples != def.MaxSamples {
		t.Errorf("expected %d, got %d", def.MaxSamples, s.cfg.MaxSamples)
	}
	if s.cfg.MaxAge != def.MaxAge {
		t.Errorf("expected %v, got %v", def.MaxAge, s.cfg.MaxAge)
	}
}

func TestSampler_Record_StoresSamples(t *testing.T) {
	s := NewSampler(SamplerConfig{MaxSamples: 10, MaxAge: time.Minute})
	s.Record("tok1", 30*time.Second)
	s.Record("tok1", 25*time.Second)

	samples := s.Samples("tok1")
	if len(samples) != 2 {
		t.Fatalf("expected 2 samples, got %d", len(samples))
	}
	if samples[0] != 30*time.Second {
		t.Errorf("expected first sample 30s, got %v", samples[0])
	}
}

func TestSampler_Record_CapsAtMaxSamples(t *testing.T) {
	s := NewSampler(SamplerConfig{MaxSamples: 3, MaxAge: time.Minute})
	for i := 0; i < 5; i++ {
		s.Record("tok1", time.Duration(i)*time.Second)
	}
	samples := s.Samples("tok1")
	if len(samples) != 3 {
		t.Fatalf("expected 3 samples, got %d", len(samples))
	}
	// Should keep the most recent 3 (2,3,4 seconds).
	if samples[0] != 2*time.Second {
		t.Errorf("expected oldest retained sample 2s, got %v", samples[0])
	}
}

func TestSampler_Samples_UnknownToken_ReturnsEmpty(t *testing.T) {
	s := NewSampler(SamplerConfig{MaxSamples: 10, MaxAge: time.Minute})
	samples := s.Samples("unknown")
	if len(samples) != 0 {
		t.Errorf("expected empty slice, got %v", samples)
	}
}

func TestSampler_Clear_RemovesSamples(t *testing.T) {
	s := NewSampler(SamplerConfig{MaxSamples: 10, MaxAge: time.Minute})
	s.Record("tok1", 10*time.Second)
	s.Clear("tok1")
	samples := s.Samples("tok1")
	if len(samples) != 0 {
		t.Errorf("expected empty after clear, got %v", samples)
	}
}

func TestSampler_IsolatesTokens(t *testing.T) {
	s := NewSampler(SamplerConfig{MaxSamples: 10, MaxAge: time.Minute})
	s.Record("tok1", 10*time.Second)
	s.Record("tok2", 20*time.Second)

	if got := s.Samples("tok1"); len(got) != 1 || got[0] != 10*time.Second {
		t.Errorf("tok1 samples wrong: %v", got)
	}
	if got := s.Samples("tok2"); len(got) != 1 || got[0] != 20*time.Second {
		t.Errorf("tok2 samples wrong: %v", got)
	}
}
