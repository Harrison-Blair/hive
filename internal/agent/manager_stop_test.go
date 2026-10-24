package agent

import (
	"context"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/records"
)

func managerStopRecordForTab(sleep sleepProcess, project, workspaceID, tabID, paneID string) records.Agent {
	record := managerStopRecord(sleep, project, workspaceID)
	record.TabID = tabID
	record.PaneID = paneID
	return record
}

func scriptManagerStopForTab(t *testing.T, fake *fakeHerdr, sleep sleepProcess, tabID, paneID string, closeError bool) {
	t.Helper()
	fake.result("pane.get", `{"type":"pane_info","pane":`+paneInfoJSON(paneID, "term_a", "w1", tabID, "/tmp/fixture")+`}`)
	fake.result("pane.process_info", `{"type":"pane_process_info","process_info":`+processInfoJSON(paneID, processEntry(sleep.pid, "claude", "claude"))+`}`)
	fake.result("workspace.get", `{"type":"workspace_info","workspace":`+workspaceJSON("w1", 1)+`}`)
	if closeError {
		fake.failure("pane.close", "confirmation_required", "closing this pane would close a worktree group")
		return
	}
	fake.result("pane.close", `{"type":"ok"}`)
	fake.result("workspace.get", `{"type":"workspace_info","workspace":`+workspaceJSON("w1", 1)+`}`)
}

func TestStopAdoptedManagerDropsClosedTabRecord(t *testing.T) {
	home := testHome(t)
	sleep := startSleep(t)
	record := managerStopRecordForTab(sleep, "demo", "w1", "w1:t2", "w1:p2")
	writeStopRecord(t, home, record)
	writeProjectWorkspaceRecord(t, home, "demo", &records.ProjectWorkspace{
		ID: "w1", RootPane: "w1:p2", TabID: "w1:t2", CreatedByHive: false, TabCreatedByHive: true,
	})
	fake := newFakeHerdr(t)
	scriptManagerStopForTab(t, fake, sleep, "w1:t2", "w1:p2", false)

	result, envelope := Stop(context.Background(), home, record)
	if envelope.Outcome != "ok" || result.Topology != "surviving" {
		t.Fatalf("result = %#v envelope = %#v; want surviving", result, envelope)
	}
	project := readProjectWorkspaceRecord(t, home, "demo")
	if project.Workspace == nil || project.Workspace.ID != "w1" || project.Workspace.CreatedByHive || project.Workspace.TabID != "" || project.Workspace.TabCreatedByHive || project.Workspace.RootPane != "" {
		t.Fatalf("project workspace = %#v; want workspace kept with no closed tab", project.Workspace)
	}
	if len(fake.recorded("workspace.close")) != 0 {
		t.Fatalf("stop sent workspace.close")
	}
}

func TestStopConfirmationRequiredKeepsStoppedRecordAndWorkspace(t *testing.T) {
	home := testHome(t)
	sleep := startSleep(t)
	record := managerStopRecordForTab(sleep, "demo", "w1", "w1:t1", "w1:p1")
	writeStopRecord(t, home, record)
	before := &records.ProjectWorkspace{ID: "w1", RootPane: "w1:p1", TabID: "w1:t1", CreatedByHive: true, TabCreatedByHive: false}
	writeProjectWorkspaceRecord(t, home, "demo", before)
	fake := newFakeHerdr(t)
	scriptManagerStopForTab(t, fake, sleep, "w1:t1", "w1:p1", true)

	result, envelope := Stop(context.Background(), home, record)
	if result.Agent.Status != records.AgentStopped {
		t.Fatalf("agent status = %q; want stopped", result.Agent.Status)
	}
	if stored := readAgent(t, home, record.Key); stored.Status != records.AgentStopped {
		t.Fatalf("stored agent status = %q; want stopped", stored.Status)
	}
	if !strings.Contains(toString(envelope.Effects), "confirmation_required") {
		t.Fatalf("effects = %#v; want the pane refusal", envelope.Effects)
	}
	if got := readProjectWorkspaceRecord(t, home, "demo").Workspace; got == nil || *got != *before {
		t.Fatalf("project workspace = %#v; want unchanged %#v", got, before)
	}
	if len(fake.recorded("workspace.close")) != 0 || len(fake.recorded("workspace.get")) != 1 {
		t.Fatalf("requests after pane refusal = %#v", fake.requests)
	}
}

func TestManagerStartFailureCleansOnlyCreatedAdoptedTab(t *testing.T) {
	home := testHome(t)
	dir := projectDirectory(t, home, "demo")
	writeManagerProject(t, home, &records.ProjectWorkspace{ID: "wBC", CreatedByHive: false})
	fake := newFakeHerdr(t)
	fake.result("workspace.get", managerWorkspaceGetJSON("wBC", "project-home"))
	fake.result("tab.create", managerTabCreatedJSON("wBC", "wBC:t2", "wBC:p2", dir))
	scriptReadyShell(fake, "wBC:p2", "term_wBC", "wBC", "wBC:t2", dir)
	fake.failure("agent.start", "invalid_agent_name", "name is too long")
	fake.result("pane.get", `{"type":"pane_info","pane":`+managerPaneInfoJSON("wBC:p2", "term_wBC", "wBC", "wBC:t2", dir, "")+`}`)
	fake.result("pane.close", `{"type":"ok"}`)

	envelope := Spawn(context.Background(), home, managerSpawnOptions("demo"))
	if envelope.Outcome != "failed" || envelope.Phase != "start" {
		t.Fatalf("envelope = %#v; want failed start", envelope)
	}
	closes := fake.recorded("pane.close")
	if len(closes) != 1 || closes[0].Params != `{"pane_id":"wBC:p2"}` {
		t.Fatalf("pane.close = %#v; want only wBC:p2", closes)
	}
	if len(fake.recorded("workspace.close")) != 0 {
		t.Fatalf("start failure closed an adopted workspace")
	}
	project := readProjectWorkspaceRecord(t, home, "demo")
	if project.Workspace == nil || project.Workspace.ID != "wBC" || project.Workspace.TabID != "" || project.Workspace.TabCreatedByHive {
		t.Fatalf("project workspace = %#v; want adopted workspace with no tab fields", project.Workspace)
	}
}
