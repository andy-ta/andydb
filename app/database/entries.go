package database

import (
	b64 "encoding/base64"
	"fmt"
	"github.com/icza/dyno"
	"strconv"
	"time"
)

type Entries struct {
	database map[string]interface{}
}

func NewEntry() Entries {
	return Entries{make(map[string]interface{})}
}

func (e *Entries) Create(value interface{}) interface{} {
	// Generate key
	now := time.Now().UnixNano()
	key := b64.StdEncoding.EncodeToString([]byte(strconv.FormatInt(now, 10)))[0:15]
	// Set ID
	if err := dyno.Set(value, key, "_id"); err != nil {
		fmt.Printf("Failed to set password: %v\n", err)
	}
	e.database[key] = value
	return e.database[key]
}

func (e *Entries) Read(key string) interface{} {
	return e.database[key]
}

func (e *Entries) Update(key string, value interface{}) interface{} {
	e.database[key] = value
	return e.database[key]
}

func (e *Entries) Del(key string) bool {
	delete(e.database, key)
	_, prs := e.database[key]
	return prs
}
