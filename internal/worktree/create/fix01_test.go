package create

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
	"github.com/Harrison-Blair/hive/internal/worktree/worktreetest"
)

func TestFixCreateTaskWriteFaultRetainsKnownRepositoryEffect(t *testing.T) {
	repo := worktreetest.NewTestRepo(t)
	home := worktreetest.NewTestHome(t, "repo-live")
	worktreetest.WriteProject(t, home, worktreetest.Project(repo))
	worktreetest.WriteTask(t, home, worktreetest.Task())
	path := worktreetest.PermittedPath(home, "repo-live", "1-live-demo")
	branch := "task/1-live-demo"
	tasksDir := filepath.Join(home, ".hive", "state", "tasks")
	faulted := false
	t.Cleanup(func() { _ = os.Chmod(tasksDir, 0o755) })
	worktreetest.StartFakeHerdr(t, func(request worktreetest.FakeRequest) worktreetest.FakeReply {
		switch request.Method {
		case "workspace.list":
			if !faulted {
				return worktreetest.FakeReply{Result: worktreetest.WorkspaceListReply(t)}
			}
			return worktreetest.FakeReply{Result: worktreetest.WorkspaceListReply(t, worktreetest.LinkedWorkspace("w2", path, repo), worktreetest.RepositoryWorkspace("w7", repo))}
		case "worktree.create":
			worktreetest.RunGit(t, repo, "worktree", "add", "-q", "-b", branch, path, "main")
			if err := os.Chmod(tasksDir, 0o500); err != nil {
				t.Fatalf("chmod tasks: %v", err)
			}
			faulted = true
			return worktreetest.FakeReply{Result: worktreetest.CreateReply(t, "w2", "w2:t1", "w2:p1", path, branch, repo)}
		}
		return worktreetest.FakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
	})

	envelope := Run(context.Background(), home, Options{Task: "1-live-demo"})
	if envelope.Outcome != "partial" || envelope.Phase != "task" {
		t.Fatalf("Create() = %+v, want partial task", envelope)
	}
	if !worktreetest.HasEffect(envelope, "workspace", "opened", "w7") {
		t.Fatalf("effects = %+v, want the observed w7 repository workspace", envelope.Effects)
	}
	result, ok := envelope.Result.(Result)
	if !ok || result.RepoWorkspaceID != "w7" || result.WorkspaceID != "w2" {
		t.Fatalf("result = %+v, want w7 and w2", envelope.Result)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("external worktree removed: %v", err)
	}
	if got := strings.TrimSpace(worktreetest.RunGit(t, path, "branch", "--show-current")); got != branch {
		t.Fatalf("branch = %q, want %q", got, branch)
	}
	if task := worktreetest.ReadTask(t, home, "1-live-demo"); task.Worktree != nil {
		t.Fatalf("task worktree = %+v, want no false metadata", task.Worktree)
	}
}

func TestFixCreateUnknownPlainFilePathIsNotWorktree(t *testing.T) {
	repo := worktreetest.NewTestRepo(t)
	home := worktreetest.NewTestHome(t, "repo-live")
	worktreetest.WriteProject(t, home, worktreetest.Project(repo))
	worktreetest.WriteTask(t, home, worktreetest.Task())
	path := worktreetest.PermittedPath(home, "repo-live", "1-live-demo")
	content := []byte("cooperating writer file\n")
	worktreetest.StartFakeHerdr(t, func(request worktreetest.FakeRequest) worktreetest.FakeReply {
		switch request.Method {
		case "workspace.list":
			return worktreetest.FakeReply{Result: worktreetest.WorkspaceListReply(t)}
		case "worktree.create":
			if err := os.WriteFile(path, content, 0o644); err != nil {
				t.Fatalf("write plain file: %v", err)
			}
			return worktreetest.FakeReply{Close: true}
		}
		return worktreetest.FakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
	})

	envelope := Run(context.Background(), home, Options{Task: "1-live-demo"})
	if envelope.Outcome != "partial" || envelope.Phase != "request" {
		t.Fatalf("Create() = %+v, want partial request", envelope)
	}
	if worktreetest.HasEffect(envelope, "worktree", "created", path) {
		t.Fatalf("effects = %+v, want no created-worktree claim for a plain file", envelope.Effects)
	}
	if task := worktreetest.ReadTask(t, home, "1-live-demo"); task.Worktree != nil {
		t.Fatalf("task worktree = %+v, want no worktree metadata", task.Worktree)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("plain file removed: %v", err)
	}
	if string(got) != string(content) {
		t.Fatalf("plain file = %q, want %q", got, content)
	}
	if branches := strings.TrimSpace(worktreetest.RunGit(t, repo, "branch", "--list", "task/1-live-demo")); branches != "" {
		t.Fatalf("branch exists = %q, want none", branches)
	}
}

