package alert

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestClassify(t *testing.T) {
	cases := []struct {
		name      string
		expires   time.Duration
		warnMins  int
		critMins  int
		expected  Level
	}{
		{"critical", 5 * time.Minute, 30, 10, LevelCritical},
		{"warning", 20 * time.Minute, 30, 10, LevelWarning},
		{"info", 60 * time.Minute, 30, 10, LevelInfo},
		{"exactly critical boundary", 10 * time.Minute, 30, 10, LevelCritical},
		{"exactly warning boundary", 30 * time.Minute, 30, 10, LevelWarning},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Classify(tc.expires, tc.warnMins, tc.critMins)
			if got != tc.expected {
				t.Errorf("Classify(%v, %d, %d) = %v; want %v",
					tc.expires, tc.warnMins, tc.critMins, got, tc.expected)
			}
		})
	}
}

func TestBuild(t *testing.T) {
	a := Build("secret/data/foo", 5*time.Minute, 30, 10)
	if a.LeaseID != "secret/data/foo" {
		t.Errorf("unexpected LeaseID: %s", a.LeaseID)
	}
	if a.Level != LevelCritical {
		t.Errorf("expected CRITICAL, got %s", a.Level)
	}
	if !strings.Contains(a.Message, "expires in") {
		t.Errorf("message missing 'expires in': %s", a.Message)
	}
}

func TestConsoleNotifier_Send(t *testing.T) {
	var buf bytes.Buffer
	n := &ConsoleNotifier{Out: &buf}
	a := Alert{
		LeaseID:   "auth/token/abc",
		ExpiresIn: 15 * time.Minute,
		Level:     LevelWarning,
		Message:   "Lease expires in 15m0s",
	}
	if err := n.Send(a); err != nil {
		t.Fatalf("Send() error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "WARNING") {
		t.Errorf("output missing WARNING: %s", out)
	}
	if !strings.Contains(out, "auth/token/abc") {
		t.Errorf("output missing lease ID: %s", out)
	}
}

func TestConsoleNotifier_Send_AllLevels(t *testing.T) {
	cases := []struct {
		level   Level
		label   string
	}{
		{LevelInfo, "INFO"},
		{LevelWarning, "WARNING"},
		{LevelCritical, "CRITICAL"},
	}

	for _, tc := range cases {
		t.Run(tc.label, func(t *testing.T) {
			var buf bytes.Buffer
			n := &ConsoleNotifier{Out: &buf}
			a := Alert{
				LeaseID:   "secret/test",
				ExpiresIn: 10 * time.Minute,
				Level:     tc.level,
				Message:   "Lease expires in 10m0s",
			}
			if err := n.Send(a); err != nil {
				t.Fatalf("Send() error: %v", err)
			}
			if !strings.Contains(buf.String(), tc.label) {
				t.Errorf("output missing %s: %s", tc.label, buf.String())
			}
		})
	}
}
