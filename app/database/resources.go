package database

import (
	"fmt"
)

type Resources struct {
	database map[string]*Entries
}

func NewDatabase() Resources {
	return Resources{make(map[string]*Entries)}
}

func (r *Resources) NewResource(name string) (*Resources, error) {
	if !r.Exists(name) {
		entry := NewEntry()
		r.database[name] = &entry
		return r, nil
	} else {
		err := fmt.Errorf("resource %q already exists", name)
		return nil, err
	}
}

func (r *Resources) Get(name string) *Entries {
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
