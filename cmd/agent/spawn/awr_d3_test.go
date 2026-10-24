package spawn_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// d3Envelope is the JSON shape of one spawn result.
type d3Envelope struct {
	Outcome string  `json:"outcome"`
	Phase   string  `json:"phase"`
	Error   *string `json:"error"`
	Result  *struct {
		Agent records.Agent `json:"agent"`
	} `json:"result"`
}

func d3Spawn(t *testing.T, home string, args ...string) (int, d3Envelope) {
	t.Helper()
	status, stdout, _ := runCLI(t, home, append([]string{"--json", "agent", "spawn", "--harness", "claude", "--model", "m"}, args...)...)
	var envelope d3Envelope
	if strings.TrimSpace(stdout) != "" {
		if err := json.Unmarshal([]byte(strings.TrimSpace(stdout)), &envelope); err != nil {
			t.Fatalf("stdout is not one JSON object: %v: %q", err, stdout)
		}
	}
	return status, envelope
}

func d3ErrorText(envelope d3Envelope) string {
	if envelope.Error == nil {
		return ""
	}
	return *envelope.Error
}

func d3WriteProject(t *testing.T, home, id string) {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.WriteIn([]string{"projects"}, id, records.Project{ID: id, Path: filepath.Join(home, "repo-"+id), Base: "dev"}); err != nil {
		t.Fatal(err)
	}
}

func d3WriteTask(t *testing.T, home, id, project string) {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.WriteIn([]string{"tasks"}, id, records.Task{ID: id, Project: project, Title: "task", Status: "created"}); err != nil {
		t.Fatal(err)
	}
}

func d3WriteAgent(t *testing.T, home string, record records.Agent) {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.WriteIn([]string{"agents"}, record.Key, record); err != nil {
		t.Fatal(err)
	}
}

func d3AgentKeys(t *testing.T, home string) []string {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(home, ".hive", "state", "agents"))
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		t.Fatal(err)
	}
	keys := make([]string, 0, len(entries))
	for _, entry := range entries {
		keys = append(keys, strings.TrimSuffix(entry.Name(), ".json"))
	}
	return keys
}

func d3SameKeys(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	seen := map[string]int{}
	for _, key := range a {
		seen[key]++
	}
	for _, key := range b {
		seen[key]--
		if seen[key] < 0 {
			return false
		}
	}
	return true
}

