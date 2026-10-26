package create

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Harrison-Blair/hive/internal/lib/herdr"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/worktree/worktreetest"
)

func effectCount(envelope output.Envelope, resource, action string) int {
	count := 0
	for _, effect := range envelope.Effects {
		if effect.Resource == resource && effect.Action == action {
			count++
		}
	}
	return count
}

// effectResources gives the sorted set of the effect resources of envelope.
func effectResources(envelope output.Envelope) []string {
	seen := map[string]bool{}
	resources := []string{}
	for _, effect := range envelope.Effects {
		if !seen[effect.Resource] {
			seen[effect.Resource] = true
			resources = append(resources, effect.Resource)
		}
	}
	sort.Strings(resources)
	return resources
}

func TestCreateRecordsWorktreeAndRepositoryWorkspace(t *testing.T) {
	repo := worktreetest.NewTestRepo(t)
	home := worktreetest.NewTestHome(t, "repo-live")
	worktreetest.WriteProject(t, home, worktreetest.Project(repo))
	worktreetest.WriteTask(t, home, worktreetest.Task())
	wantPath := filepath.Join(home, "projects", "repo-live", "worktrees", "1-live-demo")
	wantBranch := "task/1-live-demo"
	unrelated := herdrWorkspace("w9", "/tmp/unrelated")

	var listCalls atomic.Int32
	server := worktreetest.StartFakeHerdr(t, func(request worktreetest.FakeRequest) worktreetest.FakeReply {
		switch request.Method {
		case "workspace.list":
			if listCalls.Add(1) == 1 {
				return worktreetest.FakeReply{Result: worktreetest.WorkspaceListReply(t, unrelated)}
			}
			return worktreetest.FakeReply{Result: worktreetest.WorkspaceListReply(t,
				unrelated,
				worktreetest.LinkedWorkspace("w2", wantPath, repo),
				worktreetest.RepositoryWorkspace("w7", repo),
			)}
		case "worktree.create":
			return worktreetest.FakeReply{Result: worktreetest.CreateReply(t, "w2", "w2:t1", "w2:p1", wantPath, wantBranch, repo)}
		}
		return worktreetest.FakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
	})

	envelope := Run(context.Background(), home, Options{Task: "1-live-demo"})
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
		if !worktreetest.HasEffect(envelope, want.resource, want.action, want.path) {
			t.Fatalf("effects = %+v, want %s %s %s", envelope.Effects, want.action, want.resource, want.path)
		}
	}

	request, found := server.RequestFor("worktree.create")
	if !found {
		t.Fatalf("methods = %v, want worktree.create", server.Methods())
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
		if got := request.StringParam(t, check.field); got != check.want {
			t.Fatalf("worktree.create %s = %q, want %q", check.field, got, check.want)
		}
	}
	if request.BoolParam(t, "focus") {
		t.Fatalf("worktree.create focus = true, want false")
	}
	if request.HasParam("workspace_id") {
		t.Fatalf("worktree.create params = %s, want no workspace_id", request.Params)
	}
	if methods := server.Methods(); !reflect.DeepEqual(methods, []string{"workspace.list", "worktree.create", "workspace.list"}) {
		t.Fatalf("methods = %v, want [workspace.list worktree.create workspace.list]", methods)
	}

	task := worktreetest.ReadTask(t, home, "1-live-demo")
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
	project := worktreetest.ReadProject(t, home, "repo-live")
	if project.RepoWorkspaceID != "w7" {
		t.Fatalf("project repo_workspace_id = %q, want w7", project.RepoWorkspaceID)
	}
	result, ok := envelope.Result.(Result)
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
				task := worktreetest.Task()
				task.Worktree = &records.Worktree{Path: "/tmp/already", Branch: "task/old"}
				return task, worktreetest.Project(repo)
			},
			wantErr: "task has a worktree: /tmp/already",
		},
		{
			name: "repository is a linked worktree",
			setup: func(t *testing.T, home, repo string) (records.Task, records.Project) {
				linked := filepath.Join(t.TempDir(), "linked")
				worktreetest.AddLinkedWorktree(t, repo, linked, "task/other", "main")
				project := worktreetest.Project(repo)
				project.Path = linked
				return worktreetest.Task(), project
			},
			wantErr: "repository is a linked worktree: register the primary checkout",
		},
		{
			name: "missing base ref",
			setup: func(t *testing.T, home, repo string) (records.Task, records.Project) {
				project := worktreetest.Project(repo)
				project.Base = "trunk"
				return worktreetest.Task(), project
			},
			wantErr: "base ref not found: trunk",
		},
		{
			name: "existing branch",
			setup: func(t *testing.T, home, repo string) (records.Task, records.Project) {
				worktreetest.RunGit(t, repo, "branch", "task/1-live-demo")
				return worktreetest.Task(), worktreetest.Project(repo)
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
				return worktreetest.Task(), worktreetest.Project(repo)
			},
			wantErr: "path exists: ",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := worktreetest.NewTestRepo(t)
			home := worktreetest.NewTestHome(t, "repo-live")
			task, project := tc.setup(t, home, repo)
			worktreetest.WriteProject(t, home, project)
			worktreetest.WriteTask(t, home, task)
			server := worktreetest.StartFakeHerdr(t, func(request worktreetest.FakeRequest) worktreetest.FakeReply {
				return worktreetest.FakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
			})

			envelope := Run(context.Background(), home, Options{Task: task.ID})
			if envelope.Outcome != "failed" || envelope.Error == nil {
				t.Fatalf("Create() = %+v, want failed", envelope)
			}
			if got := fmt.Sprint(envelope.Error); !strings.Contains(got, tc.wantErr) {
				t.Fatalf("error = %q, want %q", got, tc.wantErr)
			}
			if methods := server.Methods(); len(methods) != 0 {
				t.Fatalf("methods = %v, want no request", methods)
			}
			if got := worktreetest.ReadTask(t, home, task.ID); !reflect.DeepEqual(got, task) {
				t.Fatalf("task changed to %+v, want %+v", got, task)
			}
			if got := worktreetest.ReadProject(t, home, project.ID); !reflect.DeepEqual(got, project) {
				t.Fatalf("project changed to %+v, want %+v", got, project)
			}
		})
	}
}

