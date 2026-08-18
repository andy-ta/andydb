package database

import (
	"errors"
	"strings"
	"testing"

	"github.com/gofrs/uuid/v5"
)

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

func TestEntriesUpdateMissingKey(t *testing.T) {
	entries := NewEntry()
	got, err := entries.Update("missing", map[string]interface{}{"name": "andy"})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil result, got %#v", got)
	}
	if len(entries.ReadAll()) != 0 {
		t.Fatal("missing-key Update must not create a row")
	}
}

func TestEntriesCreateIgnoresClientID(t *testing.T) {
	entries := NewEntry()
	created, err := entries.Create(map[string]interface{}{"name": "andy", "_id": "client-chosen"})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	record := mustEntryMap(t, created)
	id := mustServerID(t, record)
	if id == "client-chosen" {
		t.Fatal("Create honored a client-supplied _id")
	}
	if entries.Read("client-chosen") != nil {
		t.Fatal("client-chosen must not be a store key")
	}
	if entries.Read(id) == nil {
		t.Fatal("expected the minted id to be the store key")
	}
}

func TestEntriesUpdatePreservesServerID(t *testing.T) {
	entries := NewEntry()
	created, err := entries.Create(map[string]interface{}{"name": "andy"})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	id := mustServerID(t, mustEntryMap(t, created))

	updated, err := entries.Update(id, map[string]interface{}{"name": "betty", "_id": "other"})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	record := mustEntryMap(t, updated)
	if record["_id"] != id {
		t.Fatalf("expected _id %q, got %#v", id, record["_id"])
	}
	if record["name"] != "betty" {
		t.Fatalf("expected updated name, got %#v", record["name"])
	}
	if entries.Read("other") != nil {
		t.Fatal("Update must not create a row at the client _id")
	}
}

func TestEntriesCreateRejectsNonObject(t *testing.T) {
	entries := NewEntry()
	got, err := entries.Create("not-an-object")
	if err == nil || !strings.Contains(err.Error(), "failed to set _id") {
		t.Fatalf("expected set _id error, got %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil result, got %#v", got)
	}
}

func TestEntriesUpdateRejectsNonObject(t *testing.T) {
	entries := NewEntry()
	created, err := entries.Create(map[string]interface{}{"name": "andy"})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	id := mustServerID(t, mustEntryMap(t, created))
	got, err := entries.Update(id, "not-an-object")
	if err == nil || !strings.Contains(err.Error(), "failed to set _id") {
		t.Fatalf("expected set _id error, got %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil result, got %#v", got)
	}
}

func TestOverrideSetField(t *testing.T) {
	restore := OverrideSetField(func(interface{}, interface{}, ...interface{}) error {
		return errors.New("forced")
	})
	defer restore()

	entries := NewEntry()
	if _, err := entries.Create(map[string]interface{}{"name": "andy"}); err == nil || !strings.Contains(err.Error(), "forced") {
		t.Fatalf("expected forced set error, got %v", err)
	}
}

func TestGenerateEntryKeyError(t *testing.T) {
	restore := OverrideNewUUIDV7(func() (uuid.UUID, error) {
		return uuid.Nil, errors.New("rand failed")
	})
	defer restore()

	if _, err := generateEntryKey(); err == nil || !strings.Contains(err.Error(), "failed to generate uuid v7") {
		t.Fatalf("expected uuid error, got %v", err)
	}
	entries := NewEntry()
	if _, err := entries.Create(map[string]interface{}{"name": "andy"}); err == nil || !strings.Contains(err.Error(), "failed to generate uuid v7") {
		t.Fatalf("expected create uuid error, got %v", err)
	}
}

func TestNewResourceAlreadyExists(t *testing.T) {
	db := NewDatabase()
	if err := db.NewResource("contacts"); err != nil {
		t.Fatalf("NewResource failed: %v", err)
	}
	if err := db.NewResource("contacts"); err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("expected already exists, got %v", err)
	}
}

func mustEntryMap(t *testing.T, value interface{}) map[string]interface{} {
	t.Helper()
	record, ok := value.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map entry, got %T", value)
	}
	return record
}

func mustServerID(t *testing.T, record map[string]interface{}) string {
	t.Helper()
	id, ok := record["_id"].(string)
	if !ok || id == "" {
		t.Fatalf("expected non-empty _id, got %#v", record["_id"])
	}
	if _, err := uuid.FromString(id); err != nil {
		t.Fatalf("expected UUID v7 _id, got %q: %v", id, err)
	}
	return id
}
