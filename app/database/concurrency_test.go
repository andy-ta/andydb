package database

import (
	"fmt"
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
	errCh := make(chan error, 1)
	for i := 0; i < workers; i++ {
		go func(worker int) {
			defer wg.Done()
			for j := 0; j < perWorker; j++ {
				if _, err := entry.Create(map[string]interface{}{"worker": worker, "sequence": j}); err != nil {
					select {
					case errCh <- err:
					default:
					}
					return
				}
			}
		}(i)
	}
	wg.Wait()
	if err := drainErr(errCh); err != nil {
		t.Fatalf("unexpected create error: %v", err)
	}

	got := len(entry.ReadAll())
	want := workers * perWorker
	if got != want {
		t.Fatalf("expected %d entries, got %d", want, got)
	}

	results := entry.ReadAll()
	ids := make(map[string]struct{}, len(results))
	seenPairs := make(map[string]struct{}, len(results))
	for _, item := range results {
		record, ok := item.(map[string]interface{})
		if !ok {
			t.Fatalf("expected entry to be a map, got %T", item)
		}
		id, ok := record["_id"].(string)
		if !ok || id == "" {
			t.Fatalf("expected non-empty _id string, got %#v", record["_id"])
		}
		if _, exists := ids[id]; exists {
			t.Fatalf("duplicate _id detected: %q", id)
		}
		ids[id] = struct{}{}

		worker, ok := record["worker"].(float64)
		if !ok {
			t.Fatalf("expected numeric worker field, got %#v", record["worker"])
		}
		sequence, ok := record["sequence"].(float64)
		if !ok {
			t.Fatalf("expected numeric sequence field, got %#v", record["sequence"])
		}
		key := fmt.Sprintf("%d:%d", int(worker), int(sequence))
		if _, exists := seenPairs[key]; exists {
			t.Fatalf("duplicate worker/sequence pair detected: %s", key)
		}
		seenPairs[key] = struct{}{}
	}
	if len(ids) != want {
		t.Fatalf("expected %d unique ids, got %d", want, len(ids))
	}
	if len(seenPairs) != want {
		t.Fatalf("expected %d unique worker/sequence pairs, got %d", want, len(seenPairs))
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
	errCh := make(chan error, 1)
	for i := 0; i < workers; i++ {
		go func(worker int) {
			defer wg.Done()
			for j := 0; j < perWorker; j++ {
				if _, err := resource.Create(map[string]interface{}{"worker": worker, "sequence": j}); err != nil {
					select {
					case errCh <- err:
					default:
					}
					return
				}
				resource.ReadAll()
			}
		}(i)
	}
	wg.Wait()
	if err := drainErr(errCh); err != nil {
		t.Fatalf("unexpected create error: %v", err)
	}

	total := len(resource.ReadAll())
	want := workers * perWorker
	if total != want {
		t.Fatalf("expected %d records, got %d", want, total)
	}

	results := resource.ReadAll()
	seenPairs := make(map[string]struct{}, len(results))
	for _, item := range results {
		record, ok := item.(map[string]interface{})
		if !ok {
			t.Fatalf("expected entry to be a map, got %T", item)
		}
		worker, ok := record["worker"].(float64)
		if !ok {
			t.Fatalf("expected numeric worker field, got %#v", record["worker"])
		}
		sequence, ok := record["sequence"].(float64)
		if !ok {
			t.Fatalf("expected numeric sequence field, got %#v", record["sequence"])
		}
		key := fmt.Sprintf("%d:%d", int(worker), int(sequence))
		if _, exists := seenPairs[key]; exists {
			t.Fatalf("duplicate worker/sequence pair detected: %s", key)
		}
		seenPairs[key] = struct{}{}
	}
	if len(seenPairs) != want {
		t.Fatalf("expected %d unique worker/sequence pairs, got %d", want, len(seenPairs))
	}
}

func drainErr(errCh <-chan error) error {
	select {
	case err := <-errCh:
		return err
	default:
		return nil
	}
}
