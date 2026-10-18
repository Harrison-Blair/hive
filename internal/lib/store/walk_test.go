package store

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWalkRecordsFindsNestedAndSkipsOnlyTemporary(t *testing.T) {
	store := openStore(t, t.TempDir())
	if err := store.Write("projects", "hive", map[string]string{"id": "hive"}); err != nil {
		t.Fatal(err)
	}
	if err := store.Write("tasks", "12-spawn-worker", map[string]string{"id": "12-spawn-worker"}); err != nil {
		t.Fatal(err)
	}
	if err := store.WriteIn([]string{"inbox", "coordinator"}, "1", map[string]string{"body": "hello"}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Next("inbox:coordinator"); err != nil {
		t.Fatal(err)
	}
	stateDir := filepath.Join(store.home, ".hive", "state")
	if err := os.WriteFile(filepath.Join(stateDir, "counters", ".hidden.json.tmp-123"), []byte(`{"value":0}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stateDir, "counters", ".other.json"), []byte(`{"value":0}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stateDir, "tasks", "notes.txt"), []byte("not a record"), 0o644); err != nil {
		t.Fatal(err)
	}

	var found []StoredRecord
	if err := store.WalkRecords(context.Background(), func(record StoredRecord) error {
		found = append(found, record)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	got := make([]string, 0, len(found))
	for _, record := range found {
		parts := append(append([]string(nil), record.Collection...), record.Name)
		got = append(got, strings.Join(parts, "/"))
	}
	want := []string{
		"counters/.other.json",
		"counters/inbox:coordinator.json",
		"inbox/coordinator/1.json",
		"projects/hive.json",
		"tasks/12-spawn-worker.json",
	}
	if len(got) != len(want) {
		t.Fatalf("WalkRecords() paths = %q; want %q", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("WalkRecords() paths = %q; want %q", got, want)
		}
	}
	for _, record := range found {
		if record.Name == "lock" || strings.Contains(record.Name, ".tmp-") {
			t.Fatalf("WalkRecords() returned skipped name %q", record.Name)
		}
	}
}

func TestWalkStopsAtCallbackError(t *testing.T) {
	store := openStore(t, t.TempDir())
	if err := store.Write("tasks", "one", map[string]string{"id": "one"}); err != nil {
		t.Fatal(err)
	}
	sentinel := errors.New("stop the walk")
	err := store.WalkRecords(context.Background(), func(StoredRecord) error {
		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("WalkRecords() error = %v; want the callback error", err)
	}
}

func TestWalkRejectsLinkedCollection(t *testing.T) {
	store := openStore(t, t.TempDir())
	stateDir := filepath.Join(store.home, ".hive", "state")
	target := t.TempDir()
	if err := os.WriteFile(filepath.Join(target, "inside.json"), []byte(`{"id":"inside"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(stateDir, "linked")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	err := store.WalkRecords(context.Background(), func(record StoredRecord) error {
		t.Errorf("WalkRecords() returned %v from a linked collection", record)
		return nil
	})
	if err == nil || !strings.Contains(err.Error(), link) {
		t.Fatalf("WalkRecords() error = %v; want the linked path %s", err, link)
	}
}

func TestWalkRejectsLinkedRecord(t *testing.T) {
	store := openStore(t, t.TempDir())
	dir := filepath.Join(store.home, ".hive", "state", "tasks")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(store.home, "target.json")
	if err := os.WriteFile(target, []byte(`{"id":"target"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "linked.json")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	err := store.WalkRecords(context.Background(), func(StoredRecord) error {
		t.Error("WalkRecords() returned a linked record")
		return nil
	})
	if err == nil || !strings.Contains(err.Error(), link) {
		t.Fatalf("WalkRecords() error = %v; want the linked path %s", err, link)
	}
}

func TestWalkBadNestedRecord(t *testing.T) {
	store := openStore(t, t.TempDir())
	dir := filepath.Join(store.home, ".hive", "state", "inbox", "coordinator")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	bad := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(bad, []byte(`{"id":`), 0o644); err != nil {
		t.Fatal(err)
	}
	err := store.WalkRecords(context.Background(), func(StoredRecord) error {
		t.Error("WalkRecords() returned a bad record")
		return nil
	})
	if err == nil || err.Error() != "record invalid: "+bad {
		t.Fatalf("WalkRecords() error = %v; want record invalid for %s", err, bad)
	}
}
