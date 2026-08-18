package database

import "testing"

func TestEntriesDelMissingKey(t *testing.T) {
	entries := NewEntry()
	if entries.Del("missing") {
		t.Fatal("expected Del of missing key to return false")
	}
}

func TestEntriesDelExistingKey(t *testing.T) {
	entries := NewEntry()
	created, err := entries.Create(map[string]interface{}{"name": "andy"})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	record, ok := created.(map[string]interface{})
	if !ok {
		t.Fatalf("expected created entry to be a map, got %T", created)
	}
	id, ok := record["_id"].(string)
	if !ok || id == "" {
		t.Fatalf("expected non-empty _id, got %#v", record["_id"])
	}

	if !entries.Del(id) {
		t.Fatal("expected Del of existing key to return true")
	}
	if entries.Read(id) != nil {
		t.Fatal("expected deleted entry to be gone")
	}
	if entries.Del(id) {
		t.Fatal("expected second Del of the same key to return false")
	}
}
