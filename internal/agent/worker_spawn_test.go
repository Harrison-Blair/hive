package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	command := exec.Command("git", args...)
	command.Dir = dir
	if out, err := command.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}
}

// taskWorktree makes a real repository with one linked worktree and returns
// the repository path, the worktree path, and the branch.
func taskWorktree(t *testing.T) (string, string, string) {
	t.Helper()
	dir := t.TempDir()
	repo := filepath.Join(dir, "repo")
	runGit(t, "", "init", "-q", repo)
	runGit(t, repo, "config", "user.email", "t@example.com")
	runGit(t, repo, "config", "user.name", "t")
	runGit(t, repo, "commit", "-q", "--allow-empty", "-m", "init")
	runGit(t, repo, "branch", "-M", "dev")
	branch := "task/1-worker-demo"
	worktree := filepath.Join(dir, "wt")
	runGit(t, repo, "worktree", "add", "-q", "-b", branch, worktree)
	return repo, worktree, branch
}

func writeProjectPath(t *testing.T, home, id, path string) {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	record := records.Project{ID: id, Path: path, Base: "dev", RegisteredAt: "2026-01-01T00:00:00Z"}
	if err := registry.WriteIn([]string{"projects"}, id, record); err != nil {
		t.Fatal(err)
	}
}

func writeTaskRecord(t *testing.T, home string, task records.Task) {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.WriteIn([]string{"tasks"}, task.ID, task); err != nil {
		t.Fatal(err)
	}
}

func readTaskRecord(t *testing.T, home, id string) records.Task {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	var task records.Task
	if err := registry.ReadIn([]string{"tasks"}, id, &task); err != nil {
		t.Fatal(err)
	}
	return task
}

func setTaskStatus(t *testing.T, home, id, status string) {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.UpdateIn([]string{"tasks"}, id, func(raw json.RawMessage) (json.RawMessage, error) {
		var task records.Task
		if err := json.Unmarshal(raw, &task); err != nil {
			return nil, err
		}
		task.Status = status
		return json.Marshal(task)
	}); err != nil {
		t.Fatal(err)
	}
}

func workerTask(home, repo, worktree, branch, status string) records.Task {
	return records.Task{
		ID:        "1-worker-demo",
		Project:   "demo",
		Title:     "worker demo",
		Status:    status,
		CreatedAt: "2026-01-01T00:00:00Z",
		UpdatedAt: "2026-01-01T00:00:00Z",
		Worktree: &records.Worktree{
			WorkspaceID: "w5",
			TabID:       "w5:t1",
			RootPane:    "w5:p1",
			Path:        worktree,
			Branch:      branch,
			CreatedAt:   "2026-01-01T00:00:00Z",
		},
	}
}

func workerWorkspaceJSON(workspaceID, label, repoRoot, checkoutPath string) string {
	return `{"workspace_id":"` + workspaceID + `","number":1,"label":"` + label + `","focused":false,"pane_count":1,"tab_count":1,"active_tab_id":"` + workspaceID + `:t1","agent_status":"unknown","worktree":{"repo_key":"` + repoRoot + `/.git","repo_name":"repo","repo_root":"` + repoRoot + `","checkout_path":"` + checkoutPath + `","is_linked_worktree":true}}`
}

func workerWorkspaceGetJSON(workspaceID, label, repoRoot, checkoutPath string) string {
	return `{"type":"workspace_info","workspace":` + workerWorkspaceJSON(workspaceID, label, repoRoot, checkoutPath) + `}`
}

func workerPaneGetJSON(paneID, workspaceID, tabID, cwd, extra string) string {
	return `{"type":"pane_info","pane":` + managerPaneInfoJSON(paneID, "term_"+paneID, workspaceID, tabID, cwd, extra) + `}`
}

func workerPaneStatusJSON(paneID, workspaceID, tabID, cwd, status, extra string) string {
	return `{"type":"pane_info","pane":` + managerPaneStatusJSON(paneID, "term_"+paneID, workspaceID, tabID, cwd, status, extra) + `}`
}

func workerAgentStartJSON(paneID, workspaceID, tabID, name string) string {
	return `{"type":"agent_started","agent":` + agentJSON(paneID, "term_"+paneID, workspaceID, tabID, "unknown", name) + `,"argv":["claude","--model","m"]}`
}

func workerAgentInfoJSON(paneID, workspaceID, tabID, status, name string) string {
	return `{"type":"agent_info","agent":` + agentJSON(paneID, "term_"+paneID, workspaceID, tabID, status, name) + `}`
}

func workerAgentPromptedJSON(paneID, workspaceID, tabID, name string) string {
	return `{"type":"agent_prompted","agent":` + agentJSON(paneID, "term_"+paneID, workspaceID, tabID, "idle", name) + `}`
}

func workerProcessJSON(paneID string, sleep sleepProcess) string {
	return `{"type":"pane_process_info","process_info":` + processInfoJSON(paneID, processEntry(sleep.pid, "claude", "claude")) + `}`
}

func worktreeOpenJSON(workspaceID, label, repoRoot, tabID, paneID, path, branch string, alreadyOpen bool) string {
	flag := "false"
	if alreadyOpen {
		flag = "true"
	}
	return `{"type":"worktree_opened","workspace":` + workerWorkspaceJSON(workspaceID, label, repoRoot, path) +
		`,"tab":` + managerTabJSON(workspaceID, tabID, 1, 1) +
		`,"root_pane":` + managerPaneInfoJSON(paneID, "term_"+paneID, workspaceID, tabID, path, "") +
		`,"worktree":{"path":"` + path + `","label":"repo","is_bare":false,"is_detached":false,"is_prunable":false,"is_linked_worktree":true,"branch":"` + branch + `"}` +
		`,"already_open":` + flag + `}`
}

