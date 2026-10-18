package worktree

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/Harrison-Blair/hive/internal/lib/herdr"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

type fakeRequest struct {
	ID     string          `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

func (r fakeRequest) stringParam(t *testing.T, field string) string {
	t.Helper()
	var values map[string]json.RawMessage
	if err := json.Unmarshal(r.Params, &values); err != nil {
		t.Fatalf("params %s: %v", r.Params, err)
	}
	raw, found := values[field]
	if !found {
		return ""
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		t.Fatalf("param %s: %v", field, err)
	}
	return value
}

func (r fakeRequest) hasParam(field string) bool {
	var values map[string]json.RawMessage
	if err := json.Unmarshal(r.Params, &values); err != nil {
		return false
	}
	_, found := values[field]
	return found
}

func (r fakeRequest) boolParam(t *testing.T, field string) bool {
	t.Helper()
	var values map[string]json.RawMessage
	if err := json.Unmarshal(r.Params, &values); err != nil {
		t.Fatalf("params %s: %v", r.Params, err)
	}
	raw, found := values[field]
	if !found {
		return false
	}
	var value bool
	if err := json.Unmarshal(raw, &value); err != nil {
		t.Fatalf("param %s: %v", field, err)
	}
	return value
}

type fakeReply struct {
	Result string
	Error  string
	Delay  time.Duration
	Hold   <-chan struct{}
	Close  bool
}

type fakeHerdr struct {
	t       *testing.T
	path    string
	ln      net.Listener
	done    chan struct{}
	once    sync.Once
	wg      sync.WaitGroup
	mu      sync.Mutex
	reqs    []fakeRequest
	handler func(fakeRequest) fakeReply
}

func startFakeHerdr(t *testing.T, handler func(fakeRequest) fakeReply) *fakeHerdr {
	t.Helper()
	dir, err := os.MkdirTemp("", "t09sock-")
	if err != nil {
		t.Fatalf("temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	path := filepath.Join(dir, "herdr.sock")
	listener, err := net.Listen("unix", path)
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	fake := &fakeHerdr{t: t, path: path, ln: listener, done: make(chan struct{}), handler: handler}
	fake.wg.Add(1)
	go func() {
		defer fake.wg.Done()
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			fake.wg.Add(1)
			go fake.serve(conn)
		}
	}()
	t.Cleanup(fake.Close)
	t.Setenv("HERDR_SOCKET_PATH", path)
	return fake
}

func (f *fakeHerdr) serve(conn net.Conn) {
	defer f.wg.Done()
	defer func() { _ = conn.Close() }()
	var request fakeRequest
	if err := json.NewDecoder(conn).Decode(&request); err != nil {
		return
	}
	if request.Method == "ping" {
		f.write(conn, request.ID, `{"type":"pong","version":"0.9.5","protocol":22}`, "")
		return
	}
	f.mu.Lock()
	f.reqs = append(f.reqs, request)
	f.mu.Unlock()
	reply := f.handler(request)
	if reply.Hold != nil {
		select {
		case <-reply.Hold:
		case <-f.done:
			return
		}
	}
	if reply.Delay > 0 {
		time.Sleep(reply.Delay)
	}
	if reply.Close {
		return
	}
	f.write(conn, request.ID, reply.Result, reply.Error)
}

func (f *fakeHerdr) write(conn net.Conn, id, result, errText string) {
	switch {
	case errText != "":
		_, _ = fmt.Fprintf(conn, `{"id":%q,"error":%s}`+"\n", id, errText)
	case result != "":
		_, _ = fmt.Fprintf(conn, `{"id":%q,"result":%s}`+"\n", id, result)
	}
}

func (f *fakeHerdr) Close() {
	f.once.Do(func() {
		close(f.done)
		_ = f.ln.Close()
		f.wg.Wait()
	})
}

func (f *fakeHerdr) requests() []fakeRequest {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]fakeRequest(nil), f.reqs...)
}

func (f *fakeHerdr) methods() []string {
	var methods []string
	for _, request := range f.requests() {
		methods = append(methods, request.Method)
	}
	return methods
}

func (f *fakeHerdr) requestFor(method string) (fakeRequest, bool) {
	for _, request := range f.requests() {
		if request.Method == method {
			return request, true
		}
	}
	return fakeRequest{}, false
}

func jsonText(t *testing.T, value any) string {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return string(data)
}

func workspaceListReply(t *testing.T, workspaces ...herdr.WorkspaceInfo) string {
	t.Helper()
	return jsonText(t, struct {
		Type       string                `json:"type"`
		Workspaces []herdr.WorkspaceInfo `json:"workspaces"`
	}{Type: "workspace_list", Workspaces: workspaces})
}

func workspaceInfoReply(t *testing.T, workspace herdr.WorkspaceInfo) string {
	t.Helper()
	return jsonText(t, struct {
		Type      string              `json:"type"`
		Workspace herdr.WorkspaceInfo `json:"workspace"`
	}{Type: "workspace_info", Workspace: workspace})
}

func linkedWorkspace(id, path, repo string) herdr.WorkspaceInfo {
	return herdr.WorkspaceInfo{
		WorkspaceID: id,
		Label:       filepath.Base(path),
		ActiveTabID: id + ":t1",
		AgentStatus: "unknown",
		Worktree: &herdr.WorkspaceWorktreeInfo{
			RepoKey:          filepath.Join(repo, ".git"),
			RepoName:         filepath.Base(repo),
			RepoRoot:         repo,
			CheckoutPath:     path,
			IsLinkedWorktree: true,
		},
	}
}

func repositoryWorkspace(id, repo string) herdr.WorkspaceInfo {
	return herdr.WorkspaceInfo{
		WorkspaceID: id,
		Label:       filepath.Base(repo),
		ActiveTabID: id + ":t1",
		AgentStatus: "unknown",
		Worktree: &herdr.WorkspaceWorktreeInfo{
			RepoKey:          filepath.Join(repo, ".git"),
			RepoName:         filepath.Base(repo),
			RepoRoot:         repo,
			CheckoutPath:     repo,
			IsLinkedWorktree: false,
		},
	}
}

func createReply(t *testing.T, workspaceID, tabID, paneID, path, branch, repo string) string {
	t.Helper()
	return jsonText(t, herdr.WorktreeCreateResult{
		Type: "worktree_created",
		Workspace: herdr.WorkspaceInfo{
			WorkspaceID: workspaceID,
			Label:       filepath.Base(path),
			ActiveTabID: tabID,
			AgentStatus: "unknown",
			Worktree: &herdr.WorkspaceWorktreeInfo{
				RepoKey:          filepath.Join(repo, ".git"),
				RepoName:         filepath.Base(repo),
				RepoRoot:         repo,
				CheckoutPath:     path,
				IsLinkedWorktree: true,
			},
		},
		Tab: herdr.TabInfo{
			TabID:       tabID,
			WorkspaceID: workspaceID,
			Number:      1,
			Label:       "1",
			AgentStatus: "unknown",
		},
		RootPane: herdr.PaneInfo{
			PaneID:      paneID,
			TerminalID:  "term_test",
			WorkspaceID: workspaceID,
			TabID:       tabID,
			AgentStatus: "unknown",
			Cwd:         path,
		},
		Worktree: herdr.WorktreeInfo{
			Path:             path,
			Label:            filepath.Base(repo),
			IsLinkedWorktree: true,
			Branch:           branch,
		},
	})
}

func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_AUTHOR_NAME=Hive Test",
		"GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=Hive Test",
		"GIT_COMMITTER_EMAIL=test@example.com",
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", args, err, output)
	}
	return string(output)
}

func newTestRepo(t *testing.T) string {
	t.Helper()
	repo := filepath.Join(t.TempDir(), "repo")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatalf("mkdir repo: %v", err)
	}
	runGit(t, repo, "init", "-q", "-b", "main")
	runGit(t, repo, "config", "user.name", "Hive Test")
	runGit(t, repo, "config", "user.email", "test@example.com")
	if err := os.WriteFile(filepath.Join(repo, "README.md"), []byte("base\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	runGit(t, repo, "add", "README.md")
	runGit(t, repo, "commit", "-q", "-m", "init")
	return repo
}

func newTestHome(t *testing.T, projectID string) string {
	t.Helper()
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, ".hive"), 0o755); err != nil {
		t.Fatalf("mkdir .hive: %v", err)
	}
	if err := os.WriteFile(filepath.Join(home, ".hive", "README.md"), []byte("home\n"), 0o644); err != nil {
		t.Fatalf("write marker: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(home, "projects", projectID, "worktrees"), 0o755); err != nil {
		t.Fatalf("mkdir project worktrees: %v", err)
	}
	return home
}

func writeProject(t *testing.T, home string, project records.Project) {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	if err := registry.Write("projects", project.ID, project); err != nil {
		t.Fatalf("write project: %v", err)
	}
}

func writeTask(t *testing.T, home string, task records.Task) {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	if err := registry.WriteIn([]string{"tasks"}, task.ID, task); err != nil {
		t.Fatalf("write task: %v", err)
	}
}

func readTask(t *testing.T, home, taskID string) records.Task {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	var task records.Task
	if err := registry.ReadIn([]string{"tasks"}, taskID, &task); err != nil {
		t.Fatalf("read task: %v", err)
	}
	return task
}

func readProject(t *testing.T, home, projectID string) records.Project {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	var project records.Project
	if err := registry.Read("projects", projectID, &project); err != nil {
		t.Fatalf("read project: %v", err)
	}
	return project
}

func addLinkedWorktree(t *testing.T, repo, path, branch, base string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir worktree parent: %v", err)
	}
	runGit(t, repo, "worktree", "add", "-q", "-b", branch, path, base)
}
