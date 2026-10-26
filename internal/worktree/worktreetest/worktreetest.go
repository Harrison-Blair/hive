// Package worktreetest holds the fixture helpers that the tests of the worktree
// operation packages share. Only tests import it.
package worktreetest

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Harrison-Blair/hive/internal/lib/herdr"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// Project returns the fixture project record of the repository.
func Project(repo string) records.Project {
	return records.Project{ID: "repo-live", Path: repo, Base: "main", RegisteredAt: "2026-10-18T00:00:00Z"}
}

// Task returns the fixture task record with no worktree.
func Task() records.Task {
	return records.Task{
		ID:        "1-live-demo",
		Project:   "repo-live",
		Title:     "live demo",
		Status:    "created",
		DependsOn: []string{},
		CreatedAt: "2026-10-18T00:00:00Z",
		UpdatedAt: "2026-10-18T00:00:00Z",
	}
}

// HasEffect reports whether the envelope has the effect.
func HasEffect(envelope output.Envelope, resource, action, path string) bool {
	for _, effect := range envelope.Effects {
		if effect.Resource == resource && effect.Action == action && effect.Path == path {
			return true
		}
	}
	return false
}

// WorktreeTask returns the fixture verified task record with a worktree.
func WorktreeTask(path, branch string) records.Task {
	task := Task()
	task.Status = "verified"
	task.Worktree = &records.Worktree{
		WorkspaceID: "w2",
		TabID:       "w2:t1",
		RootPane:    "w2:p1",
		Path:        path,
		Branch:      branch,
		CreatedAt:   "2026-10-18T00:00:00Z",
	}
	return task
}

// WriteTaskAt writes a task record at the key.
func WriteTaskAt(t *testing.T, home, key string, task records.Task) {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	if err := registry.WriteIn([]string{"tasks"}, key, task); err != nil {
		t.Fatalf("write task at %s: %v", key, err)
	}
}

// WriteProjectAt writes a project record at the key.
func WriteProjectAt(t *testing.T, home, key string, project records.Project) {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	if err := registry.Write("projects", key, project); err != nil {
		t.Fatalf("write project at %s: %v", key, err)
	}
}

// PermittedPath returns the worktree path that hive permits for the task.
func PermittedPath(home, projectID, taskID string) string {
	return filepath.Join(home, "projects", projectID, "worktrees", taskID)
}

// LogEvents returns each log event of the home.
func LogEvents(t *testing.T, home string) []map[string]any {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(home, ".hive", "logs", "*.jsonl"))
	if err != nil {
		t.Fatalf("glob logs: %v", err)
	}
	events := []map[string]any{}
	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read log %s: %v", path, err)
		}
		for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
			if line == "" {
				continue
			}
			var event map[string]any
			if err := json.Unmarshal([]byte(line), &event); err != nil {
				t.Fatalf("parse log line: %v", err)
			}
			events = append(events, event)
		}
	}
	return events
}

// ExternalEvents returns each herdr and process log event of the home.
func ExternalEvents(t *testing.T, home string) []map[string]any {
	t.Helper()
	external := []map[string]any{}
	for _, event := range LogEvents(t, home) {
		message, _ := event["msg"].(string)
		switch message {
		case "herdr request", "herdr reply", "process start", "process end":
			external = append(external, event)
		}
	}
	return external
}

