package database

import (
	b64 "encoding/base64"
	"github.com/bennyscetbun/jsongo"
	"strconv"
	"time"
)

type Entries struct {
	database map[string]jsongo.Node
}

func NewEntry() Entries {
	return Entries{make(map[string]jsongo.Node)}
}

func (e *Entries) Create(value jsongo.Node) jsongo.Node {
	now := time.Now().UnixNano()
	key := b64.StdEncoding.EncodeToString([]byte(strconv.FormatInt(now, 10)))[0:15]
	e.database[key] = value
	return e.database[key]
}

func (e *Entries) Read(key string) jsongo.Node {
	return e.database[key]
}

func (e *Entries) Update(key string, value jsongo.Node) jsongo.Node {
	e.database[key] = value
	return e.database[key]
}

func (e *Entries) Del(key string) bool {
	delete(e.database, key)
	_, prs := e.database[key]
	return prs
}
