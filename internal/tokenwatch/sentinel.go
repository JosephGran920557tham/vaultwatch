package tokenwatch

import (
	"fmt"
	"sync"
	"time"

	"github.com/vaultwatch/internal/alert"
)

// DefaultSentinelConfig returns a SentinelConfig with sensible defaults.
func DefaultSentinelConfig() SentinelConfig {
	return SentinelConfig{
		MissWindow:   5 * time.Minute,
		MissThreshold: 3,
	}
}

// SentinelConfig controls how the SentinelDetector identifies tokens
// that have repeatedly failed to report within the expected window.
type SentinelConfig struct {
	// MissWindow is the duration within which a token must check in.
	MissWindow time.Duration
	// MissThreshold is the number of consecutive misses before an alert fires.
	MissThreshold int
}

type sentinelEntry struct {
	lastSeen time.Time
	misses   int
}

// SentinelDetector tracks consecutive check-in misses per token.
type SentinelDetector struct {
	cfg     SentinelConfig
	mu      sync.Mutex
	entries map[string]*sentinelEntry
}

// NewSentinelDetector constructs a SentinelDetector. Zero-value config
// fields are replaced with defaults.
func NewSentinelDetector(cfg SentinelConfig) *SentinelDetector {
	def := DefaultSentinelConfig()
	if cfg.MissWindow <= 0 {
		cfg.MissWindow = def.MissWindow
	}
	if cfg.MissThreshold <= 0 {
		cfg.MissThreshold = def.MissThreshold
	}
	return &SentinelDetector{
		cfg:     cfg,
		entries: make(map[string]*sentinelEntry),
	}
}

// Ping records a successful check-in for the given token ID.
func (s *SentinelDetector) Ping(tokenID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[tokenID] = &sentinelEntry{lastSeen: time.Now(), misses: 0}
}

// Check evaluates whether the token has exceeded the miss threshold.
// It returns a critical alert if so, nil otherwise.
func (s *SentinelDetector) Check(tokenID string, now time.Time) *alert.Alert {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.entries[tokenID]
	if !ok {
		s.entries[tokenID] = &sentinelEntry{lastSeen: now}
		return nil
	}

	if now.Sub(entry.lastSeen) >= s.cfg.MissWindow {
		entry.misses++
		entry.lastSeen = now
	}

	if entry.misses >= s.cfg.MissThreshold {
		return &alert.Alert{
			LeaseID: tokenID,
			Level:   alert.Critical,
			Message: fmt.Sprintf("token %s missed %d consecutive check-ins", tokenID, entry.misses),
		}
	}
	return nil
}