func TestFixCreateUnknownRepositoryOnlyPersistsProject(t *testing.T) {
	repo := worktreetest.NewTestRepo(t)
	home := worktreetest.NewTestHome(t, "repo-live")
	worktreetest.WriteProject(t, home, worktreetest.Project(repo))
	worktreetest.WriteTask(t, home, worktreetest.Task())
	var lists int
	worktreetest.StartFakeHerdr(t, func(request worktreetest.FakeRequest) worktreetest.FakeReply {
		switch request.Method {
		case "workspace.list":
			lists++
			if lists == 1 {
				return worktreetest.FakeReply{Result: worktreetest.WorkspaceListReply(t)}
			}
			return worktreetest.FakeReply{Result: worktreetest.WorkspaceListReply(t, worktreetest.RepositoryWorkspace("w7", repo))}
		case "worktree.create":
			return worktreetest.FakeReply{Close: true}
		}
		return worktreetest.FakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
	})

	envelope := Run(context.Background(), home, Options{Task: "1-live-demo"})
	if envelope.Outcome != "partial" {
		t.Fatalf("Create() = %+v, want partial", envelope)
	}
	if !worktreetest.HasEffect(envelope, "workspace", "opened", "w7") {
		t.Fatalf("effects = %+v, want w7 opened", envelope.Effects)
	}
	if project := worktreetest.ReadProject(t, home, "repo-live"); project.RepoWorkspaceID != "w7" {
		t.Fatalf("repo_workspace_id = %q, want w7 without a linked path", project.RepoWorkspaceID)
	}
	if task := worktreetest.ReadTask(t, home, "1-live-demo"); task.Worktree != nil {
		t.Fatalf("task worktree = %+v, want none", task.Worktree)
	}
}

func TestFixCreateUnknownDiagnosticsStayExplicit(t *testing.T) {
	t.Run("branch inspection", func(t *testing.T) {
		repo := worktreetest.NewTestRepo(t)
		home := worktreetest.NewTestHome(t, "repo-live")
		worktreetest.WriteProject(t, home, worktreetest.Project(repo))
		worktreetest.WriteTask(t, home, worktreetest.Task())
		moved := repo + "-moved"
		t.Cleanup(func() {
			if _, err := os.Stat(moved); err == nil {
				_ = os.Rename(moved, repo)
			}
		})
		worktreetest.StartFakeHerdr(t, func(request worktreetest.FakeRequest) worktreetest.FakeReply {
			switch request.Method {
			case "workspace.list":
				return worktreetest.FakeReply{Result: worktreetest.WorkspaceListReply(t)}
			case "worktree.create":
				if err := os.Rename(repo, moved); err != nil {
					t.Fatalf("move repo: %v", err)
				}
				return worktreetest.FakeReply{Close: true}
			}
			return worktreetest.FakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
		})

		envelope := Run(context.Background(), home, Options{Task: "1-live-demo"})
		if envelope.Outcome != "partial" || !strings.Contains(fmt.Sprint(envelope.Error), "branch") {
			t.Fatalf("Create() = %+v, want a bounded branch inspection diagnostic", envelope)
		}
	})

	t.Run("path inspection", func(t *testing.T) {
		repo := worktreetest.NewTestRepo(t)
		home := worktreetest.NewTestHome(t, "repo-live")
		worktreetest.WriteProject(t, home, worktreetest.Project(repo))
		worktreetest.WriteTask(t, home, worktreetest.Task())
		parent := filepath.Join(home, "projects", "repo-live", "worktrees")
		t.Cleanup(func() { _ = os.Chmod(parent, 0o755) })
		worktreetest.StartFakeHerdr(t, func(request worktreetest.FakeRequest) worktreetest.FakeReply {
			switch request.Method {
			case "workspace.list":
				return worktreetest.FakeReply{Result: worktreetest.WorkspaceListReply(t)}
			case "worktree.create":
				if err := os.Chmod(parent, 0o000); err != nil {
					t.Fatalf("chmod worktrees: %v", err)
				}
				return worktreetest.FakeReply{Close: true}
			}
			return worktreetest.FakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
		})

		envelope := Run(context.Background(), home, Options{Task: "1-live-demo"})
		if envelope.Outcome != "partial" || !strings.Contains(fmt.Sprint(envelope.Error), "inspect path") {
			t.Fatalf("Create() = %+v, want a non-absence path diagnostic", envelope)
		}
	})

	t.Run("repository ambiguity", func(t *testing.T) {
		repo := worktreetest.NewTestRepo(t)
		home := worktreetest.NewTestHome(t, "repo-live")
		worktreetest.WriteProject(t, home, worktreetest.Project(repo))
		worktreetest.WriteTask(t, home, worktreetest.Task())
		var lists int
		worktreetest.StartFakeHerdr(t, func(request worktreetest.FakeRequest) worktreetest.FakeReply {
			switch request.Method {
			case "workspace.list":
				lists++
				if lists == 1 {
					return worktreetest.FakeReply{Result: worktreetest.WorkspaceListReply(t)}
				}
				return worktreetest.FakeReply{Result: worktreetest.WorkspaceListReply(t, worktreetest.RepositoryWorkspace("w7", repo), worktreetest.RepositoryWorkspace("w8", repo))}
			case "worktree.create":
				return worktreetest.FakeReply{Close: true}
			}
			return worktreetest.FakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
		})

		envelope := Run(context.Background(), home, Options{Task: "1-live-demo"})
		if envelope.Outcome != "partial" || !strings.Contains(fmt.Sprint(envelope.Error), "uncertain") {
			t.Fatalf("Create() = %+v, want an explicit ambiguity diagnostic", envelope)
		}
		if effectCount(envelope, "workspace", "opened") != 0 {
			t.Fatalf("effects = %+v, want no invented repository ID", envelope.Effects)
		}
		if project := worktreetest.ReadProject(t, home, "repo-live"); project.RepoWorkspaceID != "" {
			t.Fatalf("repo_workspace_id = %q, want empty", project.RepoWorkspaceID)
		}
	})
}

