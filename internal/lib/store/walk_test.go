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

func TestWalkTemporaryPatternChecksTypeFirst(t *testing.T) {
	t.Run("directory descends", func(t *testing.T) {
		store := openStore(t, t.TempDir())
		dir := filepath.Join(store.home, ".hive", "state", ".probe.tmp-123")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "one.json"), []byte(`{"id":"one"}`), 0o644); err != nil {
			t.Fatal(err)
		}
		var found []string
		if err := store.WalkRecords(context.Background(), func(record StoredRecord) error {
			found = append(found, strings.Join(append(append([]string(nil), record.Collection...), record.Name), "/"))
			return nil
		}); err != nil {
			t.Fatal(err)
		}
		want := ".probe.tmp-123/one.json"
		if len(found) != 1 || found[0] != want {
			t.Fatalf("WalkRecords() paths = %q; want [%q]", found, want)
		}
	})

	t.Run("linked collection fails", func(t *testing.T) {
		store := openStore(t, t.TempDir())
		target := t.TempDir()
		if err := os.WriteFile(filepath.Join(target, "one.json"), []byte(`{"id":"one"}`), 0o644); err != nil {
			t.Fatal(err)
		}
		link := filepath.Join(store.home, ".hive", "state", ".probe.tmp-123")
		if err := os.Symlink(target, link); err != nil {
			t.Fatal(err)
		}
		err := store.WalkRecords(context.Background(), func(record StoredRecord) error {
			t.Errorf("WalkRecords() returned %v from a linked collection", record)
			return nil
		})
		if err == nil || err.Error() != "store path invalid: "+link {
			t.Fatalf("WalkRecords() error = %v; want store path invalid for %s", err, link)
		}
	})

	t.Run("regular atomic temporary file skips", func(t *testing.T) {
		store := openStore(t, t.TempDir())
		stateDir := filepath.Join(store.home, ".hive", "state")
		if err := os.WriteFile(filepath.Join(stateDir, ".probe.tmp-123"), []byte(`{"value":1}`), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := store.WalkRecords(context.Background(), func(record StoredRecord) error {
			t.Errorf("WalkRecords() returned a temporary file %q", record.Name)
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	})
}

func TestCollectWalkNodesChecksLockTypeFirst(t *testing.T) {
	stateDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(stateDir, "nested", "lock"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stateDir, "nested", "lock", "one.json"), []byte(`{"id":"one"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stateDir, "plain.json"), []byte(`{"id":"plain"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	nodes, err := collectWalkNodes(stateDir, nil, true)
	if err != nil {
		t.Fatal(err)
	}
	var found []string
	for _, node := range nodes {
		found = append(found, node.rel)
	}
	want := []string{"nested", "nested/lock", "nested/lock/one.json", "plain.json"}
	if len(found) != len(want) {
		t.Fatalf("collectWalkNodes() paths = %q; want %q", found, want)
	}
	for i := range want {
		if found[i] != want[i] {
			t.Fatalf("collectWalkNodes() paths = %q; want %q", found, want)
		}
	}

	if err := os.RemoveAll(filepath.Join(stateDir, "nested")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stateDir, "lock"), []byte(`{"value":1}`), 0o644); err != nil {
		t.Fatal(err)
	}
	nodes, err = collectWalkNodes(stateDir, nil, true)
	if err != nil {
		t.Fatal(err)
	}
	for _, node := range nodes {
		if node.name == "lock" {
			t.Fatalf("collectWalkNodes() returned the regular lock file: %q", node.rel)
		}
	}
}
