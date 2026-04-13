// Package metrics provides lightweight in-process counters and gauges
// for tracking VaultWatch operational statistics.
package metrics

import (
	"sync"
	"sync/atomic"
)

// Counter is a monotonically increasing integer counter.
type Counter struct{ v uint64 }

// Inc increments the counter by 1.
func (c *Counter) Inc() { atomic.AddUint64(&c.v, 1) }

// Add increments the counter by n.
func (c *Counter) Add(n uint64) { atomic.AddUint64(&c.v, n) }

// Value returns the current counter value.
func (c *Counter) Value() uint64 { return atomic.LoadUint64(&c.v) }

// Gauge is a value that can go up or down.
type Gauge struct {
	mu sync.Mutex
	v  float64
}

// Set assigns the gauge to v.
func (g *Gauge) Set(v float64) {
	g.mu.Lock()
	g.v = v
	g.mu.Unlock()
}

// Value returns the current gauge value.
func (g *Gauge) Value() float64 {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.v
}

// Registry holds a named set of counters and gauges.
type Registry struct {
	mu       sync.RWMutex
	counters map[string]*Counter
	gauges   map[string]*Gauge
}

// NewRegistry creates an empty Registry.
func NewRegistry() *Registry {
	return &Registry{
		counters: make(map[string]*Counter),
		gauges:   make(map[string]*Gauge),
	}
}

// Counter returns (or creates) the named counter.
func (r *Registry) Counter(name string) *Counter {
	r.mu.Lock()
	defer r.mu.Unlock()
	if c, ok := r.counters[name]; ok {
		return c
	}
	c := &Counter{}
	r.counters[name] = c
	return c
}

// Gauge returns (or creates) the named gauge.
func (r *Registry) Gauge(name string) *Gauge {
	r.mu.Lock()
	defer r.mu.Unlock()
	if g, ok := r.gauges[name]; ok {
		return g
	}
	g := &Gauge{}
	r.gauges[name] = g
	return g
}

// Snapshot returns a point-in-time copy of all metric values.
func (r *Registry) Snapshot() map[string]float64 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[string]float64, len(r.counters)+len(r.gauges))
	for k, c := range r.counters {
		out[k] = float64(c.Value())
	}
	for k, g := range r.gauges {
		out[k] = g.Value()
	}
	return out
}
