package worktree

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Harrison-Blair/hive/internal/lib/herdr"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
)

func testProject(repo string) records.Project {
	return records.Project{ID: "repo-live", Path: repo, Base: "main", RegisteredAt: "2026-10-18T00:00:00Z"}
}

func testTask() records.Task {
	return records.Task{
		ID:              "1-live-demo",
		Project:         "repo-live",
		Title:           "live demo",
		ControllingRole: "manager-repo-live",
		Status:          "created",
		DependsOn:       []string{},
		CreatedAt:       "2026-10-18T00:00:00Z",
		UpdatedAt:       "2026-10-18T00:00:00Z",
	}
}

func hasEffect(envelope output.Envelope, resource, action, path string) bool {
	for _, effect := range envelope.Effects {
		if effect.Resource == resource && effect.Action == action && effect.Path == path {
			return true
		}
	}
	return false
}

func effectCount(envelope output.Envelope, resource, action string) int {
	count := 0
	for _, effect := range envelope.Effects {
		if effect.Resource == resource && effect.Action == action {
			count++
		}
	}
	return count
}

func TestCreateRecordsWorktreeAndRepositoryWorkspace(t *testing.T) {
	repo := newTestRepo(t)
	home := newTestHome(t, "repo-live")
	writeProject(t, home, testProject(repo))
	writeTask(t, home, testTask())
	wantPath := filepath.Join(home, "projects", "repo-live", "worktrees", "1-live-demo")
	wantBranch := "task/1-live-demo"
	unrelated := herdrWorkspace("w9", "/tmp/unrelated")

	var listCalls atomic.Int32
	server := startFakeHerdr(t, func(request fakeRequest) fakeReply {
		switch request.Method {
		case "workspace.list":
			if listCalls.Add(1) == 1 {
				return fakeReply{Result: workspaceListReply(t, unrelated)}
			}
			return fakeReply{Result: workspaceListReply(t,
				unrelated,
				linkedWorkspace("w2", wantPath, repo),
				repositoryWorkspace("w7", repo),
			)}
		case "worktree.create":
			return fakeReply{Result: createReply(t, "w2", "w2:t1", "w2:p1", wantPath, wantBranch, repo)}
		}
		return fakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
	})

	envelope := Create(context.Background(), home, CreateOptions{Task: "1-live-demo"})
	if envelope.Outcome != "ok" || envelope.Phase != "complete" || envelope.Error != nil {
		t.Fatalf("Create() = %+v, want ok complete", envelope)
	}
	for _, want := range []struct {
		resource string
		action   string
		path     string
	}{
		{"workspace", "created", "w2"},
		{"worktree", "created", wantPath},
		{"branch", "created", wantBranch},
		{"workspace", "opened", "w7"},
	} {
		if !hasEffect(envelope, want.resource, want.action, want.path) {
			t.Fatalf("effects = %+v, want %s %s %s", envelope.Effects, want.action, want.resource, want.path)
		}
	}

	request, found := server.requestFor("worktree.create")
	if !found {
		t.Fatalf("methods = %v, want worktree.create", server.methods())
	}
	checks := []struct {
		field string
		want  string
	}{
		{"cwd", repo},
		{"branch", wantBranch},
		{"base", "main"},
		{"path", wantPath},
		{"label", "1-live-demo"},
	}
	for _, check := range checks {
		if got := request.stringParam(t, check.field); got != check.want {
			t.Fatalf("worktree.create %s = %q, want %q", check.field, got, check.want)
		}
	}
	if request.boolParam(t, "focus") {
		t.Fatalf("worktree.create focus = true, want false")
	}
	if request.hasParam("workspace_id") {
		t.Fatalf("worktree.create params = %s, want no workspace_id", request.Params)
	}
	if methods := server.methods(); !reflect.DeepEqual(methods, []string{"workspace.list", "worktree.create", "workspace.list"}) {
		t.Fatalf("methods = %v, want [workspace.list worktree.create workspace.list]", methods)
	}

	task := readTask(t, home, "1-live-demo")
	if task.Worktree == nil {
		t.Fatalf("task worktree = nil, want a record")
	}
	if task.Worktree.WorkspaceID != "w2" || task.Worktree.TabID != "w2:t1" || task.Worktree.RootPane != "w2:p1" {
		t.Fatalf("task worktree topology = %+v, want w2/w2:t1/w2:p1", task.Worktree)
	}
	if task.Worktree.Path != wantPath || task.Worktree.Branch != wantBranch {
		t.Fatalf("task worktree path/branch = %q/%q, want %q/%q", task.Worktree.Path, task.Worktree.Branch, wantPath, wantBranch)
	}
	if task.Worktree.CreatedAt == "" || task.Worktree.RemovedAt != "" {
		t.Fatalf("task worktree times = %+v, want created_at set and removed_at empty", task.Worktree)
	}
	project := readProject(t, home, "repo-live")
	if project.RepoWorkspaceID != "w7" {
		t.Fatalf("project repo_workspace_id = %q, want w7", project.RepoWorkspaceID)
	}
	result, ok := envelope.Result.(CreateResult)
	if !ok {
		t.Fatalf("result = %T, want CreateResult", envelope.Result)
	}
	if result.WorkspaceID != "w2" || result.RootPane != "w2:p1" || result.RepoWorkspaceID != "w7" {
		t.Fatalf("result = %+v, want w2/w2:p1/w7", result)
	}
}

