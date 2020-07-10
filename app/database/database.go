package database

import (
	b64 "encoding/base64"
	"github.com/bennyscetbun/jsongo"
	"strconv"
	"time"
)

var Database = make(map[string]jsongo.Node)

func create(value jsongo.Node) jsongo.Node {
	now := time.Now().UnixNano()
	key := b64.StdEncoding.EncodeToString([]byte(strconv.FormatInt(now, 10)))[0:15]
	Database[key] = value
	return Database[key]
}

func read(key string) jsongo.Node {
	return Database[key]
}

func update(key string, value jsongo.Node) jsongo.Node {
	Database[key] = value
	return Database[key]
}

func del(key string) bool {
	delete(Database, key)
	_, prs := Database[key]
	return prs
}
