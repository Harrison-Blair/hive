package agent

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

func managerSpawnOptions(project string) SpawnOptions {
	return SpawnOptions{Role: "manager", Project: project, Harness: "claude", Model: "m"}
}

func projectDirectory(t *testing.T, home, id string) string {
	t.Helper()
	dir := filepath.Join(home, "projects", id)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	resolved, err := filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatal(err)
	}
	return resolved
}

func writeProjectWorkspaceRecord(t *testing.T, home, id string, workspace *records.ProjectWorkspace) {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	record := records.Project{ID: id, Path: "/tmp/" + id, Base: "dev", RegisteredAt: "2026-01-01T00:00:00Z", Workspace: workspace}
	if err := registry.WriteIn([]string{"projects"}, id, record); err != nil {
		t.Fatal(err)
	}
}

func readProjectWorkspaceRecord(t *testing.T, home, id string) records.Project {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	var project records.Project
	if err := registry.ReadIn([]string{"projects"}, id, &project); err != nil {
		t.Fatal(err)
	}
	return project
}

func managerWorkspaceJSON(id, label string) string {
	return `{"workspace_id":"` + id + `","number":1,"label":"` + label + `","focused":false,"pane_count":1,"tab_count":1,"active_tab_id":"` + id + `:t1","agent_status":"unknown"}`
}

func managerWorkspaceGetJSON(id, label string) string {
	return `{"type":"workspace_info","workspace":` + managerWorkspaceJSON(id, label) + `}`
}

func managerWorkspaceListJSON(items ...string) string {
	return `{"type":"workspace_list","workspaces":[` + strings.Join(items, ",") + `]}`
}

func managerTabJSON(workspaceID, tabID string, number, paneCount int) string {
	return `{"tab_id":"` + tabID + `","workspace_id":"` + workspaceID + `","number":` + jsonNumber(int64(number)) + `,"label":"1","focused":false,"pane_count":` + jsonNumber(int64(paneCount)) + `,"agent_status":"unknown"}`
}

func managerTabListJSON(tabs ...string) string {
	return `{"type":"tab_list","tabs":[` + strings.Join(tabs, ",") + `]}`
}

func managerPaneListJSON(panes ...string) string {
	return `{"type":"pane_list","panes":[` + strings.Join(panes, ",") + `]}`
}

// managerPaneInfoJSON is one pane with the unknown status and optional extra
// fields, such as restore_error or agent.
func managerPaneInfoJSON(paneID, terminalID, workspaceID, tabID, cwd, extra string) string {
	return `{"pane_id":"` + paneID + `","terminal_id":"` + terminalID + `","workspace_id":"` + workspaceID + `","tab_id":"` + tabID + `","focused":false,"agent_status":"unknown","revision":1,"cwd":"` + cwd + `"` + extra + `}`
}

func managerPaneStatusJSON(paneID, terminalID, workspaceID, tabID, cwd, status, extra string) string {
	return `{"pane_id":"` + paneID + `","terminal_id":"` + terminalID + `","workspace_id":"` + workspaceID + `","tab_id":"` + tabID + `","focused":false,"agent_status":"` + status + `","revision":1,"cwd":"` + cwd + `"` + extra + `}`
}

func managerWorkspaceCreateJSON(workspaceID, label, cwd, paneID string) string {
	return `{"type":"workspace_created","workspace":` + managerWorkspaceJSON(workspaceID, label) + `,"tab":` + managerTabJSON(workspaceID, workspaceID+":t1", 1, 1) + `,"root_pane":` + managerPaneInfoJSON(paneID, "term_"+workspaceID, workspaceID, workspaceID+":t1", cwd, "") + `}`
}

