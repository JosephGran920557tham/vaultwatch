package tokenwatch

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type healthResponse struct {
	Status    string    `json:"status"`
	Healthy   int       `json:"healthy"`
	Warning   int       `json:"warning"`
	Critical  int       `json:"critical"`
	Total     int       `json:"total"`
	CheckedAt time.Time `json:"checked_at"`
}

// HealthHandler returns an http.Handler that exposes token health over HTTP.
// It responds with 200 when all tokens are healthy and 503 when any are critical.
func HealthHandler(hc *HealthChecker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		status, err := hc.Check(ctx)
		if err != nil {
			http.Error(w, "health check error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		statusStr := "healthy"
		httpCode := http.StatusOK
		if !status.IsHealthy() {
			statusStr = "unhealthy"
			httpCode = http.StatusServiceUnavailable
		}

		resp := healthResponse{
			Status:    statusStr,
			Healthy:   status.Healthy,
			Warning:   status.Warning,
			Critical:  status.Critical,
			Total:     status.Total,
			CheckedAt: status.CheckedAt,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(httpCode)
		_ = json.NewEncoder(w).Encode(resp)
	})
}