func TestCreatePathMismatchIsPartial(t *testing.T) {
	repo := worktreetest.NewTestRepo(t)
	home := worktreetest.NewTestHome(t, "repo-live")
	worktreetest.WriteProject(t, home, worktreetest.Project(repo))
	worktreetest.WriteTask(t, home, worktreetest.Task())
	wantPath := filepath.Join(home, "projects", "repo-live", "worktrees", "1-live-demo")
	actualPath := filepath.Join(t.TempDir(), "other")
	server := worktreetest.StartFakeHerdr(t, func(request worktreetest.FakeRequest) worktreetest.FakeReply {
		switch request.Method {
		case "workspace.list":
			return worktreetest.FakeReply{Result: worktreetest.WorkspaceListReply(t)}
		case "worktree.create":
			return worktreetest.FakeReply{Result: worktreetest.CreateReply(t, "w2", "w2:t1", "w2:p1", actualPath, "task/1-live-demo", repo)}
		}
		return worktreetest.FakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
	})

	envelope := Run(context.Background(), home, Options{Task: "1-live-demo"})
	if envelope.Outcome != "partial" || envelope.Error == nil {
		t.Fatalf("Create() = %+v, want partial", envelope)
	}
	if !strings.Contains(fmt.Sprint(envelope.Error), "path") {
		t.Fatalf("error = %q, want a path mismatch", fmt.Sprint(envelope.Error))
	}
	if !worktreetest.HasEffect(envelope, "worktree", "created", actualPath) {
		t.Fatalf("effects = %+v, want the actual worktree path", envelope.Effects)
	}
	if task := worktreetest.ReadTask(t, home, "1-live-demo"); task.Worktree != nil {
		t.Fatalf("task worktree = %+v, want nil", task.Worktree)
	}
	if project := worktreetest.ReadProject(t, home, "repo-live"); project.RepoWorkspaceID != "" {
		t.Fatalf("project repo_workspace_id = %q, want empty", project.RepoWorkspaceID)
	}
	if methods := server.Methods(); !reflect.DeepEqual(methods, []string{"workspace.list", "worktree.create"}) {
		t.Fatalf("methods = %v, want the precheck list and the create", methods)
	}
	_ = wantPath
}

