package tokenwatch

import (
	"errors"
	"log"
	"time"
)

// WindowGuardConfig configures the WindowGuard.
type WindowGuardConfig struct {
	// WindowSize is the sliding window duration.
	WindowSize time.Duration
	// MaxAlerts is the maximum alerts allowed per token within the window.
	MaxAlerts int
}

// DefaultWindowGuardConfig returns a WindowGuardConfig with sensible defaults.
func DefaultWindowGuardConfig() WindowGuardConfig {
	return WindowGuardConfig{
		WindowSize: 10 * time.Minute,
		MaxAlerts:  5,
	}
}

// WindowGuard wraps a Window to gate alert dispatch per token using a sliding
// window strategy. It logs suppressed alerts at debug level.
type WindowGuard struct {
	win    *Window
	logger *log.Logger
}

// NewWindowGuard creates a WindowGuard from the given config.
func NewWindowGuard(cfg WindowGuardConfig, logger *log.Logger) (*WindowGuard, error) {
	if cfg.WindowSize <= 0 {
		return nil, errors.New("window_guard: WindowSize must be positive")
	}
	if cfg.MaxAlerts <= 0 {
		return nil, errors.New("window_guard: MaxAlerts must be positive")
	}
	win, err := NewWindow(WindowConfig{
		Size:      cfg.WindowSize,
		MaxEvents: cfg.MaxAlerts,
	})
	if err != nil {
		return nil, err
	}
	if logger == nil {
		logger = log.Default()
	}
	return &WindowGuard{win: win, logger: logger}, nil
}

// Allow returns true if the token has not exceeded its alert quota within the
// current sliding window. Denied tokens are logged.
func (g *WindowGuard) Allow(tokenID string) bool {
	if g.win.Allow(tokenID) {
		return true
	}
	g.logger.Printf("window_guard: suppressed alert for token %q (window limit reached)", tokenID)
	return false
}

// Count returns the current event count for tokenID within the window.
func (g *WindowGuard) Count(tokenID string) int {
	return g.win.Count(tokenID)
}

// Reset clears the window state for tokenID.
func (g *WindowGuard) Reset(tokenID string) {
	g.win.Reset(tokenID)
}
