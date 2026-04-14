package tokenwatch

import (
	"fmt"
	"sync"
)

// Registry manages a dynamic set of token accessors to watch.
type Registry struct {
	mu        sync.RWMutex
	accessors map[string]struct{}
}

// NewRegistry returns an empty Registry.
func NewRegistry() *Registry {
	return &Registry{
		accessors: make(map[string]struct{}),
	}
}

// Add registers an accessor for monitoring. Returns an error if already present.
func (r *Registry) Add(accessor string) error {
	if accessor == "" {
		return fmt.Errorf("tokenwatch: accessor must not be empty")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.accessors[accessor]; exists {
		return fmt.Errorf("tokenwatch: accessor %q already registered", accessor)
	}
	r.accessors[accessor] = struct{}{}
	return nil
}

// Remove unregisters an accessor. Returns an error if not found.
func (r *Registry) Remove(accessor string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.accessors[accessor]; !exists {
		return fmt.Errorf("tokenwatch: accessor %q not found", accessor)
	}
	delete(r.accessors, accessor)
	return nil
}

// List returns a snapshot of all registered accessors.
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, 0, len(r.accessors))
	for acc := range r.accessors {
		out = append(out, acc)
	}
	return out
}

// Len returns the number of registered accessors.
func (r *Registry) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.accessors)
}
