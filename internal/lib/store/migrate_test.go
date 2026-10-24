package store

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
)

func seedState(t *testing.T, home string, parts []string, value any) string {
	t.Helper()
	path := filepath.Join(append([]string{home, ".hive", "state"}, parts...)...)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func seedHive(t *testing.T, home string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(home, ".hive"), 0o755); err != nil {
		t.Fatal(err)
	}
}

func readAgentMap(t *testing.T, home, key string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(home, ".hive", "state", "agents", key+".json"))
	if err != nil {
		t.Fatal(err)
	}
	var record map[string]any
	if err := json.Unmarshal(data, &record); err != nil {
		t.Fatal(err)
	}
	return record
}

func readSchemaVersion(t *testing.T, home string) (int, bool) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(home, ".hive", "state", "schema.json"))
	if os.IsNotExist(err) {
		return 0, false
	}
	if err != nil {
		t.Fatal(err)
	}
	var schema struct {
		Version int `json:"version"`
	}
	if err := json.Unmarshal(data, &schema); err != nil {
		t.Fatal(err)
	}
	return schema.Version, true
}

func stateFingerprints(t *testing.T, home string) []string {
	t.Helper()
	root := filepath.Join(home, ".hive", "state")
	found := []string{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(data)
		rel, err := filepath.Rel(home, path)
		if err != nil {
			return err
		}
		found = append(found, filepath.ToSlash(rel)+" "+hex.EncodeToString(sum[:]))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(found)
	return found
}

func migrationLogEvents(t *testing.T, home string) []map[string]any {
	t.Helper()
	dir := filepath.Join(home, ".hive", "logs")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read migration log: %v", err)
	}
	events := []map[string]any{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".jsonl") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
			if line == "" {
				continue
			}
			var event map[string]any
			if err := json.Unmarshal([]byte(line), &event); err != nil {
				t.Fatalf("log line = %q: %v", line, err)
			}
			events = append(events, event)
		}
	}
	return events
}

func findMigrationEvent(t *testing.T, events []map[string]any, message string) map[string]any {
	t.Helper()
	for _, event := range events {
		if event["msg"] == message {
			return event
		}
	}
	t.Fatalf("the log has no %q event: %#v", message, events)
	return nil
}

// TestMigrationVersionTwoSetsAgentFacts gives one old-shape record to each row
// of the version 2 table of D2. It also shows that a second open changes no
// file.
func TestMigrationVersionTwoSetsAgentFacts(t *testing.T) {
	cases := []struct {
		name         string
		role         string
		project      string
		worktree     string
		wantProfile  string
		wantKind     string
		wantOwns     string
		wantWorktree string
	}{
		{name: "coordinator", role: "coordinator", wantProfile: "coordinator", wantKind: "coordinator"},
		{name: "supervisor", role: "supervisor", wantKind: "supervisor"},
		{name: "manager", role: "manager", project: "hive", wantProfile: "planner", wantOwns: "hive"},
		{name: "worker", role: "worker", project: "hive", worktree: "/tmp/worktrees/1-do", wantProfile: "builder", wantWorktree: "/tmp/worktrees/1-do"},
		{name: "verifier", role: "verifier", project: "hive", worktree: "/tmp/worktrees/1-do", wantProfile: "verifier", wantWorktree: "/tmp/worktrees/1-do"},
		{name: "agent", role: "agent", wantProfile: "general"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			home := t.TempDir()
			seedHive(t, home)
			task := ""
			if c.worktree != "" {
				task = "1-do"
				seedState(t, home, []string{"tasks", task + ".json"}, map[string]any{
					"id": task, "project": c.project, "title": "do",
					"worktree": map[string]any{"path": c.worktree},
				})
			}
			key := c.name + "-g1"
			seedState(t, home, []string{"agents", key + ".json"}, map[string]any{
				"key": key, "name": c.name, "role": c.role,
				"project": c.project, "task": task, "status": "live", "generation": 1,
			})

			if _, err := Open(home); err != nil {
				t.Fatalf("Open() = %v; want the migration", err)
			}
			record := readAgentMap(t, home, key)
			for field, want := range map[string]string{
				"role":     c.role,
				"profile":  c.wantProfile,
				"kind":     c.wantKind,
				"owns":     c.wantOwns,
				"worktree": c.wantWorktree,
			} {
				if got, ok := record[field]; !ok || got != want {
					t.Errorf("%s = %#v, present %t; want %q", field, got, ok, want)
				}
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
		})
	}
}