func herdrWorkspace(id, label string) herdr.WorkspaceInfo {
	return herdr.WorkspaceInfo{WorkspaceID: id, Label: label, ActiveTabID: id + ":t1", AgentStatus: "unknown"}
}

func TestCreateRefusalsSendNoRequestAndChangeNoRecord(t *testing.T) {
	cases := []struct {
		name    string
		setup   func(t *testing.T, home, repo string) (records.Task, records.Project)
		wantErr string
	}{
		{
			name: "task has a worktree",
			setup: func(t *testing.T, home, repo string) (records.Task, records.Project) {
				task := testTask()
				task.Worktree = &records.Worktree{Path: "/tmp/already", Branch: "task/old"}
				return task, testProject(repo)
			},
			wantErr: "task has a worktree: /tmp/already",
		},
		{
			name: "repository is a linked worktree",
			setup: func(t *testing.T, home, repo string) (records.Task, records.Project) {
				linked := filepath.Join(t.TempDir(), "linked")
				addLinkedWorktree(t, repo, linked, "task/other", "main")
				project := testProject(repo)
				project.Path = linked
				return testTask(), project
			},
			wantErr: "repository is a linked worktree: register the primary checkout",
		},
		{
			name: "missing base ref",
			setup: func(t *testing.T, home, repo string) (records.Task, records.Project) {
				project := testProject(repo)
				project.Base = "trunk"
				return testTask(), project
			},
			wantErr: "base ref not found: trunk",
		},
		{
			name: "existing branch",
			setup: func(t *testing.T, home, repo string) (records.Task, records.Project) {
				runGit(t, repo, "branch", "task/1-live-demo")
				return testTask(), testProject(repo)
			},
			wantErr: "branch exists: task/1-live-demo",
		},
		{
			name: "existing path",
			setup: func(t *testing.T, home, repo string) (records.Task, records.Project) {
				path := filepath.Join(home, "projects", "repo-live", "worktrees", "1-live-demo")
				if err := os.MkdirAll(path, 0o755); err != nil {
					t.Fatalf("mkdir path: %v", err)
				}
				return testTask(), testProject(repo)
			},
			wantErr: "path exists: ",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := newTestRepo(t)
			home := newTestHome(t, "repo-live")
			task, project := tc.setup(t, home, repo)
			writeProject(t, home, project)
			writeTask(t, home, task)
			server := startFakeHerdr(t, func(request fakeRequest) fakeReply {
				return fakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
			})

			envelope := Create(context.Background(), home, CreateOptions{Task: task.ID})
			if envelope.Outcome != "failed" || envelope.Error == nil {
				t.Fatalf("Create() = %+v, want failed", envelope)
			}
			if got := fmt.Sprint(envelope.Error); !strings.Contains(got, tc.wantErr) {
				t.Fatalf("error = %q, want %q", got, tc.wantErr)
			}
			if methods := server.methods(); len(methods) != 0 {
				t.Fatalf("methods = %v, want no request", methods)
			}
			if got := readTask(t, home, task.ID); !reflect.DeepEqual(got, task) {
				t.Fatalf("task changed to %+v, want %+v", got, task)
			}
			if got := readProject(t, home, project.ID); !reflect.DeepEqual(got, project) {
				t.Fatalf("project changed to %+v, want %+v", got, project)
			}
		})
	}
}

