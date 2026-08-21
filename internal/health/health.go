package health

import (
	"context"
	"sync"
	"time"
)

type Check func(context.Context) error
type Status struct {
	Name     string
	OK       bool
	Error    string
	Duration time.Duration
}
type Registry struct {
	mu     sync.RWMutex
	checks map[string]Check
	order  []string
}

func New() *Registry { return &Registry{checks: map[string]Check{}, order: make([]string, 0)} }
func (r *Registry) Register(name string, fn Check) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.checks[name]; !exists {
		r.order = append(r.order, name)
	}
	r.checks[name] = fn
}
func (r *Registry) Run(ctx context.Context) []Status {
	r.mu.RLock()
	type namedCheck struct {
		name string
		fn   Check
	}
	checks := make([]namedCheck, 0, len(r.order))
	for _, name := range r.order {
		checks = append(checks, namedCheck{name: name, fn: r.checks[name]})
	}
	r.mu.RUnlock()
	out := make([]Status, 0, len(checks))
	var pending sync.WaitGroup
	for _, check := range checks {
		check := check
		pending.Add(1)
		go func() {
			defer pending.Done()
			start := time.Now()
			err := check.fn(ctx)
			status := Status{Name: check.name, OK: err == nil, Duration: time.Since(start)}
			if err != nil {
				status.Error = err.Error()
			}
			out = append(out, status)
		}()
	}
	pending.Wait()
	return out
}
