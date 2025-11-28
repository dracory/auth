package utils

import (
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
)

type testRecord struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func newTestDriver(t *testing.T) *Driver {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "scribbledb")
	d, err := New(dir, nil)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	return d
}

func TestNew_CreatesDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "data")
	d, err := New(dir, nil)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if d.dir != filepath.Clean(dir) {
		t.Fatalf("driver dir = %q, want %q", d.dir, filepath.Clean(dir))
	}
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("expected directory %q to exist, got error %v", dir, err)
	}
}

func TestWriteAndRead_RoundTrip(t *testing.T) {
	d := newTestDriver(t)
	original := testRecord{ID: 1, Name: "Alice"}
	if err := d.Write("users", "alice", original); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	var got testRecord
	if err := d.Read("users", "alice", &got); err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if !reflect.DeepEqual(original, got) {
		t.Fatalf("round-trip mismatch: got %+v, want %+v", got, original)
	}
}

func TestWrite_ErrorsOnMissingCollectionOrResource(t *testing.T) {
	d := newTestDriver(t)
	if err := d.Write("", "res", struct{}{}); err != ErrMissingCollection {
		t.Fatalf("Write() error = %v, want %v", err, ErrMissingCollection)
	}
	if err := d.Write("col", "", struct{}{}); err != ErrMissingResource {
		t.Fatalf("Write() error = %v, want %v", err, ErrMissingResource)
	}
}

func TestRead_ErrorsOnMissingCollectionOrResource(t *testing.T) {
	d := newTestDriver(t)
	var v struct{}
	if err := d.Read("", "res", &v); err != ErrMissingCollection {
		t.Fatalf("Read() error = %v, want %v", err, ErrMissingCollection)
	}
	if err := d.Read("col", "", &v); err != ErrMissingResource {
		t.Fatalf("Read() error = %v, want %v", err, ErrMissingResource)
	}
}

func TestReadAll_ReturnsAllRecords(t *testing.T) {
	d := newTestDriver(t)
	if _, err := d.ReadAll(""); err != ErrMissingCollection {
		t.Fatalf("ReadAll() error = %v, want %v", err, ErrMissingCollection)
	}
	rec1 := testRecord{ID: 1, Name: "Alice"}
	rec2 := testRecord{ID: 2, Name: "Bob"}
	if err := d.Write("users", "alice", rec1); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if err := d.Write("users", "bob", rec2); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	records, err := d.ReadAll("users")
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("ReadAll() len = %d, want 2", len(records))
	}
}

func TestDelete_RemovesFileAndCollection(t *testing.T) {
	d := newTestDriver(t)
	rec := testRecord{ID: 1, Name: "Alice"}
	if err := d.Write("users", "alice", rec); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if err := d.Delete("users", "alice"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(d.dir, "users", "alice.json")); !os.IsNotExist(err) {
		t.Fatalf("expected alice.json to be removed, got error %v", err)
	}
	if err := d.Write("users", "alice", rec); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if err := d.Write("users", "bob", rec); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if err := d.Delete("users", ""); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(d.dir, "users")); !os.IsNotExist(err) {
		t.Fatalf("expected users directory to be removed, got error %v", err)
	}
}

func TestGetOrCreateMutex_ReusesMutexPerCollection(t *testing.T) {
	d := &Driver{mutexes: make(map[string]*sync.Mutex)}
	m1 := d.getOrCreateMutex("users")
	m2 := d.getOrCreateMutex("users")
	if m1 != m2 {
		t.Fatalf("expected same mutex instance for collection 'users'")
	}
	m3 := d.getOrCreateMutex("posts")
	if m1 == m3 {
		t.Fatalf("expected different mutex instances for different collections")
	}
}