// scriptManagerAgent scripts the spawn tail after the pane choice.
func scriptManagerAgent(t *testing.T, fake *fakeHerdr, sleep sleepProcess, paneID, workspaceID, key, cwd string) {
	t.Helper()
	tabID := workspaceID + ":t1"
	fake.result("agent.start", `{"type":"agent_started","agent":`+agentJSON(paneID, "term_x", workspaceID, tabID, "unknown", key)+`,"argv":["claude","--model","m"]}`)
	fake.result("pane.get", `{"type":"pane_info","pane":`+managerPaneInfoJSON(paneID, "term_x", workspaceID, tabID, cwd, "")+`}`)
	fake.result("agent.wait", `{"type":"agent_info","agent":`+agentJSON(paneID, "term_x", workspaceID, tabID, "idle", key)+`}`)
	fake.result("agent.get", `{"type":"agent_info","agent":`+agentJSON(paneID, "term_x", workspaceID, tabID, "idle", key)+`}`)
	fake.result("pane.process_info", `{"type":"pane_process_info","process_info":`+processInfoJSON(paneID, processEntry(sleep.pid, "claude", "claude"))+`}`)
}

func TestManagerWorkspaceReuseOK(t *testing.T) {
	home := testHome(t)
	dir := projectDirectory(t, home, "demo")
	writeProjectWorkspaceRecord(t, home, "demo", &records.ProjectWorkspace{ID: "w5", RootPane: "w5:p1", CreatedByHive: true, CreatedAt: "2026-01-01T00:00:00Z"})
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	fake.result("workspace.get", managerWorkspaceGetJSON("w5", "hive:demo"))
	fake.result("pane.get", `{"type":"pane_info","pane":`+managerPaneInfoJSON("w5:p1", "term_w5", "w5", "w5:t1", dir, "")+`}`)
	fake.result("tab.list", managerTabListJSON(managerTabJSON("w5", "w5:t1", 1, 1)))
	scriptManagerAgent(t, fake, sleep, "w5:p1", "w5", "manager-demo-g1", dir)
	envelope := Spawn(context.Background(), home, managerSpawnOptions("demo"))
	if envelope.Outcome != "ok" || envelope.Error != nil {
		t.Fatalf("envelope = %#v; want ok", envelope)
	}
	if len(fake.recorded("workspace.create")) != 0 || len(fake.recorded("workspace.close")) != 0 {
		t.Fatalf("reuse made or closed a workspace")
	}
	project := readProjectWorkspaceRecord(t, home, "demo")
	if project.Workspace == nil || project.Workspace.ID != "w5" || project.Workspace.RootPane != "w5:p1" {
		t.Fatalf("project workspace = %#v", project.Workspace)
	}
}

func TestManagerWorkspaceReuseGone(t *testing.T) {
	home := testHome(t)
	dir := projectDirectory(t, home, "demo")
	writeProjectWorkspaceRecord(t, home, "demo", &records.ProjectWorkspace{ID: "w5", RootPane: "w5:p1", CreatedByHive: true, CreatedAt: "2026-01-01T00:00:00Z"})
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	fake.failure("workspace.get", "workspace_not_found", "workspace w5 not found")
	fake.result("workspace.list", managerWorkspaceListJSON())
	fake.result("workspace.create", managerWorkspaceCreateJSON("w6", "hive:demo", dir, "w6:p1"))
	scriptManagerAgent(t, fake, sleep, "w6:p1", "w6", "manager-demo-g1", dir)
	envelope := Spawn(context.Background(), home, managerSpawnOptions("demo"))
	if envelope.Outcome != "ok" || envelope.Error != nil {
		t.Fatalf("envelope = %#v; want ok", envelope)
	}
	project := readProjectWorkspaceRecord(t, home, "demo")
	if project.Workspace == nil || project.Workspace.ID != "w6" || !project.Workspace.CreatedByHive {
		t.Fatalf("project workspace = %#v", project.Workspace)
	}
	creates := fake.recorded("workspace.create")
	if len(creates) != 1 || !strings.Contains(creates[0].Params, `"cwd":"`+dir+`"`) {
		t.Fatalf("workspace.create = %#v", creates)
	}
}

