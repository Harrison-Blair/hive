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

func scriptManagerSpawnTail(t *testing.T, fake *fakeHerdr, sleep sleepProcess, paneID, workspaceID, tabID string) {
	t.Helper()
	key := "manager-demo-g1"
	fake.result("agent.start", `{"type":"agent_started","agent":`+managerAgentInfoJSON(paneID, workspaceID, tabID, "unknown", key)+`,"argv":["claude","--model","m"]}`)
	fake.result("pane.get", `{"type":"pane_info","pane":`+managerPaneInfoJSON(paneID, "term_"+workspaceID, workspaceID, tabID, managerRepoPath, "")+`}`)
	fake.result("agent.wait", `{"type":"agent_info","agent":`+managerAgentInfoJSON(paneID, workspaceID, tabID, "idle", key)+`}`)
	for range 3 {
		fake.result("agent.get", `{"type":"agent_info","agent":`+managerAgentInfoJSON(paneID, workspaceID, tabID, "idle", key)+`}`)
	}
	fake.result("pane.process_info", managerProcessInfoJSON(paneID, sleep))
	fake.result("pane.process_info", `{"type":"pane_process_info","process_info":`+processInfoJSON(paneID, processEntry(sleep.pid, "claude", "claude"))+`}`)
	fake.result("agent.prompt", `{"type":"agent_prompted","agent":`+managerAgentInfoJSON(paneID, workspaceID, tabID, "working", key)+`}`)
}

func TestManagerAdoptsParentWorkspaceAndCreatesNamedTab(t *testing.T) {
	home := testHome(t)
	writeManagerProject(t, home, nil)
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	fake.result("worktree.list", managerWorktreeListJSON("wBC"))
	fake.result("workspace.get", managerWorkspaceGetJSON("wBC", "project-home"))
	fake.result("tab.create", managerTabCreatedJSON("wBC", "wBC:t2", "wBC:p2", managerRepoPath))
	// These replies let the old layout complete its path so the assertion below
	// fails on the behavior, rather than on a missing fake reply.
	fake.result("workspace.list", managerWorkspaceListJSON())
	fake.result("workspace.create", managerWorkspaceCreateJSON("wBD", "hive:demo", managerRepoPath, "wBD:p1"))
	scriptManagerSpawnTail(t, fake, sleep, "wBC:p2", "wBC", "wBC:t2")

	envelope := Spawn(context.Background(), home, managerSpawnOptions("demo"))
	if envelope.Outcome != "ok" || envelope.Error != nil {
		t.Fatalf("envelope = %#v; want ok", envelope)
	}
	created := fake.recorded("tab.create")
	if len(created) != 1 || created[0].Params != `{"workspace_id":"wBC","cwd":"`+managerRepoPath+`","label":"manager-demo","focus":false}` {
		t.Fatalf("tab.create = %#v; want one named tab in wBC", created)
	}
	if len(fake.recorded("workspace.create")) != 0 {
		t.Fatalf("manager created a second workspace")
	}
	starts := fake.recorded("agent.start")
	if len(starts) != 1 || !strings.Contains(starts[0].Params, `"pane_id":"wBC:p2"`) {
		t.Fatalf("agent.start = %#v; want root pane wBC:p2", starts)
	}
	project := readProjectWorkspaceRecord(t, home, "demo")
	if project.Workspace == nil || project.Workspace.ID != "wBC" || project.Workspace.CreatedByHive || project.Workspace.TabID != "wBC:t2" || !project.Workspace.TabCreatedByHive || project.Workspace.RootPane != "wBC:p2" {
		t.Fatalf("project workspace = %#v; want adopted workspace and created tab", project.Workspace)
	}
}

