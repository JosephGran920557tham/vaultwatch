package tokenwatch

import "context"

// WatermarkScanner scans all registered tokens through a WatermarkDetector.
type WatermarkScanner struct {
	registry *Registry
	detector *WatermarkDetector
	lookup   func(ctx context.Context, tokenID string) (*TokenInfo, error)
}

// NewWatermarkScanner constructs a WatermarkScanner. Panics if any argument is nil.
func NewWatermarkScanner(
	reg *Registry,
	det *WatermarkDetector,
	lookup func(ctx context.Context, tokenID string) (*TokenInfo, error),
) *WatermarkScanner {
	if reg == nil {
		panic("watermark scanner: registry is nil")
	}
	if det == nil {
		panic("watermark scanner: detector is nil")
	}
	if lookup == nil {
		panic("watermark scanner: lookup is nil")
	}
	return &WatermarkScanner{registry: reg, detector: det, lookup: lookup}
}

// Scan checks every registered token and returns all watermark alerts.
func (s *WatermarkScanner) Scan(ctx context.Context) []Alert {
	var alerts []Alert
	for _, id := range s.registry.List() {
		info, err := s.lookup(ctx, id)
		if err != nil || info == nil {
			continue
		}
		s.detector.Record(id, info.TTL)
		if a := s.detector.Check(id, info.TTL); a != nil {
			alerts = append(alerts, *a)
		}
	}
	return alerts
}
