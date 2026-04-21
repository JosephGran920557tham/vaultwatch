package tokenwatch

import (
	"errors"
	"testing"
)

func newTestFingerprintScanner(lookup func(string) (map[string]string, error)) *FingerprintScanner {
	r := NewRegistry()
	fp := NewFingerprint(DefaultFingerprintConfig())
	return NewFingerprintScanner(r, fp, lookup)
}

func TestNewFingerprintScanner_NilRegistry_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil registry")
		}
	}()
	NewFingerprintScanner(nil, NewFingerprint(DefaultFingerprintConfig()), func(string) (map[string]string, error) { return nil, nil })
}

func TestNewFingerprintScanner_NilFingerprint_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil fingerprint")
		}
	}()
	NewFingerprintScanner(NewRegistry(), nil, func(string) (map[string]string, error) { return nil, nil })
}

func TestNewFingerprintScanner_NilLookup_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil lookup")
		}
	}()
	NewFingerprintScanner(NewRegistry(), NewFingerprint(DefaultFingerprintConfig()), nil)
}

func TestFingerprintScanner_Scan_NoTokens_ReturnsEmpty(t *testing.T) {
	s := newTestFingerprintScanner(func(string) (map[string]string, error) {
		return map[string]string{"role": "admin"}, nil
	})
	alerts := s.Scan()
	if len(alerts) != 0 {
		t.Fatalf("expected no alerts for empty registry, got %d", len(alerts))
	}
}

func TestFingerprintScanner_Scan_FirstObservation_EmitsAlert(t *testing.T) {
	r := NewRegistry()
	_ = r.Add("tok-1")
	fp := NewFingerprint(DefaultFingerprintConfig())
	s := NewFingerprintScanner(r, fp, func(string) (map[string]string, error) {
		return map[string]string{"role": "admin"}, nil
	})
	// First scan: token is new, Track returns true → alert emitted.
	alerts := s.Scan()
	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert on first observation, got %d", len(alerts))
	}
}

func TestFingerprintScanner_Scan_StableHash_NoAlert(t *testing.T) {
	r := NewRegistry()
	_ = r.Add("tok-stable")
	fp := NewFingerprint(DefaultFingerprintConfig())
	s := NewFingerprintScanner(r, fp, func(string) (map[string]string, error) {
		return map[string]string{"role": "viewer"}, nil
	})
	s.Scan() // seed
	alerts := s.Scan()
	if len(alerts) != 0 {
		t.Fatalf("expected no alerts for stable fingerprint, got %d", len(alerts))
	}
}

func TestFingerprintScanner_Scan_ChangedHash_EmitsAlert(t *testing.T) {
	r := NewRegistry()
	_ = r.Add("tok-change")
	fp := NewFingerprint(DefaultFingerprintConfig())
	call := 0
	s := NewFingerprintScanner(r, fp, func(string) (map[string]string, error) {
		call++
		if call == 1 {
			return map[string]string{"role": "admin"}, nil
		}
		return map[string]string{"role": "superuser"}, nil
	})
	s.Scan() // seed with "admin"
	alerts := s.Scan() // should detect change to "superuser"
	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert on fingerprint change, got %d", len(alerts))
	}
}

func TestFingerprintScanner_Scan_LookupError_Skips(t *testing.T) {
	r := NewRegistry()
	_ = r.Add("tok-err")
	s := newTestFingerprintScanner(func(string) (map[string]string, error) {
		return nil, errors.New("vault unavailable")
	})
	alerts := s.Scan()
	if len(alerts) != 0 {
		t.Fatalf("expected no alerts on lookup error, got %d", len(alerts))
	}
}
