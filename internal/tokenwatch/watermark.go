package tokenwatch

import (
	"sync"
	"time"
)

// DefaultWatermarkConfig returns sensible defaults for WatermarkDetector.
func DefaultWatermarkConfig() WatermarkConfig {
	return WatermarkConfig{
		LowWatermark:  300 * time.Second,
		HighWatermark: 3600 * time.Second,
	}
}

// WatermarkConfig configures low/high TTL watermarks.
type WatermarkConfig struct {
	LowWatermark  time.Duration
	HighWatermark time.Duration
}

// WatermarkDetector tracks per-token TTL high-water marks and alerts when
// the current TTL drops below the low watermark.
type WatermarkDetector struct {
	cfg  WatermarkConfig
	mu   sync.Mutex
	peaks map[string]time.Duration
}

// NewWatermarkDetector creates a WatermarkDetector. Zero config values use defaults.
func NewWatermarkDetector(cfg WatermarkConfig) *WatermarkDetector {
	d := DefaultWatermarkConfig()
	if cfg.LowWatermark > 0 {
		d.LowWatermark = cfg.LowWatermark
	}
	if cfg.HighWatermark > 0 {
		d.HighWatermark = cfg.HighWatermark
	}
	return &WatermarkDetector{
		cfg:  d,
		peaks: make(map[string]time.Duration),
	}
}

// Record updates the observed peak TTL for a token.
func (w *WatermarkDetector) Record(tokenID string, ttl time.Duration) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if ttl > w.peaks[tokenID] {
		w.peaks[tokenID] = ttl
	}
}

// Check returns a warning alert if ttl is below the low watermark and a peak
// has been established above the high watermark.
func (w *WatermarkDetector) Check(tokenID string, ttl time.Duration) *Alert {
	w.mu.Lock()
	peak := w.peaks[tokenID]
	w.mu.Unlock()

	if peak < w.cfg.HighWatermark {
		return nil
	}
	if ttl >= w.cfg.LowWatermark {
		return nil
	}
	return &Alert{
		LeaseID:  tokenID,
		Level:    LevelWarning,
		Message:  "token TTL dropped below low watermark",
		ExpireAt: time.Now().Add(ttl),
	}
}