func TestCreateFailedBranchStays(t *testing.T) {
	repo := worktreetest.NewTestRepo(t)
	home := worktreetest.NewTestHome(t, "repo-live")
	worktreetest.WriteProject(t, home, worktreetest.Project(repo))
	worktreetest.WriteTask(t, home, worktreetest.Task())
	wantBranch := "task/1-live-demo"
	server := worktreetest.StartFakeHerdr(t, func(request worktreetest.FakeRequest) worktreetest.FakeReply {
		switch request.Method {
		case "workspace.list":
			return worktreetest.FakeReply{Result: worktreetest.WorkspaceListReply(t)}
		case "worktree.create":
			worktreetest.RunGit(t, repo, "branch", wantBranch)
			return worktreetest.FakeReply{Error: `{"code":"worktree_create_failed","message":"fatal: test failure"}`}
		}
		return worktreetest.FakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
	})

	envelope := Run(context.Background(), home, Options{Task: "1-live-demo"})
	if envelope.Outcome != "failed" || envelope.Phase != "request" {
		t.Fatalf("Create() = %+v, want failed request", envelope)
	}
	message := fmt.Sprint(envelope.Error)
	if !strings.Contains(message, "fatal: test failure") || !strings.Contains(message, "branch "+wantBranch+" stays") {
		t.Fatalf("error = %q, want the herdr message and the branch note", message)
	}
	if !worktreetest.HasEffect(envelope, "branch", "created", wantBranch) {
		t.Fatalf("effects = %+v, want branch created", envelope.Effects)
	}
	if task := worktreetest.ReadTask(t, home, "1-live-demo"); task.Worktree != nil {
		t.Fatalf("task worktree = %+v, want nil", task.Worktree)
	}
	_ = server
}

func TestCreateOperationInProgress(t *testing.T) {
	repo := worktreetest.NewTestRepo(t)
	home := worktreetest.NewTestHome(t, "repo-live")
	worktreetest.WriteProject(t, home, worktreetest.Project(repo))
	worktreetest.WriteTask(t, home, worktreetest.Task())
	server := worktreetest.StartFakeHerdr(t, func(request worktreetest.FakeRequest) worktreetest.FakeReply {
		switch request.Method {
		case "workspace.list":
			return worktreetest.FakeReply{Result: worktreetest.WorkspaceListReply(t)}
		case "worktree.create":
			return worktreetest.FakeReply{Error: `{"code":"worktree_operation_in_progress","message":"worktree operation is already in progress for this checkout"}`}
		}
		return worktreetest.FakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
	})

	envelope := Run(context.Background(), home, Options{Task: "1-live-demo"})
	if envelope.Outcome != "failed" || fmt.Sprint(envelope.Error) != "worktree operation in progress: retry later" {
		t.Fatalf("Create() = %+v, want the retry message", envelope)
	}
	if task := worktreetest.ReadTask(t, home, "1-live-demo"); task.Worktree != nil {
		t.Fatalf("task worktree = %+v, want nil", task.Worktree)
	}
	if project := worktreetest.ReadProject(t, home, "repo-live"); project.RepoWorkspaceID != "" {
		t.Fatalf("project repo_workspace_id = %q, want empty", project.RepoWorkspaceID)
	}
	_ = server
}

