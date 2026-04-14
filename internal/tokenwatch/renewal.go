package tokenwatch

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// RenewalRecord tracks the last renewal attempt for a token.
type RenewalRecord struct {
	Token     string
	RenewedAt time.Time
	Err       error
}

// Renewer is a function that renews a token and returns its new TTL.
type Renewer func(token string) (time.Duration, error)

// RenewalManager manages token renewals based on expiry classification.
type RenewalManager struct {
	mu       sync.Mutex
	records  map[string]RenewalRecord
	renewer  Renewer
	classify ExpiryClassifier
	now      func() time.Time
}

// NewRenewalManager creates a RenewalManager with the given renewer and classifier.
func NewRenewalManager(renewer Renewer, classify ExpiryClassifier) (*RenewalManager, error) {
	if renewer == nil {
		return nil, errors.New("tokenwatch: renewer must not be nil")
	}
	if classify == nil {
		return nil, errors.New("tokenwatch: classifier must not be nil")
	}
	return &RenewalManager{
		records: make(map[string]RenewalRecord),
		renewer: renewer,
		classify: classify,
		now:     time.Now,
	}, nil
}

// MaybeRenew renews the token if its TTL is within the warning or critical window.
// It skips renewal if the token was already renewed within the last minute.
func (m *RenewalManager) MaybeRenew(token string, ttl time.Duration) error {
	level := m.classify(ttl)
	if level == LevelInfo {
		return nil
	}

	m.mu.Lock()
	rec, seen := m.records[token]
	if seen && m.now().Sub(rec.RenewedAt) < time.Minute {
		m.mu.Unlock()
		return nil
	}
	m.mu.Unlock()

	_, err := m.renewer(token)

	m.mu.Lock()
	m.records[token] = RenewalRecord{
		Token:     token,
		RenewedAt: m.now(),
		Err:       err,
	}
	m.mu.Unlock()

	if err != nil {
		return fmt.Errorf("tokenwatch: renewal failed for token %q: %w", token, err)
	}
	return nil
}

// LastRecord returns the most recent RenewalRecord for a token, if any.
func (m *RenewalManager) LastRecord(token string) (RenewalRecord, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	rec, ok := m.records[token]
	return rec, ok
}