func TestManagerWorkspaceReuseOtherDirectory(t *testing.T) {
	home := testHome(t)
	dir := projectDirectory(t, home, "demo")
	writeProjectWorkspaceRecord(t, home, "demo", &records.ProjectWorkspace{ID: "w5", RootPane: "w5:p1", CreatedByHive: true, CreatedAt: "2026-01-01T00:00:00Z"})
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	fake.result("workspace.get", managerWorkspaceGetJSON("w5", "hive:demo"))
	fake.result("pane.get", `{"type":"pane_info","pane":`+managerPaneInfoJSON("w5:p1", "term_w5", "w5", "w5:t1", "/tmp/elsewhere", "")+`}`)
	fake.result("workspace.list", managerWorkspaceListJSON())
	fake.result("workspace.create", managerWorkspaceCreateJSON("w6", "hive:demo", dir, "w6:p1"))
	scriptManagerAgent(t, fake, sleep, "w6:p1", "w6", "manager-demo-g1", dir)
	envelope := Spawn(context.Background(), home, managerSpawnOptions("demo"))
	if envelope.Outcome != "ok" {
		t.Fatalf("envelope = %#v; want ok", envelope)
	}
	project := readProjectWorkspaceRecord(t, home, "demo")
	if project.Workspace == nil || project.Workspace.ID != "w6" {
		t.Fatalf("project workspace = %#v", project.Workspace)
	}
}

func TestManagerWorkspaceReuseLabelDiffers(t *testing.T) {
	home := testHome(t)
	dir := projectDirectory(t, home, "demo")
	writeProjectWorkspaceRecord(t, home, "demo", &records.ProjectWorkspace{ID: "w5", RootPane: "w5:p1", CreatedByHive: true, CreatedAt: "2026-01-01T00:00:00Z"})
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	fake.result("workspace.get", managerWorkspaceGetJSON("w5", "other"))
	fake.result("workspace.list", managerWorkspaceListJSON())
	fake.result("workspace.create", managerWorkspaceCreateJSON("w6", "hive:demo", dir, "w6:p1"))
	scriptManagerAgent(t, fake, sleep, "w6:p1", "w6", "manager-demo-g1", dir)
	envelope := Spawn(context.Background(), home, managerSpawnOptions("demo"))
	if envelope.Outcome != "ok" {
		t.Fatalf("envelope = %#v; want ok", envelope)
	}
	if project := readProjectWorkspaceRecord(t, home, "demo"); project.Workspace == nil || project.Workspace.ID != "w6" {
		t.Fatalf("project workspace = %#v", project.Workspace)
	}
}

func TestManagerWorkspaceNoRecordedRoot(t *testing.T) {
	home := testHome(t)
	writeProjectWorkspaceRecord(t, home, "demo", &records.ProjectWorkspace{ID: "w5", RootPane: "", CreatedByHive: true, CreatedAt: "2026-01-01T00:00:00Z"})
	fake := newFakeHerdr(t)
	fake.result("workspace.get", managerWorkspaceGetJSON("w5", "hive:demo"))
	envelope := Spawn(context.Background(), home, managerSpawnOptions("demo"))
	if envelope.Outcome != "failed" {
		t.Fatalf("envelope = %#v; want failed", envelope)
	}
	if !strings.Contains(toString(envelope.Error), "workspace w5 has no recorded root pane: pass --pane <shell pane>") {
		t.Fatalf("error = %v", envelope.Error)
	}
	if len(fake.recorded("pane.get")) != 0 || len(fake.recorded("agent.start")) != 0 {
		t.Fatalf("refusal sent a request")
	}
}