func tabCreateJSON(workspaceID, tabID, paneID, cwd string) string {
	return `{"type":"tab_created","tab":` + managerTabJSON(workspaceID, tabID, 2, 1) +
		`,"root_pane":` + managerPaneInfoJSON(paneID, "term_"+paneID, workspaceID, tabID, cwd, "") + `}`
}

func requestMethods(f *fakeHerdr) []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	methods := make([]string, 0, len(f.requests))
	for _, request := range f.requests {
		methods = append(methods, request.Method)
	}
	return methods
}

func requireOrder(t *testing.T, f *fakeHerdr, want ...string) {
	t.Helper()
	methods := requestMethods(f)
	index := 0
	for _, method := range methods {
		if index < len(want) && method == want[index] {
			index++
		}
	}
	if index != len(want) {
		t.Fatalf("methods = %v; want the order %v", methods, want)
	}
}

func scriptWorkerSuccess(fake *fakeHerdr, sleep sleepProcess, paneID, workspaceID, tabID, cwd, name string) {
	fake.result("pane.get", workerPaneGetJSON(paneID, workspaceID, tabID, cwd, ""))
	fake.result("workspace.get", workerWorkspaceGetJSON(workspaceID, "1-worker-demo", workspaceRoot(cwd), cwd))
	scriptReadyShell(fake, paneID, "term_"+paneID, workspaceID, tabID, cwd)
	fake.result("agent.start", workerAgentStartJSON(paneID, workspaceID, tabID, name))
	fake.result("pane.get", workerPaneGetJSON(paneID, workspaceID, tabID, cwd, ""))
	fake.result("agent.wait", workerAgentInfoJSON(paneID, workspaceID, tabID, "idle", name))
	fake.result("agent.get", workerAgentInfoJSON(paneID, workspaceID, tabID, "idle", name))
	fake.result("pane.process_info", workerProcessJSON(paneID, sleep))
	fake.result("agent.get", workerAgentInfoJSON(paneID, workspaceID, tabID, "idle", name))
	fake.result("agent.prompt", workerAgentPromptedJSON(paneID, workspaceID, tabID, name))
	fake.result("agent.get", workerAgentInfoJSON(paneID, workspaceID, tabID, "working", name))
}

// workspaceRoot returns the repository root from one worktree path. The tests
// place both under one temporary directory.
func workspaceRoot(worktree string) string {
	return filepath.Join(filepath.Dir(worktree), "repo")
}

func TestWorkerSpawnAssignsAndPromptsIdle(t *testing.T) {
	home := testHome(t)
	repo, worktree, branch := taskWorktree(t)
	writeProjectPath(t, home, "demo", repo)
	writeTaskRecord(t, home, workerTask(home, repo, worktree, branch, "created"))
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	scriptWorkerSuccess(fake, sleep, "w5:p1", "w5", "w5:t1", worktree, "worker-1")

	envelope := Spawn(context.Background(), home, SpawnOptions{
		Role: "worker", Task: "1-worker-demo", Harness: "claude", Model: "m", Name: "worker-1", NameSet: true,
	})
	if envelope.Outcome != "ok" || envelope.Error != nil {
		t.Fatalf("envelope = %#v; want ok", envelope)
	}
	requireOrder(t, fake, "agent.start", "agent.wait", "agent.prompt")
	task := readTaskRecord(t, home, "1-worker-demo")
	if task.Status != "assigned" || task.ActingAgent != "worker-1" {
		t.Fatalf("task = %#v; want assigned to worker-1", task)
	}
	prompts := fake.recorded("agent.prompt")
	if len(prompts) != 1 {
		t.Fatalf("agent.prompt requests = %d; want 1", len(prompts))
	}
	var params struct {
		Target string `json:"target"`
		Text   string `json:"text"`
	}
	if err := json.Unmarshal([]byte(prompts[0].Params), &params); err != nil {
		t.Fatal(err)
	}
	want := workerPrompt(home, "1-worker-demo", worktree, branch)
	if params.Text != want {
		t.Fatalf("prompt = %q; want %q", params.Text, want)
	}
	if params.Target != "w5:p1" {
		t.Fatalf("prompt target = %q; want w5:p1", params.Target)
	}
	record, found, err := loadAgentByName(home, "worker-1")
	if err != nil || !found {
		t.Fatalf("find agent = %v, %v", found, err)
	}
	if record.Project != "demo" || record.Task != "1-worker-demo" || record.Harness != "claude" || record.Model != "m" {
		t.Fatalf("record = %#v", record)
	}
	if record.WorkspaceID != "w5" || record.TabID != "w5:t1" || record.Cwd != worktree || record.QuietAfterMin != 10 {
		t.Fatalf("record topology = %#v", record)
	}
}

func TestWorkerSpawnAssignsAndPromptsDone(t *testing.T) {
	home := testHome(t)
	repo, worktree, branch := taskWorktree(t)
	writeProjectPath(t, home, "demo", repo)
	writeTaskRecord(t, home, workerTask(home, repo, worktree, branch, "assigned"))
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	fake.result("pane.get", workerPaneGetJSON("w5:p1", "w5", "w5:t1", worktree, ""))
	fake.result("workspace.get", workerWorkspaceGetJSON("w5", "1-worker-demo", repo, worktree))
	scriptReadyShell(fake, "w5:p1", "term_w5:p1", "w5", "w5:t1", worktree)
	fake.result("agent.start", workerAgentStartJSON("w5:p1", "w5", "w5:t1", "worker-1"))
	fake.result("pane.get", workerPaneGetJSON("w5:p1", "w5", "w5:t1", worktree, ""))
	fake.result("agent.wait", workerAgentInfoJSON("w5:p1", "w5", "w5:t1", "done", "worker-1"))
	fake.result("agent.get", workerAgentInfoJSON("w5:p1", "w5", "w5:t1", "done", "worker-1"))
	fake.result("pane.process_info", workerProcessJSON("w5:p1", sleep))
	fake.result("agent.get", workerAgentInfoJSON("w5:p1", "w5", "w5:t1", "done", "worker-1"))
	fake.result("agent.prompt", workerAgentPromptedJSON("w5:p1", "w5", "w5:t1", "worker-1"))
	fake.result("agent.get", workerAgentInfoJSON("w5:p1", "w5", "w5:t1", "working", "worker-1"))

	envelope := Spawn(context.Background(), home, SpawnOptions{
		Role: "worker", Task: "1-worker-demo", Harness: "claude", Model: "m", Name: "worker-1", NameSet: true,
	})
	if envelope.Outcome != "ok" {
		t.Fatalf("envelope = %#v; want ok", envelope)
	}
	if task := readTaskRecord(t, home, "1-worker-demo"); task.Status != "assigned" {
		t.Fatalf("task = %#v; want assigned", task)
	}
}

