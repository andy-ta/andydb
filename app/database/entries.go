package database

import (
	"errors"
	"fmt"
	"sync"

	"github.com/gofrs/uuid/v5"
	"github.com/icza/dyno"
)

var ErrNotFound = errors.New("entry not found")

type Entries struct {
	mu       sync.RWMutex
	database map[string]interface{}
}

func NewEntry() Entries {
	return Entries{database: make(map[string]interface{})}
}

func (e *Entries) Create(value interface{}) (interface{}, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	stripClientID(value)
	key, err := generateEntryKey()
	if err != nil {
		return nil, err
	}
	if err := dyno.Set(value, key, "_id"); err != nil {
		return nil, fmt.Errorf("failed to set _id: %w", err)
	}
	e.database[key] = value
	return e.database[key], nil
}

func (e *Entries) Read(key string) interface{} {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.database[key]
}

func (e *Entries) ReadAll() []interface{} {
	e.mu.RLock()
	defer e.mu.RUnlock()
	values := make([]interface{}, 0, len(e.database))
	for _, val := range e.database {
		values = append(values, val)
	}
	return values
}

func (e *Entries) Update(key string, value interface{}) (interface{}, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if _, exists := e.database[key]; !exists {
		return nil, ErrNotFound
	}
	stripClientID(value)
	if err := dyno.Set(value, key, "_id"); err != nil {
		return nil, fmt.Errorf("failed to set _id: %w", err)
	}
	e.database[key] = value
	return e.database[key], nil
}

func (e *Entries) Del(key string) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	_, existed := e.database[key]
	if !existed {
		return false
	}
	delete(e.database, key)
	return true
}

func stripClientID(value interface{}) {
	if entry, ok := value.(map[string]interface{}); ok {
		delete(entry, "_id")
	}
}

func generateEntryKey() (string, error) {
	u, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("failed to generate uuid v7: %w", err)
	}
	return u.String(), nil
}
