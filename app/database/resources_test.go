package database

import (
	"sync"
	"testing"
)

func TestGetOrCreateNew(t *testing.T) {
	db := NewDatabase()
	resource, err := db.GetOrCreate("contacts")
	if err != nil {
		t.Fatalf("GetOrCreate failed: %v", err)
	}
	if resource == nil {
		t.Fatal("expected resource")
	}
	if !db.Exists("contacts") {
		t.Fatal("expected contacts to exist")
	}
	if db.Get("contacts") != resource {
		t.Fatal("Get should return the same resource")
	}
}

func TestGetOrCreateExisting(t *testing.T) {
	db := NewDatabase()
	if err := db.NewResource("contacts"); err != nil {
		t.Fatalf("NewResource failed: %v", err)
	}
	first := db.Get("contacts")
	second, err := db.GetOrCreate("contacts")
	if err != nil {
		t.Fatalf("GetOrCreate failed: %v", err)
	}
	if first != second {
		t.Fatal("GetOrCreate must return the existing collection")
	}
}

func TestRemoveMissing(t *testing.T) {
	db := NewDatabase()
	if db.Remove("missing") {
		t.Fatal("expected Remove of missing resource to return false")
	}
}

func TestRemoveExisting(t *testing.T) {
	db := NewDatabase()
	if err := db.NewResource("contacts"); err != nil {
		t.Fatalf("NewResource failed: %v", err)
	}
	if !db.Remove("contacts") {
		t.Fatal("expected Remove of existing resource to return true")
	}
	if db.Exists("contacts") {
		t.Fatal("expected resource to be gone")
	}
	if db.Remove("contacts") {
		t.Fatal("expected second Remove to return false")
	}
}

func TestGetOrCreateConcurrentSameName(t *testing.T) {
	db := NewDatabase()
	const workers = 20
	got := make([]*Entries, workers)
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func(i int) {
			defer wg.Done()
			resource, err := db.GetOrCreate("contacts")
			if err != nil {
				t.Errorf("GetOrCreate failed: %v", err)
				return
			}
			got[i] = resource
		}(i)
	}
	wg.Wait()

	first := got[0]
	if first == nil {
		t.Fatal("expected a resource")
	}
	for i, resource := range got {
		if resource != first {
			t.Fatalf("worker %d got a different collection pointer", i)
		}
	}
}
