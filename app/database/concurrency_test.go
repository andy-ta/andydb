package database

import (
	"strings"
	"sync"
	"testing"
)

func TestEntriesConcurrentCreate(t *testing.T) {
	entry := NewEntry()
	const workers = 10
	const perWorker = 50

	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func(worker int) {
			defer wg.Done()
			for j := 0; j < perWorker; j++ {
				entry.Create(map[string]interface{}{"worker": worker, "sequence": j})
			}
		}(i)
	}
	wg.Wait()

	got := len(entry.ReadAll())
	want := workers * perWorker
	if got != want {
		t.Fatalf("expected %d entries, got %d", want, got)
	}
}

func TestResourcesConcurrentNewResource(t *testing.T) {
	db := NewDatabase()
	names := []string{"alpha", "beta", "gamma"}

	var wg sync.WaitGroup
	const goroutines = 15
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(iter int) {
			defer wg.Done()
			name := names[iter%len(names)]
			if err := db.NewResource(name); err != nil && !strings.Contains(err.Error(), "already exists") {
				t.Errorf("unexpected error: %v", err)
			}
		}(i)
	}
	wg.Wait()

	for _, name := range names {
		if !db.Exists(name) {
			t.Fatalf("resource %q should exist", name)
		}
		if db.Get(name) == nil {
			t.Fatalf("resource %q should be retrievable", name)
		}
	}
}

func TestResourcesConcurrentWrites(t *testing.T) {
	db := NewDatabase()
	const name = "records"
	if err := db.NewResource(name); err != nil {
		t.Fatalf("failed to create resource: %v", err)
	}
	resource := db.Get(name)
	if resource == nil {
		t.Fatalf("resource %q unexpectedly nil", name)
	}

	const workers = 6
	const perWorker = 40
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func(worker int) {
			defer wg.Done()
			for j := 0; j < perWorker; j++ {
				resource.Create(map[string]interface{}{"worker": worker, "sequence": j})
				resource.ReadAll()
			}
		}(i)
	}
	wg.Wait()

	total := len(resource.ReadAll())
	want := workers * perWorker
	if total != want {
		t.Fatalf("expected %d records, got %d", want, total)
	}
}
