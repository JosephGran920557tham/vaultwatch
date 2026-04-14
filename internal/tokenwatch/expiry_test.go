package tokenwatch

import (
	"strings"
	"testing"
	"time"

	"github.com/your-org/vaultwatch/internal/alert"
)

func TestDefaultExpiryClassifier_HasSensibleDefaults(t *testing.T) {
	ec := DefaultExpiryClassifier()
	if ec.WarnThreshold != 24*time.Hour {
		t.Errorf("expected warn threshold 24h, got %s", ec.WarnThreshold)
	}
	if ec.CriticalThreshold != 4*time.Hour {
		t.Errorf("expected critical threshold 4h, got %s", ec.CriticalThreshold)
	}
}

func TestClassify_Info(t *testing.T) {
	ec := DefaultExpiryClassifier()
	level := ec.Classify(48 * time.Hour)
	if level != alert.LevelInfo {
		t.Errorf("expected Info, got %s", level)
	}
}

func TestClassify_Warning(t *testing.T) {
	ec := DefaultExpiryClassifier()
	level := ec.Classify(12 * time.Hour)
	if level != alert.LevelWarning {
		t.Errorf("expected Warning, got %s", level)
	}
}

func TestClassify_Critical(t *testing.T) {
	ec := DefaultExpiryClassifier()
	level := ec.Classify(1 * time.Hour)
	if level != alert.LevelCritical {
		t.Errorf("expected Critical, got %s", level)
	}
}

func TestClassify_ExactBoundary_Critical(t *testing.T) {
	ec := DefaultExpiryClassifier()
	level := ec.Classify(4 * time.Hour)
	if level != alert.LevelCritical {
		t.Errorf("expected Critical at exact boundary, got %s", level)
	}
}

func TestSummary_ContainsTokenID(t *testing.T) {
	ec := DefaultExpiryClassifier()
	s := ec.Summary("tok-abc123", 2*time.Hour)
	if !strings.Contains(s, "tok-abc123") {
		t.Errorf("summary missing token ID: %s", s)
	}
}

func TestSummary_ContainsLevel(t *testing.T) {
	ec := DefaultExpiryClassifier()
	s := ec.Summary("tok-xyz", 1*time.Hour)
	if !strings.Contains(s, "critical") {
		t.Errorf("summary missing level: %s", s)
	}
}

func TestValidate_Valid(t *testing.T) {
	ec := DefaultExpiryClassifier()
	if err := ec.Validate(); err != nil {
		t.Errorf("unexpected validation error: %v", err)
	}
}

func TestValidate_CriticalGEWarn(t *testing.T) {
	ec := &ExpiryClassifier{
		WarnThreshold:     4 * time.Hour,
		CriticalThreshold: 8 * time.Hour,
	}
	if err := ec.Validate(); err == nil {
		t.Error("expected validation error when critical >= warn")
	}
}

func TestValidate_NegativeCritical(t *testing.T) {
	ec := &ExpiryClassifier{
		WarnThreshold:     24 * time.Hour,
		CriticalThreshold: -1 * time.Hour,
	}
	if err := ec.Validate(); err == nil {
		t.Error("expected validation error for negative critical threshold")
	}
}