func TestManagerWorkspaceMovedRoot(t *testing.T) {
	home := testHome(t)
	dir := projectDirectory(t, home, "demo")
	writeProjectWorkspaceRecord(t, home, "demo", &records.ProjectWorkspace{ID: "w5", RootPane: "w5:p1", CreatedByHive: true, CreatedAt: "2026-01-01T00:00:00Z"})
	fake := newFakeHerdr(t)
	fake.result("workspace.get", managerWorkspaceGetJSON("w5", "hive:demo"))
	fake.result("pane.get", `{"type":"pane_info","pane":`+managerPaneInfoJSON("w5:p9", "term_w5", "w5", "w5:t2", dir, "")+`}`)
	fake.result("tab.list", managerTabListJSON(managerTabJSON("w5", "w5:t1", 1, 1)))
	envelope := Spawn(context.Background(), home, managerSpawnOptions("demo"))
	if envelope.Outcome != "failed" {
		t.Fatalf("envelope = %#v; want failed", envelope)
	}
	if !strings.Contains(toString(envelope.Error), "workspace w5 has no recorded root pane: pass --pane <shell pane>") {
		t.Fatalf("error = %v", envelope.Error)
	}
	if len(fake.recorded("agent.start")) != 0 || len(fake.recorded("workspace.close")) != 0 {
		t.Fatalf("moved root sent a request")
	}
}

func TestManagerWorkspaceUserTabStays(t *testing.T) {
	home := testHome(t)
	dir := projectDirectory(t, home, "demo")
	writeProjectWorkspaceRecord(t, home, "demo", &records.ProjectWorkspace{ID: "w5", RootPane: "w5:p1", CreatedByHive: true, CreatedAt: "2026-01-01T00:00:00Z"})
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	fake.result("workspace.get", managerWorkspaceGetJSON("w5", "hive:demo"))
	fake.result("pane.get", `{"type":"pane_info","pane":`+managerPaneInfoJSON("w5:p1", "term_w5", "w5", "w5:t1", dir, "")+`}`)
	fake.result("tab.list", managerTabListJSON(managerTabJSON("w5", "w5:t1", 1, 1), managerTabJSON("w5", "w5:t2", 2, 1)))
	scriptManagerAgent(t, fake, sleep, "w5:p1", "w5", "manager-demo-g1", dir)
	envelope := Spawn(context.Background(), home, managerSpawnOptions("demo"))
	if envelope.Outcome != "ok" {
		t.Fatalf("envelope = %#v; want ok", envelope)
	}
	if len(fake.recorded("workspace.close")) != 0 || len(fake.recorded("tab.close")) != 0 {
		t.Fatalf("reuse closed the user tab")
	}
}

func TestManagerWorkspaceOwnerOneMatch(t *testing.T) {
	home := testHome(t)
	writeProjectWorkspaceRecord(t, home, "demo", nil)
	fake := newFakeHerdr(t)
	fake.result("workspace.list", managerWorkspaceListJSON(managerWorkspaceJSON("w7", "hive:demo")))
	envelope := Spawn(context.Background(), home, managerSpawnOptions("demo"))
	if envelope.Outcome != "failed" {
		t.Fatalf("envelope = %#v; want failed", envelope)
	}
	want := "workspace w7 has the label hive:demo but hive did not make it: close it or pass --adopt w7"
	if !strings.Contains(toString(envelope.Error), want) {
		t.Fatalf("error = %v; want %s", envelope.Error, want)
	}
	if len(fake.recorded("workspace.create")) != 0 {
		t.Fatalf("made a workspace with an owner present")
	}
}

func TestManagerWorkspaceOwnerTwoMatches(t *testing.T) {
	home := testHome(t)
	writeProjectWorkspaceRecord(t, home, "demo", nil)
	fake := newFakeHerdr(t)
	fake.result("workspace.list", managerWorkspaceListJSON(managerWorkspaceJSON("w7", "hive:demo"), managerWorkspaceJSON("w8", "hive:demo")))
	envelope := Spawn(context.Background(), home, managerSpawnOptions("demo"))
	if envelope.Outcome != "failed" {
		t.Fatalf("envelope = %#v; want failed", envelope)
	}
	if !strings.Contains(toString(envelope.Error), "pass --adopt w7") {
		t.Fatalf("error = %v; want the first match", envelope.Error)
	}
}