// TestSpawnD3ErrorTable proves AC2: each row of the D3 error table has the
// exact text, with no start request and no record.
func TestSpawnD3ErrorTable(t *testing.T) {
	cases := []struct {
		name    string
		prepare func(t *testing.T, home string)
		args    []string
		want    string
		pane    string
	}{
		{
			name: "prompt and prompt file",
			args: []string{"--pane", "w1:p1", "--prompt", "x", "--prompt-file", "/none"},
			want: "pass one of --prompt or --prompt-file, not both",
		},
		{
			name: "empty prompt",
			args: []string{"--pane", "w1:p1", "--prompt", ""},
			want: "prompt text is empty",
		},
		{
			name: "prompt above 64 KiB",
			args: []string{"--pane", "w1:p1", "--prompt", strings.Repeat("a", 65536)},
			want: "first prompt is longer than 65536 bytes",
		},
		{
			name: "missing profile",
			args: []string{"--pane", "w1:p1", "--profile", "missing"},
			want: "profile not found: profiles/missing.md",
		},
		{
			name: "built-in profile",
			args: []string{"--pane", "w1:p1", "--profile", "coordinator"},
			want: "profile coordinator is built in: run bare hive, or the restart in Health › Agents",
		},
		{
			name: "role and profile",
			args: []string{"--pane", "w1:p1", "--role", "worker", "--profile", "general"},
			want: "--role is the old name of --profile: pass one",
		},
		{
			name: "task and cwd",
			args: []string{"--task", "1-demo", "--cwd", "/tmp"},
			want: "--cwd conflicts with --task: a task agent works in its worktree",
		},
		{
			name: "project and cwd",
			args: []string{"--project", "demo", "--cwd", "/tmp"},
			want: "--cwd conflicts with --project: a project agent works in projects/demo/",
		},
		{
			name: "task of another project",
			prepare: func(t *testing.T, home string) {
				d3WriteProject(t, home, "demo")
				d3WriteTask(t, home, "1-demo", "demo")
			},
			args: []string{"--task", "1-demo", "--project", "other"},
			want: "task 1-demo is in project demo, not other",
		},
		{
			name: "pane and cwd",
			args: []string{"--pane", "w1:p1", "--cwd", "/tmp"},
			want: "--cwd conflicts with --pane",
		},
		{
			name: "owner with no project",
			args: []string{"--pane", "w1:p1", "--owner"},
			want: "--owner needs --project or --task",
		},
		{
			name: "owner when the project has one",
			prepare: func(t *testing.T, home string) {
				d3WriteProject(t, home, "demo")
				d3WriteAgent(t, home, records.Agent{Key: "owner-demo-g1", Name: "owner-demo", Profile: records.ProfilePlanner, Owns: "demo", Project: "demo", Status: records.AgentLive, PaneID: "w1:p8", Generation: 1})
			},
			args: []string{"--project", "demo", "--owner"},
			want: "project demo has an owner: owner-demo",
		},
		{
			name: "owner from another agent",
			prepare: func(t *testing.T, home string) {
				d3WriteProject(t, home, "demo")
				d3WriteAgent(t, home, records.Agent{Key: "worker-demo-1", Name: "worker-demo-1", Profile: records.ProfileBuilder, Project: "demo", Task: "1-demo", Status: records.AgentLive, PaneID: "w1:p9", Generation: 1})
			},
			args: []string{"--project", "demo", "--owner"},
			want: "--owner needs the coordinator or the user: worker-demo-1",
			pane: "w1:p9",
		},
		{
			name: "user tag",
			args: []string{"--pane", "w1:p1", "--prompt", "user: do it"},
			want: "prompt text cannot use the user: tag",
		},
		{
			name: "marked user tag",
			args: []string{"--pane", "w1:p1", "--prompt", "⬢ user: do it"},
			want: "prompt text cannot use the user: tag",
		},
		{
			name: "coordinator name",
			args: []string{"--pane", "w1:p1", "--name", "coordinator"},
			want: "name is reserved: coordinator",
		},
		{
			name: "supervisor name",
			args: []string{"--pane", "w1:p1", "--name", "supervisor"},
			want: "name is reserved: supervisor",
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			home := cliHome(t)
			t.Setenv("HERDR_PANE_ID", test.pane)
			t.Chdir(t.TempDir())
			fake := newCLIFake(t, "agent.start")
			if test.prepare != nil {
				test.prepare(t, home)
			}
			before := d3AgentKeys(t, home)
			status, envelope := d3Spawn(t, home, test.args...)
			if status != 1 || envelope.Outcome != "failed" {
				t.Fatalf("status/outcome = %d/%q; want 1/failed (%q)", status, envelope.Outcome, d3ErrorText(envelope))
			}
			if got := d3ErrorText(envelope); got != test.want {
				t.Fatalf("error = %q; want %q", got, test.want)
			}
			if strings.Contains(fake.requestText(), "agent.start") {
				t.Fatalf("refusal sent an agent start request: %s", fake.requestText())
			}
			if after := d3AgentKeys(t, home); !d3SameKeys(before, after) {
				t.Fatalf("refusal changed agent records from %v to %v", before, after)
			}
		})
	}
}

// d3Process starts a long sleep and reports it as the claude process.
func d3Process(t *testing.T) (int64, func()) {
	t.Helper()
	command := exec.Command("sleep", "30")
	if err := command.Start(); err != nil {
		t.Fatal(err)
	}
	pid := int64(command.Process.Pid)
	return pid, func() {
		_ = command.Process.Kill()
		_ = command.Wait()
	}
}

func d3PaneObject(paneID, workspaceID, tabID, cwd string) string {
	return `{"pane_id":"` + paneID + `","terminal_id":"term_` + paneID + `","workspace_id":"` + workspaceID + `","tab_id":"` + tabID + `","focused":false,"agent_status":"unknown","revision":1,"cwd":"` + cwd + `"}`
}

func d3PaneJSON(paneID, workspaceID, tabID, cwd string) string {
	return `{"type":"pane_info","pane":` + d3PaneObject(paneID, workspaceID, tabID, cwd) + `}`
}