// TestMigrationWaitsOnceOnAHeldLock proves that repeated store opens do not
// each wait for a held lock. The first open uses the two-second wait. Each
// later open tries once. After the holder releases the lock, the next open
// completes the migration.
func TestMigrationWaitsOnceOnAHeldLock(t *testing.T) {
	home := t.TempDir()
	seedHive(t, home)
	seedState(t, home, []string{"agents", "manager-hive-g1.json"}, map[string]any{
		"key": "manager-hive-g1", "name": "manager-hive", "role": "manager", "project": "hive", "status": "live", "generation": 1,
	})
	release := holdStateLock(t, home)
	defer release()

	started := time.Now()
	for attempt := 1; attempt <= 3; attempt++ {
		if _, err := Open(home); !errors.Is(err, ErrLockTimeout) {
			t.Fatalf("Open() attempt %d = %v; want ErrLockTimeout", attempt, err)
		}
	}
	if elapsed := time.Since(started); elapsed > 5*time.Second {
		t.Fatalf("three opens with a held lock took %v; want one bounded wait", elapsed)
	}

	release()
	if _, err := Open(home); err != nil {
		t.Fatalf("Open() after release = %v; want the migration", err)
	}
	if version, ok := readSchemaVersion(t, home); !ok || version != 3 {
		t.Fatalf("schema version = %d, present %t; want 3", version, ok)
	}
}