func TestFixCreateSelectedKeyMustMatchLoadedTask(t *testing.T) {
	cases := []struct {
		name  string
		build func() (records.Task, records.Project)
	}{
		{
			name: "invalid record id",
			build: func() (records.Task, records.Project) {
				task := worktreetest.Task()
				task.ID = "../../../../escape"
				return task, worktreetest.Project("/tmp/repo")
			},
		},
		{
			name: "different valid record id",
			build: func() (records.Task, records.Project) {
				task := worktreetest.Task()
				task.ID = "2-other"
				return task, worktreetest.Project("/tmp/repo")
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := worktreetest.NewTestRepo(t)
			home := worktreetest.NewTestHome(t, "repo-live")
			task, project := tc.build()
			project.Path = repo
			worktreetest.WriteProjectAt(t, home, "repo-live", project)
			worktreetest.WriteTaskAt(t, home, "1-live-demo", task)
			server := worktreetest.StartFakeHerdr(t, func(request worktreetest.FakeRequest) worktreetest.FakeReply {
				return worktreetest.FakeReply{Error: `{"code":"invalid_request","message":"no request allowed"}`}
			})

			envelope := Run(context.Background(), home, Options{Task: "1-live-demo"})
			if envelope.Outcome != "failed" {
				t.Fatalf("Create() = %+v, want failed", envelope)
			}
			if methods := server.Methods(); len(methods) != 0 {
				t.Fatalf("methods = %v, want no request", methods)
			}
			if got := worktreetest.ReadTask(t, home, "1-live-demo"); !reflect.DeepEqual(got, task) {
				t.Fatalf("task changed to %+v, want %+v", got, task)
			}
			if got := worktreetest.ReadProject(t, home, "repo-live"); !reflect.DeepEqual(got, project) {
				t.Fatalf("project changed to %+v, want %+v", got, project)
			}
		})
	}
}

func TestFixCreateSelectedKeyMustMatchLoadedProject(t *testing.T) {
	repo := worktreetest.NewTestRepo(t)
	home := worktreetest.NewTestHome(t, "repo-live")
	project := worktreetest.Project(repo)
	project.ID = "other-project"
	worktreetest.WriteProjectAt(t, home, "repo-live", project)
	worktreetest.WriteTask(t, home, worktreetest.Task())
	server := worktreetest.StartFakeHerdr(t, func(request worktreetest.FakeRequest) worktreetest.FakeReply {
		return worktreetest.FakeReply{Error: `{"code":"invalid_request","message":"no request allowed"}`}
	})

	envelope := Run(context.Background(), home, Options{Task: "1-live-demo"})
	if envelope.Outcome != "failed" {
		t.Fatalf("Create() = %+v, want failed", envelope)
	}
	if methods := server.Methods(); len(methods) != 0 {
		t.Fatalf("methods = %v, want no request", methods)
	}
	if _, err := os.Lstat(filepath.Join(home, "projects", "other-project")); !os.IsNotExist(err) {
		t.Fatalf("other-project namespace touched: %v", err)
	}
}

