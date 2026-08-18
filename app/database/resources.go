package database

import (
	"fmt"
	"sync"
)

type Resources struct {
	mu       sync.RWMutex
	database map[string]*Entries
}

func NewDatabase() *Resources {
	return &Resources{database: make(map[string]*Entries)}
}

func (r *Resources) NewResource(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.database[name]; exists {
		return fmt.Errorf("resource %q already exists", name)
	}
	entry := NewEntry()
	r.database[name] = &entry
	return nil
}

// GetOrCreate returns the named collection, creating it if needed.
// If a later disk backend is added, this should return an error so
// persistence failures can be handled by callers.
func (r *Resources) GetOrCreate(name string) *Entries {
	r.mu.Lock()
	defer r.mu.Unlock()
	if existing, ok := r.database[name]; ok {
		return existing
	}
	entry := NewEntry()
	r.database[name] = &entry
	return &entry
}

func (r *Resources) Get(name string) *Entries {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.database[name]
}

func (r *Resources) Exists(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, prs := r.database[name]
	return prs
}

func (r *Resources) Remove(name string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.database[name]; !exists {
		return false
	}
	delete(r.database, name)
	return true
}
