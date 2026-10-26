package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
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
	if version, ok := readSchemaVersion(t, home); !ok || version != 4 {
		t.Fatalf("schema version = %d, present %t; want 4", version, ok)
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
		"seq": float64(1), "inbox": "owner-demo", "kind": "completed",
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
	if moved["text"] != "completed" || moved["agent"] != "worker-1" || moved["inbox"] != "coordinator" {
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

// TestMigrationVersionThreeKeepsAnEqualUnreadSourceRecord proves that the
// step moves an unread source record even when the owner inbox already has a
// record with equal content. A record without a marker is not a copy of the
// source record.
func TestMigrationVersionThreeKeepsAnEqualUnreadSourceRecord(t *testing.T) {
	home := seedVersionTwoHome(t)
	seedState(t, home, []string{"inbox", "owner-demo", "1.json"}, map[string]any{
		"seq": 1, "role": "owner-demo", "kind": "completed", "task": "1-task",
		"agent": "worker-1", "text": "completed", "created_at": "2026-01-01T00:00:00Z", "read_at": "",
	})
	seedState(t, home, []string{"counters", "inbox:owner-demo.json"}, map[string]any{"value": 1})

	if _, err := Open(home); err != nil {
		t.Fatalf("Open() = %v; want the migration", err)
	}
	if version, ok := readSchemaVersion(t, home); !ok || version != 4 {
		t.Fatalf("schema version = %d, present %t; want 4", version, ok)
	}
	if !fileMissing(t, home, "inbox", "manager-demo") {
		t.Fatal("the old manager inbox folder is still present")
	}
	want := []map[string]any{
		{
			"seq": float64(1), "inbox": "owner-demo", "kind": "completed", "task": "1-task",
			"agent": "worker-1", "text": "completed", "created_at": "2026-01-01T00:00:00Z", "read_at": "",
		},
		{
			"seq": float64(2), "inbox": "owner-demo", "kind": "completed", "task": "1-task",
			"agent": "worker-1", "text": "completed", "created_at": "2026-01-01T00:00:00Z", "read_at": "",
		},
	}
	entries, err := os.ReadDir(filepath.Join(home, ".hive", "state", "inbox", "owner-demo"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != len(want) {
		t.Fatalf("owner inbox records = %d; want %d", len(entries), len(want))
	}
	for index, entry := range entries {
		record := readJSONMap(t, home, "inbox", "owner-demo", entry.Name())
		for field, value := range want[index] {
			if record[field] != value {
				t.Errorf("record %d field %s = %#v; want %#v", index+1, field, record[field], value)
			}
		}
		if _, ok := record["migrated_from"]; ok {
			t.Errorf("record %d = %#v; want no migration marker", index+1, record)
		}
	}
}

// TestMigrationVersionThreeKeepsAnEqualReadCoordinatorRecord proves that the
// coordinator fallback keeps a read record that already has equal content and
// appends the unread source record.
func TestMigrationVersionThreeKeepsAnEqualReadCoordinatorRecord(t *testing.T) {
	home := seedVersionTwoHome(t)
	seedState(t, home, []string{"agents", "owner-demo-g1.json"}, map[string]any{
		"key": "owner-demo-g1", "name": "owner-demo", "role": "manager", "kind": "",
		"profile": "planner", "owns": "demo", "worktree": "", "project": "demo", "task": "",
		"status": "stopped", "generation": 1,
	})
	seedState(t, home, []string{"inbox", "coordinator", "1.json"}, map[string]any{
		"seq": 1, "role": "coordinator", "kind": "completed", "task": "1-task",
		"agent": "worker-1", "text": "completed", "created_at": "2026-01-01T00:00:00Z",
		"read_at": "2026-01-02T00:00:00Z",
	})
	seedState(t, home, []string{"counters", "inbox:coordinator.json"}, map[string]any{"value": 1})

	if _, err := Open(home); err != nil {
		t.Fatalf("Open() = %v; want the migration", err)
	}
	if !fileMissing(t, home, "inbox", "manager-demo") {
		t.Fatal("the old manager inbox folder is still present")
	}
	want := []map[string]any{
		{
			"seq": float64(1), "inbox": "coordinator", "kind": "completed", "task": "1-task",
			"agent": "worker-1", "text": "completed", "created_at": "2026-01-01T00:00:00Z",
			"read_at": "2026-01-02T00:00:00Z",
		},
		{
			"seq": float64(2), "inbox": "coordinator", "kind": "completed", "task": "1-task",
			"agent": "worker-1", "text": "completed", "created_at": "2026-01-01T00:00:00Z",
			"read_at": "",
		},
	}
	entries, err := os.ReadDir(filepath.Join(home, ".hive", "state", "inbox", "coordinator"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != len(want) {
		t.Fatalf("coordinator inbox records = %d; want %d", len(entries), len(want))
	}
	for index, entry := range entries {
		record := readJSONMap(t, home, "inbox", "coordinator", entry.Name())
		for field, value := range want[index] {
			if record[field] != value {
				t.Errorf("record %d field %s = %#v; want %#v", index+1, field, record[field], value)
			}
		}
		if _, ok := record["migrated_from"]; ok {
			t.Errorf("record %d = %#v; want no migration marker", index+1, record)
		}
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
		if record["seq"] != float64(len(seen)+1) || record["inbox"] != "owner-demo" || record["text"] != "same text" {
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

// seedOwnerAliasHome writes one version 2 home where the live owner of project
// one is named manager-two, the old inbox key of project two. Project two has
// no live owner. Each old inbox holds one unread record.
func seedOwnerAliasHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	seedHive(t, home)
	seedState(t, home, []string{"schema.json"}, map[string]any{"version": 2})
	seedState(t, home, []string{"agents", "manager-two-g1.json"}, map[string]any{
		"key": "manager-two-g1", "name": "manager-two", "role": "manager", "kind": "",
		"profile": "planner", "owns": "one", "worktree": "", "project": "one", "task": "",
		"status": "live", "generation": 1,
	})
	seedState(t, home, []string{"claims", "manager-one.json"}, map[string]any{
		"key": "manager-one", "agent": "manager-two-g1", "generation": 1,
	})
	for _, project := range []string{"one", "two"} {
		inbox := "manager-" + project
		seedState(t, home, []string{"inbox", inbox, "1.json"}, map[string]any{
			"seq": 1, "role": inbox, "kind": "message", "task": "",
			"agent": "user", "text": "mail for " + project,
			"created_at": "2026-01-01T00:00:00Z", "read_at": "",
		})
		seedState(t, home, []string{"counters", "inbox:" + inbox + ".json"}, map[string]any{"value": 1})
	}
	return home
}

// TestMigrationVersionThreeKeepsEachSourceDestination proves that each source
// record keeps the destination of its own project when the live owner of one
// project is named after the old inbox of another project. The owner inbox
// manager-two is also the old inbox of project two, so the step must not read
// its copied record as a source record of project two.
func TestMigrationVersionThreeKeepsEachSourceDestination(t *testing.T) {
	home := seedOwnerAliasHome(t)

	if _, err := Open(home); err != nil {
		t.Fatalf("Open() = %v; want the migration", err)
	}
	if version, ok := readSchemaVersion(t, home); !ok || version != 4 {
		t.Fatalf("schema version = %d, present %t; want 4", version, ok)
	}

	owner := readJSONMap(t, home, "inbox", "manager-two", "2.json")
	wantOwner := map[string]any{
		"seq": float64(2), "inbox": "manager-two", "kind": "message", "task": "",
		"agent": "user", "text": "mail for one", "created_at": "2026-01-01T00:00:00Z",
		"read_at": "", "nudged_at": "", "nudged_generation": float64(0),
	}
	if !reflect.DeepEqual(owner, wantOwner) {
		t.Fatalf("owner record = %#v; want %#v", owner, wantOwner)
	}
	entries, err := os.ReadDir(filepath.Join(home, ".hive", "state", "inbox", "manager-two"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("owner inbox records = %d; want one record", len(entries))
	}

	coordinator := readJSONMap(t, home, "inbox", "coordinator", "1.json")
	wantCoordinator := map[string]any{
		"seq": float64(1), "inbox": "coordinator", "kind": "message", "task": "",
		"agent": "user", "text": "mail for two", "created_at": "2026-01-01T00:00:00Z",
		"read_at": "", "nudged_at": "", "nudged_generation": float64(0),
	}
	if !reflect.DeepEqual(coordinator, wantCoordinator) {
		t.Fatalf("coordinator record = %#v; want %#v", coordinator, wantCoordinator)
	}
	entries, err = os.ReadDir(filepath.Join(home, ".hive", "state", "inbox", "coordinator"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("coordinator inbox records = %d; want one record", len(entries))
	}

	if !fileMissing(t, home, "inbox", "manager-one") {
		t.Fatal("the old inbox of project one is still present")
	}
	if !fileMissing(t, home, "counters", "inbox:manager-one.json") {
		t.Fatal("the old counter of project one is still present")
	}
	counter := readJSONMap(t, home, "counters", "inbox:manager-two.json")
	if counter["value"] != float64(2) {
		t.Fatalf("owner inbox counter = %#v; want value 2", counter)
	}
}

// TestMigrationVersionThreeKeepsEachSourceDestinationOnRetry proves that a
// retry after a partial copy keeps the copied record in the owner inbox. The
// marked copy in manager-two is not a source record of project two.
func TestMigrationVersionThreeKeepsEachSourceDestinationOnRetry(t *testing.T) {
	home := seedOwnerAliasHome(t)
	seedState(t, home, []string{"inbox", "manager-two", "2.json"}, map[string]any{
		"seq": 2, "role": "manager-two", "kind": "message", "task": "",
		"agent": "user", "text": "mail for one", "created_at": "2026-01-01T00:00:00Z",
		"read_at": "", "nudged_at": "", "nudged_generation": 0,
		"migrated_from": "manager-one/1",
	})
	seedState(t, home, []string{"counters", "inbox:manager-two.json"}, map[string]any{"value": 2})

	if _, err := Open(home); err != nil {
		t.Fatalf("Open() = %v; want the migration", err)
	}
	owner := readJSONMap(t, home, "inbox", "manager-two", "2.json")
	wantOwner := map[string]any{
		"seq": float64(2), "inbox": "manager-two", "kind": "message", "task": "",
		"agent": "user", "text": "mail for one", "created_at": "2026-01-01T00:00:00Z",
		"read_at": "", "nudged_at": "", "nudged_generation": float64(0),
	}
	if !reflect.DeepEqual(owner, wantOwner) {
		t.Fatalf("owner record = %#v; want %#v", owner, wantOwner)
	}
	coordinator := readJSONMap(t, home, "inbox", "coordinator", "1.json")
	wantCoordinator := map[string]any{
		"seq": float64(1), "inbox": "coordinator", "kind": "message", "task": "",
		"agent": "user", "text": "mail for two", "created_at": "2026-01-01T00:00:00Z",
		"read_at": "", "nudged_at": "", "nudged_generation": float64(0),
	}
	if !reflect.DeepEqual(coordinator, wantCoordinator) {
		t.Fatalf("coordinator record = %#v; want %#v", coordinator, wantCoordinator)
	}
	entries, err := os.ReadDir(filepath.Join(home, ".hive", "state", "inbox", "coordinator"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("coordinator inbox records = %d; want one record", len(entries))
	}
	if !fileMissing(t, home, "inbox", "manager-one") {
		t.Fatal("the old inbox of project one is still present")
	}
	if !fileMissing(t, home, "counters", "inbox:manager-one.json") {
		t.Fatal("the old counter of project one is still present")
	}
}

// checkOwnerAliasDestinations proves the D2 and D7 result of the alias home:
// the owner inbox of project one holds only its own record, and the
// coordinator inbox holds only the record of project two.
func checkOwnerAliasDestinations(t *testing.T, home string) {
	t.Helper()
	owner := readJSONMap(t, home, "inbox", "manager-two", "2.json")
	wantOwner := map[string]any{
		"seq": float64(2), "inbox": "manager-two", "kind": "message", "task": "",
		"agent": "user", "text": "mail for one", "created_at": "2026-01-01T00:00:00Z",
		"read_at": "", "nudged_at": "", "nudged_generation": float64(0),
	}
	if !reflect.DeepEqual(owner, wantOwner) {
		t.Fatalf("owner record = %#v; want %#v", owner, wantOwner)
	}
	entries, err := os.ReadDir(filepath.Join(home, ".hive", "state", "inbox", "manager-two"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("owner inbox records = %d; want one record", len(entries))
	}
	coordinator := readJSONMap(t, home, "inbox", "coordinator", "1.json")
	wantCoordinator := map[string]any{
		"seq": float64(1), "inbox": "coordinator", "kind": "message", "task": "",
		"agent": "user", "text": "mail for two", "created_at": "2026-01-01T00:00:00Z",
		"read_at": "", "nudged_at": "", "nudged_generation": float64(0),
	}
	if !reflect.DeepEqual(coordinator, wantCoordinator) {
		t.Fatalf("coordinator record = %#v; want %#v", coordinator, wantCoordinator)
	}
	entries, err = os.ReadDir(filepath.Join(home, ".hive", "state", "inbox", "coordinator"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("coordinator inbox records = %d; want one record", len(entries))
	}
	if !fileMissing(t, home, "inbox", "manager-one") {
		t.Fatal("the old inbox of project one is still present")
	}
	if !fileMissing(t, home, "counters", "inbox:manager-one.json") {
		t.Fatal("the old counter of project one is still present")
	}
}

// TestMigrationVersionThreeKeepsEachSourceDestinationAfterTaskFailure proves
// that a retry after a failed task step keeps the destination of each moved
// record. The failed run copies the mail and clears the markers, so the retry
// must not read the owner record as a source record of project two.
func TestMigrationVersionThreeKeepsEachSourceDestinationAfterTaskFailure(t *testing.T) {
	home := seedOwnerAliasHome(t)
	taskPath := filepath.Join(home, ".hive", "state", "tasks", "1-bad.json")
	seedState(t, home, []string{"tasks", "1-bad.json"}, []any{})

	_, err := Open(home)
	if err == nil || err.Error() != "record invalid: "+taskPath {
		t.Fatalf("Open() = %v; want record invalid: %s", err, taskPath)
	}
	if version, ok := readSchemaVersion(t, home); !ok || version != 2 {
		t.Fatalf("schema version = %d, present %t; want 2", version, ok)
	}
	owner := readJSONMap(t, home, "inbox", "manager-two", "2.json")
	if owner["text"] != "mail for one" || owner["inbox"] != "manager-two" {
		t.Fatalf("owner record = %#v; want mail for one in the owner inbox", owner)
	}
	coordinator := readJSONMap(t, home, "inbox", "coordinator", "1.json")
	if coordinator["text"] != "mail for two" || coordinator["inbox"] != "coordinator" {
		t.Fatalf("coordinator record = %#v; want mail for two in the coordinator inbox", coordinator)
	}

	seedState(t, home, []string{"tasks", "1-bad.json"}, map[string]any{
		"id": "1-bad", "project": "one", "title": "Task", "controlling_role": "manager-one",
		"status": "created", "acting_agent": "", "depends_on": []string{},
	})
	if _, err := Open(home); err != nil {
		t.Fatalf("Open() after the repair = %v; want the migration", err)
	}
	if version, ok := readSchemaVersion(t, home); !ok || version != 4 {
		t.Fatalf("schema version = %d, present %t; want 4", version, ok)
	}
	checkOwnerAliasDestinations(t, home)
	task := readJSONMap(t, home, "tasks", "1-bad.json")
	if _, ok := task["controlling_role"]; ok {
		t.Fatalf("task = %#v; want no controlling_role", task)
	}
	if !fileMissing(t, home, "migrate-v3.json") {
		t.Fatal("the version 3 plan is still present")
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

// TestMigrationVersionThreeKeepsEachSourceDestinationAfterMarkerFailure
// proves that a failure during the marker cleanup keeps the destination of a
// record whose marker is already gone. The destinations of the failed run
// decide the retry, so the owner record is not a source record of project two.
func TestMigrationVersionThreeKeepsEachSourceDestinationAfterMarkerFailure(t *testing.T) {
	home := seedOwnerAliasHome(t)
	badPath := filepath.Join(home, ".hive", "state", "inbox", "zzz", "1.json")
	seedState(t, home, []string{"inbox", "zzz", "1.json"}, []any{})

	_, err := Open(home)
	if err == nil || err.Error() != "record invalid: "+badPath {
		t.Fatalf("Open() = %v; want record invalid: %s", err, badPath)
	}
	if version, ok := readSchemaVersion(t, home); !ok || version != 2 {
		t.Fatalf("schema version = %d, present %t; want 2", version, ok)
	}
	owner := readJSONMap(t, home, "inbox", "manager-two", "2.json")
	if _, ok := owner["migrated_from"]; ok {
		t.Fatalf("owner record = %#v; want the marker cleanup before the failing folder", owner)
	}

	if err := os.RemoveAll(filepath.Join(home, ".hive", "state", "inbox", "zzz")); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(home); err != nil {
		t.Fatalf("Open() after the repair = %v; want the migration", err)
	}
	if version, ok := readSchemaVersion(t, home); !ok || version != 4 {
		t.Fatalf("schema version = %d, present %t; want 4", version, ok)
	}
	checkOwnerAliasDestinations(t, home)
	if !fileMissing(t, home, "migrate-v3.json") {
		t.Fatal("the version 3 plan is still present")
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

// TestMigrationVersionThreeResumesFromTheSourceMarker proves that a retry
// after a partial copy uses the source marker of the moved record and copies
// each remaining unread source record once.
func TestMigrationVersionThreeResumesFromTheSourceMarker(t *testing.T) {
	home := seedEqualUnreadHome(t)
	seedState(t, home, []string{"inbox", "owner-demo", "1.json"}, map[string]any{
		"seq": 1, "role": "owner-demo", "kind": "message", "task": "",
		"agent": "user", "text": "same text", "created_at": "2026-01-01T00:00:00Z", "read_at": "",
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
	if len(entries) != 2 {
		t.Fatalf("owner inbox records = %d; want the first copy and the remaining record", len(entries))
	}
	for _, entry := range entries {
		record := readJSONMap(t, home, "inbox", "owner-demo", entry.Name())
		if _, ok := record["migrated_from"]; ok {
			t.Fatalf("record %s = %#v; want no migration marker", entry.Name(), record)
		}
		if record["inbox"] != "owner-demo" || record["text"] != "same text" {
			t.Fatalf("record %s = %#v; want the owner copy", entry.Name(), record)
		}
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
	if version, ok := readSchemaVersion(t, home); !ok || version != 4 {
		t.Fatalf("schema version = %d, present %t; want 4", version, ok)
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