func TestCreatePathMismatchIsPartial(t *testing.T) {
	repo := newTestRepo(t)
	home := newTestHome(t, "repo-live")
	writeProject(t, home, testProject(repo))
	writeTask(t, home, testTask())
	wantPath := filepath.Join(home, "projects", "repo-live", "worktrees", "1-live-demo")
	actualPath := filepath.Join(t.TempDir(), "other")
	server := startFakeHerdr(t, func(request fakeRequest) fakeReply {
		switch request.Method {
		case "workspace.list":
			return fakeReply{Result: workspaceListReply(t)}
		case "worktree.create":
			return fakeReply{Result: createReply(t, "w2", "w2:t1", "w2:p1", actualPath, "task/1-live-demo", repo)}
		}
		return fakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
	})

	envelope := Create(context.Background(), home, CreateOptions{Task: "1-live-demo"})
	if envelope.Outcome != "partial" || envelope.Error == nil {
		t.Fatalf("Create() = %+v, want partial", envelope)
	}
	if !strings.Contains(fmt.Sprint(envelope.Error), "path") {
		t.Fatalf("error = %q, want a path mismatch", fmt.Sprint(envelope.Error))
	}
	if !hasEffect(envelope, "worktree", "created", actualPath) {
		t.Fatalf("effects = %+v, want the actual worktree path", envelope.Effects)
	}
	if task := readTask(t, home, "1-live-demo"); task.Worktree != nil {
		t.Fatalf("task worktree = %+v, want nil", task.Worktree)
	}
	if project := readProject(t, home, "repo-live"); project.RepoWorkspaceID != "" {
		t.Fatalf("project repo_workspace_id = %q, want empty", project.RepoWorkspaceID)
	}
	if methods := server.methods(); !reflect.DeepEqual(methods, []string{"workspace.list", "worktree.create"}) {
		t.Fatalf("methods = %v, want the precheck list and the create", methods)
	}
	_ = wantPath
}

func TestCreateFailedBranchStays(t *testing.T) {
	repo := newTestRepo(t)
	home := newTestHome(t, "repo-live")
	writeProject(t, home, testProject(repo))
	writeTask(t, home, testTask())
	wantBranch := "task/1-live-demo"
	server := startFakeHerdr(t, func(request fakeRequest) fakeReply {
		switch request.Method {
		case "workspace.list":
			return fakeReply{Result: workspaceListReply(t)}
		case "worktree.create":
			runGit(t, repo, "branch", wantBranch)
			return fakeReply{Error: `{"code":"worktree_create_failed","message":"fatal: test failure"}`}
		}
		return fakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
	})

	envelope := Create(context.Background(), home, CreateOptions{Task: "1-live-demo"})
	if envelope.Outcome != "failed" || envelope.Phase != "request" {
		t.Fatalf("Create() = %+v, want failed request", envelope)
	}
	message := fmt.Sprint(envelope.Error)
	if !strings.Contains(message, "fatal: test failure") || !strings.Contains(message, "branch "+wantBranch+" stays") {
		t.Fatalf("error = %q, want the herdr message and the branch note", message)
	}
	if !hasEffect(envelope, "branch", "created", wantBranch) {
		t.Fatalf("effects = %+v, want branch created", envelope.Effects)
	}
	if task := readTask(t, home, "1-live-demo"); task.Worktree != nil {
		t.Fatalf("task worktree = %+v, want nil", task.Worktree)
	}
	_ = server
}