func TestManagerCreatesWorkspaceWhenNoParentExists(t *testing.T) {
	home := testHome(t)
	writeManagerProject(t, home, nil)
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	fake.result("worktree.list", managerWorktreeListJSON(""))
	fake.result("workspace.create", managerWorkspaceCreateJSON("wBD", "hive:demo", managerRepoPath, "wBD:p1"))
	fake.result("workspace.list", managerWorkspaceListJSON())
	scriptManagerSpawnTail(t, fake, sleep, "wBD:p1", "wBD", "wBD:t1")

	envelope := Spawn(context.Background(), home, managerSpawnOptions("demo"))
	if envelope.Outcome != "ok" || envelope.Error != nil {
		t.Fatalf("envelope = %#v; want ok", envelope)
	}
	creates := fake.recorded("workspace.create")
	if len(creates) != 1 || creates[0].Params != `{"cwd":"`+managerRepoPath+`","label":"hive:demo","focus":false}` {
		t.Fatalf("workspace.create = %#v; want one repo workspace", creates)
	}
	if len(fake.recorded("tab.create")) != 0 {
		t.Fatalf("created workspace also created a tab")
	}
	project := readProjectWorkspaceRecord(t, home, "demo")
	if project.Workspace == nil || project.Workspace.ID != "wBD" || !project.Workspace.CreatedByHive || project.Workspace.TabID != "wBD:t1" || project.Workspace.TabCreatedByHive {
		t.Fatalf("project workspace = %#v; want hive-created root tab", project.Workspace)
	}
}

func TestManagerReusesRecordedFreeShellTab(t *testing.T) {
	home := testHome(t)
	writeManagerProject(t, home, &records.ProjectWorkspace{ID: "wBC", RootPane: "wBC:p1", TabID: "wBC:t1", CreatedByHive: true, TabCreatedByHive: true})
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	fake.result("workspace.get", managerWorkspaceGetJSON("wBC", "user-label"))
	fake.result("worktree.list", managerWorktreeListJSON("wBC"))
	fake.result("tab.list", managerTabListJSON(managerTabJSON("wBC", "wBC:t1", 1, 1)))
	fake.result("pane.list", managerPaneListJSON(managerPaneInfoJSON("wBC:p1", "term_wBC", "wBC", "wBC:t1", managerRepoPath, "")))
	fake.result("pane.process_info", managerProcessInfoJSON("wBC:p1", sleep))
	scriptManagerSpawnTail(t, fake, sleep, "wBC:p1", "wBC", "wBC:t1")

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
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	fake.result("workspace.get", managerWorkspaceGetJSON("wBC", "project-home"))
	fake.result("worktree.list", managerWorktreeListJSON("wBC"))
	fake.result("tab.list", managerTabListJSON())
	fake.result("tab.create", managerTabCreatedJSON("wBC", "wBC:t2", "wBC:p2", managerRepoPath))
	scriptManagerSpawnTail(t, fake, sleep, "wBC:p2", "wBC", "wBC:t2")

	envelope := Spawn(context.Background(), home, managerSpawnOptions("demo"))
	if envelope.Outcome != "ok" || envelope.Error != nil {
		t.Fatalf("envelope = %#v; want ok", envelope)
	}
	created := fake.recorded("tab.create")
	if len(created) != 1 || created[0].Params != `{"workspace_id":"wBC","cwd":"`+managerRepoPath+`","label":"manager-demo","focus":false}` {
		t.Fatalf("tab.create = %#v; want one named tab", created)
	}
	project := readProjectWorkspaceRecord(t, home, "demo")
	if project.Workspace == nil || project.Workspace.ID != "wBC" || !project.Workspace.CreatedByHive || project.Workspace.TabID != "wBC:t2" || !project.Workspace.TabCreatedByHive {
		t.Fatalf("project workspace = %#v; want old workspace ownership and new tab ownership", project.Workspace)
	}
}

