package database

import (
	b64 "encoding/base64"
	"fmt"
	"strconv"
	"sync"
	"time"

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
	// Generate key
	now := time.Now().UnixNano()
	key := b64.StdEncoding.EncodeToString([]byte(strconv.FormatInt(now, 10)))[0:15]
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