func TestCreateOperationInProgress(t *testing.T) {
	repo := newTestRepo(t)
	home := newTestHome(t, "repo-live")
	writeProject(t, home, testProject(repo))
	writeTask(t, home, testTask())
	server := startFakeHerdr(t, func(request fakeRequest) fakeReply {
		switch request.Method {
		case "workspace.list":
			return fakeReply{Result: workspaceListReply(t)}
		case "worktree.create":
			return fakeReply{Error: `{"code":"worktree_operation_in_progress","message":"worktree operation is already in progress for this checkout"}`}
		}
		return fakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
	})

	envelope := Create(context.Background(), home, CreateOptions{Task: "1-live-demo"})
	if envelope.Outcome != "failed" || fmt.Sprint(envelope.Error) != "worktree operation in progress: retry later" {
		t.Fatalf("Create() = %+v, want the retry message", envelope)
	}
	if task := readTask(t, home, "1-live-demo"); task.Worktree != nil {
		t.Fatalf("task worktree = %+v, want nil", task.Worktree)
	}
	if project := readProject(t, home, "repo-live"); project.RepoWorkspaceID != "" {
		t.Fatalf("project repo_workspace_id = %q, want empty", project.RepoWorkspaceID)
	}
	_ = server
}

func TestCreateClosedConnectionIsPartial(t *testing.T) {
	repo := newTestRepo(t)
	home := newTestHome(t, "repo-live")
	writeProject(t, home, testProject(repo))
	writeTask(t, home, testTask())
	wantPath := filepath.Join(home, "projects", "repo-live", "worktrees", "1-live-demo")
	wantBranch := "task/1-live-demo"
	var listCalls atomic.Int32
	server := startFakeHerdr(t, func(request fakeRequest) fakeReply {
		switch request.Method {
		case "workspace.list":
			if listCalls.Add(1) == 1 {
				return fakeReply{Result: workspaceListReply(t)}
			}
			return fakeReply{Result: workspaceListReply(t,
				linkedWorkspace("w2", wantPath, repo),
				repositoryWorkspace("w7", repo),
			)}
		case "worktree.create":
			runGit(t, repo, "worktree", "add", "-q", "-b", wantBranch, wantPath, "main")
			return fakeReply{Close: true}
		}
		return fakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
	})

	envelope := Create(context.Background(), home, CreateOptions{Task: "1-live-demo"})
	if envelope.Outcome != "partial" || envelope.Phase != "request" {
		t.Fatalf("Create() = %+v, want partial request", envelope)
	}
	for _, want := range []struct {
		resource string
		action   string
		path     string
	}{
		{"workspace", "created", "w2"},
		{"worktree", "created", wantPath},
		{"branch", "created", wantBranch},
		{"workspace", "opened", "w7"},
	} {
		if !hasEffect(envelope, want.resource, want.action, want.path) {
			t.Fatalf("effects = %+v, want %s %s %s", envelope.Effects, want.action, want.resource, want.path)
		}
	}
	task := readTask(t, home, "1-live-demo")
	if task.Worktree == nil {
		t.Fatalf("task worktree = nil, want the observed fields")
	}
	if task.Worktree.WorkspaceID != "w2" || task.Worktree.TabID != "w2:t1" || task.Worktree.RootPane != "" {
		t.Fatalf("task worktree topology = %+v, want w2/w2:t1 and no root pane", task.Worktree)
	}
	if task.Worktree.Path != wantPath || task.Worktree.Branch != wantBranch || task.Worktree.CreatedAt == "" {
		t.Fatalf("task worktree = %+v, want path, branch, and created_at", task.Worktree)
	}
	if project := readProject(t, home, "repo-live"); project.RepoWorkspaceID != "w7" {
		t.Fatalf("project repo_workspace_id = %q, want w7", project.RepoWorkspaceID)
	}
	_ = server
}