func TestManagerWorkspaceAdoptOnePane(t *testing.T) {
	home := testHome(t)
	dir := projectDirectory(t, home, "demo")
	writeProjectWorkspaceRecord(t, home, "demo", nil)
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	fake.result("workspace.get", managerWorkspaceGetJSON("w5", "hive:demo"))
	fake.result("tab.list", managerTabListJSON(managerTabJSON("w5", "w5:t1", 1, 1)))
	fake.result("pane.list", managerPaneListJSON(managerPaneInfoJSON("w5:p1", "term_w5", "w5", "w5:t1", dir, "")))
	scriptManagerAgent(t, fake, sleep, "w5:p1", "w5", "manager-demo-g1", dir)
	options := managerSpawnOptions("demo")
	options.Adopt = "w5"
	envelope := Spawn(context.Background(), home, options)
	if envelope.Outcome != "ok" || envelope.Error != nil {
		t.Fatalf("envelope = %#v; want ok", envelope)
	}
	project := readProjectWorkspaceRecord(t, home, "demo")
	if project.Workspace == nil || project.Workspace.ID != "w5" || project.Workspace.RootPane != "w5:p1" || project.Workspace.CreatedByHive {
		t.Fatalf("project workspace = %#v; want adopted w5", project.Workspace)
	}
	if len(fake.recorded("workspace.create")) != 0 {
		t.Fatalf("adopt made a workspace")
	}
}

func TestManagerWorkspaceAdoptTwoPanes(t *testing.T) {
	home := testHome(t)
	dir := projectDirectory(t, home, "demo")
	writeProjectWorkspaceRecord(t, home, "demo", nil)
	fake := newFakeHerdr(t)
	fake.result("workspace.get", managerWorkspaceGetJSON("w5", "hive:demo"))
	fake.result("tab.list", managerTabListJSON(managerTabJSON("w5", "w5:t1", 1, 2)))
	fake.result("pane.list", managerPaneListJSON(
		managerPaneInfoJSON("w5:p1", "term_w5", "w5", "w5:t1", dir, ""),
		managerPaneInfoJSON("w5:p2", "term_w5", "w5", "w5:t1", dir, ""),
	))
	options := managerSpawnOptions("demo")
	options.Adopt = "w5"
	envelope := Spawn(context.Background(), home, options)
	if envelope.Outcome != "failed" {
		t.Fatalf("envelope = %#v; want failed", envelope)
	}
	if !strings.Contains(toString(envelope.Error), "first tab has 2 panes: pass --pane") {
		t.Fatalf("error = %v", envelope.Error)
	}
	if len(fake.recorded("agent.start")) != 0 {
		t.Fatalf("adopt with two panes started an agent")
	}
}

func TestManagerWorkspaceCreateOK(t *testing.T) {
	home := testHome(t)
	dir := projectDirectory(t, home, "demo")
	writeProjectWorkspaceRecord(t, home, "demo", nil)
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	fake.result("workspace.list", managerWorkspaceListJSON())
	fake.result("workspace.create", managerWorkspaceCreateJSON("w6", "hive:demo", dir, "w6:p1"))
	scriptManagerAgent(t, fake, sleep, "w6:p1", "w6", "manager-demo-g1", dir)
	envelope := Spawn(context.Background(), home, managerSpawnOptions("demo"))
	if envelope.Outcome != "ok" || envelope.Error != nil {
		t.Fatalf("envelope = %#v; want ok", envelope)
	}
	project := readProjectWorkspaceRecord(t, home, "demo")
	if project.Workspace == nil || project.Workspace.ID != "w6" || project.Workspace.RootPane != "w6:p1" || !project.Workspace.CreatedByHive || project.Workspace.CreatedAt == "" {
		t.Fatalf("project workspace = %#v", project.Workspace)
	}
	creates := fake.recorded("workspace.create")
	if len(creates) != 1 {
		t.Fatalf("workspace.create requests = %d; want 1", len(creates))
	}
	want := `{"cwd":"` + dir + `","label":"hive:demo","focus":false}`
	if creates[0].Params != want {
		t.Fatalf("workspace.create params = %s; want %s", creates[0].Params, want)
	}
}