func TestCreateClosedConnectionIsPartial(t *testing.T) {
	repo := worktreetest.NewTestRepo(t)
	home := worktreetest.NewTestHome(t, "repo-live")
	worktreetest.WriteProject(t, home, worktreetest.Project(repo))
	worktreetest.WriteTask(t, home, worktreetest.Task())
	wantPath := filepath.Join(home, "projects", "repo-live", "worktrees", "1-live-demo")
	wantBranch := "task/1-live-demo"
	var listCalls atomic.Int32
	server := worktreetest.StartFakeHerdr(t, func(request worktreetest.FakeRequest) worktreetest.FakeReply {
		switch request.Method {
		case "workspace.list":
			if listCalls.Add(1) == 1 {
				return worktreetest.FakeReply{Result: worktreetest.WorkspaceListReply(t)}
			}
			return worktreetest.FakeReply{Result: worktreetest.WorkspaceListReply(t,
				worktreetest.LinkedWorkspace("w2", wantPath, repo),
				worktreetest.RepositoryWorkspace("w7", repo),
			)}
		case "worktree.create":
			worktreetest.RunGit(t, repo, "worktree", "add", "-q", "-b", wantBranch, wantPath, "main")
			return worktreetest.FakeReply{Close: true}
		}
		return worktreetest.FakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
	})

	envelope := Run(context.Background(), home, Options{Task: "1-live-demo"})
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
		if !worktreetest.HasEffect(envelope, want.resource, want.action, want.path) {
			t.Fatalf("effects = %+v, want %s %s %s", envelope.Effects, want.action, want.resource, want.path)
		}
	}
	task := worktreetest.ReadTask(t, home, "1-live-demo")
	if task.Worktree == nil {
		t.Fatalf("task worktree = nil, want the observed fields")
	}
	if task.Worktree.WorkspaceID != "w2" || task.Worktree.TabID != "w2:t1" || task.Worktree.RootPane != "" {
		t.Fatalf("task worktree topology = %+v, want w2/w2:t1 and no root pane", task.Worktree)
	}
	if task.Worktree.Path != wantPath || task.Worktree.Branch != wantBranch || task.Worktree.CreatedAt == "" {
		t.Fatalf("task worktree = %+v, want path, branch, and created_at", task.Worktree)
	}
	if project := worktreetest.ReadProject(t, home, "repo-live"); project.RepoWorkspaceID != "w7" {
		t.Fatalf("project repo_workspace_id = %q, want w7", project.RepoWorkspaceID)
	}
	_ = server
}

