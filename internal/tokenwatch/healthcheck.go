package tokenwatch

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// HealthStatus represents the overall health of monitored tokens.
type HealthStatus struct {
	Healthy   int
	Warning   int
	Critical  int
	Total     int
	CheckedAt time.Time
}

// IsHealthy returns true when no tokens are in a critical state.
func (h HealthStatus) IsHealthy() bool {
	return h.Critical == 0
}

// String returns a human-readable summary of the health status.
func (h HealthStatus) String() string {
	return fmt.Sprintf("tokens: total=%d healthy=%d warning=%d critical=%d",
		h.Total, h.Healthy, h.Warning, h.Critical)
}

// HealthChecker evaluates the current health of all registered tokens.
type HealthChecker struct {
	alerter   *Alerter
	classifer *ExpiryClassifier
	mu        sync.Mutex
	last      *HealthStatus
}

// NewHealthChecker creates a HealthChecker backed by the given Alerter.
func NewHealthChecker(a *Alerter, c *ExpiryClassifier) (*HealthChecker, error) {
	if a == nil {
		return nil, fmt.Errorf("tokenwatch: alerter must not be nil")
	}
	if c == nil {
		return nil, fmt.Errorf("tokenwatch: classifier must not be nil")
	}
	return &HealthChecker{alerter: a, classifer: c}, nil
}

// Check inspects all tokens and returns the current HealthStatus.
func (h *HealthChecker) Check(ctx context.Context) (HealthStatus, error) {
	alerts, err := h.alerter.CheckAll(ctx)
	if err != nil {
		return HealthStatus{}, fmt.Errorf("tokenwatch: health check failed: %w", err)
	}

	status := HealthStatus{CheckedAt: time.Now()}
	for _, a := range alerts {
		status.Total++
		switch a.Level {
		case "critical":
			status.Critical++
		case "warning":
			status.Warning++
		default:
			status.Healthy++
		}
	}

	h.mu.Lock()
	h.last = &status
	h.mu.Unlock()

	return status, nil
}

// Last returns the most recently computed HealthStatus, or nil if Check
// has not yet been called.
func (h *HealthChecker) Last() *HealthStatus {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.last
}
