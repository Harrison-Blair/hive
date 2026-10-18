package worktree

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

func fixWriteTaskAt(t *testing.T, home, key string, task records.Task) {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	if err := registry.WriteIn([]string{"tasks"}, key, task); err != nil {
		t.Fatalf("write task at %s: %v", key, err)
	}
}

func fixWriteProjectAt(t *testing.T, home, key string, project records.Project) {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	if err := registry.Write("projects", key, project); err != nil {
		t.Fatalf("write project at %s: %v", key, err)
	}
}

func fixPermittedPath(home, projectID, taskID string) string {
	return filepath.Join(home, "projects", projectID, "worktrees", taskID)
}

func fixLogEvents(t *testing.T, home string) []map[string]any {
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

func fixExternalEvents(t *testing.T, home string) []map[string]any {
	t.Helper()
	external := []map[string]any{}
	for _, event := range fixLogEvents(t, home) {
		message, _ := event["msg"].(string)
		switch message {
		case "herdr request", "herdr reply", "process start", "process end":
			external = append(external, event)
		}
	}
	return external
}

func TestFixCreateTaskWriteFaultRetainsKnownRepositoryEffect(t *testing.T) {
	repo := newTestRepo(t)
	home := newTestHome(t, "repo-live")
	writeProject(t, home, testProject(repo))
	writeTask(t, home, testTask())
	path := fixPermittedPath(home, "repo-live", "1-live-demo")
	branch := "task/1-live-demo"
	tasksDir := filepath.Join(home, ".hive", "state", "tasks")
	faulted := false
	t.Cleanup(func() { _ = os.Chmod(tasksDir, 0o755) })
	startFakeHerdr(t, func(request fakeRequest) fakeReply {
		switch request.Method {
		case "workspace.list":
			if !faulted {
				return fakeReply{Result: workspaceListReply(t)}
			}
			return fakeReply{Result: workspaceListReply(t, linkedWorkspace("w2", path, repo), repositoryWorkspace("w7", repo))}
		case "worktree.create":
			runGit(t, repo, "worktree", "add", "-q", "-b", branch, path, "main")
			if err := os.Chmod(tasksDir, 0o500); err != nil {
				t.Fatalf("chmod tasks: %v", err)
			}
			faulted = true
			return fakeReply{Result: createReply(t, "w2", "w2:t1", "w2:p1", path, branch, repo)}
		}
		return fakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
	})

	envelope := Create(context.Background(), home, CreateOptions{Task: "1-live-demo"})
	if envelope.Outcome != "partial" || envelope.Phase != "task" {
		t.Fatalf("Create() = %+v, want partial task", envelope)
	}
	if !hasEffect(envelope, "workspace", "opened", "w7") {
		t.Fatalf("effects = %+v, want the observed w7 repository workspace", envelope.Effects)
	}
	result, ok := envelope.Result.(CreateResult)
	if !ok || result.RepoWorkspaceID != "w7" || result.WorkspaceID != "w2" {
		t.Fatalf("result = %+v, want w7 and w2", envelope.Result)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("external worktree removed: %v", err)
	}
	if got := strings.TrimSpace(runGit(t, path, "branch", "--show-current")); got != branch {
		t.Fatalf("branch = %q, want %q", got, branch)
	}
	if task := readTask(t, home, "1-live-demo"); task.Worktree != nil {
		t.Fatalf("task worktree = %+v, want no false metadata", task.Worktree)
	}
}

func TestFixCreateUnknownPlainFilePathIsNotWorktree(t *testing.T) {
	repo := newTestRepo(t)
	home := newTestHome(t, "repo-live")
	writeProject(t, home, testProject(repo))
	writeTask(t, home, testTask())
	path := fixPermittedPath(home, "repo-live", "1-live-demo")
	content := []byte("cooperating writer file\n")
	startFakeHerdr(t, func(request fakeRequest) fakeReply {
		switch request.Method {
		case "workspace.list":
			return fakeReply{Result: workspaceListReply(t)}
		case "worktree.create":
			if err := os.WriteFile(path, content, 0o644); err != nil {
				t.Fatalf("write plain file: %v", err)
			}
			return fakeReply{Close: true}
		}
		return fakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
	})

	envelope := Create(context.Background(), home, CreateOptions{Task: "1-live-demo"})
	if envelope.Outcome != "partial" || envelope.Phase != "request" {
		t.Fatalf("Create() = %+v, want partial request", envelope)
	}
	if hasEffect(envelope, "worktree", "created", path) {
		t.Fatalf("effects = %+v, want no created-worktree claim for a plain file", envelope.Effects)
	}
	if task := readTask(t, home, "1-live-demo"); task.Worktree != nil {
		t.Fatalf("task worktree = %+v, want no worktree metadata", task.Worktree)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("plain file removed: %v", err)
	}
	if string(got) != string(content) {
		t.Fatalf("plain file = %q, want %q", got, content)
	}
	if branches := strings.TrimSpace(runGit(t, repo, "branch", "--list", "task/1-live-demo")); branches != "" {
		t.Fatalf("branch exists = %q, want none", branches)
	}
}

func TestFixCreateUnknownRepositoryOnlyPersistsProject(t *testing.T) {
	repo := newTestRepo(t)
	home := newTestHome(t, "repo-live")
	writeProject(t, home, testProject(repo))
	writeTask(t, home, testTask())
	var lists int
	startFakeHerdr(t, func(request fakeRequest) fakeReply {
		switch request.Method {
		case "workspace.list":
			lists++
			if lists == 1 {
				return fakeReply{Result: workspaceListReply(t)}
			}
			return fakeReply{Result: workspaceListReply(t, repositoryWorkspace("w7", repo))}
		case "worktree.create":
			return fakeReply{Close: true}
		}
		return fakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
	})

	envelope := Create(context.Background(), home, CreateOptions{Task: "1-live-demo"})
	if envelope.Outcome != "partial" {
		t.Fatalf("Create() = %+v, want partial", envelope)
	}
	if !hasEffect(envelope, "workspace", "opened", "w7") {
		t.Fatalf("effects = %+v, want w7 opened", envelope.Effects)
	}
	if project := readProject(t, home, "repo-live"); project.RepoWorkspaceID != "w7" {
		t.Fatalf("repo_workspace_id = %q, want w7 without a linked path", project.RepoWorkspaceID)
	}
	if task := readTask(t, home, "1-live-demo"); task.Worktree != nil {
		t.Fatalf("task worktree = %+v, want none", task.Worktree)
	}
}

func TestFixCreateUnknownDiagnosticsStayExplicit(t *testing.T) {
	t.Run("branch inspection", func(t *testing.T) {
		repo := newTestRepo(t)
		home := newTestHome(t, "repo-live")
		writeProject(t, home, testProject(repo))
		writeTask(t, home, testTask())
		moved := repo + "-moved"
		t.Cleanup(func() {
			if _, err := os.Stat(moved); err == nil {
				_ = os.Rename(moved, repo)
			}
		})
		startFakeHerdr(t, func(request fakeRequest) fakeReply {
			switch request.Method {
			case "workspace.list":
				return fakeReply{Result: workspaceListReply(t)}
			case "worktree.create":
				if err := os.Rename(repo, moved); err != nil {
					t.Fatalf("move repo: %v", err)
				}
				return fakeReply{Close: true}
			}
			return fakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
		})

		envelope := Create(context.Background(), home, CreateOptions{Task: "1-live-demo"})
		if envelope.Outcome != "partial" || !strings.Contains(fmt.Sprint(envelope.Error), "branch") {
			t.Fatalf("Create() = %+v, want a bounded branch inspection diagnostic", envelope)
		}
	})

	t.Run("path inspection", func(t *testing.T) {
		repo := newTestRepo(t)
		home := newTestHome(t, "repo-live")
		writeProject(t, home, testProject(repo))
		writeTask(t, home, testTask())
		parent := filepath.Join(home, "projects", "repo-live", "worktrees")
		t.Cleanup(func() { _ = os.Chmod(parent, 0o755) })
		startFakeHerdr(t, func(request fakeRequest) fakeReply {
			switch request.Method {
			case "workspace.list":
				return fakeReply{Result: workspaceListReply(t)}
			case "worktree.create":
				if err := os.Chmod(parent, 0o000); err != nil {
					t.Fatalf("chmod worktrees: %v", err)
				}
				return fakeReply{Close: true}
			}
			return fakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
		})

		envelope := Create(context.Background(), home, CreateOptions{Task: "1-live-demo"})
		if envelope.Outcome != "partial" || !strings.Contains(fmt.Sprint(envelope.Error), "inspect path") {
			t.Fatalf("Create() = %+v, want a non-absence path diagnostic", envelope)
		}
	})

	t.Run("repository ambiguity", func(t *testing.T) {
		repo := newTestRepo(t)
		home := newTestHome(t, "repo-live")
		writeProject(t, home, testProject(repo))
		writeTask(t, home, testTask())
		var lists int
		startFakeHerdr(t, func(request fakeRequest) fakeReply {
			switch request.Method {
			case "workspace.list":
				lists++
				if lists == 1 {
					return fakeReply{Result: workspaceListReply(t)}
				}
				return fakeReply{Result: workspaceListReply(t, repositoryWorkspace("w7", repo), repositoryWorkspace("w8", repo))}
			case "worktree.create":
				return fakeReply{Close: true}
			}
			return fakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
		})

		envelope := Create(context.Background(), home, CreateOptions{Task: "1-live-demo"})
		if envelope.Outcome != "partial" || !strings.Contains(fmt.Sprint(envelope.Error), "uncertain") {
			t.Fatalf("Create() = %+v, want an explicit ambiguity diagnostic", envelope)
		}
		if effectCount(envelope, "workspace", "opened") != 0 {
			t.Fatalf("effects = %+v, want no invented repository ID", envelope.Effects)
		}
		if project := readProject(t, home, "repo-live"); project.RepoWorkspaceID != "" {
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
				task := testTask()
				task.ID = "../../../../escape"
				return task, testProject("/tmp/repo")
			},
		},
		{
			name: "different valid record id",
			build: func() (records.Task, records.Project) {
				task := testTask()
				task.ID = "2-other"
				return task, testProject("/tmp/repo")
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := newTestRepo(t)
			home := newTestHome(t, "repo-live")
			task, project := tc.build()
			project.Path = repo
			fixWriteProjectAt(t, home, "repo-live", project)
			fixWriteTaskAt(t, home, "1-live-demo", task)
			server := startFakeHerdr(t, func(request fakeRequest) fakeReply {
				return fakeReply{Error: `{"code":"invalid_request","message":"no request allowed"}`}
			})

			envelope := Create(context.Background(), home, CreateOptions{Task: "1-live-demo"})
			if envelope.Outcome != "failed" {
				t.Fatalf("Create() = %+v, want failed", envelope)
			}
			if methods := server.methods(); len(methods) != 0 {
				t.Fatalf("methods = %v, want no request", methods)
			}
			if got := readTask(t, home, "1-live-demo"); !reflect.DeepEqual(got, task) {
				t.Fatalf("task changed to %+v, want %+v", got, task)
			}
			if got := readProject(t, home, "repo-live"); !reflect.DeepEqual(got, project) {
				t.Fatalf("project changed to %+v, want %+v", got, project)
			}
		})
	}
}

func TestFixCreateSelectedKeyMustMatchLoadedProject(t *testing.T) {
	repo := newTestRepo(t)
	home := newTestHome(t, "repo-live")
	project := testProject(repo)
	project.ID = "other-project"
	fixWriteProjectAt(t, home, "repo-live", project)
	writeTask(t, home, testTask())
	server := startFakeHerdr(t, func(request fakeRequest) fakeReply {
		return fakeReply{Error: `{"code":"invalid_request","message":"no request allowed"}`}
	})

	envelope := Create(context.Background(), home, CreateOptions{Task: "1-live-demo"})
	if envelope.Outcome != "failed" {
		t.Fatalf("Create() = %+v, want failed", envelope)
	}
	if methods := server.methods(); len(methods) != 0 {
		t.Fatalf("methods = %v, want no request", methods)
	}
	if _, err := os.Lstat(filepath.Join(home, "projects", "other-project")); !os.IsNotExist(err) {
		t.Fatalf("other-project namespace touched: %v", err)
	}
}

func TestFixRemoveSelectedKeyMustMatchLoadedTask(t *testing.T) {
	repo := newTestRepo(t)
	home := newTestHome(t, "repo-live")
	path := fixPermittedPath(home, "repo-live", "1-live-demo")
	branch := "task/1-live-demo"
	addLinkedWorktree(t, repo, path, branch, "main")
	writeProject(t, home, testProject(repo))
	mismatched := worktreeTask(path, branch)
	mismatched.ID = "2-other"
	mismatched.Worktree.WorkspaceID = ""
	fixWriteTaskAt(t, home, "1-live-demo", mismatched)
	other := worktreeTask("/unrelated/history", "task/2-other")
	other.ID = "2-other"
	writeTask(t, home, other)
	server := startFakeHerdr(t, func(request fakeRequest) fakeReply {
		return fakeReply{Error: `{"code":"invalid_request","message":"no request allowed"}`}
	})

	envelope := Remove(context.Background(), home, RemoveOptions{Task: "1-live-demo"})
	if envelope.Outcome != "failed" {
		t.Fatalf("Remove() = %+v, want failed", envelope)
	}
	if methods := server.methods(); len(methods) != 0 {
		t.Fatalf("methods = %v, want no request", methods)
	}
	if _, err := os.Lstat(path); err != nil {
		t.Fatalf("worktree removed for a mismatched record: %v", err)
	}
	if got := readTask(t, home, "2-other"); !reflect.DeepEqual(got, other) {
		t.Fatalf("different record changed to %+v, want %+v", got, other)
	}
}

func TestFixRemoveSelectedKeyMustMatchLoadedProject(t *testing.T) {
	repo := newTestRepo(t)
	home := newTestHome(t, "repo-live")
	path := fixPermittedPath(home, "repo-live", "1-live-demo")
	branch := "task/1-live-demo"
	addLinkedWorktree(t, repo, path, branch, "main")
	project := testProject(repo)
	project.ID = "other-project"
	fixWriteProjectAt(t, home, "repo-live", project)
	task := worktreeTask(path, branch)
	task.Worktree.WorkspaceID = ""
	writeTask(t, home, task)
	server := startFakeHerdr(t, func(request fakeRequest) fakeReply {
		return fakeReply{Error: `{"code":"invalid_request","message":"no request allowed"}`}
	})

	envelope := Remove(context.Background(), home, RemoveOptions{Task: "1-live-demo"})
	if envelope.Outcome != "failed" {
		t.Fatalf("Remove() = %+v, want failed", envelope)
	}
	if methods := server.methods(); len(methods) != 0 {
		t.Fatalf("methods = %v, want no request", methods)
	}
	if _, err := os.Lstat(path); err != nil {
		t.Fatalf("worktree removed for a mismatched project record: %v", err)
	}
}

func TestFixRemoveRejectsLinkedRegisteredSource(t *testing.T) {
	for _, force := range []bool{false, true} {
		t.Run(fmt.Sprint(force), func(t *testing.T) {
			repo := newTestRepo(t)
			home := newTestHome(t, "repo-live")
			path := fixPermittedPath(home, "repo-live", "1-live-demo")
			branch := "task/1-live-demo"
			addLinkedWorktree(t, repo, path, branch, "main")
			source := filepath.Join(t.TempDir(), "linked-source")
			addLinkedWorktree(t, repo, source, "task/source", "main")
			project := testProject(repo)
			project.Path = source
			writeProject(t, home, project)
			task := worktreeTask(path, branch)
			task.Worktree.WorkspaceID = ""
			writeTask(t, home, task)
			server := startFakeHerdr(t, func(request fakeRequest) fakeReply {
				return fakeReply{Error: `{"code":"invalid_request","message":"no request allowed"}`}
			})

			envelope := Remove(context.Background(), home, RemoveOptions{Task: "1-live-demo", Force: force})
			if envelope.Outcome != "failed" || fmt.Sprint(envelope.Error) != "repository is a linked worktree: register the primary checkout" {
				t.Fatalf("Remove() = %+v, want the registered-primary refusal", envelope)
			}
			if methods := server.methods(); len(methods) != 0 {
				t.Fatalf("methods = %v, want no request", methods)
			}
			if _, err := os.Lstat(path); err != nil {
				t.Fatalf("worktree removed through a linked registered source: %v", err)
			}
		})
	}
}

func TestFixRemoveRejectsPathAndBranchMetadata(t *testing.T) {
	for _, force := range []bool{false, true} {
		t.Run(fmt.Sprintf("path/%v", force), func(t *testing.T) {
			repo := newTestRepo(t)
			home := newTestHome(t, "repo-live")
			path := filepath.Join(t.TempDir(), "unrelated-checkout")
			branch := "user/precious"
			addLinkedWorktree(t, repo, path, branch, "main")
			writeProject(t, home, testProject(repo))
			task := worktreeTask(path, branch)
			task.Worktree.WorkspaceID = ""
			writeTask(t, home, task)
			server := startFakeHerdr(t, func(request fakeRequest) fakeReply {
				return fakeReply{Error: `{"code":"invalid_request","message":"no request allowed"}`}
			})

			envelope := Remove(context.Background(), home, RemoveOptions{Task: "1-live-demo", Force: force})
			if envelope.Outcome != "failed" || fmt.Sprint(envelope.Error) != "worktree mismatch: path" {
				t.Fatalf("Remove() = %+v, want the permitted-path refusal", envelope)
			}
			if methods := server.methods(); len(methods) != 0 {
				t.Fatalf("methods = %v, want no request", methods)
			}
			if _, err := os.Lstat(path); err != nil {
				t.Fatalf("unrelated checkout removed: %v", err)
			}
		})

		t.Run(fmt.Sprintf("branch/%v", force), func(t *testing.T) {
			repo := newTestRepo(t)
			home := newTestHome(t, "repo-live")
			path := fixPermittedPath(home, "repo-live", "1-live-demo")
			branch := "user/precious"
			addLinkedWorktree(t, repo, path, branch, "main")
			writeProject(t, home, testProject(repo))
			task := worktreeTask(path, branch)
			task.Worktree.WorkspaceID = ""
			writeTask(t, home, task)
			server := startFakeHerdr(t, func(request fakeRequest) fakeReply {
				return fakeReply{Error: `{"code":"invalid_request","message":"no request allowed"}`}
			})

			envelope := Remove(context.Background(), home, RemoveOptions{Task: "1-live-demo", Force: force})
			if envelope.Outcome != "failed" || fmt.Sprint(envelope.Error) != "worktree mismatch: branch" {
				t.Fatalf("Remove() = %+v, want the task-branch refusal", envelope)
			}
			if methods := server.methods(); len(methods) != 0 {
				t.Fatalf("methods = %v, want no request", methods)
			}
			if _, err := os.Lstat(path); err != nil {
				t.Fatalf("non-task branch checkout removed: %v", err)
			}
		})
	}
}

func TestFixExternalLogsCarryApplicableIDs(t *testing.T) {
	repo := newTestRepo(t)
	home := newTestHome(t, "repo-live")
	writeProject(t, home, testProject(repo))
	writeTask(t, home, testTask())
	path := fixPermittedPath(home, "repo-live", "1-live-demo")
	branch := "task/1-live-demo"
	var lists int
	startFakeHerdr(t, func(request fakeRequest) fakeReply {
		switch request.Method {
		case "workspace.list":
			lists++
			if lists == 1 {
				return fakeReply{Result: workspaceListReply(t)}
			}
			return fakeReply{Result: workspaceListReply(t, linkedWorkspace("w2", path, repo), repositoryWorkspace("w7", repo))}
		case "worktree.create":
			runGit(t, repo, "worktree", "add", "-q", "-b", branch, path, "main")
			return fakeReply{Result: createReply(t, "w2", "w2:t1", "w2:p1", path, branch, repo)}
		}
		return fakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
	})

	envelope := Create(context.Background(), home, CreateOptions{Task: "1-live-demo"})
	if envelope.Outcome != "ok" {
		t.Fatalf("Create() = %+v, want ok", envelope)
	}
	events := fixExternalEvents(t, home)
	if len(events) == 0 {
		t.Fatal("no external log events")
	}
	for _, event := range events {
		if event["project_id"] != "repo-live" || event["task_id"] != "1-live-demo" {
			t.Fatalf("external event = %+v, want project/task scope", event)
		}
	}
}

func TestFixRemoveExternalLogsCarryApplicableIDs(t *testing.T) {
	repo := newTestRepo(t)
	home := newTestHome(t, "repo-live")
	path := fixPermittedPath(home, "repo-live", "1-live-demo")
	branch := "task/1-live-demo"
	addLinkedWorktree(t, repo, path, branch, "main")
	writeProject(t, home, testProject(repo))
	task := worktreeTask(path, branch)
	writeTask(t, home, task)
	startFakeHerdr(t, func(request fakeRequest) fakeReply {
		switch request.Method {
		case "workspace.get":
			return fakeReply{Result: workspaceInfoReply(t, linkedWorkspace("w2", path, repo))}
		case "workspace.close":
			return fakeReply{Result: `{"type":"ok"}`}
		}
		return fakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
	})

	envelope := Remove(context.Background(), home, RemoveOptions{Task: "1-live-demo"})
	if envelope.Outcome != "ok" {
		t.Fatalf("Remove() = %+v, want ok", envelope)
	}
	events := fixExternalEvents(t, home)
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
	repo := newTestRepo(t)
	home := newTestHome(t, "repo-live")
	writeProject(t, home, testProject(repo))
	writeTask(t, home, testTask())
	large := strings.Repeat("x", 9000)
	startFakeHerdr(t, func(request fakeRequest) fakeReply {
		switch request.Method {
		case "workspace.list":
			return fakeReply{Result: workspaceListReply(t)}
		case "worktree.create":
			return fakeReply{Error: jsonText(t, map[string]any{"code": "invalid_request", "message": large})}
		}
		return fakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
	})

	envelope := Create(context.Background(), home, CreateOptions{Task: "1-live-demo"})
	if envelope.Outcome != "failed" {
		t.Fatalf("Create() = %+v, want failed", envelope)
	}
	found := false
	for _, event := range fixLogEvents(t, home) {
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
	repo := newTestRepo(t)
	home := newTestHome(t, "repo-live")
	writeProject(t, home, testProject(repo))
	writeTask(t, home, testTask())
	path := fixPermittedPath(home, "repo-live", "1-live-demo")
	branch := "task/1-live-demo"
	tasksDir := filepath.Join(home, ".hive", "state", "tasks")
	faulted := false
	t.Cleanup(func() { _ = os.Chmod(tasksDir, 0o755) })
	startFakeHerdr(t, func(request fakeRequest) fakeReply {
		switch request.Method {
		case "workspace.list":
			if !faulted {
				return fakeReply{Result: workspaceListReply(t)}
			}
			return fakeReply{Result: workspaceListReply(t, linkedWorkspace("w2", path, repo), repositoryWorkspace("w7", repo))}
		case "worktree.create":
			runGit(t, repo, "worktree", "add", "-q", "-b", branch, path, "main")
			if err := os.Chmod(tasksDir, 0o500); err != nil {
				t.Fatalf("chmod tasks: %v", err)
			}
			faulted = true
			return fakeReply{Result: createReply(t, "w2", "w2:t1", "w2:p1", path, branch, repo)}
		}
		return fakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
	})

	envelope := Create(context.Background(), home, CreateOptions{Task: "1-live-demo"})
	if envelope.Outcome != "partial" {
		t.Fatalf("Create() = %+v, want partial", envelope)
	}
	found := false
	for _, event := range fixLogEvents(t, home) {
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
		repo := newTestRepo(t)
		home := newTestHome(t, "repo-live")
		writeProject(t, home, testProject(repo))
		writeTask(t, home, testTask())
		path := fixPermittedPath(home, "repo-live", "1-live-demo")
		branch := "task/1-live-demo"
		arrived := make(chan struct{})
		release := make(chan struct{})
		var lists int
		startFakeHerdr(t, func(request fakeRequest) fakeReply {
			switch request.Method {
			case "workspace.list":
				lists++
				if lists == 1 {
					return fakeReply{Result: workspaceListReply(t)}
				}
				return fakeReply{Result: workspaceListReply(t, linkedWorkspace("w2", path, repo))}
			case "worktree.create":
				close(arrived)
				return fakeReply{Result: createReply(t, "w2", "w2:t1", "w2:p1", path, branch, repo), Hold: release}
			}
			return fakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
		})

		done := make(chan struct{})
		go func() {
			Create(context.Background(), home, CreateOptions{Task: "1-live-demo"})
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
			changed := testTask()
			changed.Title = "held update"
			return locked.WriteIn([]string{"tasks"}, "1-live-demo", changed)
		})
		if writeErr != nil {
			t.Fatalf("store lock was not available during the external wait: %v", writeErr)
		}
		close(release)
		<-done
	})

	t.Run("remove", func(t *testing.T) {
		repo := newTestRepo(t)
		home := newTestHome(t, "repo-live")
		path := fixPermittedPath(home, "repo-live", "1-live-demo")
		branch := "task/1-live-demo"
		addLinkedWorktree(t, repo, path, branch, "main")
		writeProject(t, home, testProject(repo))
		task := worktreeTask(path, branch)
		writeTask(t, home, task)
		arrived := make(chan struct{})
		release := make(chan struct{})
		startFakeHerdr(t, func(request fakeRequest) fakeReply {
			switch request.Method {
			case "workspace.get":
				close(arrived)
				return fakeReply{Result: workspaceInfoReply(t, linkedWorkspace("w2", path, repo)), Hold: release}
			case "workspace.close":
				return fakeReply{Result: `{"type":"ok"}`}
			}
			return fakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
		})

		done := make(chan struct{})
		go func() {
			Remove(context.Background(), home, RemoveOptions{Task: "1-live-demo"})
			close(done)
		}()
		select {
		case <-arrived:
		case <-time.After(2 * time.Second):
			t.Fatal("remove sent no held request")
		}
		registry, err := store.Open(home)
		if err != nil {
			t.Fatalf("open store: %v", err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()
		writeErr := registry.WithLock(ctx, func(locked *store.LockedStore) error {
			changed := worktreeTask(path, branch)
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
	repo := newTestRepo(t)
	home := newTestHome(t, "repo-live")
	project := testProject(repo)
	project.Base = "missing"
	writeProject(t, home, project)
	writeTask(t, home, testTask())
	startFakeHerdr(t, func(request fakeRequest) fakeReply {
		return fakeReply{Error: `{"code":"invalid_request","message":"no request allowed"}`}
	})

	envelope := Create(context.Background(), home, CreateOptions{Task: "1-live-demo"})
	if envelope.Outcome != "failed" || fmt.Sprint(envelope.Error) != "base ref not found: missing" {
		t.Fatalf("Create() = %+v, want the base refusal", envelope)
	}
	found := false
	for _, event := range fixLogEvents(t, home) {
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
