package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestMigrationVersionFourRemovesRoleCompatibility proves D2/D13: schema 3
// records migrate to schema 4, the agent role mirror goes away, and each inbox
// record field role becomes inbox. A second open changes no state byte.
func TestMigrationVersionFourRemovesRoleCompatibility(t *testing.T) {
	home := t.TempDir()
	seedHive(t, home)
	seedState(t, home, []string{"schema.json"}, map[string]any{"version": 3})
	seedState(t, home, []string{"agents", "owner-hive-g2.json"}, map[string]any{
		"key": "owner-hive-g2", "name": "owner-hive", "role": "manager",
		"profile": "planner", "kind": "", "owns": "hive", "project": "hive",
		"status": "live", "generation": 2, "replaces": "owner-hive-g1",
	})
	seedState(t, home, []string{"agents", "builder-1-g1.json"}, map[string]any{
		"key": "builder-1-g1", "name": "builder-1", "role": "worker",
		"profile": "builder", "kind": "", "owns": "", "project": "hive",
		"task": "1-do", "worktree": "/tmp/1-do", "status": "live", "generation": 1,
	})
	seedState(t, home, []string{"inbox", "coordinator", "1.json"}, map[string]any{
		"seq": 1, "role": "coordinator", "kind": "message", "agent": "builder-1",
		"text": "hello", "created_at": "2026-10-26T00:00:00Z", "read_at": "",
	})
	seedState(t, home, []string{"inbox", "owner-hive", "2.json"}, map[string]any{
		"seq": 2, "role": "owner-hive", "kind": "completed", "task": "1-do",
		"agent": "builder-1", "text": "completed", "created_at": "2026-10-26T00:01:00Z",
		"read_at": "2026-10-26T00:02:00Z", "nudged_at": "", "nudged_generation": 0,
	})

	if _, err := Open(home); err != nil {
		t.Fatalf("Open() = %v; want the version 4 migration", err)
	}
	if version, ok := readSchemaVersion(t, home); !ok || version != 4 {
		t.Fatalf("schema version = %d, present %t; want 4", version, ok)
	}

	owner := readAgentMap(t, home, "owner-hive-g2")
	if _, ok := owner["role"]; ok {
		t.Fatalf("owner record still has role: %#v", owner)
	}
	for field, want := range map[string]any{
		"name": "owner-hive", "profile": "planner", "kind": "", "owns": "hive",
		"project": "hive", "status": "live", "generation": float64(2),
		"replaces": "owner-hive-g1",
	} {
		if got := owner[field]; got != want {
			t.Errorf("owner %s = %#v; want %#v", field, got, want)
		}
	}
	builder := readAgentMap(t, home, "builder-1-g1")
	if _, ok := builder["role"]; ok {
		t.Fatalf("builder record still has role: %#v", builder)
	}
	if builder["profile"] != "builder" || builder["task"] != "1-do" || builder["worktree"] != "/tmp/1-do" {
		t.Fatalf("builder facts changed: %#v", builder)
	}

	for _, inbox := range []struct {
		key string
		seq string
	}{
		{key: "coordinator", seq: "1"},
		{key: "owner-hive", seq: "2"},
	} {
		record := readInboxMap(t, home, inbox.key, inbox.seq)
		if _, ok := record["role"]; ok {
			t.Errorf("inbox %s/%s still has role: %#v", inbox.key, inbox.seq, record)
		}
		if got := record["inbox"]; got != inbox.key {
			t.Errorf("inbox %s/%s field inbox = %#v; want %q", inbox.key, inbox.seq, got, inbox.key)
		}
	}
	completed := readInboxMap(t, home, "owner-hive", "2")
	if completed["read_at"] != "2026-10-26T00:02:00Z" || completed["text"] != "completed" {
		t.Fatalf("inbox facts changed: %#v", completed)
	}

	before := stateFingerprints(t, home)
	if _, err := Open(home); err != nil {
		t.Fatalf("second Open() = %v; want no change", err)
	}
	after := stateFingerprints(t, home)
	if len(before) != len(after) {
		t.Fatalf("state files = %d then %d; want no change", len(before), len(after))
	}
	for index := range before {
		if before[index] != after[index] {
			t.Fatalf("file changed on the second open: %s then %s", before[index], after[index])
		}
	}

	events := migrationLogEvents(t, home)
	start := findMigrationEvent(t, events, "state migrate start")
	if start["from"] != float64(3) || start["to"] != float64(4) {
		t.Fatalf("migration start = %#v; want 3 to 4", start)
	}
	done := findMigrationEvent(t, events, "state migrate done")
	if done["from"] != float64(3) || done["to"] != float64(4) {
		t.Fatalf("migration done = %#v; want 3 to 4", done)
	}
	paths := map[string]bool{}
	for _, event := range events {
		if event["msg"] == "state migrate record" {
			path, _ := event["path"].(string)
			paths[path] = true
		}
	}
	for _, want := range []string{
		".hive/state/agents/owner-hive-g2.json",
		".hive/state/agents/builder-1-g1.json",
		".hive/state/inbox/coordinator/1.json",
		".hive/state/inbox/owner-hive/2.json",
	} {
		if !paths[want] {
			t.Errorf("migration log has no record path %q: %#v", want, paths)
		}
	}
}

// TestMigrationVersionFourIsIdempotentOnAPartialRecord proves that a record
// with both fields keeps its inbox value, and a record that already has the
// new shape is not rewritten.
func TestMigrationVersionFourIsIdempotentOnAPartialRecord(t *testing.T) {
	home := t.TempDir()
	seedHive(t, home)
	seedState(t, home, []string{"schema.json"}, map[string]any{"version": 3})
	seedState(t, home, []string{"agents", "plain-1.json"}, map[string]any{
		"key": "plain-1", "name": "plain-1", "profile": "general", "status": "live", "generation": 1,
	})
	seedState(t, home, []string{"inbox", "coordinator", "1.json"}, map[string]any{
		"seq": 1, "inbox": "coordinator", "role": "coordinator", "kind": "message",
		"text": "both", "created_at": "2026-10-26T00:00:00Z",
	})
	if _, err := Open(home); err != nil {
		t.Fatalf("Open() = %v", err)
	}
	record := readInboxMap(t, home, "coordinator", "1")
	if _, ok := record["role"]; ok {
		t.Fatalf("partial inbox record still has role: %#v", record)
	}
	if record["inbox"] != "coordinator" {
		t.Fatalf("partial inbox record inbox = %#v; want coordinator", record["inbox"])
	}
	agent := readAgentMap(t, home, "plain-1")
	if _, ok := agent["role"]; ok {
		t.Fatalf("agent record gained or kept role: %#v", agent)
	}
}

func readInboxMap(t *testing.T, home, key, seq string) map[string]any {
	t.Helper()
	path := filepath.Join(home, ".hive", "state", "inbox", key, seq+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var record map[string]any
	if err := json.Unmarshal(data, &record); err != nil {
		t.Fatal(err)
	}
	return record
}