func TestFixExternalLogsCarryApplicableIDs(t *testing.T) {
	repo := worktreetest.NewTestRepo(t)
	home := worktreetest.NewTestHome(t, "repo-live")
	worktreetest.WriteProject(t, home, worktreetest.Project(repo))
	worktreetest.WriteTask(t, home, worktreetest.Task())
	path := worktreetest.PermittedPath(home, "repo-live", "1-live-demo")
	branch := "task/1-live-demo"
	var lists int
	worktreetest.StartFakeHerdr(t, func(request worktreetest.FakeRequest) worktreetest.FakeReply {
		switch request.Method {
		case "workspace.list":
			lists++
			if lists == 1 {
				return worktreetest.FakeReply{Result: worktreetest.WorkspaceListReply(t)}
			}
			return worktreetest.FakeReply{Result: worktreetest.WorkspaceListReply(t, worktreetest.LinkedWorkspace("w2", path, repo), worktreetest.RepositoryWorkspace("w7", repo))}
		case "worktree.create":
			worktreetest.RunGit(t, repo, "worktree", "add", "-q", "-b", branch, path, "main")
			return worktreetest.FakeReply{Result: worktreetest.CreateReply(t, "w2", "w2:t1", "w2:p1", path, branch, repo)}
		}
		return worktreetest.FakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
	})

	envelope := Run(context.Background(), home, Options{Task: "1-live-demo"})
	if envelope.Outcome != "ok" {
		t.Fatalf("Create() = %+v, want ok", envelope)
	}
	events := worktreetest.ExternalEvents(t, home)
	if len(events) == 0 {
		t.Fatal("no external log events")
	}
	for _, event := range events {
		if event["project_id"] != "repo-live" || event["task_id"] != "1-live-demo" {
			t.Fatalf("external event = %+v, want project/task scope", event)
		}
	}
}

func TestFixLargeDiagnosticLogIsBounded(t *testing.T) {
	repo := worktreetest.NewTestRepo(t)
	home := worktreetest.NewTestHome(t, "repo-live")
	worktreetest.WriteProject(t, home, worktreetest.Project(repo))
	worktreetest.WriteTask(t, home, worktreetest.Task())
	large := strings.Repeat("x", 9000)
	worktreetest.StartFakeHerdr(t, func(request worktreetest.FakeRequest) worktreetest.FakeReply {
		switch request.Method {
		case "workspace.list":
			return worktreetest.FakeReply{Result: worktreetest.WorkspaceListReply(t)}
		case "worktree.create":
			return worktreetest.FakeReply{Error: worktreetest.JSONText(t, map[string]any{"code": "invalid_request", "message": large})}
		}
		return worktreetest.FakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
	})

	envelope := Run(context.Background(), home, Options{Task: "1-live-demo"})
	if envelope.Outcome != "failed" {
		t.Fatalf("Create() = %+v, want failed", envelope)
	}
	found := false
	for _, event := range worktreetest.LogEvents(t, home) {
		if event["msg"] != "worktree error" {
			continue
		}
		found = true
		value, _ := event["error"].(string)
		if len(value) > 4096 {
			t.Fatalf("logged error bytes = %d, want at most 4096", len(value))
		}
		if event["truncated"] != true {
			t.Fatalf("truncated = %v, want true for a shortened diagnostic", event["truncated"])
		}
	}
	if !found {
		t.Fatal("no worktree error event")
	}
}