func TestWorkerSpawnRefusals(t *testing.T) {
	cases := []struct {
		name    string
		task    func(home, repo, worktree, branch string) records.Task
		project string
		script  func(fake *fakeHerdr, home, repo, worktree string)
		records func(t *testing.T, home string)
		want    string
	}{
		{
			name: "no worktree",
			task: func(home, repo, worktree, branch string) records.Task {
				task := workerTask(home, repo, worktree, branch, "created")
				task.Worktree = nil
				return task
			},
			want: "task has no worktree: run hive worktree create --task 1-worker-demo",
		},
		{
			name:    "another project",
			project: "other",
			task: func(home, repo, worktree, branch string) records.Task {
				return workerTask(home, repo, worktree, branch, "created")
			},
			want: "task 1-worker-demo is in project demo",
		},
		{
			name: "another workspace",
			task: func(home, repo, worktree, branch string) records.Task {
				return workerTask(home, repo, worktree, branch, "created")
			},
			script: func(fake *fakeHerdr, home, repo, worktree string) {
				fake.result("pane.get", workerPaneGetJSON("w5:p1", "w9", "w5:t1", worktree, ""))
			},
			want: "worktree pane mismatch: workspace_id",
		},
		{
			name: "another cwd",
			task: func(home, repo, worktree, branch string) records.Task {
				return workerTask(home, repo, worktree, branch, "created")
			},
			script: func(fake *fakeHerdr, home, repo, worktree string) {
				fake.result("pane.get", workerPaneGetJSON("w5:p1", "w5", "w5:t1", "/tmp/elsewhere", ""))
			},
			want: "worktree pane mismatch: cwd",
		},
		{
			name: "moved pane",
			task: func(home, repo, worktree, branch string) records.Task {
				return workerTask(home, repo, worktree, branch, "created")
			},
			script: func(fake *fakeHerdr, home, repo, worktree string) {
				fake.result("pane.get", workerPaneGetJSON("w5:p9", "w5", "w5:t1", worktree, ""))
			},
			want: "worktree pane mismatch: pane_id",
		},
		{
			name: "restore error",
			task: func(home, repo, worktree, branch string) records.Task {
				return workerTask(home, repo, worktree, branch, "created")
			},
			script: func(fake *fakeHerdr, home, repo, worktree string) {
				fake.result("pane.get", workerPaneGetJSON("w5:p1", "w5", "w5:t1", worktree, `,"restore_error":"Saved directory is unavailable."`))
			},
			want: "pane not restored: Saved directory is unavailable.",
		},
		{
			name: "busy with name",
			task: func(home, repo, worktree, branch string) records.Task {
				return workerTask(home, repo, worktree, branch, "created")
			},
			script: func(fake *fakeHerdr, home, repo, worktree string) {
				fake.result("pane.get", workerPaneStatusJSON("w5:p1", "w5", "w5:t1", worktree, "idle", `,"agent":"claude"`))
				fake.result("agent.get", workerAgentInfoJSON("w5:p1", "w5", "w5:t1", "idle", "worker-9"))
			},
			want: "pane busy: worker-9",
		},
		{
			name: "busy without name",
			task: func(home, repo, worktree, branch string) records.Task {
				return workerTask(home, repo, worktree, branch, "created")
			},
			script: func(fake *fakeHerdr, home, repo, worktree string) {
				fake.result("pane.get", workerPaneStatusJSON("w5:p1", "w5", "w5:t1", worktree, "idle", `,"agent":"claude"`))
				fake.result("agent.get", workerAgentInfoJSON("w5:p1", "w5", "w5:t1", "idle", ""))
			},
			want: "pane busy: claude agent",
		},
		{
			name: "unresolved record",
			task: func(home, repo, worktree, branch string) records.Task {
				return workerTask(home, repo, worktree, branch, "created")
			},
			script: func(fake *fakeHerdr, home, repo, worktree string) {
				fake.result("pane.get", workerPaneGetJSON("w5:p1", "w5", "w5:t1", worktree, ""))
			},
			records: func(t *testing.T, home string) {
				writeRecord(t, home, records.Agent{Key: "worker-7", Name: "worker-7", Role: "worker", Project: "demo", Task: "1-worker-demo", PaneID: "w5:p1", Status: records.AgentUnresolved})
			},
			want: "pane has an unresolved record: worker-7",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			home := testHome(t)
			repo, worktree, branch := taskWorktree(t)
			writeProjectPath(t, home, "demo", repo)
			writeTaskRecord(t, home, testCase.task(home, repo, worktree, branch))
			if testCase.records != nil {
				testCase.records(t, home)
			}
			fake := newFakeHerdr(t)
			if testCase.script != nil {
				testCase.script(fake, home, repo, worktree)
			}
			options := SpawnOptions{Role: "worker", Task: "1-worker-demo", Harness: "claude", Model: "m", Project: testCase.project}
			envelope := Spawn(context.Background(), home, options)
			if envelope.Outcome != "failed" {
				t.Fatalf("envelope = %#v; want failed", envelope)
			}
			if !strings.Contains(fmt.Sprint(envelope.Error), testCase.want) {
				t.Fatalf("error = %v; want %q", envelope.Error, testCase.want)
			}
			if len(fake.recorded("agent.start")) != 0 {
				t.Fatalf("refusal sent agent.start")
			}
		})
	}
}

