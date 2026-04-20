package tokenwatch

import (
	"fmt"
	"sync"
	"time"

	"github.com/your-org/vaultwatch/internal/alert"
)

// DefaultIsolationConfig returns sensible defaults for IsolationDetector.
func DefaultIsolationConfig() IsolationConfig {
	return IsolationConfig{
		MinPeers:       2,
		Window:         5 * time.Minute,
		CriticalFactor: 0.5,
	}
}

// IsolationConfig controls how the IsolationDetector identifies tokens
// whose TTL is significantly lower than their peers, suggesting they are
// isolated from the normal renewal cycle.
type IsolationConfig struct {
	// MinPeers is the minimum number of peer TTL samples required before
	// an isolation check is performed.
	MinPeers int
	// Window is the rolling duration over which peer samples are considered.
	Window time.Duration
	// CriticalFactor is the ratio (0–1) below which a token's TTL relative
	// to the peer median triggers a Critical alert.
	CriticalFactor float64
}

// IsolationDetector flags tokens whose TTL is an outlier compared to peers.
type IsolationDetector struct {
	cfg     IsolationConfig
	mu      sync.Mutex
	samples []peerSample
}

type peerSample struct {
	ttl       time.Duration
	recordedAt time.Time
}

// NewIsolationDetector creates an IsolationDetector, applying defaults for
// zero-value fields.
func NewIsolationDetector(cfg IsolationConfig) *IsolationDetector {
	def := DefaultIsolationConfig()
	if cfg.MinPeers <= 0 {
		cfg.MinPeers = def.MinPeers
	}
	if cfg.Window <= 0 {
		cfg.Window = def.Window
	}
	if cfg.CriticalFactor <= 0 || cfg.CriticalFactor > 1 {
		cfg.CriticalFactor = def.CriticalFactor
	}
	return &IsolationDetector{cfg: cfg}
}

// RecordPeer adds a peer TTL observation to the rolling window.
func (d *IsolationDetector) RecordPeer(ttl time.Duration) {
	d.mu.Lock()
	defer d.mu.Unlock()
	cutoff := time.Now().Add(-d.cfg.Window)
	filtered := d.samples[:0]
	for _, s := range d.samples {
		if s.recordedAt.After(cutoff) {
			filtered = append(filtered, s)
		}
	}
	d.samples = append(filtered, peerSample{ttl: ttl, recordedAt: time.Now()})
}

// Check returns a Critical alert if tokenTTL is below CriticalFactor of the
// peer median, or nil if the token appears healthy relative to its peers.
func (d *IsolationDetector) Check(tokenID string, tokenTTL time.Duration) *alert.Alert {
	d.mu.Lock()
	defer d.mu.Unlock()
	if len(d.samples) < d.cfg.MinPeers {
		return nil
	}
	median := medianTTL(d.samples)
	threshold := time.Duration(float64(median) * d.cfg.CriticalFactor)
	if tokenTTL >= threshold {
		return nil
	}
	return &alert.Alert{
		LeaseID:  tokenID,
		Level:    alert.Critical,
		Message:  fmt.Sprintf("token TTL %s is isolated (peer median %s, threshold %s)", tokenTTL, median, threshold),
	}
}

func medianTTL(samples []peerSample) time.Duration {
	vals := make([]time.Duration, len(samples))
	for i, s := range samples {
		vals[i] = s.ttl
	}
	// simple insertion sort for small N
	for i := 1; i < len(vals); i++ {
		for j := i; j > 0 && vals[j] < vals[j-1]; j-- {
			vals[j], vals[j-1] = vals[j-1], vals[j]
		}
	}
	mid := len(vals) / 2
	if len(vals)%2 == 0 {
		return (vals[mid-1] + vals[mid]) / 2
	}
	return vals[mid]
}
