package agent

import (
	"context"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

const managerRepoPath = "/tmp/demo-repository"

func writeManagerProject(t *testing.T, home string, workspace *records.ProjectWorkspace) {
	t.Helper()
	projectDirectory(t, home, "demo")
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	project := records.Project{ID: "demo", Path: managerRepoPath, Base: "dev", RegisteredAt: "2026-01-01T00:00:00Z", Workspace: workspace}
	if err := registry.WriteIn([]string{"projects"}, "demo", project); err != nil {
		t.Fatal(err)
	}
}

func managerWorktreeListJSON(parentID string) string {
	source := `"repo_key":"/tmp/demo-repository/.git","repo_name":"demo-repository","repo_root":"/tmp/demo-repository","source_checkout_path":"/tmp/demo-repository"`
	if parentID != "" {
		source += `,"source_workspace_id":"` + parentID + `"`
	}
	return `{"type":"worktree_list","source":{` + source + `},"worktrees":[]}`
}

func managerTabCreatedJSON(workspaceID, tabID, paneID, cwd string) string {
	return `{"type":"tab_created","tab":{"tab_id":"` + tabID + `","workspace_id":"` + workspaceID + `","number":2,"label":"manager-demo","focused":false,"pane_count":1,"agent_status":"unknown"},"root_pane":` + managerPaneInfoJSON(paneID, "term_"+workspaceID, workspaceID, tabID, cwd, "") + `}`
}

func managerProcessInfoJSON(paneID string, sleep sleepProcess) string {
	return `{"type":"pane_process_info","process_info":{"pane_id":"` + paneID + `","shell_pid":` + jsonNumber(sleep.pid) + `,"foreground_process_group_id":` + jsonNumber(sleep.pid) + `,"foreground_processes":[` + processEntry(sleep.pid, "claude", "claude") + `]}}`
}

func managerAgentInfoJSON(paneID, workspaceID, tabID, status, name string) string {
	return `{"pane_id":"` + paneID + `","terminal_id":"term_` + workspaceID + `","workspace_id":"` + workspaceID + `","tab_id":"` + tabID + `","agent_status":"` + status + `","focused":false,"revision":2,"name":"` + name + `"}`
}

func scriptManagerSpawnTail(t *testing.T, home string, fake *fakeHerdr, sleep sleepProcess, paneID, workspaceID, tabID string, readyShell ...bool) {
	t.Helper()
	cwd := projectDirectory(t, home, "demo")
	key := "manager-demo-g1"
	if len(readyShell) == 0 || readyShell[0] {
		scriptReadyShell(fake, paneID, "term_"+workspaceID, workspaceID, tabID, cwd)
	}
	fake.result("agent.start", `{"type":"agent_started","agent":`+managerAgentInfoJSON(paneID, workspaceID, tabID, "unknown", key)+`,"argv":["claude","--model","m"]}`)
	fake.result("pane.get", `{"type":"pane_info","pane":`+managerPaneInfoJSON(paneID, "term_"+workspaceID, workspaceID, tabID, cwd, "")+`}`)
	fake.result("agent.wait", `{"type":"agent_info","agent":`+managerAgentInfoJSON(paneID, workspaceID, tabID, "idle", key)+`}`)
	for range 3 {
		fake.result("agent.get", `{"type":"agent_info","agent":`+managerAgentInfoJSON(paneID, workspaceID, tabID, "idle", key)+`}`)
	}
	fake.result("pane.process_info", managerProcessInfoJSON(paneID, sleep))
	fake.result("pane.process_info", `{"type":"pane_process_info","process_info":`+processInfoJSON(paneID, processEntry(sleep.pid, "claude", "claude"))+`}`)
	fake.result("agent.prompt", `{"type":"agent_prompted","agent":`+managerAgentInfoJSON(paneID, workspaceID, tabID, "working", key)+`}`)
}

func TestManagerCreatesWorkspaceAtProjectDirectory(t *testing.T) {
	home := testHome(t)
	writeManagerProject(t, home, nil)
	dir := projectDirectory(t, home, "demo")
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	fake.result("workspace.create", managerWorkspaceCreateJSON("wBD", "hive:demo", dir, "wBD:p1"))
	scriptManagerSpawnTail(t, home, fake, sleep, "wBD:p1", "wBD", "wBD:t1")

	envelope := Spawn(context.Background(), home, managerSpawnOptions("demo"))
	if envelope.Outcome != "ok" || envelope.Error != nil {
		t.Fatalf("envelope = %#v; want ok", envelope)
	}
	creates := fake.recorded("workspace.create")
	if len(creates) != 1 || creates[0].Params != `{"cwd":"`+dir+`","label":"hive:demo","focus":false}` {
		t.Fatalf("workspace.create = %#v; want one workspace at the project directory", creates)
	}
	if len(fake.recorded("tab.create")) != 0 {
		t.Fatalf("created workspace also created a tab")
	}
	project := readProjectWorkspaceRecord(t, home, "demo")
	if project.Workspace == nil || project.Workspace.ID != "wBD" || !project.Workspace.CreatedByHive || project.Workspace.TabID != "wBD:t1" || project.Workspace.TabCreatedByHive {
		t.Fatalf("project workspace = %#v; want hive-created root tab", project.Workspace)
	}
	if len(fake.recorded("worktree.list")) != 0 {
		t.Fatalf("project placement used the parent checkout: %#v", fake.requests)
	}
}

func TestManagerCreatesTabInRecordedWorkspace(t *testing.T) {
	home := testHome(t)
	writeManagerProject(t, home, &records.ProjectWorkspace{ID: "wBC", RootPane: "wBC:p1", TabID: "wBC:t1", CreatedByHive: true, TabCreatedByHive: true})
	dir := projectDirectory(t, home, "demo")
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	fake.result("workspace.get", managerWorkspaceGetJSON("wBC", "hive:demo"))
	fake.result("tab.list", managerTabListJSON())
	fake.result("tab.create", managerTabCreatedJSON("wBC", "wBC:t2", "wBC:p2", dir))
	scriptManagerSpawnTail(t, home, fake, sleep, "wBC:p2", "wBC", "wBC:t2")

	envelope := Spawn(context.Background(), home, managerSpawnOptions("demo"))
	if envelope.Outcome != "ok" || envelope.Error != nil {
		t.Fatalf("envelope = %#v; want ok", envelope)
	}
	created := fake.recorded("tab.create")
	if len(created) != 1 || created[0].Params != `{"workspace_id":"wBC","cwd":"`+dir+`","label":"manager-demo","focus":false}` {
		t.Fatalf("tab.create = %#v; want one named tab", created)
	}
	if len(fake.recorded("workspace.create")) != 0 {
		t.Fatalf("recorded workspace was replaced")
	}
}

func TestManagerReusesRecordedFreeShellTab(t *testing.T) {
	home := testHome(t)
	writeManagerProject(t, home, &records.ProjectWorkspace{ID: "wBC", RootPane: "wBC:p1", TabID: "wBC:t1", CreatedByHive: true, TabCreatedByHive: true})
	dir := projectDirectory(t, home, "demo")
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	fake.result("workspace.get", managerWorkspaceGetJSON("wBC", "user-label"))
	fake.result("tab.list", managerTabListJSON(managerTabJSON("wBC", "wBC:t1", 1, 1)))
	fake.result("pane.list", managerPaneListJSON(managerPaneInfoJSON("wBC:p1", "term_wBC", "wBC", "wBC:t1", dir, "")))
	fake.result("pane.process_info", managerProcessInfoJSON("wBC:p1", sleep))
	scriptManagerSpawnTail(t, home, fake, sleep, "wBC:p1", "wBC", "wBC:t1")

	envelope := Spawn(context.Background(), home, managerSpawnOptions("demo"))
	if envelope.Outcome != "ok" || envelope.Error != nil {
		t.Fatalf("envelope = %#v; want ok", envelope)
	}
	if len(fake.recorded("tab.create")) != 0 || len(fake.recorded("workspace.create")) != 0 {
		t.Fatalf("reused tab created topology: requests %#v", fake.requests)
	}
	project := readProjectWorkspaceRecord(t, home, "demo")
	if project.Workspace == nil || project.Workspace.ID != "wBC" || !project.Workspace.CreatedByHive || project.Workspace.TabID != "wBC:t1" || !project.Workspace.TabCreatedByHive {
		t.Fatalf("project workspace = %#v; want recorded ownership", project.Workspace)
	}
}

func TestManagerCreatesNewTabWhenRecordedTabIsGone(t *testing.T) {
	home := testHome(t)
	writeManagerProject(t, home, &records.ProjectWorkspace{ID: "wBC", RootPane: "wBC:p1", TabID: "wBC:t1", CreatedByHive: true, TabCreatedByHive: false})
	dir := projectDirectory(t, home, "demo")
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	fake.result("workspace.get", managerWorkspaceGetJSON("wBC", "project-home"))
	fake.result("tab.list", managerTabListJSON())
	fake.result("tab.create", managerTabCreatedJSON("wBC", "wBC:t2", "wBC:p2", dir))
	scriptManagerSpawnTail(t, home, fake, sleep, "wBC:p2", "wBC", "wBC:t2")

	envelope := Spawn(context.Background(), home, managerSpawnOptions("demo"))
	if envelope.Outcome != "ok" || envelope.Error != nil {
		t.Fatalf("envelope = %#v; want ok", envelope)
	}
	created := fake.recorded("tab.create")
	if len(created) != 1 || created[0].Params != `{"workspace_id":"wBC","cwd":"`+dir+`","label":"manager-demo","focus":false}` {
		t.Fatalf("tab.create = %#v; want one named tab", created)
	}
	project := readProjectWorkspaceRecord(t, home, "demo")
	if project.Workspace == nil || project.Workspace.ID != "wBC" || !project.Workspace.CreatedByHive || project.Workspace.TabID != "wBC:t2" || !project.Workspace.TabCreatedByHive {
		t.Fatalf("project workspace = %#v; want old workspace ownership and new tab ownership", project.Workspace)
	}
}

func TestManagerKeepsRecordedWorkspaceWithoutClosingIt(t *testing.T) {
	home := testHome(t)
	writeManagerProject(t, home, &records.ProjectWorkspace{ID: "wOld", RootPane: "wOld:p1", CreatedByHive: true})
	dir := projectDirectory(t, home, "demo")
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	fake.result("workspace.get", managerWorkspaceGetJSON("wOld", "hive:demo"))
	fake.result("tab.create", managerTabCreatedJSON("wOld", "wOld:t2", "wOld:p2", dir))
	scriptManagerSpawnTail(t, home, fake, sleep, "wOld:p2", "wOld", "wOld:t2")

	envelope := Spawn(context.Background(), home, managerSpawnOptions("demo"))
	if envelope.Outcome != "ok" || envelope.Error != nil {
		t.Fatalf("envelope = %#v; want ok", envelope)
	}
	if len(fake.recorded("workspace.close")) != 0 {
		t.Fatalf("recorded workspace was closed")
	}
	project := readProjectWorkspaceRecord(t, home, "demo")
	if project.Workspace == nil || project.Workspace.ID != "wOld" || project.Workspace.TabID != "wOld:t2" {
		t.Fatalf("project workspace = %#v; want the recorded workspace", project.Workspace)
	}
}

func TestManagerAdoptFlagUsesTheWorkspaceAndKeepsItsLabel(t *testing.T) {
	home := testHome(t)
	writeManagerProject(t, home, nil)
	dir := projectDirectory(t, home, "demo")
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	fake.result("workspace.get", managerWorkspaceGetJSON("wBC", "not-hive-label"))
	fake.result("tab.create", managerTabCreatedJSON("wBC", "wBC:t2", "wBC:p2", dir))
	scriptManagerSpawnTail(t, home, fake, sleep, "wBC:p2", "wBC", "wBC:t2")
	options := managerSpawnOptions("demo")
	options.Adopt = "wBC"

	envelope := Spawn(context.Background(), home, options)
	if envelope.Outcome != "ok" || envelope.Error != nil {
		t.Fatalf("envelope = %#v; want ok", envelope)
	}
	if len(fake.recorded("workspace.create")) != 0 || len(fake.recorded("tab.create")) != 1 {
		t.Fatalf("adopt topology requests = %#v", fake.requests)
	}
	if len(fake.recorded("workspace.get")) != 1 || fake.recorded("workspace.get")[0].Params != `{"workspace_id":"wBC"}` {
		t.Fatalf("workspace.get = %#v", fake.recorded("workspace.get"))
	}
}

func TestManagerAdoptFlagRejectsDifferentWorkspace(t *testing.T) {
	home := testHome(t)
	writeManagerProject(t, home, nil)
	fake := newFakeHerdr(t)
	fake.failure("workspace.get", "workspace_not_found", "workspace wElsewhere not found")
	options := managerSpawnOptions("demo")
	options.Adopt = "wElsewhere"
	envelope := Spawn(context.Background(), home, options)
	if envelope.Outcome != "failed" || !strings.Contains(toString(envelope.Error), "workspace get:") {
		t.Fatalf("envelope = %#v; want a workspace read refusal", envelope)
	}
	if len(fake.recorded("tab.create")) != 0 || len(fake.recorded("workspace.create")) != 0 || len(fake.recorded("agent.start")) != 0 {
		t.Fatalf("wrong workspace caused a change: %#v", fake.requests)
	}
}

func TestManagerNotRepositoryReportsWorktreeListError(t *testing.T) {
	home := testHome(t)
	writeManagerProject(t, home, nil)
	fake := newFakeHerdr(t)
	fake.failure("workspace.create", "internal_error", "workspace create failed")
	envelope := Spawn(context.Background(), home, managerSpawnOptions("demo"))
	if envelope.Outcome != "failed" || !strings.Contains(toString(envelope.Error), "workspace create:") {
		t.Fatalf("envelope = %#v; want workspace create error", envelope)
	}
	if len(fake.recorded("agent.start")) != 0 {
		t.Fatalf("workspace failure caused a start: %#v", fake.requests)
	}
}

func TestManagerWorkspaceCreatedAtProjectDirectory(t *testing.T) {
	home := testHome(t)
	writeManagerProject(t, home, nil)
	dir := projectDirectory(t, home, "demo")
	fake := newFakeHerdr(t)
	fake.result("workspace.create", managerWorkspaceCreateJSON("wBD", "hive:demo", dir, "wBD:p1"))
	fake.result("workspace.close", `{"type":"ok"}`)
	// A manager start failure ensures this test only checks workspace resolution.
	scriptReadyShell(fake, "wBD:p1", "term_wBD", "wBD", "wBD:t1", dir)
	fake.failure("agent.start", "invalid_agent_name", "name is too long")
	envelope := Spawn(context.Background(), home, managerSpawnOptions("demo"))
	if envelope.Outcome != "failed" || envelope.Phase != "start" {
		t.Fatalf("envelope = %#v; want failed start", envelope)
	}
	creates := fake.recorded("workspace.create")
	if len(creates) != 1 || creates[0].Params != `{"cwd":"`+dir+`","label":"hive:demo","focus":false}` {
		t.Fatalf("workspace.create = %#v; want project directory", creates)
	}
}