func TestWorkerSpawnBlockedKeepsTask(t *testing.T) {
	home := testHome(t)
	repo, worktree, branch := taskWorktree(t)
	writeProjectPath(t, home, "demo", repo)
	writeTaskRecord(t, home, workerTask(home, repo, worktree, branch, "created"))
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	fake.result("pane.get", workerPaneGetJSON("w5:p1", "w5", "w5:t1", worktree, ""))
	fake.result("workspace.get", workerWorkspaceGetJSON("w5", "1-worker-demo", repo, worktree))
	scriptReadyShell(fake, "w5:p1", "term_w5:p1", "w5", "w5:t1", worktree)
	fake.result("agent.start", workerAgentStartJSON("w5:p1", "w5", "w5:t1", "worker-1"))
	fake.result("pane.get", workerPaneGetJSON("w5:p1", "w5", "w5:t1", worktree, ""))
	fake.result("agent.wait", workerAgentInfoJSON("w5:p1", "w5", "w5:t1", "blocked", "worker-1"))
	fake.result("agent.get", workerAgentInfoJSON("w5:p1", "w5", "w5:t1", "blocked", "worker-1"))
	fake.result("pane.process_info", workerProcessJSON("w5:p1", sleep))

	envelope := Spawn(context.Background(), home, SpawnOptions{
		Role: "worker", Task: "1-worker-demo", Harness: "claude", Model: "m", Name: "worker-1", NameSet: true,
	})
	if envelope.Outcome != "partial" {
		t.Fatalf("envelope = %#v; want partial", envelope)
	}
	if !strings.Contains(fmt.Sprint(envelope.Error), "agent blocked: accept the dialog in w5:p1") {
		t.Fatalf("error = %v", envelope.Error)
	}
	if !strings.Contains(fmt.Sprint(envelope.Error), "then run: hive task assign --task 1-worker-demo --agent worker-1") {
		t.Fatalf("error = %v; want the by-hand hint", envelope.Error)
	}
	if len(fake.recorded("agent.prompt")) != 0 {
		t.Fatalf("blocked start sent a prompt")
	}
	if task := readTaskRecord(t, home, "1-worker-demo"); task.Status != "created" || task.ActingAgent != "" {
		t.Fatalf("task = %#v; want unchanged", task)
	}
}

func TestWorkerSpawnTimeoutKeepsTask(t *testing.T) {
	home := testHome(t)
	repo, worktree, branch := taskWorktree(t)
	writeProjectPath(t, home, "demo", repo)
	writeTaskRecord(t, home, workerTask(home, repo, worktree, branch, "created"))
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	fake.result("pane.get", workerPaneGetJSON("w5:p1", "w5", "w5:t1", worktree, ""))
	fake.result("workspace.get", workerWorkspaceGetJSON("w5", "1-worker-demo", repo, worktree))
	scriptReadyShell(fake, "w5:p1", "term_w5:p1", "w5", "w5:t1", worktree)
	fake.result("agent.start", workerAgentStartJSON("w5:p1", "w5", "w5:t1", "worker-1"))
	fake.result("pane.get", workerPaneGetJSON("w5:p1", "w5", "w5:t1", worktree, ""))
	fake.failure("agent.wait", "timeout", "no settled status")
	fake.result("agent.get", workerAgentInfoJSON("w5:p1", "w5", "w5:t1", "working", "worker-1"))
	fake.result("pane.process_info", workerProcessJSON("w5:p1", sleep))

	envelope := Spawn(context.Background(), home, SpawnOptions{
		Role: "worker", Task: "1-worker-demo", Harness: "claude", Model: "m", Name: "worker-1", NameSet: true,
	})
	if envelope.Outcome != "partial" {
		t.Fatalf("envelope = %#v; want partial", envelope)
	}
	if len(fake.recorded("agent.prompt")) != 0 {
		t.Fatalf("timeout sent a prompt")
	}
	if task := readTaskRecord(t, home, "1-worker-demo"); task.Status != "created" || task.ActingAgent != "" {
		t.Fatalf("task = %#v; want unchanged", task)
	}
}

func TestWorkerSpawnChangedStatusSkipsAssignment(t *testing.T) {
	home := testHome(t)
	repo, worktree, branch := taskWorktree(t)
	writeProjectPath(t, home, "demo", repo)
	writeTaskRecord(t, home, workerTask(home, repo, worktree, branch, "created"))
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	scriptWorkerSuccess(fake, sleep, "w5:p1", "w5", "w5:t1", worktree, "worker-1")
	fake.setBefore(func(method string, _ json.RawMessage) {
		if method == "agent.wait" {
			setTaskStatus(t, home, "1-worker-demo", "completed")
		}
	})

	envelope := Spawn(context.Background(), home, SpawnOptions{
		Role: "worker", Task: "1-worker-demo", Harness: "claude", Model: "m", Name: "worker-1", NameSet: true,
	})
	if envelope.Outcome != "partial" {
		t.Fatalf("envelope = %#v; want partial", envelope)
	}
	if !strings.Contains(fmt.Sprint(envelope.Error), "task is completed") {
		t.Fatalf("error = %v; want task is completed", envelope.Error)
	}
	if len(fake.recorded("agent.prompt")) != 0 {
		t.Fatalf("changed status sent a prompt")
	}
	if task := readTaskRecord(t, home, "1-worker-demo"); task.ActingAgent != "" {
		t.Fatalf("task = %#v; want no acting agent", task)
	}
}