func TestManagerWorkspaceCreateCwdMismatch(t *testing.T) {
	home := testHome(t)
	writeProjectWorkspaceRecord(t, home, "demo", nil)
	fake := newFakeHerdr(t)
	fake.result("workspace.list", managerWorkspaceListJSON())
	fake.result("workspace.create", managerWorkspaceCreateJSON("w6", "hive:demo", "/tmp/wrong", "w6:p1"))
	fake.result("workspace.close", `{"type":"ok"}`)
	envelope := Spawn(context.Background(), home, managerSpawnOptions("demo"))
	if envelope.Outcome != "failed" {
		t.Fatalf("envelope = %#v; want failed", envelope)
	}
	if !strings.Contains(toString(envelope.Error), "workspace cwd mismatch") {
		t.Fatalf("error = %v", envelope.Error)
	}
	closes := fake.recorded("workspace.close")
	if len(closes) != 1 || closes[0].Params != `{"workspace_id":"w6","close_group":false}` {
		t.Fatalf("workspace.close = %#v", closes)
	}
	if project := readProjectWorkspaceRecord(t, home, "demo"); project.Workspace != nil {
		t.Fatalf("project workspace = %#v; want nil", project.Workspace)
	}
}

func TestManagerWorkspaceCreateRecordFailure(t *testing.T) {
	home := testHome(t)
	dir := projectDirectory(t, home, "demo")
	writeProjectWorkspaceRecord(t, home, "demo", nil)
	fake := newFakeHerdr(t)
	fake.result("workspace.list", managerWorkspaceListJSON())
	fake.result("workspace.create", managerWorkspaceCreateJSON("w6", "hive:demo", dir, "w6:p1"))
	fake.result("workspace.close", `{"type":"ok"}`)
	projectsDir := filepath.Join(home, ".hive", "state", "projects")
	if err := os.Chmod(projectsDir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(projectsDir, 0o755) })
	envelope := Spawn(context.Background(), home, managerSpawnOptions("demo"))
	if envelope.Outcome != "failed" {
		t.Fatalf("envelope = %#v; want failed", envelope)
	}
	if !strings.Contains(toString(envelope.Error), "store:") {
		t.Fatalf("error = %v; want a store error", envelope.Error)
	}
	if len(fake.recorded("workspace.close")) != 1 {
		t.Fatalf("workspace.close requests = %d; want 1", len(fake.recorded("workspace.close")))
	}
	if len(fake.recorded("agent.start")) != 0 {
		t.Fatalf("record failure started an agent")
	}
}

func TestManagerWorkspacePaneBusyFromStart(t *testing.T) {
	home := testHome(t)
	dir := projectDirectory(t, home, "demo")
	writeProjectWorkspaceRecord(t, home, "demo", nil)
	fake := newFakeHerdr(t)
	fake.result("workspace.list", managerWorkspaceListJSON())
	fake.result("workspace.create", managerWorkspaceCreateJSON("w6", "hive:demo", dir, "w6:p1"))
	fake.failure("agent.start", "agent_pane_busy", "pane w6:p1 is busy")
	envelope := Spawn(context.Background(), home, managerSpawnOptions("demo"))
	if envelope.Outcome != "failed" {
		t.Fatalf("envelope = %#v; want failed", envelope)
	}
	if !strings.Contains(toString(envelope.Error), "first tab busy: stop the agent or pass --pane") {
		t.Fatalf("error = %v", envelope.Error)
	}
}

