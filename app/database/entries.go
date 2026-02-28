package database

import (
	"fmt"
	"sync"

	"github.com/gofrs/uuid/v5"
	"github.com/icza/dyno"
)

type Entries struct {
	mu       sync.RWMutex
	database map[string]interface{}
}

func NewEntry() Entries {
	return Entries{database: make(map[string]interface{})}
}

func (e *Entries) Create(value interface{}) interface{} {
	e.mu.Lock()
	defer e.mu.Unlock()
	key := generateEntryKey()
	if err := dyno.Set(value, key, "_id"); err != nil {
		fmt.Printf("Failed to set _id: %v\n", err)
	}
	e.database[key] = value
	return e.database[key]
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

func (e *Entries) Update(key string, value interface{}) interface{} {
	e.mu.Lock()
	defer e.mu.Unlock()
	if err := dyno.Set(value, key, "_id"); err != nil {
		fmt.Printf("Failed to set _id: %v\n", err)
	}
	e.database[key] = value
	return e.database[key]
}

func (e *Entries) Del(key string) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.database, key)
	_, prs := e.database[key]
	return !prs
}

func generateEntryKey() string {
	u, err := uuid.NewV7()
	if err != nil {
		fmt.Printf("failed to generate uuid v7: %v\n", err)
		return uuid.Must(uuid.NewV7()).String()
	}
	return u.String()
}