func TestWorkerSpawnPromptFailureIsPartial(t *testing.T) {
	home := testHome(t)
	repo, worktree, branch := taskWorktree(t)
	writeProjectPath(t, home, "demo", repo)
	writeTaskRecord(t, home, workerTask(home, repo, worktree, branch, "created"))
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	fake.result("pane.get", workerPaneGetJSON("w5:p1", "w5", "w5:t1", worktree, ""))
	fake.result("workspace.get", workerWorkspaceGetJSON("w5", "1-worker-demo", repo, worktree))
	scriptReadyShell(fake, "w5:p1", "term_w5:p1", "w5", "w5:t1", worktree)
	fake.result("agent.start", workerAgentStartJSON("w5:p1", "w5", "w5:t1", "worker-1"))
	fake.result("pane.get", workerPaneGetJSON("w5:p1", "w5", "w5:t1", worktree, ""))
	fake.result("agent.wait", workerAgentInfoJSON("w5:p1", "w5", "w5:t1", "idle", "worker-1"))
	fake.result("agent.get", workerAgentInfoJSON("w5:p1", "w5", "w5:t1", "idle", "worker-1"))
	fake.result("pane.process_info", workerProcessJSON("w5:p1", sleep))
	fake.result("agent.get", workerAgentInfoJSON("w5:p1", "w5", "w5:t1", "idle", "worker-1"))
	fake.failure("agent.prompt", "agent_prompt_stalled", "the prompt did not settle")

	envelope := Spawn(context.Background(), home, SpawnOptions{
		Role: "worker", Task: "1-worker-demo", Harness: "claude", Model: "m", Name: "worker-1", NameSet: true,
	})
	if envelope.Outcome != "partial" {
		t.Fatalf("envelope = %#v; want partial", envelope)
	}
	if !hasEffectAction(envelope.Effects, "assigned") {
		t.Fatalf("effects = %#v; want task assigned", envelope.Effects)
	}
	if !strings.Contains(fmt.Sprint(envelope.Error), "then run: hive agent send --agent worker-1") {
		t.Fatalf("error = %v; want the by-hand hint", envelope.Error)
	}
	if task := readTaskRecord(t, home, "1-worker-demo"); task.Status != "assigned" || task.ActingAgent != "worker-1" {
		t.Fatalf("task = %#v; want assigned", task)
	}
}

func hasEffectAction(effects []output.Effect, action string) bool {
	for _, effect := range effects {
		if effect.Action == action {
			return true
		}
	}
	return false
}

func loadAgentByName(home, name string) (records.Agent, bool, error) {
	registry, err := store.Open(home)
	if err != nil {
		return records.Agent{}, false, err
	}
	found, err := loadAgentRecords(registry)
	if err != nil {
		return records.Agent{}, false, err
	}
	for _, record := range found {
		if record.Name == name {
			return record, true, nil
		}
	}
	return records.Agent{}, false, nil
}

func TestVerifierSpawnReopensWorktree(t *testing.T) {
	home := testHome(t)
	repo, worktree, branch := taskWorktree(t)
	writeProjectPath(t, home, "demo", repo)
	writeTaskRecord(t, home, workerTask(home, repo, worktree, branch, "completed"))
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	fake.failure("workspace.get", "workspace_not_found", "workspace w5 not found")
	fake.result("worktree.open", worktreeOpenJSON("w6", "1-worker-demo", repo, "w6:t1", "w6:p1", worktree, branch, false))
	fake.result("pane.get", workerPaneGetJSON("w6:p1", "w6", "w6:t1", worktree, ""))
	fake.result("workspace.get", workerWorkspaceGetJSON("w6", "1-worker-demo", repo, worktree))
	scriptReadyShell(fake, "w6:p1", "term_w6:p1", "w6", "w6:t1", worktree)
	fake.result("agent.start", workerAgentStartJSON("w6:p1", "w6", "w6:t1", "verifier-1"))
	fake.result("pane.get", workerPaneGetJSON("w6:p1", "w6", "w6:t1", worktree, ""))
	fake.result("agent.wait", workerAgentInfoJSON("w6:p1", "w6", "w6:t1", "idle", "verifier-1"))
	fake.result("agent.get", workerAgentInfoJSON("w6:p1", "w6", "w6:t1", "idle", "verifier-1"))
	fake.result("pane.process_info", workerProcessJSON("w6:p1", sleep))
	fake.result("agent.get", workerAgentInfoJSON("w6:p1", "w6", "w6:t1", "idle", "verifier-1"))
	fake.result("agent.prompt", workerAgentPromptedJSON("w6:p1", "w6", "w6:t1", "verifier-1"))
	fake.result("agent.get", workerAgentInfoJSON("w6:p1", "w6", "w6:t1", "working", "verifier-1"))
	startStatus := ""
	fake.setBefore(func(method string, _ json.RawMessage) {
		if method == "agent.start" {
			startStatus = readAgent(t, home, "verifier-1").Status
		}
	})

	envelope := Spawn(context.Background(), home, SpawnOptions{
		Role: "verifier", Task: "1-worker-demo", Harness: "claude", Model: "m", Name: "verifier-1", NameSet: true,
	})
	if envelope.Outcome != "ok" {
		t.Fatalf("envelope = %#v; want ok", envelope)
	}
	if startStatus != "starting" || spawnResultAgent(envelope).Status != "live" {
		t.Fatalf("status at agent.start/final = %q/%q; want starting/live", startStatus, spawnResultAgent(envelope).Status)
	}
	opens := fake.recorded("worktree.open")
	if len(opens) != 1 {
		t.Fatalf("worktree.open requests = %d; want 1", len(opens))
	}
	want := `{"cwd":"` + repo + `","path":"` + worktree + `","label":"1-worker-demo","focus":false}`
	if opens[0].Params != want {
		t.Fatalf("worktree.open = %s; want %s", opens[0].Params, want)
	}
	task := readTaskRecord(t, home, "1-worker-demo")
	if task.Worktree == nil || task.Worktree.WorkspaceID != "w6" || task.Worktree.TabID != "w6:t1" || task.Worktree.RootPane != "w6:p1" {
		t.Fatalf("worktree = %#v; want the reopened topology", task.Worktree)
	}
	if task.Status != "completed" || task.ActingAgent != "" {
		t.Fatalf("task = %#v; want completed with no acting agent", task)
	}
	prompts := fake.recorded("agent.prompt")
	if len(prompts) != 1 {
		t.Fatalf("agent.prompt requests = %d; want 1", len(prompts))
	}
	var params struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal([]byte(prompts[0].Params), &params); err != nil {
		t.Fatal(err)
	}
	if params.Text != verifierPrompt(home, "1-worker-demo") {
		t.Fatalf("prompt = %q", params.Text)
	}
	if len(fake.recorded("workspace.close")) != 0 {
		t.Fatalf("success closed the workspace")
	}
}

