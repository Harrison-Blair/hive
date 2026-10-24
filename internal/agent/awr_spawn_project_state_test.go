package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

const fixedRegisteredAt = "2000-01-02T03:04:05Z"

// writeFixedProject writes the project fixture through the typed records.Project
// writer with one fixed synthetic registered_at value.
func writeFixedProject(t *testing.T, home string) string {
	t.Helper()
	projectDirectory(t, home, "demo")
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	project := records.Project{ID: "demo", Path: managerRepoPath, Base: "dev", RegisteredAt: fixedRegisteredAt}
	if err := registry.WriteIn([]string{"projects"}, "demo", project); err != nil {
		t.Fatal(err)
	}
	return filepath.Join(home, ".hive", "state", "projects", "demo.json")
}

// TestSpawnProjectPlacementKeepsRegisteredAt proves that a project placement
// keeps the id, path, base, and registered_at of the project record and adds
// only the workspace field.
func TestSpawnProjectPlacementKeepsRegisteredAt(t *testing.T) {
	home := testHome(t)
	path := writeFixedProject(t, home)
	before := readProjectJSON(t, path)
	dir := projectDirectory(t, home, "demo")
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	fake.result("workspace.create", managerWorkspaceCreateJSON("wBD", "hive:demo", dir, "wBD:p1"))
	scriptManagerSpawnTail(t, home, fake, sleep, "wBD:p1", "wBD", "wBD:t1")

	if envelope := Spawn(context.Background(), home, managerSpawnOptions("demo")); envelope.Outcome != "ok" {
		t.Fatalf("envelope = %#v; want ok", envelope)
	}
	project := readProjectWorkspaceRecord(t, home, "demo")
	if project.ID != "demo" || project.Path != managerRepoPath || project.Base != "dev" || project.RegisteredAt != fixedRegisteredAt {
		t.Fatalf("project fields = %#v; want id, path, base, and registered_at kept", project)
	}
	if project.Workspace == nil || project.Workspace.ID != "wBD" {
		t.Fatalf("project workspace = %#v; want wBD", project.Workspace)
	}
	after := readProjectJSON(t, path)
	delete(after, "workspace")
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("placement changed project state other than workspace: before=%v after=%v", before, after)
	}
}

// TestSpawnProjectCleanupKeepsExactState proves that a failed start closes the
// workspace and leaves the project record byte-for-byte equal to its state
// before placement.
func TestSpawnProjectCleanupKeepsExactState(t *testing.T) {
	home := testHome(t)
	path := writeFixedProject(t, home)
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	dir := projectDirectory(t, home, "demo")
	fake := newFakeHerdr(t)
	fake.result("workspace.create", managerWorkspaceCreateJSON("wBD", "hive:demo", dir, "wBD:p1"))
	fake.result("workspace.close", `{"type":"ok"}`)
	scriptReadyShell(fake, "wBD:p1", "term_wBD", "wBD", "wBD:t1", dir)
	fake.failure("agent.start", "invalid_agent_name", "name is too long")

	if envelope := Spawn(context.Background(), home, managerSpawnOptions("demo")); envelope.Outcome != "failed" || envelope.Phase != "start" {
		t.Fatalf("envelope = %#v; want failed start", envelope)
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatalf("cleanup changed project state:\nbefore=%s\nafter=%s", before, after)
	}
	project := readProjectWorkspaceRecord(t, home, "demo")
	if project.ID != "demo" || project.Path != managerRepoPath || project.Base != "dev" || project.RegisteredAt != fixedRegisteredAt || project.Workspace != nil {
		t.Fatalf("project after cleanup = %#v; want id, path, base, and registered_at kept with no workspace", project)
	}
}

func readProjectJSON(t *testing.T, path string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]any
	if err := json.Unmarshal(data, &fields); err != nil {
		t.Fatal(err)
	}
	return fields
}
