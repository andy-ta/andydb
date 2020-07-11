package database

import (
	"fmt"
)

type Resources struct {
	database map[string]entries
}

func NewDatabase() Resources {
	return Resources{make(map[string]entries)}
}

func (r *Resources) NewResource(name string) (Resources, error) {
	if r.Exists(name) {
		r.database[name] = NewEntry()
		return *r, _
	} else {
		err := fmt.Errorf("resource %q already exists", name)
		return _, err
	}
}

func (r *Resources) Get(name string) entries {
	return r.database[name]
}

func (r *Resources) Exists(name string) bool {
	_, prs := r.database[name]
	return prs
}

func (r *Resources) Remove(name string) bool {
	delete(r.database, name)
	return r.Exists(name)
}