func TestVerifierSpawnMakesTab(t *testing.T) {
	home := testHome(t)
	repo, worktree, branch := taskWorktree(t)
	writeProjectPath(t, home, "demo", repo)
	writeTaskRecord(t, home, workerTask(home, repo, worktree, branch, "completed"))
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	fake.result("workspace.get", workerWorkspaceGetJSON("w5", "1-worker-demo", repo, worktree))
	fake.result("tab.create", tabCreateJSON("w5", "w5:t2", "w5:p2", worktree))
	fake.result("pane.get", workerPaneGetJSON("w5:p2", "w5", "w5:t2", worktree, ""))
	fake.result("workspace.get", workerWorkspaceGetJSON("w5", "1-worker-demo", repo, worktree))
	scriptReadyShell(fake, "w5:p2", "term_w5:p2", "w5", "w5:t2", worktree)
	fake.result("agent.start", workerAgentStartJSON("w5:p2", "w5", "w5:t2", "verifier-1"))
	fake.result("pane.get", workerPaneGetJSON("w5:p2", "w5", "w5:t2", worktree, ""))
	fake.result("agent.wait", workerAgentInfoJSON("w5:p2", "w5", "w5:t2", "idle", "verifier-1"))
	fake.result("agent.get", workerAgentInfoJSON("w5:p2", "w5", "w5:t2", "idle", "verifier-1"))
	fake.result("pane.process_info", workerProcessJSON("w5:p2", sleep))
	fake.result("agent.get", workerAgentInfoJSON("w5:p2", "w5", "w5:t2", "idle", "verifier-1"))
	fake.result("agent.prompt", workerAgentPromptedJSON("w5:p2", "w5", "w5:t2", "verifier-1"))
	fake.result("agent.get", workerAgentInfoJSON("w5:p2", "w5", "w5:t2", "working", "verifier-1"))

	envelope := Spawn(context.Background(), home, SpawnOptions{
		Role: "verifier", Task: "1-worker-demo", Harness: "claude", Model: "m", Name: "verifier-1", NameSet: true,
	})
	if envelope.Outcome != "ok" {
		t.Fatalf("envelope = %#v; want ok", envelope)
	}
	creates := fake.recorded("tab.create")
	if len(creates) != 1 {
		t.Fatalf("tab.create requests = %d; want 1", len(creates))
	}
	want := `{"workspace_id":"w5","cwd":"` + worktree + `","focus":false}`
	if creates[0].Params != want {
		t.Fatalf("tab.create = %s; want %s", creates[0].Params, want)
	}
	if len(fake.recorded("worktree.open")) != 0 {
		t.Fatalf("found workspace called worktree.open")
	}
	task := readTaskRecord(t, home, "1-worker-demo")
	if task.Worktree == nil || task.Worktree.RootPane != "w5:p2" || task.Worktree.TabID != "w5:t2" {
		t.Fatalf("worktree = %#v", task.Worktree)
	}
}

func TestVerifierSpawnRefusals(t *testing.T) {
	t.Run("status", func(t *testing.T) {
		home := testHome(t)
		repo, worktree, branch := taskWorktree(t)
		writeProjectPath(t, home, "demo", repo)
		writeTaskRecord(t, home, workerTask(home, repo, worktree, branch, "assigned"))
		fake := newFakeHerdr(t)
		envelope := Spawn(context.Background(), home, SpawnOptions{Role: "verifier", Task: "1-worker-demo", Harness: "claude", Model: "m"})
		if envelope.Outcome != "failed" || !strings.Contains(fmt.Sprint(envelope.Error), "task is assigned") {
			t.Fatalf("envelope = %#v; want task is assigned", envelope)
		}
		if len(fake.recorded("workspace.get")) != 0 {
			t.Fatalf("status refusal read a workspace")
		}
	})
	t.Run("live worker", func(t *testing.T) {
		home := testHome(t)
		repo, worktree, branch := taskWorktree(t)
		writeProjectPath(t, home, "demo", repo)
		writeTaskRecord(t, home, workerTask(home, repo, worktree, branch, "completed"))
		writeRecord(t, home, records.Agent{Key: "worker-1", Name: "worker-1", Role: "worker", Project: "demo", Task: "1-worker-demo", PaneID: "w5:p1", Status: records.AgentLive})
		fake := newFakeHerdr(t)
		envelope := Spawn(context.Background(), home, SpawnOptions{Role: "verifier", Task: "1-worker-demo", Harness: "claude", Model: "m"})
		if envelope.Outcome != "failed" || !strings.Contains(fmt.Sprint(envelope.Error), "worker still live: worker-1") {
			t.Fatalf("envelope = %#v; want worker still live", envelope)
		}
		if len(fake.recorded("workspace.get")) != 0 {
			t.Fatalf("live refusal read a workspace")
		}
	})
}

