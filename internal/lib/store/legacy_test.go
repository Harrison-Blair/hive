package store

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

func TestLegacyListHidesLongID(t *testing.T) {
	store := openStore(t, t.TempDir())
	longID := strings.Repeat("a", 19)
	if err := store.Write("projects", longID, map[string]string{"id": longID}); err != nil {
		t.Fatal(err)
	}
	list, err := store.List("projects")
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 0 {
		t.Fatalf("List() returned %d records; want the long ID hidden", len(list))
	}
	in, err := store.ListIn([]string{"projects"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(in) != 1 || !strings.Contains(string(in[0]), longID) {
		t.Fatalf("ListIn() records = %q; want the long ID", in)
	}
}

func TestLegacyLongTaskIDAndColonCounter(t *testing.T) {
	store := openStore(t, t.TempDir())
	longID := strings.Repeat("b", 40)
	if err := store.Write("tasks", longID, map[string]string{"id": longID}); err != nil {
		t.Fatal(err)
	}
	var record map[string]string
	if err := store.Read("tasks", longID, &record); err != nil {
		t.Fatal(err)
	}
	if record["id"] != longID {
		t.Fatalf("Read() id = %q; want %q", record["id"], longID)
	}
	for want := uint64(1); want <= 2; want++ {
		got, err := store.Next("inbox:coordinator")
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Fatalf("Next() = %d; want %d", got, want)
		}
	}
	path := filepath.Join(store.home, ".hive", "state", "counters", "inbox:coordinator.json")
	if _, err := readRecordFile(path); err != nil {
		t.Fatalf("colon counter record missing: %v", err)
	}
}

func TestLegacyUnsafeNamesStillRejected(t *testing.T) {
	store := openStore(t, t.TempDir())
	for _, test := range []struct {
		kind, id string
	}{
		{"", "ok"}, {"../bad", "ok"}, {"kind", ""}, {"kind", "../bad"}, {"kind", "with/slash"},
	} {
		if err := store.Read(test.kind, test.id, &map[string]any{}); err == nil {
			t.Errorf("Read(%q, %q) succeeded", test.kind, test.id)
		}
	}
	if err := store.Write("kind", "", map[string]any{}); err == nil {
		t.Error("Write with an empty ID succeeded")
	}
	if err := store.Update("kind", "../bad", func(json.RawMessage) (json.RawMessage, error) {
		return json.RawMessage(`{"value":1}`), nil
	}); err == nil {
		t.Error("Update with a parent path succeeded")
	}
	for _, name := range []string{"", "..", "../bad", "with/slash"} {
		if _, err := store.Next(name); err == nil {
			t.Errorf("Next(%q) succeeded", name)
		}
	}
}