func TestCreatePreservesConcurrentTaskChange(t *testing.T) {
	repo := worktreetest.NewTestRepo(t)
	home := worktreetest.NewTestHome(t, "repo-live")
	worktreetest.WriteProject(t, home, worktreetest.Project(repo))
	worktreetest.WriteTask(t, home, worktreetest.Task())
	wantPath := filepath.Join(home, "projects", "repo-live", "worktrees", "1-live-demo")
	wantBranch := "task/1-live-demo"
	arrived := make(chan struct{})
	release := make(chan struct{})
	var listCalls atomic.Int32
	worktreetest.StartFakeHerdr(t, func(request worktreetest.FakeRequest) worktreetest.FakeReply {
		switch request.Method {
		case "workspace.list":
			if listCalls.Add(1) == 1 {
				return worktreetest.FakeReply{Result: worktreetest.WorkspaceListReply(t)}
			}
			return worktreetest.FakeReply{Result: worktreetest.WorkspaceListReply(t, worktreetest.LinkedWorkspace("w2", wantPath, repo))}
		case "worktree.create":
			close(arrived)
			return worktreetest.FakeReply{Result: worktreetest.CreateReply(t, "w2", "w2:t1", "w2:p1", wantPath, wantBranch, repo), Hold: release}
		}
		return worktreetest.FakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
	})

	done := make(chan output.Envelope, 1)
	go func() { done <- Run(context.Background(), home, Options{Task: "1-live-demo"}) }()
	select {
	case <-arrived:
	case <-time.After(2 * time.Second):
		t.Fatal("create sent no worktree.create request")
	}
	changed := worktreetest.Task()
	changed.Title = "changed title"
	changed.Status = "cancelled"
	changed.ActingAgent = "worker-a"
	changed.CancelReason = "user stopped it"
	changed.DependsOn = []string{"9-other"}
	changed.Verification = &records.Verification{Agent: "verifier-a", Result: "pass", Reason: "checked", At: "2026-10-18T01:00:00Z"}
	worktreetest.WriteTask(t, home, changed)
	close(release)
	envelope := <-done
	if envelope.Outcome != "ok" {
		t.Fatalf("Create() = %+v, want ok", envelope)
	}
	task := worktreetest.ReadTask(t, home, "1-live-demo")
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

func TestCreateCopiesFeatureBriefIntoUntrackedWorktree(t *testing.T) {
	repo := worktreetest.NewTestRepo(t)
	if err := os.WriteFile(filepath.Join(repo, ".gitignore"), []byte("*\n!.gitignore\n.hive/*\n!.hive/README.md\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	worktreetest.RunGit(t, repo, "add", "-f", ".gitignore")
	worktreetest.RunGit(t, repo, "commit", "-q", "-m", "ignore the feature spec")
	home := worktreetest.NewTestHome(t, "repo-live")
	task := worktreetest.Task()
	task.Feature = "f"
	task.Spec = true
	worktreetest.WriteProject(t, home, worktreetest.Project(repo))
	worktreetest.WriteTask(t, home, task)
	stateBrief := filepath.Join(home, ".hive", "state", "briefs", task.ID+".md")
	if err := os.MkdirAll(filepath.Dir(stateBrief), 0o755); err != nil {
		t.Fatal(err)
	}
	const brief = "# Feature task\n"
	if err := os.WriteFile(stateBrief, []byte(brief), 0o644); err != nil {
		t.Fatal(err)
	}
	worktreePath := filepath.Join(home, "projects", task.Project, "worktrees", task.ID)
	branch := "task/" + task.ID
	var listCalls atomic.Int32
	worktreetest.StartFakeHerdr(t, func(request worktreetest.FakeRequest) worktreetest.FakeReply {
		switch request.Method {
		case "workspace.list":
			if listCalls.Add(1) == 1 {
				return worktreetest.FakeReply{Result: worktreetest.WorkspaceListReply(t)}
			}
			return worktreetest.FakeReply{Result: worktreetest.WorkspaceListReply(t, worktreetest.LinkedWorkspace("w2", worktreePath, repo))}
		case "worktree.create":
			worktreetest.RunGit(t, repo, "worktree", "add", "-q", "-b", branch, worktreePath, "main")
			return worktreetest.FakeReply{Result: worktreetest.CreateReply(t, "w2", "w2:t1", "w2:p1", worktreePath, branch, repo)}
		}
		return worktreetest.FakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
	})

	envelope := Run(context.Background(), home, Options{Task: task.ID})
	if envelope.Outcome != "ok" {
		t.Fatalf("Create() = %+v, want ok", envelope)
	}
	featureBrief := filepath.Join(worktreePath, ".hive", "spec", "f", "tasks", task.ID, "brief.md")
	if got, err := os.ReadFile(featureBrief); err != nil || string(got) != brief {
		t.Fatalf("worktree brief = %q, %v; want %q", got, err, brief)
	}
	if !worktreetest.HasEffect(envelope, "brief", "copied", featureBrief) {
		t.Fatalf("effects = %+v, want copied brief at %s", envelope.Effects, featureBrief)
	}
	// AC7: the effects have no allowlist step.
	if got, want := effectResources(envelope), []string{"branch", "brief", "spec", "spec_copy", "workspace", "worktree"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("effect resources = %v, want %v", got, want)
	}
	// AC1: the repository ignores the copied brief, so the worktree is clean.
	if status := worktreetest.RunGit(t, worktreePath, "status", "--short", "--untracked-files=all"); status != "" {
		t.Fatalf("git status = %q, want a clean worktree", status)
	}
	worktreetest.RunGit(t, worktreePath, "check-ignore", "-q", ".hive/spec/f/tasks/1-live-demo/brief.md")
}

func TestCreateWithoutFeatureSkipsFeatureIgnoreCheckAndCopy(t *testing.T) {
	repo := worktreetest.NewTestRepo(t)
	if err := os.WriteFile(filepath.Join(repo, ".gitignore"), []byte("*\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	worktreetest.RunGit(t, repo, "add", "-f", ".gitignore")
	worktreetest.RunGit(t, repo, "commit", "-q", "-m", "ignore files")
	home := worktreetest.NewTestHome(t, "repo-live")
	task := worktreetest.Task()
	worktreetest.WriteProject(t, home, worktreetest.Project(repo))
	worktreetest.WriteTask(t, home, task)
	worktreePath := filepath.Join(home, "projects", task.Project, "worktrees", task.ID)
	branch := "task/" + task.ID
	var listCalls atomic.Int32
	server := worktreetest.StartFakeHerdr(t, func(request worktreetest.FakeRequest) worktreetest.FakeReply {
		switch request.Method {
		case "workspace.list":
			if listCalls.Add(1) == 1 {
				return worktreetest.FakeReply{Result: worktreetest.WorkspaceListReply(t)}
			}
			return worktreetest.FakeReply{Result: worktreetest.WorkspaceListReply(t, worktreetest.LinkedWorkspace("w2", worktreePath, repo))}
		case "worktree.create":
			worktreetest.RunGit(t, repo, "worktree", "add", "-q", "-b", branch, worktreePath, "main")
			return worktreetest.FakeReply{Result: worktreetest.CreateReply(t, "w2", "w2:t1", "w2:p1", worktreePath, branch, repo)}
		}
		return worktreetest.FakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
	})

	envelope := Run(context.Background(), home, Options{Task: task.ID})
	if envelope.Outcome != "ok" {
		t.Fatalf("Create() = %+v, want ok", envelope)
	}
	if len(server.Methods()) != 3 {
		t.Fatalf("methods = %v, want the workspace lists and worktree create", server.Methods())
	}
	featureBrief := filepath.Join(worktreePath, ".hive", "spec", "f", "tasks", task.ID, "brief.md")
	if _, err := os.Lstat(featureBrief); !os.IsNotExist(err) {
		t.Fatalf("feature brief status = %v, want no file", err)
	}
}

func TestCreateAmbiguousRepositoryWorkspaceIsPartial(t *testing.T) {
	repo := worktreetest.NewTestRepo(t)
	home := worktreetest.NewTestHome(t, "repo-live")
	worktreetest.WriteProject(t, home, worktreetest.Project(repo))
	worktreetest.WriteTask(t, home, worktreetest.Task())
	wantPath := filepath.Join(home, "projects", "repo-live", "worktrees", "1-live-demo")
	wantBranch := "task/1-live-demo"
	var listCalls atomic.Int32
	worktreetest.StartFakeHerdr(t, func(request worktreetest.FakeRequest) worktreetest.FakeReply {
		switch request.Method {
		case "workspace.list":
			if listCalls.Add(1) == 1 {
				return worktreetest.FakeReply{Result: worktreetest.WorkspaceListReply(t)}
			}
			return worktreetest.FakeReply{Result: worktreetest.WorkspaceListReply(t,
				worktreetest.LinkedWorkspace("w2", wantPath, repo),
				worktreetest.RepositoryWorkspace("w7", repo),
				worktreetest.RepositoryWorkspace("w8", repo),
			)}
		case "worktree.create":
			return worktreetest.FakeReply{Result: worktreetest.CreateReply(t, "w2", "w2:t1", "w2:p1", wantPath, wantBranch, repo)}
		}
		return worktreetest.FakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
	})

	envelope := Run(context.Background(), home, Options{Task: "1-live-demo"})
	if envelope.Outcome != "partial" || envelope.Phase != "workspace" {
		t.Fatalf("Create() = %+v, want partial workspace", envelope)
	}
	if effectCount(envelope, "workspace", "opened") != 0 {
		t.Fatalf("effects = %+v, want no repository workspace claim", envelope.Effects)
	}
	if !worktreetest.HasEffect(envelope, "workspace", "created", "w2") || !worktreetest.HasEffect(envelope, "worktree", "created", wantPath) {
		t.Fatalf("effects = %+v, want only the known effects", envelope.Effects)
	}
	if project := worktreetest.ReadProject(t, home, "repo-live"); project.RepoWorkspaceID != "" {
		t.Fatalf("project repo_workspace_id = %q, want empty", project.RepoWorkspaceID)
	}
	if task := worktreetest.ReadTask(t, home, "1-live-demo"); task.Worktree == nil || task.Worktree.WorkspaceID != "w2" {
		t.Fatalf("task worktree = %+v, want the returned workspace", task.Worktree)
	}
}