func TestFirstTabReordered(t *testing.T) {
	home := testHome(t)
	dir := projectDirectory(t, home, "demo")
	writeProjectWorkspaceRecord(t, home, "demo", &records.ProjectWorkspace{ID: "w5", RootPane: "w5:p2", CreatedByHive: true, CreatedAt: "2026-01-01T00:00:00Z"})
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	fake.result("workspace.get", managerWorkspaceGetJSON("w5", "hive:demo"))
	fake.result("pane.get", `{"type":"pane_info","pane":`+managerPaneInfoJSON("w5:p2", "term_w5", "w5", "w5:t2", dir, "")+`}`)
	fake.result("tab.list", managerTabListJSON(managerTabJSON("w5", "w5:t2", 2, 1), managerTabJSON("w5", "w5:t1", 1, 1)))
	scriptManagerAgent(t, fake, sleep, "w5:p2", "w5", "manager-demo-g1", dir)
	envelope := Spawn(context.Background(), home, managerSpawnOptions("demo"))
	if envelope.Outcome != "ok" {
		t.Fatalf("envelope = %#v; want ok", envelope)
	}
}

func TestFirstTabRestoreError(t *testing.T) {
	home := testHome(t)
	dir := projectDirectory(t, home, "demo")
	writeProjectWorkspaceRecord(t, home, "demo", &records.ProjectWorkspace{ID: "w5", RootPane: "w5:p1", CreatedByHive: true, CreatedAt: "2026-01-01T00:00:00Z"})
	fake := newFakeHerdr(t)
	fake.result("workspace.get", managerWorkspaceGetJSON("w5", "hive:demo"))
	fake.result("pane.get", `{"type":"pane_info","pane":`+managerPaneInfoJSON("w5:p1", "term_w5", "w5", "w5:t1", dir, `,"restore_error":"Saved directory is unavailable."`)+`}`)
	fake.result("tab.list", managerTabListJSON(managerTabJSON("w5", "w5:t1", 1, 1)))
	envelope := Spawn(context.Background(), home, managerSpawnOptions("demo"))
	if envelope.Outcome != "failed" {
		t.Fatalf("envelope = %#v; want failed", envelope)
	}
	if !strings.Contains(toString(envelope.Error), "first tab busy: stop the agent or pass --pane") {
		t.Fatalf("error = %v", envelope.Error)
	}
	if len(fake.recorded("agent.start")) != 0 {
		t.Fatalf("restore_error pane started an agent")
	}
}

func TestFirstTabAgentPresent(t *testing.T) {
	home := testHome(t)
	dir := projectDirectory(t, home, "demo")
	writeProjectWorkspaceRecord(t, home, "demo", &records.ProjectWorkspace{ID: "w5", RootPane: "w5:p1", CreatedByHive: true, CreatedAt: "2026-01-01T00:00:00Z"})
	fake := newFakeHerdr(t)
	fake.result("workspace.get", managerWorkspaceGetJSON("w5", "hive:demo"))
	fake.result("pane.get", `{"type":"pane_info","pane":`+managerPaneStatusJSON("w5:p1", "term_w5", "w5", "w5:t1", dir, "idle", `,"agent":"claude"`)+`}`)
	fake.result("tab.list", managerTabListJSON(managerTabJSON("w5", "w5:t1", 1, 1)))
	envelope := Spawn(context.Background(), home, managerSpawnOptions("demo"))
	if envelope.Outcome != "failed" {
		t.Fatalf("envelope = %#v; want failed", envelope)
	}
	if !strings.Contains(toString(envelope.Error), "first tab busy: stop the agent or pass --pane") {
		t.Fatalf("error = %v", envelope.Error)
	}
	if len(fake.recorded("agent.start")) != 0 {
		t.Fatalf("busy pane started an agent")
	}
}
