package tokenwatch

import (
	"fmt"
	"sync"
	"time"

	"github.com/vaultwatch/internal/alert"
)

// DefaultBudgetConfig returns sensible defaults for the budget detector.
func DefaultBudgetConfig() BudgetConfig {
	return BudgetConfig{
		MaxRenewals:  10,
		Window:       time.Hour,
		WarningRatio: 0.75,
	}
}

// BudgetConfig controls how many renewals are permitted per token per window.
type BudgetConfig struct {
	MaxRenewals  int
	Window       time.Duration
	WarningRatio float64 // fraction of MaxRenewals that triggers a warning
}

type budgetEntry struct {
	count     int
	windowEnd time.Time
}

// BudgetDetector tracks renewal counts per token and alerts when the budget
// is approaching or has been exhausted within a rolling window.
type BudgetDetector struct {
	cfg     BudgetConfig
	mu      sync.Mutex
	entries map[string]*budgetEntry
}

// NewBudgetDetector creates a BudgetDetector, applying defaults for zero values.
func NewBudgetDetector(cfg BudgetConfig) *BudgetDetector {
	def := DefaultBudgetConfig()
	if cfg.MaxRenewals <= 0 {
		cfg.MaxRenewals = def.MaxRenewals
	}
	if cfg.Window <= 0 {
		cfg.Window = def.Window
	}
	if cfg.WarningRatio <= 0 || cfg.WarningRatio > 1 {
		cfg.WarningRatio = def.WarningRatio
	}
	return &BudgetDetector{cfg: cfg, entries: make(map[string]*budgetEntry)}
}

// Record increments the renewal counter for the given token.
func (b *BudgetDetector) Record(tokenID string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	now := time.Now()
	e, ok := b.entries[tokenID]
	if !ok || now.After(e.windowEnd) {
		b.entries[tokenID] = &budgetEntry{count: 1, windowEnd: now.Add(b.cfg.Window)}
		return
	}
	e.count++
}

// Check returns an alert if the token's renewal budget is exhausted or near
// exhaustion, or nil if usage is within acceptable bounds.
func (b *BudgetDetector) Check(tokenID string) *alert.Alert {
	b.mu.Lock()
	defer b.mu.Unlock()
	e, ok := b.entries[tokenID]
	if !ok || time.Now().After(e.windowEnd) {
		return nil
	}
	ratio := float64(e.count) / float64(b.cfg.MaxRenewals)
	switch {
	case e.count >= b.cfg.MaxRenewals:
		return &alert.Alert{
			LeaseID: tokenID,
			Level:   alert.LevelCritical,
			Message: fmt.Sprintf("renewal budget exhausted: %d/%d renewals in window", e.count, b.cfg.MaxRenewals),
		}
	case ratio >= b.cfg.WarningRatio:
		return &alert.Alert{
			LeaseID: tokenID,
			Level:   alert.LevelWarning,
			Message: fmt.Sprintf("renewal budget nearly exhausted: %d/%d renewals in window", e.count, b.cfg.MaxRenewals),
		}
	default:
		return nil
	}
}