func TestVerifierSpawnAlreadyOpenKeepsTopology(t *testing.T) {
	home := testHome(t)
	repo, worktree, branch := taskWorktree(t)
	writeProjectPath(t, home, "demo", repo)
	writeTaskRecord(t, home, workerTask(home, repo, worktree, branch, "completed"))
	fake := newFakeHerdr(t)
	fake.failure("workspace.get", "workspace_not_found", "workspace w5 not found")
	fake.result("worktree.open", worktreeOpenJSON("w6", "1-worker-demo", repo, "w6:t1", "w6:p1", worktree, branch, true))
	fake.result("pane.get", workerPaneGetJSON("w6:p1", "w6", "w6:t1", worktree, ""))
	fake.result("workspace.get", workerWorkspaceGetJSON("w6", "1-worker-demo", repo, worktree))
	scriptReadyShell(fake, "w6:p1", "term_w6:p1", "w6", "w6:t1", worktree)
	fake.failure("agent.start", "agent_pane_busy", "pane is busy")

	envelope := Spawn(context.Background(), home, SpawnOptions{Role: "verifier", Task: "1-worker-demo", Harness: "claude", Model: "m", Name: "verifier-1", NameSet: true})
	if envelope.Outcome != "failed" {
		t.Fatalf("envelope = %#v; want failed", envelope)
	}
	if len(fake.recorded("workspace.close")) != 0 {
		t.Fatalf("already_open spawned a close: %#v", fake.recorded("workspace.close"))
	}
}

func TestVerifierSpawnCleanupAfterNoStart(t *testing.T) {
	home := testHome(t)
	repo, worktree, branch := taskWorktree(t)
	writeProjectPath(t, home, "demo", repo)
	writeTaskRecord(t, home, workerTask(home, repo, worktree, branch, "completed"))
	fake := newFakeHerdr(t)
	fake.failure("workspace.get", "workspace_not_found", "workspace w5 not found")
	fake.result("worktree.open", worktreeOpenJSON("w6", "1-worker-demo", repo, "w6:t1", "w6:p1", worktree, branch, false))
	fake.result("pane.get", workerPaneGetJSON("w6:p1", "w6", "w6:t1", worktree, ""))
	fake.result("workspace.get", workerWorkspaceGetJSON("w6", "1-worker-demo", repo, worktree))
	scriptReadyShell(fake, "w6:p1", "term_w6:p1", "w6", "w6:t1", worktree)
	fake.failure("agent.start", "agent_pane_busy", "pane is busy")
	fake.result("workspace.get", workerWorkspaceGetJSON("w6", "1-worker-demo", repo, worktree))
	fake.result("workspace.close", `{"type":"ok"}`)

	envelope := Spawn(context.Background(), home, SpawnOptions{Role: "verifier", Task: "1-worker-demo", Harness: "claude", Model: "m", Name: "verifier-1", NameSet: true})
	if envelope.Outcome != "failed" {
		t.Fatalf("envelope = %#v; want failed", envelope)
	}
	closes := fake.recorded("workspace.close")
	if len(closes) != 1 || closes[0].Params != `{"workspace_id":"w6","close_group":false}` {
		t.Fatalf("workspace.close = %#v; want one close of w6", closes)
	}
}

func TestVerifierSpawnKeepsTopologyAfterLostStart(t *testing.T) {
	home := testHome(t)
	repo, worktree, branch := taskWorktree(t)
	writeProjectPath(t, home, "demo", repo)
	writeTaskRecord(t, home, workerTask(home, repo, worktree, branch, "completed"))
	fake := newFakeHerdr(t)
	fake.failure("workspace.get", "workspace_not_found", "workspace w5 not found")
	fake.result("worktree.open", worktreeOpenJSON("w6", "1-worker-demo", repo, "w6:t1", "w6:p1", worktree, branch, false))
	fake.result("pane.get", workerPaneGetJSON("w6:p1", "w6", "w6:t1", worktree, ""))
	fake.result("workspace.get", workerWorkspaceGetJSON("w6", "1-worker-demo", repo, worktree))
	scriptReadyShell(fake, "w6:p1", "term_w6:p1", "w6", "w6:t1", worktree)
	fake.drop("agent.start")

	envelope := Spawn(context.Background(), home, SpawnOptions{Role: "verifier", Task: "1-worker-demo", Harness: "claude", Model: "m", Name: "verifier-1", NameSet: true})
	if envelope.Outcome != "partial" {
		t.Fatalf("envelope = %#v; want partial", envelope)
	}
	if len(fake.recorded("workspace.close")) != 0 || len(fake.recorded("pane.close")) != 0 {
		t.Fatalf("uncertain start cleaned up the topology")
	}
	task := readTaskRecord(t, home, "1-worker-demo")
	if task.Worktree == nil || task.Worktree.WorkspaceID != "w6" {
		t.Fatalf("worktree = %#v; want the preserved topology", task.Worktree)
	}
}

func TestVerifierSpawnTopologyWriteFailureCleansUp(t *testing.T) {
	home := testHome(t)
	repo, worktree, branch := taskWorktree(t)
	writeProjectPath(t, home, "demo", repo)
	writeTaskRecord(t, home, workerTask(home, repo, worktree, branch, "completed"))
	fake := newFakeHerdr(t)
	fake.failure("workspace.get", "workspace_not_found", "workspace w5 not found")
	fake.result("worktree.open", worktreeOpenJSON("w6", "1-worker-demo", repo, "w6:t1", "w6:p1", worktree, branch, false))
	fake.result("workspace.get", workerWorkspaceGetJSON("w6", "1-worker-demo", repo, worktree))
	fake.result("workspace.close", `{"type":"ok"}`)
	tasksDir := filepath.Join(home, ".hive", "state", "tasks")
	if err := os.Chmod(tasksDir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(tasksDir, 0o755) })

	envelope := Spawn(context.Background(), home, SpawnOptions{Role: "verifier", Task: "1-worker-demo", Harness: "claude", Model: "m", Name: "verifier-1", NameSet: true})
	if envelope.Outcome != "failed" || !strings.Contains(fmt.Sprint(envelope.Error), "store:") {
		t.Fatalf("envelope = %#v; want a store failure", envelope)
	}
	if len(fake.recorded("agent.start")) != 0 {
		t.Fatalf("topology write failure started an agent")
	}
	if len(fake.recorded("workspace.close")) != 1 {
		t.Fatalf("workspace.close requests = %d; want 1", len(fake.recorded("workspace.close")))
	}
}

