package service

import (
	"context"
	"errors"
	"sort"
	"sync"
)

type HealthStatus struct {
	Status string `json:"status"`
}
type Service interface {
	Provider() string
	Name() string
	Health(context.Context) HealthStatus
	Reset(context.Context) error
}
type Registry struct {
	mu       sync.RWMutex
	services map[string]Service
}

func NewRegistry() *Registry           { return &Registry{services: map[string]Service{}} }
func key(provider, name string) string { return provider + "." + name }
func (r *Registry) Register(s Service) error {
	if s == nil || s.Provider() == "" || s.Name() == "" {
		return errors.New("service identity is required")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	id := key(s.Provider(), s.Name())
	if _, ok := r.services[id]; ok {
		return errors.New("duplicate service " + id)
	}
	r.services[id] = s
	return nil
}
func (r *Registry) Find(provider, name string) (Service, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.services[key(provider, name)]
	return s, ok
}
func (r *Registry) Services() []Service {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ids := make([]string, 0, len(r.services))
	for id := range r.services {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	out := make([]Service, 0, len(ids))
	for _, id := range ids {
		out = append(out, r.services[id])
	}
	return out
}
func (r *Registry) Health(ctx context.Context) map[string]HealthStatus {
	out := map[string]HealthStatus{}
	for _, s := range r.Services() {
		out[key(s.Provider(), s.Name())] = s.Health(ctx)
	}
	return out
}
func (r *Registry) Reset(ctx context.Context) error {
	for _, s := range r.Services() {
		if err := s.Reset(ctx); err != nil {
			return err
		}
	}
	return nil
}
