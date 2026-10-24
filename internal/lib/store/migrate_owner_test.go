package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

// readJSONMap reads one state file as a JSON object.
func readJSONMap(t *testing.T, home string, parts ...string) map[string]any {
	t.Helper()
	path := filepath.Join(append([]string{home, ".hive", "state"}, parts...)...)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var value map[string]any
	if err := json.Unmarshal(data, &value); err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
	return value
}

func fileMissing(t *testing.T, home string, parts ...string) bool {
	t.Helper()
	path := filepath.Join(append([]string{home, ".hive", "state"}, parts...)...)
	_, err := os.Lstat(path)
	if err == nil {
		return false
	}
	if !os.IsNotExist(err) {
		t.Fatalf("stat %s: %v", path, err)
	}
	return true
}

// seedVersionTwoHome writes one version 2 home with an old fixed manager
// claim, an old fixed manager inbox, and one task with controlling_role.
func seedVersionTwoHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	seedHive(t, home)
	seedState(t, home, []string{"schema.json"}, map[string]any{"version": 2})
	seedState(t, home, []string{"agents", "owner-demo-g1.json"}, map[string]any{
		"key": "owner-demo-g1", "name": "owner-demo", "role": "manager", "kind": "",
		"profile": "planner", "owns": "demo", "worktree": "", "project": "demo", "task": "",
		"status": "live", "generation": 1,
	})
	seedState(t, home, []string{"claims", "manager-demo.json"}, map[string]any{
		"key": "manager-demo", "agent": "owner-demo-g1", "generation": 3,
		"pid": 0, "pid_start": 0, "claimed_at": "2026-01-01T00:00:00Z",
	})
	seedState(t, home, []string{"inbox", "manager-demo", "1.json"}, map[string]any{
		"seq": 1, "role": "manager-demo", "kind": "completed", "task": "1-task",
		"agent": "worker-1", "text": "completed", "created_at": "2026-01-01T00:00:00Z", "read_at": "",
	})
	seedState(t, home, []string{"inbox", "manager-demo", "2.json"}, map[string]any{
		"seq": 2, "role": "manager-demo", "kind": "message", "task": "",
		"agent": "user", "text": "read already", "created_at": "2026-01-01T00:01:00Z", "read_at": "2026-01-02T00:00:00Z",
	})
	seedState(t, home, []string{"counters", "inbox:manager-demo.json"}, map[string]any{"value": 2})
	seedState(t, home, []string{"tasks", "1-task.json"}, map[string]any{
		"id": "1-task", "project": "demo", "title": "task", "controlling_role": "manager-demo",
		"acting_agent": "worker-1", "status": "completed", "depends_on": []string{},
	})
	return home
}

