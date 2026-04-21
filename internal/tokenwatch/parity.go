package tokenwatch

import (
	"fmt"
	"sync"
	"time"

	"github.com/vaultwatch/internal/alert"
)

// DefaultParityConfig returns a ParityConfig with sensible defaults.
func DefaultParityConfig() ParityConfig {
	return ParityConfig{
		MaxSkew:          30 * time.Second,
		WarningThreshold: 0.10,
		CriticalThreshold: 0.25,
	}
}

// ParityConfig controls how TTL parity between token pairs is evaluated.
type ParityConfig struct {
	// MaxSkew is the absolute TTL difference tolerated before alerting.
	MaxSkew time.Duration
	// WarningThreshold is the relative skew ratio (0–1) for a warning alert.
	WarningThreshold float64
	// CriticalThreshold is the relative skew ratio (0–1) for a critical alert.
	CriticalThreshold float64
}

// ParityDetector compares TTLs between paired tokens and emits alerts when
// they diverge beyond configured thresholds.
type ParityDetector struct {
	cfg   ParityConfig
	mu    sync.Mutex
	pairs map[string]string // tokenID -> peerID
}

// NewParityDetector constructs a ParityDetector. Zero-value fields in cfg are
// replaced with defaults.
func NewParityDetector(cfg ParityConfig) *ParityDetector {
	def := DefaultParityConfig()
	if cfg.MaxSkew <= 0 {
		cfg.MaxSkew = def.MaxSkew
	}
	if cfg.WarningThreshold <= 0 {
		cfg.WarningThreshold = def.WarningThreshold
	}
	if cfg.CriticalThreshold <= 0 {
		cfg.CriticalThreshold = def.CriticalThreshold
	}
	return &ParityDetector{
		cfg:   cfg,
		pairs: make(map[string]string),
	}
}

// Pair registers two token IDs as peers for parity checking.
func (p *ParityDetector) Pair(tokenID, peerID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.pairs[tokenID] = peerID
}

// Check compares tokenTTL against peerTTL and returns an alert if the skew
// exceeds configured thresholds. Returns nil when parity is acceptable.
func (p *ParityDetector) Check(tokenID string, tokenTTL, peerTTL time.Duration) *alert.Alert {
	if tokenTTL <= 0 || peerTTL <= 0 {
		return nil
	}
	skew := tokenTTL - peerTTL
	if skew < 0 {
		skew = -skew
	}
	base := peerTTL
	if base <= 0 {
		return nil
	}
	ratio := float64(skew) / float64(base)

	var level alert.Level
	switch {
	case skew >= p.cfg.MaxSkew && ratio >= p.cfg.CriticalThreshold:
		level = alert.LevelCritical
	case skew >= p.cfg.MaxSkew || ratio >= p.cfg.WarningThreshold:
		level = alert.LevelWarning
	default:
		return nil
	}

	return &alert.Alert{
		LeaseID:   tokenID,
		Level:     level,
		Message:   fmt.Sprintf("token %s TTL parity skew %.1fs (%.0f%%) vs peer", tokenID, skew.Seconds(), ratio*100),
		ExpiresAt: time.Now().Add(tokenTTL),
	}
}