func TestCreatePreservesConcurrentTaskChange(t *testing.T) {
	repo := newTestRepo(t)
	home := newTestHome(t, "repo-live")
	writeProject(t, home, testProject(repo))
	writeTask(t, home, testTask())
	wantPath := filepath.Join(home, "projects", "repo-live", "worktrees", "1-live-demo")
	wantBranch := "task/1-live-demo"
	arrived := make(chan struct{})
	release := make(chan struct{})
	var listCalls atomic.Int32
	startFakeHerdr(t, func(request fakeRequest) fakeReply {
		switch request.Method {
		case "workspace.list":
			if listCalls.Add(1) == 1 {
				return fakeReply{Result: workspaceListReply(t)}
			}
			return fakeReply{Result: workspaceListReply(t, linkedWorkspace("w2", wantPath, repo))}
		case "worktree.create":
			close(arrived)
			return fakeReply{Result: createReply(t, "w2", "w2:t1", "w2:p1", wantPath, wantBranch, repo), Hold: release}
		}
		return fakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
	})

	done := make(chan output.Envelope, 1)
	go func() { done <- Create(context.Background(), home, CreateOptions{Task: "1-live-demo"}) }()
	select {
	case <-arrived:
	case <-time.After(2 * time.Second):
		t.Fatal("create sent no worktree.create request")
	}
	changed := testTask()
	changed.Title = "changed title"
	changed.Status = "cancelled"
	changed.ActingAgent = "worker-a"
	changed.CancelReason = "user stopped it"
	changed.DependsOn = []string{"9-other"}
	changed.Verification = &records.Verification{Agent: "verifier-a", Result: "pass", Reason: "checked", At: "2026-10-18T01:00:00Z"}
	writeTask(t, home, changed)
	close(release)
	envelope := <-done
	if envelope.Outcome != "ok" {
		t.Fatalf("Create() = %+v, want ok", envelope)
	}
	task := readTask(t, home, "1-live-demo")
	if task.Title != "changed title" || task.Status != "cancelled" || task.ActingAgent != "worker-a" || task.CancelReason != "user stopped it" {
		t.Fatalf("task = %+v, want the concurrent fields", task)
	}
	if !reflect.DeepEqual(task.DependsOn, []string{"9-other"}) || task.Verification == nil || task.Verification.Agent != "verifier-a" {
		t.Fatalf("task = %+v, want the concurrent dependencies and verification", task)
	}
	if task.Worktree == nil || task.Worktree.WorkspaceID != "w2" || task.Worktree.Path != wantPath {
		t.Fatalf("task worktree = %+v, want the new worktree", task.Worktree)
	}
}

func TestCreateAmbiguousRepositoryWorkspaceIsPartial(t *testing.T) {
	repo := newTestRepo(t)
	home := newTestHome(t, "repo-live")
	writeProject(t, home, testProject(repo))
	writeTask(t, home, testTask())
	wantPath := filepath.Join(home, "projects", "repo-live", "worktrees", "1-live-demo")
	wantBranch := "task/1-live-demo"
	var listCalls atomic.Int32
	startFakeHerdr(t, func(request fakeRequest) fakeReply {
		switch request.Method {
		case "workspace.list":
			if listCalls.Add(1) == 1 {
				return fakeReply{Result: workspaceListReply(t)}
			}
			return fakeReply{Result: workspaceListReply(t,
				linkedWorkspace("w2", wantPath, repo),
				repositoryWorkspace("w7", repo),
				repositoryWorkspace("w8", repo),
			)}
		case "worktree.create":
			return fakeReply{Result: createReply(t, "w2", "w2:t1", "w2:p1", wantPath, wantBranch, repo)}
		}
		return fakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
	})

	envelope := Create(context.Background(), home, CreateOptions{Task: "1-live-demo"})
	if envelope.Outcome != "partial" || envelope.Phase != "workspace" {
		t.Fatalf("Create() = %+v, want partial workspace", envelope)
	}
	if effectCount(envelope, "workspace", "opened") != 0 {
		t.Fatalf("effects = %+v, want no repository workspace claim", envelope.Effects)
	}
	if !hasEffect(envelope, "workspace", "created", "w2") || !hasEffect(envelope, "worktree", "created", wantPath) {
		t.Fatalf("effects = %+v, want only the known effects", envelope.Effects)
	}
	if project := readProject(t, home, "repo-live"); project.RepoWorkspaceID != "" {
		t.Fatalf("project repo_workspace_id = %q, want empty", project.RepoWorkspaceID)
	}
	if task := readTask(t, home, "1-live-demo"); task.Worktree == nil || task.Worktree.WorkspaceID != "w2" {
		t.Fatalf("task worktree = %+v, want the returned workspace", task.Worktree)
	}
}