// FakeRequest is one request that the fake herdr received.
type FakeRequest struct {
	ID     string          `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

// StringParam returns one string parameter of the request.
func (r FakeRequest) StringParam(t *testing.T, field string) string {
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

// HasParam reports whether the request has the parameter.
func (r FakeRequest) HasParam(field string) bool {
	var values map[string]json.RawMessage
	if err := json.Unmarshal(r.Params, &values); err != nil {
		return false
	}
	_, found := values[field]
	return found
}

// BoolParam returns one boolean parameter of the request.
func (r FakeRequest) BoolParam(t *testing.T, field string) bool {
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

// FakeReply is the reply of the fake herdr to one request.
type FakeReply struct {
	Result string
	Error  string
	Delay  time.Duration
	Hold   <-chan struct{}
	Close  bool
}

// FakeHerdr is a herdr socket server that a test controls.
type FakeHerdr struct {
	t       *testing.T
	path    string
	ln      net.Listener
	done    chan struct{}
	once    sync.Once
	wg      sync.WaitGroup
	mu      sync.Mutex
	reqs    []FakeRequest
	handler func(FakeRequest) FakeReply
}

// StartFakeHerdr starts a fake herdr and points HERDR_SOCKET_PATH at it.
func StartFakeHerdr(t *testing.T, handler func(FakeRequest) FakeReply) *FakeHerdr {
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
	fake := &FakeHerdr{t: t, path: path, ln: listener, done: make(chan struct{}), handler: handler}
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

func (f *FakeHerdr) serve(conn net.Conn) {
	defer f.wg.Done()
	defer func() { _ = conn.Close() }()
	var request FakeRequest
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

func (f *FakeHerdr) write(conn net.Conn, id, result, errText string) {
	switch {
	case errText != "":
		_, _ = fmt.Fprintf(conn, `{"id":%q,"error":%s}`+"\n", id, errText)
	case result != "":
		_, _ = fmt.Fprintf(conn, `{"id":%q,"result":%s}`+"\n", id, result)
	}
}

// Close stops the fake herdr and waits for each connection.
func (f *FakeHerdr) Close() {
	f.once.Do(func() {
		close(f.done)
		_ = f.ln.Close()
		f.wg.Wait()
	})
}

// Requests returns each request other than ping.
func (f *FakeHerdr) Requests() []FakeRequest {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]FakeRequest(nil), f.reqs...)
}

// Methods returns the method of each request other than ping.
func (f *FakeHerdr) Methods() []string {
	var methods []string
	for _, request := range f.Requests() {
		methods = append(methods, request.Method)
	}
	return methods
}

// RequestFor returns the first request of the method.
func (f *FakeHerdr) RequestFor(method string) (FakeRequest, bool) {
	for _, request := range f.Requests() {
		if request.Method == method {
			return request, true
		}
	}
	return FakeRequest{}, false
}

// JSONText returns the JSON text of the value.
func JSONText(t *testing.T, value any) string {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return string(data)
}

// WorkspaceListReply returns a workspace.list result.
func WorkspaceListReply(t *testing.T, workspaces ...herdr.WorkspaceInfo) string {
	t.Helper()
	return JSONText(t, struct {
		Type       string                `json:"type"`
		Workspaces []herdr.WorkspaceInfo `json:"workspaces"`
	}{Type: "workspace_list", Workspaces: workspaces})
}

// WorkspaceInfoReply returns a workspace.get result.
func WorkspaceInfoReply(t *testing.T, workspace herdr.WorkspaceInfo) string {
	t.Helper()
	return JSONText(t, struct {
		Type      string              `json:"type"`
		Workspace herdr.WorkspaceInfo `json:"workspace"`
	}{Type: "workspace_info", Workspace: workspace})
}

// LinkedWorkspace returns a workspace of a linked worktree.
func LinkedWorkspace(id, path, repo string) herdr.WorkspaceInfo {
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

// RepositoryWorkspace returns a workspace of the primary checkout.
func RepositoryWorkspace(id, repo string) herdr.WorkspaceInfo {
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

// CreateReply returns a worktree.create result.
func CreateReply(t *testing.T, workspaceID, tabID, paneID, path, branch, repo string) string {
	t.Helper()
	return JSONText(t, herdr.WorktreeCreateResult{
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

// RunGit runs one Git command in the directory and returns its output.
func RunGit(t *testing.T, dir string, args ...string) string {
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

// NewTestRepo makes a Git repository with one commit on main.
func NewTestRepo(t *testing.T) string {
	t.Helper()
	repo := filepath.Join(t.TempDir(), "repo")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatalf("mkdir repo: %v", err)
	}
	RunGit(t, repo, "init", "-q", "-b", "main")
	RunGit(t, repo, "config", "user.name", "Hive Test")
	RunGit(t, repo, "config", "user.email", "test@example.com")
	if err := os.WriteFile(filepath.Join(repo, "README.md"), []byte("base\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	RunGit(t, repo, "add", "README.md")
	RunGit(t, repo, "commit", "-q", "-m", "init")
	return repo
}

// NewTestHome makes a fixture home with its marker and the project worktrees folder.
func NewTestHome(t *testing.T, projectID string) string {
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

// WriteProject writes one project record.
func WriteProject(t *testing.T, home string, project records.Project) {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	if err := registry.Write("projects", project.ID, project); err != nil {
		t.Fatalf("write project: %v", err)
	}
}

// WriteTask writes one task record.
func WriteTask(t *testing.T, home string, task records.Task) {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	if err := registry.WriteIn([]string{"tasks"}, task.ID, task); err != nil {
		t.Fatalf("write task: %v", err)
	}
}

// ReadTask reads one task record.
func ReadTask(t *testing.T, home, taskID string) records.Task {
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

// ReadProject reads one project record.
func ReadProject(t *testing.T, home, projectID string) records.Project {
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

// AddLinkedWorktree adds a linked worktree on a new branch.
func AddLinkedWorktree(t *testing.T, repo, path, branch, base string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir worktree parent: %v", err)
	}
	RunGit(t, repo, "worktree", "add", "-q", "-b", branch, path, base)
}