func TestManagerClearsOldLayoutRecordWithoutClosingWorkspace(t *testing.T) {
	home := testHome(t)
	writeManagerProject(t, home, &records.ProjectWorkspace{ID: "wOld", RootPane: "wOld:p1", CreatedByHive: true})
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	fake.result("workspace.get", managerWorkspaceGetJSON("wOld", "hive:demo"))
	fake.result("worktree.list", managerWorktreeListJSON("wBC"))
	fake.result("workspace.get", managerWorkspaceGetJSON("wBC", "project-home"))
	fake.result("tab.create", managerTabCreatedJSON("wBC", "wBC:t2", "wBC:p2", managerRepoPath))
	fake.result("workspace.list", managerWorkspaceListJSON())
	fake.result("workspace.create", managerWorkspaceCreateJSON("wBD", "hive:demo", managerRepoPath, "wBD:p1"))
	scriptManagerSpawnTail(t, fake, sleep, "wBC:p2", "wBC", "wBC:t2")

	envelope := Spawn(context.Background(), home, managerSpawnOptions("demo"))
	if envelope.Outcome != "ok" || envelope.Error != nil {
		t.Fatalf("envelope = %#v; want ok", envelope)
	}
	if len(fake.recorded("workspace.close")) != 0 {
		t.Fatalf("old-layout workspace was closed")
	}
	if len(fake.recorded("worktree.list")) != 1 || fake.recorded("worktree.list")[0].Params != `{"cwd":"`+managerRepoPath+`"}` {
		t.Fatalf("worktree.list = %#v; want the repository path", fake.recorded("worktree.list"))
	}
	project := readProjectWorkspaceRecord(t, home, "demo")
	if project.Workspace == nil || project.Workspace.ID != "wBC" || project.Workspace.TabID != "wBC:t2" {
		t.Fatalf("project workspace = %#v; want the parent workspace", project.Workspace)
	}
}

func TestManagerAdoptFlagNeedsParentWorkspaceButKeepsItsLabel(t *testing.T) {
	home := testHome(t)
	writeManagerProject(t, home, nil)
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	fake.result("worktree.list", managerWorktreeListJSON("wBC"))
	fake.result("workspace.get", managerWorkspaceGetJSON("wBC", "not-hive-label"))
	fake.result("tab.create", managerTabCreatedJSON("wBC", "wBC:t2", "wBC:p2", managerRepoPath))
	scriptManagerSpawnTail(t, fake, sleep, "wBC:p2", "wBC", "wBC:t2")
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
	fake.result("worktree.list", managerWorktreeListJSON("wBC"))
	options := managerSpawnOptions("demo")
	options.Adopt = "wElsewhere"
	envelope := Spawn(context.Background(), home, options)
	if envelope.Outcome != "failed" || !strings.Contains(toString(envelope.Error), "does not hold the parent checkout") {
		t.Fatalf("envelope = %#v; want parent-checkout refusal", envelope)
	}
	if len(fake.recorded("tab.create")) != 0 || len(fake.recorded("workspace.create")) != 0 || len(fake.recorded("agent.start")) != 0 {
		t.Fatalf("wrong workspace caused a change: %#v", fake.requests)
	}
}

func TestManagerNotRepositoryReportsWorktreeListError(t *testing.T) {
	home := testHome(t)
	writeManagerProject(t, home, nil)
	fake := newFakeHerdr(t)
	fake.failure("worktree.list", "not_git_worktree", "not a git worktree")
	envelope := Spawn(context.Background(), home, managerSpawnOptions("demo"))
	if envelope.Outcome != "failed" || !strings.Contains(toString(envelope.Error), "not_git_worktree") {
		t.Fatalf("envelope = %#v; want not_git_worktree error", envelope)
	}
	if len(fake.recorded("workspace.create")) != 0 || len(fake.recorded("agent.start")) != 0 {
		t.Fatalf("non-repository caused a change: %#v", fake.requests)
	}
}

func TestManagerWorkspaceCreatedAtRepositoryPath(t *testing.T) {
	home := testHome(t)
	writeManagerProject(t, home, nil)
	fake := newFakeHerdr(t)
	fake.result("worktree.list", managerWorktreeListJSON(""))
	fake.result("workspace.create", managerWorkspaceCreateJSON("wBD", "hive:demo", managerRepoPath, "wBD:p1"))
	fake.result("workspace.close", `{"type":"ok"}`)
	// A manager start failure ensures this test only checks workspace resolution.
	fake.failure("agent.start", "invalid_agent_name", "name is too long")
	envelope := Spawn(context.Background(), home, managerSpawnOptions("demo"))
	if envelope.Outcome != "failed" || envelope.Phase != "start" {
		t.Fatalf("envelope = %#v; want failed start", envelope)
	}
	creates := fake.recorded("workspace.create")
	if len(creates) != 1 || creates[0].Params != `{"cwd":"`+managerRepoPath+`","label":"hive:demo","focus":false}` {
		t.Fatalf("workspace.create = %#v; want repository path", creates)
	}
}