func TestFixCreateTaskFaultLogsKnownEffect(t *testing.T) {
	repo := worktreetest.NewTestRepo(t)
	home := worktreetest.NewTestHome(t, "repo-live")
	worktreetest.WriteProject(t, home, worktreetest.Project(repo))
	worktreetest.WriteTask(t, home, worktreetest.Task())
	path := worktreetest.PermittedPath(home, "repo-live", "1-live-demo")
	branch := "task/1-live-demo"
	tasksDir := filepath.Join(home, ".hive", "state", "tasks")
	faulted := false
	t.Cleanup(func() { _ = os.Chmod(tasksDir, 0o755) })
	worktreetest.StartFakeHerdr(t, func(request worktreetest.FakeRequest) worktreetest.FakeReply {
		switch request.Method {
		case "workspace.list":
			if !faulted {
				return worktreetest.FakeReply{Result: worktreetest.WorkspaceListReply(t)}
			}
			return worktreetest.FakeReply{Result: worktreetest.WorkspaceListReply(t, worktreetest.LinkedWorkspace("w2", path, repo), worktreetest.RepositoryWorkspace("w7", repo))}
		case "worktree.create":
			worktreetest.RunGit(t, repo, "worktree", "add", "-q", "-b", branch, path, "main")
			if err := os.Chmod(tasksDir, 0o500); err != nil {
				t.Fatalf("chmod tasks: %v", err)
			}
			faulted = true
			return worktreetest.FakeReply{Result: worktreetest.CreateReply(t, "w2", "w2:t1", "w2:p1", path, branch, repo)}
		}
		return worktreetest.FakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
	})

	envelope := Run(context.Background(), home, Options{Task: "1-live-demo"})
	if envelope.Outcome != "partial" {
		t.Fatalf("Create() = %+v, want partial", envelope)
	}
	found := false
	for _, event := range worktreetest.LogEvents(t, home) {
		if event["msg"] != "worktree effect" {
			continue
		}
		if event["resource"] == "workspace" && event["path"] == "w7" && event["action"] == "opened" {
			found = true
		}
	}
	if !found {
		t.Fatal("known w7 effect missing from logs after a persistence failure")
	}
}

func TestFixHeldCreateAndRemoveReleaseStoreLock(t *testing.T) {
	t.Run("create", func(t *testing.T) {
		repo := worktreetest.NewTestRepo(t)
		home := worktreetest.NewTestHome(t, "repo-live")
		worktreetest.WriteProject(t, home, worktreetest.Project(repo))
		worktreetest.WriteTask(t, home, worktreetest.Task())
		path := worktreetest.PermittedPath(home, "repo-live", "1-live-demo")
		branch := "task/1-live-demo"
		arrived := make(chan struct{})
		release := make(chan struct{})
		var lists int
		worktreetest.StartFakeHerdr(t, func(request worktreetest.FakeRequest) worktreetest.FakeReply {
			switch request.Method {
			case "workspace.list":
				lists++
				if lists == 1 {
					return worktreetest.FakeReply{Result: worktreetest.WorkspaceListReply(t)}
				}
				return worktreetest.FakeReply{Result: worktreetest.WorkspaceListReply(t, worktreetest.LinkedWorkspace("w2", path, repo))}
			case "worktree.create":
				close(arrived)
				return worktreetest.FakeReply{Result: worktreetest.CreateReply(t, "w2", "w2:t1", "w2:p1", path, branch, repo), Hold: release}
			}
			return worktreetest.FakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
		})

		done := make(chan struct{})
		go func() {
			Run(context.Background(), home, Options{Task: "1-live-demo"})
			close(done)
		}()
		select {
		case <-arrived:
		case <-time.After(2 * time.Second):
			t.Fatal("create sent no held request")
		}
		registry, err := store.Open(home)
		if err != nil {
			t.Fatalf("open store: %v", err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()
		writeErr := registry.WithLock(ctx, func(locked *store.LockedStore) error {
			changed := worktreetest.Task()
			changed.Title = "held update"
			return locked.WriteIn([]string{"tasks"}, "1-live-demo", changed)
		})
		if writeErr != nil {
			t.Fatalf("store lock was not available during the external wait: %v", writeErr)
		}
		close(release)
		<-done
	})
}

func TestFixPrecheckFailureLogCarriesIDs(t *testing.T) {
	repo := worktreetest.NewTestRepo(t)
	home := worktreetest.NewTestHome(t, "repo-live")
	project := worktreetest.Project(repo)
	project.Base = "missing"
	worktreetest.WriteProject(t, home, project)
	worktreetest.WriteTask(t, home, worktreetest.Task())
	worktreetest.StartFakeHerdr(t, func(request worktreetest.FakeRequest) worktreetest.FakeReply {
		return worktreetest.FakeReply{Error: `{"code":"invalid_request","message":"no request allowed"}`}
	})

	envelope := Run(context.Background(), home, Options{Task: "1-live-demo"})
	if envelope.Outcome != "failed" || fmt.Sprint(envelope.Error) != "base ref not found: missing" {
		t.Fatalf("Create() = %+v, want the base refusal", envelope)
	}
	found := false
	for _, event := range worktreetest.LogEvents(t, home) {
		if event["msg"] != "worktree error" || event["phase"] != "precheck" {
			continue
		}
		found = true
		if event["project_id"] != "repo-live" || event["task_id"] != "1-live-demo" {
			t.Fatalf("precheck error event = %+v, want project/task scope", event)
		}
	}
	if !found {
		t.Fatal("no precheck error event")
	}
}