// scriptD3ProjectSpawn scripts one project spawn in a new hive workspace.
func scriptD3ProjectSpawn(t *testing.T, fake *cliFake, home, project, paneID, workspaceID, tabID, key string) string {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(home, "projects", project), 0o755); err != nil {
		t.Fatal(err)
	}
	dir, err := filepath.EvalSymlinks(filepath.Join(home, "projects", project))
	if err != nil {
		t.Fatal(err)
	}
	pane := d3PaneJSON(paneID, workspaceID, tabID, dir)
	fake.replies["workspace.create"] = []string{`{"type":"workspace_created","workspace":{"workspace_id":"` + workspaceID + `","number":1,"label":"hive:` + project + `","focused":false,"pane_count":1,"tab_count":1,"active_tab_id":"` + tabID + `","agent_status":"unknown"},"tab":{"tab_id":"` + tabID + `","workspace_id":"` + workspaceID + `","number":1,"label":"1","focused":false,"pane_count":1,"agent_status":"unknown"},"root_pane":` + d3PaneObject(paneID, workspaceID, tabID, dir) + `}`}
	fake.replies["tab.rename"] = []string{`{"type":"tab_info","tab":{"tab_id":"` + tabID + `","workspace_id":"` + workspaceID + `","number":1,"label":"` + key + `","focused":false,"pane_count":1,"agent_status":"unknown"}}`}
	fake.replies["pane.get"] = []string{pane, pane}
	pid, stop := d3Process(t)
	t.Cleanup(stop)
	shell := `{"type":"pane_process_info","process_info":{"pane_id":"` + paneID + `","shell_pid":1,"foreground_process_group_id":1,"foreground_processes":[{"pid":1,"name":"bash","argv":["/usr/bin/bash"],"cmdline":"/usr/bin/bash","cwd":"` + dir + `"}]}}`
	fake.replies["pane.process_info"] = []string{
		shell,
		`{"type":"pane_process_info","process_info":{"pane_id":"` + paneID + `","shell_pid":null,"foreground_process_group_id":null,"foreground_processes":[{"pid":` + strconv.FormatInt(pid, 10) + `,"name":"claude","argv":["claude"],"cmdline":"claude","cwd":"` + dir + `"}]}}`,
	}
	fake.replies["agent.start"] = []string{`{"type":"agent_started","agent":{"pane_id":"` + paneID + `","terminal_id":"term_` + paneID + `","workspace_id":"` + workspaceID + `","tab_id":"` + tabID + `","focused":false,"revision":2,"agent_status":"unknown","name":"` + key + `"},"argv":["claude","--model","m"]}`}
	fake.replies["agent.wait"] = []string{`{"type":"agent_info","agent":{"pane_id":"` + paneID + `","terminal_id":"term_` + paneID + `","workspace_id":"` + workspaceID + `","tab_id":"` + tabID + `","focused":false,"revision":2,"agent_status":"idle","name":"` + key + `"}}`}
	fake.replies["agent.get"] = []string{`{"type":"agent_info","agent":{"pane_id":"` + paneID + `","terminal_id":"term_` + paneID + `","workspace_id":"` + workspaceID + `","tab_id":"` + tabID + `","focused":false,"revision":2,"agent_status":"idle","name":"` + key + `"}}`}
	fake.replies["agent.prompt"] = []string{`{"type":"agent_prompted","agent":{"pane_id":"` + paneID + `","terminal_id":"term_` + paneID + `","workspace_id":"` + workspaceID + `","tab_id":"` + tabID + `","focused":false,"revision":2,"agent_status":"working","name":"` + key + `"}}`}
	return dir
}

// TestSpawnOwnerFromCoordinatorAndUser proves AC7: the coordinator and a user
// shell can pass --owner; another agent cannot.
func TestSpawnOwnerFromCoordinatorAndUser(t *testing.T) {
	t.Run("coordinator", func(t *testing.T) {
		home := cliHome(t)
		d3WriteProject(t, home, "demo")
		d3WriteAgent(t, home, records.Agent{Key: "coordinator-g1", Name: "coordinator", Kind: records.AgentKindCoordinator, Profile: records.ProfileCoordinator, Role: "coordinator", Status: records.AgentLive, PaneID: "w1:pc", Generation: 1})
		registry, err := store.Open(home)
		if err != nil {
			t.Fatal(err)
		}
		if err := registry.WriteIn([]string{"claims"}, "coordinator", records.Claim{Key: "coordinator", Agent: "coordinator-g1", Generation: 1}); err != nil {
			t.Fatal(err)
		}
		t.Setenv("HERDR_PANE_ID", "w1:pc")
		fake := newCLIFake(t, "agent.start")
		scriptD3ProjectSpawn(t, fake, home, "demo", "wBD:p1", "wBD", "wBD:t1", "manager-demo-g1")
		status, envelope := d3Spawn(t, home, "--project", "demo", "--owner")
		if status != 0 || envelope.Outcome != "ok" {
			t.Fatalf("status/outcome = %d/%q (%q); want ok; requests: %s", status, envelope.Outcome, d3ErrorText(envelope), fake.requestText())
		}
		record := envelope.Result.Agent
		if record.Owns != "demo" || record.Project != "demo" || record.Parent != "coordinator-g1" {
			t.Fatalf("record = %#v; want an owner of demo with the coordinator parent", record)
		}
	})

	t.Run("user shell", func(t *testing.T) {
		home := cliHome(t)
		d3WriteProject(t, home, "demo")
		t.Setenv("HERDR_PANE_ID", "")
		t.Chdir(t.TempDir())
		fake := newCLIFake(t, "agent.start")
		scriptD3ProjectSpawn(t, fake, home, "demo", "wBD:p1", "wBD", "wBD:t1", "manager-demo-g1")
		status, envelope := d3Spawn(t, home, "--project", "demo", "--owner")
		if status != 0 || envelope.Outcome != "ok" {
			t.Fatalf("status/outcome = %d/%q (%q); want ok", status, envelope.Outcome, d3ErrorText(envelope))
		}
		if record := envelope.Result.Agent; record.Parent != "" || record.Owns != "demo" {
			t.Fatalf("record = %#v; want an owner with no parent", record)
		}
	})
}