func workerAgentInteractiveJSON(paneID, workspaceID, tabID, status, name string) string {
	return `{"type":"agent_info","agent":{"pane_id":"` + paneID + `","terminal_id":"term_` + paneID + `","workspace_id":"` + workspaceID + `","tab_id":"` + tabID + `","agent_status":"` + status + `","focused":false,"revision":2,"name":"` + name + `","interactive_ready":true}}`
}

func TestWorkerSpawnWaitsForInteractiveAgent(t *testing.T) {
	home := testHome(t)
	repo, worktree, branch := taskWorktree(t)
	writeProjectPath(t, home, "demo", repo)
	writeTaskRecord(t, home, workerTask(home, repo, worktree, branch, "created"))
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	fake.result("pane.get", workerPaneGetJSON("w5:p1", "w5", "w5:t1", worktree, ""))
	fake.result("workspace.get", workerWorkspaceGetJSON("w5", "1-worker-demo", repo, worktree))
	scriptReadyShell(fake, "w5:p1", "term_w5:p1", "w5", "w5:t1", worktree)
	fake.result("agent.start", workerAgentStartJSON("w5:p1", "w5", "w5:t1", "worker-1"))
	fake.result("pane.get", workerPaneGetJSON("w5:p1", "w5", "w5:t1", worktree, ""))
	fake.result("agent.wait", workerAgentInfoJSON("w5:p1", "w5", "w5:t1", "idle", "worker-1"))
	fake.result("agent.get", workerAgentInfoJSON("w5:p1", "w5", "w5:t1", "idle", "worker-1"))
	fake.result("pane.process_info", workerProcessJSON("w5:p1", sleep))
	fake.result("agent.get", workerAgentInfoJSON("w5:p1", "w5", "w5:t1", "idle", "worker-1"))
	fake.failure("agent.prompt", "agent_not_ready", "agent worker-1 is not an active named agent")
	fake.result("agent.get", workerAgentInteractiveJSON("w5:p1", "w5", "w5:t1", "idle", "worker-1"))
	fake.result("agent.prompt", workerAgentPromptedJSON("w5:p1", "w5", "w5:t1", "worker-1"))
	fake.result("agent.get", workerAgentInfoJSON("w5:p1", "w5", "w5:t1", "working", "worker-1"))

	envelope := Spawn(context.Background(), home, SpawnOptions{
		Role: "worker", Task: "1-worker-demo", Harness: "claude", Model: "m", Name: "worker-1", NameSet: true,
	})
	if envelope.Outcome != "ok" {
		t.Fatalf("envelope = %#v; want ok", envelope)
	}
	if len(fake.recorded("agent.prompt")) != 2 {
		t.Fatalf("agent.prompt requests = %d; want 2", len(fake.recorded("agent.prompt")))
	}
	if task := readTaskRecord(t, home, "1-worker-demo"); task.Status != "assigned" {
		t.Fatalf("task = %#v; want assigned", task)
	}
}

func TestWorkerSpawnResendsLostPrompt(t *testing.T) {
	home := testHome(t)
	repo, worktree, branch := taskWorktree(t)
	writeProjectPath(t, home, "demo", repo)
	writeTaskRecord(t, home, workerTask(home, repo, worktree, branch, "created"))
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	fake.result("pane.get", workerPaneGetJSON("w5:p1", "w5", "w5:t1", worktree, ""))
	fake.result("workspace.get", workerWorkspaceGetJSON("w5", "1-worker-demo", repo, worktree))
	scriptReadyShell(fake, "w5:p1", "term_w5:p1", "w5", "w5:t1", worktree)
	fake.result("agent.start", workerAgentStartJSON("w5:p1", "w5", "w5:t1", "worker-1"))
	fake.result("pane.get", workerPaneGetJSON("w5:p1", "w5", "w5:t1", worktree, ""))
	fake.result("agent.wait", workerAgentInfoJSON("w5:p1", "w5", "w5:t1", "idle", "worker-1"))
	fake.result("agent.get", workerAgentInfoJSON("w5:p1", "w5", "w5:t1", "idle", "worker-1"))
	fake.result("pane.process_info", workerProcessJSON("w5:p1", sleep))
	fake.result("agent.get", workerAgentInfoJSON("w5:p1", "w5", "w5:t1", "unknown", "worker-1"))
	fake.result("agent.prompt", workerAgentPromptedJSON("w5:p1", "w5", "w5:t1", "worker-1"))
	fake.result("agent.get", workerAgentInfoJSON("w5:p1", "w5", "w5:t1", "unknown", "worker-1"))
	fake.result("agent.prompt", workerAgentPromptedJSON("w5:p1", "w5", "w5:t1", "worker-1"))
	fake.result("agent.get", workerAgentInfoJSON("w5:p1", "w5", "w5:t1", "working", "worker-1"))

	envelope := Spawn(context.Background(), home, SpawnOptions{
		Role: "worker", Task: "1-worker-demo", Harness: "claude", Model: "m", Name: "worker-1", NameSet: true,
	})
	if envelope.Outcome != "ok" {
		t.Fatalf("envelope = %#v; want ok", envelope)
	}
	if len(fake.recorded("agent.prompt")) != 2 {
		t.Fatalf("agent.prompt requests = %d; want 2", len(fake.recorded("agent.prompt")))
	}
	if task := readTaskRecord(t, home, "1-worker-demo"); task.Status != "assigned" {
		t.Fatalf("task = %#v; want assigned", task)
	}
}