// TestMigrationFailedStepKeepsVersionOne shows AC3: a failed step writes no
// version, and the next open completes the step.
func TestMigrationFailedStepKeepsVersionOne(t *testing.T) {
	home := t.TempDir()
	seedHive(t, home)
	seedState(t, home, []string{"agents", "manager-hive-g1.json"}, map[string]any{
		"key": "manager-hive-g1", "name": "manager-hive", "role": "manager", "project": "hive", "status": "live", "generation": 1,
	})
	target := filepath.Join(home, "target.json")
	if err := os.WriteFile(target, []byte(`{"key":"bad"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(home, ".hive", "state", "agents", "bad.json")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}

	if _, err := Open(home); err == nil {
		t.Fatal("Open() = nil error; want the failed step")
	}
	if version, ok := readSchemaVersion(t, home); ok {
		t.Fatalf("schema version = %d; want no schema after a failed step", version)
	}
	if err := os.Remove(link); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(home); err != nil {
		t.Fatalf("Open() after the repair = %v; want success", err)
	}
	if version, ok := readSchemaVersion(t, home); !ok || version != 3 {
		t.Fatalf("schema version = %d, present %t; want 3", version, ok)
	}
	if record := readAgentMap(t, home, "manager-hive-g1"); record["owns"] != "hive" || record["profile"] != "planner" {
		t.Fatalf("record = %#v; want the completed step", record)
	}
}

// TestMigrationWritesLogEvents proves AC4: the four events of D2 carry their
// fields, including the failed event.
func TestMigrationWritesLogEvents(t *testing.T) {
	home := t.TempDir()
	seedHive(t, home)
	seedState(t, home, []string{"agents", "worker-hive-g1.json"}, map[string]any{
		"key": "worker-hive-g1", "name": "worker-hive", "role": "worker", "project": "hive", "status": "live", "generation": 1,
	})
	if _, err := Open(home); err != nil {
		t.Fatalf("Open() = %v", err)
	}
	events := migrationLogEvents(t, home)
	start := findMigrationEvent(t, events, "state migrate start")
	if start["component"] != "store" || start["from"] != float64(1) || start["to"] != float64(2) {
		t.Fatalf("start event = %#v; want store 1 to 2", start)
	}
	record := findMigrationEvent(t, events, "state migrate record")
	if record["component"] != "store" || record["path"] != ".hive/state/agents/worker-hive-g1.json" || record["change"] == "" {
		t.Fatalf("record event = %#v; want the path and a change", record)
	}
	done := findMigrationEvent(t, events, "state migrate done")
	if done["component"] != "store" || done["from"] != float64(1) || done["to"] != float64(2) || done["count"] != float64(1) {
		t.Fatalf("done event = %#v; want 1 to 2 with count 1", done)
	}
	if _, ok := done["duration_ms"]; !ok {
		t.Fatalf("done event = %#v; want duration_ms", done)
	}

	failedHome := t.TempDir()
	seedHive(t, failedHome)
	seedState(t, failedHome, []string{"agents", "worker-hive-g1.json"}, map[string]any{
		"key": "worker-hive-g1", "name": "worker-hive", "role": "worker", "project": "hive", "status": "live", "generation": 1,
	})
	bad := filepath.Join(failedHome, ".hive", "state", "agents", "bad.json")
	if err := os.Symlink(filepath.Join(failedHome, "missing.json"), bad); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(failedHome); err == nil {
		t.Fatal("Open() = nil error; want the failed step")
	}
	failed := findMigrationEvent(t, migrationLogEvents(t, failedHome), "state migrate failed")
	if failed["component"] != "store" || failed["from"] != float64(1) || failed["to"] != float64(2) || failed["path"] == "" || failed["error"] == "" {
		t.Fatalf("failed event = %#v; want from, to, path, and error", failed)
	}
}

// TestMigrationOfCopiedHomeShapes uses one record of each shape of the real
// home with scrubbed text values. It reads no file of the real home.
func TestMigrationOfCopiedHomeShapes(t *testing.T) {
	home := t.TempDir()
	seedHive(t, home)
	task := map[string]any{
		"id": "1-shape", "project": "demo", "title": "shape", "status": "ready",
		"worktree": map[string]any{"path": "/scratch/worktrees/1-shape"},
	}
	seedState(t, home, []string{"tasks", "1-shape.json"}, task)
	claim := map[string]any{"key": "manager-demo", "agent": "manager-demo-g1", "generation": 1, "pid": 1, "pid_start": 1}
	seedState(t, home, []string{"claims", "manager-demo.json"}, claim)
	inbox := map[string]any{"seq": 1, "role": "manager-demo", "kind": "message", "task": "", "agent": "t1-worker-1", "text": "shape"}
	seedState(t, home, []string{"inbox", "manager-demo", "1.json"}, inbox)
	shapes := []map[string]any{
		{"key": "coordinator-g1", "name": "coordinator", "role": "coordinator", "project": "", "task": "", "harness": "claude", "model": "m", "pane_id": "w1:p1", "workspace_id": "w1", "tab_id": "w1:t1", "terminal_id": "term_1", "cwd": "/scratch/home", "status": "stale", "generation": 1, "started_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z", "spawned": false, "started_by_hive": true},
		{"key": "supervisor-g1", "name": "supervisor", "role": "supervisor", "project": "", "task": "", "pane_id": "w1:p2", "status": "stale", "generation": 1, "started_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z", "spawned": false},
		{"key": "manager-demo-g1", "name": "manager-demo", "role": "manager", "project": "demo", "task": "", "pane_id": "w1:p3", "status": "stale", "generation": 1, "started_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z", "spawned": true},
		{"key": "t1-worker-1", "name": "t1-worker-1", "role": "worker", "project": "demo", "task": "1-shape", "pane_id": "w1:p4", "status": "stale", "generation": 1, "started_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z", "spawned": true},
		{"key": "t1-verifier-1", "name": "t1-verifier-1", "role": "verifier", "project": "demo", "task": "1-shape", "pane_id": "w1:p5", "status": "stale", "generation": 1, "started_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z", "spawned": true},
		{"key": "helper", "name": "helper", "role": "agent", "project": "", "task": "", "pane_id": "w1:p6", "status": "live", "generation": 1, "started_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z", "spawned": true},
	}
	for _, shape := range shapes {
		seedState(t, home, []string{"agents", shape["key"].(string) + ".json"}, shape)
	}

	if _, err := Open(home); err != nil {
		t.Fatalf("Open() = %v; want the migration", err)
	}
	want := map[string]map[string]string{
		"coordinator-g1":  {"profile": "coordinator", "kind": "coordinator", "owns": "", "role": "coordinator"},
		"supervisor-g1":   {"profile": "", "kind": "supervisor", "owns": "", "role": "supervisor"},
		"manager-demo-g1": {"profile": "planner", "kind": "", "owns": "demo", "role": "manager"},
		"t1-worker-1":     {"profile": "builder", "kind": "", "owns": "", "role": "worker"},
		"t1-verifier-1":   {"profile": "verifier", "kind": "", "owns": "", "role": "verifier"},
		"helper":          {"profile": "general", "kind": "", "owns": "", "role": "agent"},
	}
	for key, fields := range want {
		record := readAgentMap(t, home, key)
		for field, value := range fields {
			if record[field] != value {
				t.Errorf("%s %s = %#v; want %q", key, field, record[field], value)
			}
		}
	}
	if record := readAgentMap(t, home, "t1-worker-1"); record["worktree"] != "/scratch/worktrees/1-shape" {
		t.Fatalf("worker worktree = %#v; want the task worktree", record["worktree"])
	}
	if record := readAgentMap(t, home, "t1-verifier-1"); record["worktree"] != "/scratch/worktrees/1-shape" {
		t.Fatalf("verifier worktree = %#v; want the task worktree", record["worktree"])
	}
	data, err := os.ReadFile(filepath.Join(home, ".hive", "state", "claims", "owner-demo.json"))
	if err != nil {
		t.Fatal(err)
	}
	var gotClaim map[string]any
	if err := json.Unmarshal(data, &gotClaim); err != nil {
		t.Fatal(err)
	}
	if gotClaim["key"] != "owner-demo" {
		t.Fatalf("claim = %#v; want version 3 to rename the claim", gotClaim)
	}
	if _, err := os.Stat(filepath.Join(home, ".hive", "state", "claims", "manager-demo.json")); !os.IsNotExist(err) {
		t.Fatalf("the old manager claim is still present: %v", err)
	}
}