// TestMigrationVersionThreeMovesOwnerRecords proves AC3: the version 3 step
// renames the old claim, moves each unread record to the owner inbox, deletes
// the old folder and counter, and removes controlling_role. A second open
// changes no file.
func TestMigrationVersionThreeMovesOwnerRecords(t *testing.T) {
	home := seedVersionTwoHome(t)

	if _, err := Open(home); err != nil {
		t.Fatalf("Open() = %v; want the migration", err)
	}
	if version, ok := readSchemaVersion(t, home); !ok || version != 3 {
		t.Fatalf("schema version = %d, present %t; want 3", version, ok)
	}

	claim := readJSONMap(t, home, "claims", "owner-demo.json")
	if claim["key"] != "owner-demo" || claim["agent"] != "owner-demo-g1" || claim["generation"] != float64(3) {
		t.Fatalf("owner claim = %#v; want owner-demo, owner-demo-g1, generation 3", claim)
	}
	if !fileMissing(t, home, "claims", "manager-demo.json") {
		t.Fatal("the old manager claim is still present")
	}

	moved := readJSONMap(t, home, "inbox", "owner-demo", "1.json")
	for field, want := range map[string]any{
		"seq": float64(1), "role": "owner-demo", "kind": "completed",
		"task": "1-task", "agent": "worker-1", "text": "completed",
		"created_at": "2026-01-01T00:00:00Z",
	} {
		if moved[field] != want {
			t.Errorf("moved record %s = %#v; want %#v", field, moved[field], want)
		}
	}
	if _, err := os.Lstat(filepath.Join(home, ".hive", "state", "inbox", "owner-demo", "2.json")); !os.IsNotExist(err) {
		t.Fatal("the read record was copied to the owner inbox")
	}
	if !fileMissing(t, home, "inbox", "manager-demo") {
		t.Fatal("the old manager inbox folder is still present")
	}
	if !fileMissing(t, home, "counters", "inbox:manager-demo.json") {
		t.Fatal("the old manager inbox counter is still present")
	}

	task := readJSONMap(t, home, "tasks", "1-task.json")
	if _, ok := task["controlling_role"]; ok {
		t.Fatalf("task = %#v; want no controlling_role", task)
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
}

// TestMigrationVersionThreeUsesCoordinatorWithoutLiveOwner proves that the
// unread records go to the coordinator when the project has no live owner.
func TestMigrationVersionThreeUsesCoordinatorWithoutLiveOwner(t *testing.T) {
	home := seedVersionTwoHome(t)
	stopped := map[string]any{
		"key": "owner-demo-g1", "name": "owner-demo", "role": "manager", "kind": "",
		"profile": "planner", "owns": "demo", "worktree": "", "project": "demo", "task": "",
		"status": "stopped", "generation": 1,
	}
	seedState(t, home, []string{"agents", "owner-demo-g1.json"}, stopped)

	if _, err := Open(home); err != nil {
		t.Fatalf("Open() = %v; want the migration", err)
	}
	moved := readJSONMap(t, home, "inbox", "coordinator", "1.json")
	if moved["text"] != "completed" || moved["agent"] != "worker-1" || moved["role"] != "coordinator" {
		t.Fatalf("coordinator record = %#v; want the unread manager record", moved)
	}
}

// TestMigrationVersionThreeKeepsInboxWhenOwnerNameMatches covers the live home
// shape: the owner name is manager-PROJECT, so the owner inbox is the old
// fixed inbox. The step keeps each record and its counter.
func TestMigrationVersionThreeKeepsInboxWhenOwnerNameMatches(t *testing.T) {
	home := t.TempDir()
	seedHive(t, home)
	seedState(t, home, []string{"schema.json"}, map[string]any{"version": 2})
	seedState(t, home, []string{"agents", "manager-demo-g1.json"}, map[string]any{
		"key": "manager-demo-g1", "name": "manager-demo", "role": "manager", "kind": "",
		"profile": "planner", "owns": "demo", "worktree": "", "project": "demo", "task": "",
		"status": "live", "generation": 1,
	})
	seedState(t, home, []string{"claims", "manager-demo.json"}, map[string]any{
		"key": "manager-demo", "agent": "manager-demo-g1", "generation": 1,
	})
	seedState(t, home, []string{"inbox", "manager-demo", "1.json"}, map[string]any{
		"seq": 1, "role": "manager-demo", "kind": "completed", "task": "1-task",
		"agent": "worker-1", "text": "completed", "created_at": "2026-01-01T00:00:00Z", "read_at": "",
	})
	seedState(t, home, []string{"counters", "inbox:manager-demo.json"}, map[string]any{"value": 1})

	if _, err := Open(home); err != nil {
		t.Fatalf("Open() = %v; want the migration", err)
	}
	record := readJSONMap(t, home, "inbox", "manager-demo", "1.json")
	if record["text"] != "completed" || record["read_at"] != "" {
		t.Fatalf("record = %#v; want the original unread record", record)
	}
	counter := readJSONMap(t, home, "counters", "inbox:manager-demo.json")
	if counter["value"] != float64(1) {
		t.Fatalf("counter = %#v; want the original value", counter)
	}
	claim := readJSONMap(t, home, "claims", "owner-demo.json")
	if claim["agent"] != "manager-demo-g1" {
		t.Fatalf("claim = %#v; want the renamed claim", claim)
	}
}

// TestMigrationVersionThreeDoesNotCopyTwice shows the step is safe to run
// again after a partial copy.
func TestMigrationVersionThreeDoesNotCopyTwice(t *testing.T) {
	home := seedVersionTwoHome(t)
	seedState(t, home, []string{"inbox", "owner-demo", "1.json"}, map[string]any{
		"seq": 1, "role": "owner-demo", "kind": "completed", "task": "1-task",
		"agent": "worker-1", "text": "completed", "created_at": "2026-01-01T00:00:00Z", "read_at": "",
	})
	seedState(t, home, []string{"counters", "inbox:owner-demo.json"}, map[string]any{"value": 1})

	if _, err := Open(home); err != nil {
		t.Fatalf("Open() = %v; want the migration", err)
	}
	if version, ok := readSchemaVersion(t, home); !ok || version != 3 {
		t.Fatalf("schema version = %d, present %t; want 3", version, ok)
	}
	if !fileMissing(t, home, "inbox", "manager-demo") {
		t.Fatal("the old manager inbox folder is still present")
	}
	entries, err := os.ReadDir(filepath.Join(home, ".hive", "state", "inbox", "owner-demo"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("owner inbox records = %d; want one copy", len(entries))
	}
}

// seedEqualUnreadHome writes one version 2 home with a live owner and two
// unread manager records that differ only by sequence ID.
func seedEqualUnreadHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	seedHive(t, home)
	seedState(t, home, []string{"schema.json"}, map[string]any{"version": 2})
	seedState(t, home, []string{"agents", "owner-demo-g1.json"}, map[string]any{
		"key": "owner-demo-g1", "name": "owner-demo", "role": "manager", "kind": "",
		"profile": "planner", "owns": "demo", "worktree": "", "project": "demo", "task": "",
		"status": "live", "generation": 1,
	})
	seedState(t, home, []string{"claims", "manager-demo.json"}, map[string]any{
		"key": "manager-demo", "agent": "owner-demo-g1", "generation": 1,
	})
	for seq := 1; seq <= 2; seq++ {
		seedState(t, home, []string{"inbox", "manager-demo", strconv.Itoa(seq) + ".json"}, map[string]any{
			"seq": seq, "role": "manager-demo", "kind": "message", "task": "",
			"agent": "user", "text": "same text", "created_at": "2026-01-01T00:00:00Z", "read_at": "",
		})
	}
	seedState(t, home, []string{"counters", "inbox:manager-demo.json"}, map[string]any{"value": 2})
	return home
}

// TestMigrationVersionThreeKeepsEqualUnreadRecords proves that version 3
// copies each unread source record, including two records that have equal
// content and time. The copy count, not a content signature alone, decides
// how many records move.
func TestMigrationVersionThreeKeepsEqualUnreadRecords(t *testing.T) {
	home := seedEqualUnreadHome(t)

	if _, err := Open(home); err != nil {
		t.Fatalf("Open() = %v; want the migration", err)
	}
	entries, err := os.ReadDir(filepath.Join(home, ".hive", "state", "inbox", "owner-demo"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("owner inbox records = %d; want two copies", len(entries))
	}
	seen := map[string]bool{}
	for _, entry := range entries {
		id := entry.Name()
		record := readJSONMap(t, home, "inbox", "owner-demo", id)
		if record["seq"] != float64(len(seen)+1) || record["role"] != "owner-demo" || record["text"] != "same text" {
			t.Fatalf("moved record = %#v; want the next owner sequence, owner-demo, same text", record)
		}
		if _, ok := record["migrated_from"]; ok {
			t.Fatalf("moved record = %#v; want no migration marker", record)
		}
		seen[id] = true
	}
	if !fileMissing(t, home, "inbox", "manager-demo") {
		t.Fatal("the old manager inbox folder is still present")
	}
}

// TestMigrationVersionThreeKeepsEqualRecordsFromTwoInboxes proves that equal
// unread records of two old inboxes both move to the same target. The source
// identity, not the content, decides which records are copies.
func TestMigrationVersionThreeKeepsEqualRecordsFromTwoInboxes(t *testing.T) {
	home := t.TempDir()
	seedHive(t, home)
	seedState(t, home, []string{"schema.json"}, map[string]any{"version": 2})
	for _, project := range []string{"one", "two"} {
		inbox := "manager-" + project
		seedState(t, home, []string{"inbox", inbox, "1.json"}, map[string]any{
			"seq": 1, "role": inbox, "kind": "message", "task": "",
			"agent": "user", "text": "same text", "created_at": "2026-01-01T00:00:00Z", "read_at": "",
		})
		seedState(t, home, []string{"counters", "inbox:" + inbox + ".json"}, map[string]any{"value": 1})
	}

	if _, err := Open(home); err != nil {
		t.Fatalf("Open() = %v; want the migration", err)
	}
	entries, err := os.ReadDir(filepath.Join(home, ".hive", "state", "inbox", "coordinator"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("coordinator inbox records = %d; want two copies", len(entries))
	}
	for _, project := range []string{"one", "two"} {
		if !fileMissing(t, home, "inbox", "manager-"+project) {
			t.Fatalf("the old inbox of project %s is still present", project)
		}
	}
}

// TestMigrationVersionThreeSkipsAMarkedCopy proves that the marker of a
// source record stops a second copy after a run that stopped early.
func TestMigrationVersionThreeSkipsAMarkedCopy(t *testing.T) {
	home := seedVersionTwoHome(t)
	seedState(t, home, []string{"inbox", "owner-demo", "1.json"}, map[string]any{
		"seq": 1, "role": "owner-demo", "kind": "completed", "task": "1-task",
		"agent": "worker-1", "text": "completed", "created_at": "2026-01-01T00:00:00Z", "read_at": "",
		"migrated_from": "manager-demo/1",
	})
	seedState(t, home, []string{"counters", "inbox:owner-demo.json"}, map[string]any{"value": 1})

	if _, err := Open(home); err != nil {
		t.Fatalf("Open() = %v; want the migration", err)
	}
	entries, err := os.ReadDir(filepath.Join(home, ".hive", "state", "inbox", "owner-demo"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("owner inbox records = %d; want one copy", len(entries))
	}
	record := readJSONMap(t, home, "inbox", "owner-demo", "1.json")
	if _, ok := record["migrated_from"]; ok {
		t.Fatalf("record = %#v; want no migration marker", record)
	}
	if !fileMissing(t, home, "inbox", "manager-demo") {
		t.Fatal("the old manager inbox folder is still present")
	}
}

// TestMigrationVersionThreeResumesAfterAPartialEqualCopy proves that a retry
// after one equal record moved copies the remaining unread record.
func TestMigrationVersionThreeResumesAfterAPartialEqualCopy(t *testing.T) {
	home := seedEqualUnreadHome(t)
	seedState(t, home, []string{"inbox", "owner-demo", "1.json"}, map[string]any{
		"seq": 1, "role": "owner-demo", "kind": "message", "task": "",
		"agent": "user", "text": "same text", "created_at": "2026-01-01T00:00:00Z", "read_at": "",
	})
	seedState(t, home, []string{"counters", "inbox:owner-demo.json"}, map[string]any{"value": 1})

	if _, err := Open(home); err != nil {
		t.Fatalf("Open() = %v; want the migration", err)
	}
	entries, err := os.ReadDir(filepath.Join(home, ".hive", "state", "inbox", "owner-demo"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("owner inbox records = %d; want two copies", len(entries))
	}
	if !fileMissing(t, home, "inbox", "manager-demo") {
		t.Fatal("the old manager inbox folder is still present")
	}
}

// TestMigrationVersionThreeRunsAfterVersionTwo covers the full path from
// version 1 with an old-shape manager record.
func TestMigrationVersionThreeRunsAfterVersionTwo(t *testing.T) {
	home := t.TempDir()
	seedHive(t, home)
	seedState(t, home, []string{"agents", "manager-demo-g1.json"}, map[string]any{
		"key": "manager-demo-g1", "name": "owner-demo", "role": "manager",
		"project": "demo", "task": "", "status": "live", "generation": 1,
	})
	seedState(t, home, []string{"claims", "manager-demo.json"}, map[string]any{
		"key": "manager-demo", "agent": "manager-demo-g1", "generation": 1,
	})
	seedState(t, home, []string{"inbox", "manager-demo", "1.json"}, map[string]any{
		"seq": 1, "role": "manager-demo", "kind": "message", "agent": "user", "text": "hello",
		"created_at": "2026-01-01T00:00:00Z", "read_at": "",
	})

	if _, err := Open(home); err != nil {
		t.Fatalf("Open() = %v; want the migration", err)
	}
	if version, ok := readSchemaVersion(t, home); !ok || version != 3 {
		t.Fatalf("schema version = %d, present %t; want 3", version, ok)
	}
	record := readAgentMap(t, home, "manager-demo-g1")
	if record["owns"] != "demo" || record["profile"] != "planner" {
		t.Fatalf("record = %#v; want the version 2 facts", record)
	}
	moved := readJSONMap(t, home, "inbox", "owner-demo", "1.json")
	if moved["text"] != "hello" {
		t.Fatalf("record = %#v; want the moved unread record", moved)
	}
	if !fileMissing(t, home, "inbox", "manager-demo") {
		t.Fatal("the old manager inbox folder is still present")
	}
}
