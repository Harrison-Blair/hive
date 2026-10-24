package agent

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/records"
)

type workspacePost int

const (
	postClosed workspacePost = iota
	postSurviving
	postUnknown
)

func managerStopRecord(sleep sleepProcess, project, workspaceID string) records.Agent {
	return records.Agent{
		Key:          "manager-" + project + "-g1",
		Name:         "manager-" + project,
		Role:         "manager",
		Profile:      records.ProfilePlanner,
		Owns:         project,
		Project:      project,
		Harness:      "claude",
		PaneID:       "w1:p1",
		WorkspaceID:  workspaceID,
		TabID:        "w1:t1",
		TerminalID:   "term_a",
		Cwd:          "/tmp/fixture",
		Status:       records.AgentLive,
		Spawned:      true,
		Generation:   1,
		ProcessPID:   sleep.pid,
		ProcessName:  "claude",
		ProcessStart: sleep.start,
	}
}

func scriptManagerStop(t *testing.T, fake *fakeHerdr, sleep sleepProcess, post workspacePost) {
	t.Helper()
	fake.result("pane.get", `{"type":"pane_info","pane":`+paneInfoJSON("w1:p1", "term_a", "w1", "w1:t1", "/tmp/fixture")+`}`)
	fake.result("pane.process_info", `{"type":"pane_process_info","process_info":`+processInfoJSON("w1:p1", processEntry(sleep.pid, "claude", "claude"))+`}`)
	fake.result("workspace.get", `{"type":"workspace_info","workspace":`+workspaceJSON("w1", 1)+`}`)
	fake.result("pane.close", `{"type":"ok"}`)
	switch post {
	case postClosed:
		fake.failure("workspace.get", "workspace_not_found", "workspace w1 not found")
	case postSurviving:
		fake.result("workspace.get", `{"type":"workspace_info","workspace":`+workspaceJSON("w1", 1)+`}`)
	case postUnknown:
		fake.failure("workspace.get", "timeout", "timed out")
	}
}

func TestStopWorkspaceClosed(t *testing.T) {
	home := testHome(t)
	sleep := startSleep(t)
	record := managerStopRecord(sleep, "demo", "w1")
	writeStopRecord(t, home, record)
	writeProjectWorkspaceRecord(t, home, "demo", &records.ProjectWorkspace{ID: "w1", RootPane: "w1:p1", CreatedByHive: true, CreatedAt: "2026-01-01T00:00:00Z"})
	fake := newFakeHerdr(t)
	scriptManagerStop(t, fake, sleep, postClosed)
	result, envelope := Stop(context.Background(), home, record)
	if envelope.Outcome != "ok" || envelope.Error != nil {
		t.Fatalf("envelope = %#v; want ok", envelope)
	}
	if result.Topology != "closed" {
		t.Fatalf("topology = %q; want closed", result.Topology)
	}
	if !hasEffect(envelope.Effects, "workspace", "closed") {
		t.Fatalf("effects = %#v; want workspace closed", envelope.Effects)
	}
	if got := readProjectWorkspaceRecord(t, home, "demo").Workspace; got != nil {
		t.Fatalf("project workspace = %#v; want nil", got)
	}
	if len(fake.recorded("workspace.close")) != 0 {
		t.Fatalf("stop sent workspace.close")
	}
}

func TestStopWorkspaceSurviving(t *testing.T) {
	home := testHome(t)
	sleep := startSleep(t)
	record := managerStopRecord(sleep, "demo", "w1")
	writeStopRecord(t, home, record)
	writeProjectWorkspaceRecord(t, home, "demo", &records.ProjectWorkspace{ID: "w1", RootPane: "w1:p1", CreatedByHive: true, CreatedAt: "2026-01-01T00:00:00Z"})
	fake := newFakeHerdr(t)
	scriptManagerStop(t, fake, sleep, postSurviving)
	result, envelope := Stop(context.Background(), home, record)
	if envelope.Outcome != "ok" || result.Topology != "surviving" {
		t.Fatalf("result = %#v envelope = %#v; want ok surviving", result, envelope)
	}
	if !hasEffect(envelope.Effects, "workspace", "kept: 1 other panes") {
		t.Fatalf("effects = %#v; want workspace kept: 1 other panes", envelope.Effects)
	}
	project := readProjectWorkspaceRecord(t, home, "demo")
	if project.Workspace == nil || project.Workspace.ID != "w1" || project.Workspace.RootPane != "w1:p1" {
		t.Fatalf("project workspace = %#v; want the kept workspace", project.Workspace)
	}
	if len(fake.recorded("workspace.close")) != 0 {
		t.Fatalf("stop sent workspace.close")
	}
}

func TestStopWorkspaceUnknown(t *testing.T) {
	home := testHome(t)
	sleep := startSleep(t)
	record := managerStopRecord(sleep, "demo", "w1")
	writeStopRecord(t, home, record)
	writeProjectWorkspaceRecord(t, home, "demo", &records.ProjectWorkspace{ID: "w1", RootPane: "w1:p1", CreatedByHive: true, CreatedAt: "2026-01-01T00:00:00Z"})
	fake := newFakeHerdr(t)
	scriptManagerStop(t, fake, sleep, postUnknown)
	result, envelope := Stop(context.Background(), home, record)
	if envelope.Outcome != "ok" || result.Topology != "unknown" {
		t.Fatalf("result = %#v envelope = %#v; want ok unknown", result, envelope)
	}
	if !hasEffect(envelope.Effects, "pane", "closed") {
		t.Fatalf("effects = %#v; want pane closed", envelope.Effects)
	}
	project := readProjectWorkspaceRecord(t, home, "demo")
	if project.Workspace == nil || project.Workspace.ID != "w1" {
		t.Fatalf("project workspace = %#v; want the kept workspace", project.Workspace)
	}
}

func TestStopWorkspaceRecordFailure(t *testing.T) {
	home := testHome(t)
	sleep := startSleep(t)
	record := managerStopRecord(sleep, "demo", "w1")
	writeStopRecord(t, home, record)
	writeProjectWorkspaceRecord(t, home, "demo", &records.ProjectWorkspace{ID: "w1", RootPane: "w1:p1", CreatedByHive: true, CreatedAt: "2026-01-01T00:00:00Z"})
	projectsDir := filepath.Join(home, ".hive", "state", "projects")
	if err := os.Chmod(projectsDir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(projectsDir, 0o755) })
	fake := newFakeHerdr(t)
	scriptManagerStop(t, fake, sleep, postClosed)
	result, envelope := Stop(context.Background(), home, record)
	if envelope.Outcome != "partial" || envelope.Phase != "workspace" {
		t.Fatalf("envelope = %#v; want partial workspace", envelope)
	}
	if !hasEffect(envelope.Effects, "pane", "closed") || !hasEffect(envelope.Effects, "agent", "stopped") {
		t.Fatalf("effects = %#v; want the known effects", envelope.Effects)
	}
	if result.Agent.Status != records.AgentStopped {
		t.Fatalf("record status = %q; want stopped", result.Agent.Status)
	}
	if len(fake.recorded("workspace.close")) != 0 {
		t.Fatalf("stop sent workspace.close")
	}
}
