package remove

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Harrison-Blair/hive/internal/lib/gitx"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/worktree/worktreetest"
)

func TestRemoveSuccessClosesWorkspaceAndRemovesWorktree(t *testing.T) {
	repo := worktreetest.NewTestRepo(t)
	home := worktreetest.NewTestHome(t, "repo-live")
	path := filepath.Join(home, "projects", "repo-live", "worktrees", "1-live-demo")
	branch := "task/1-live-demo"
	worktreetest.AddLinkedWorktree(t, repo, path, branch, "main")
	worktreetest.WriteProject(t, home, worktreetest.Project(repo))
	worktreetest.WriteTask(t, home, worktreetest.WorktreeTask(path, branch))
	server := worktreetest.StartFakeHerdr(t, func(request worktreetest.FakeRequest) worktreetest.FakeReply {
		switch request.Method {
		case "workspace.get":
			return worktreetest.FakeReply{Result: worktreetest.WorkspaceInfoReply(t, worktreetest.LinkedWorkspace("w2", path, repo))}
		case "agent.list":
			return worktreetest.FakeReply{Result: `{"type":"agent_list","agents":[]}`}
		case "workspace.close":
			return worktreetest.FakeReply{Result: `{"type":"ok"}`}
		}
		return worktreetest.FakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
	})

	envelope := Run(context.Background(), home, Options{Task: "1-live-demo"})
	if envelope.Outcome != "ok" || envelope.Phase != "complete" || envelope.Error != nil {
		t.Fatalf("Remove() = %+v, want ok complete", envelope)
	}
	if !worktreetest.HasEffect(envelope, "workspace", "closed", "w2") || !worktreetest.HasEffect(envelope, "worktree", "removed", path) {
		t.Fatalf("effects = %+v, want workspace closed and worktree removed", envelope.Effects)
	}
	if _, err := os.Lstat(path); !os.IsNotExist(err) {
		t.Fatalf("worktree still exists: %v", err)
	}
	exists, err := gitBranchExists(repo, branch)
	if err != nil || !exists {
		t.Fatalf("branch %s exists = %v, %v, want true", branch, exists, err)
	}
	task := worktreetest.ReadTask(t, home, "1-live-demo")
	if task.Worktree == nil {
		t.Fatalf("task worktree = nil, want the history record")
	}
	if task.Worktree.RemovedAt == "" {
		t.Fatalf("task worktree = %+v, want removed_at", task.Worktree)
	}
	if task.Worktree.WorkspaceID != "" || task.Worktree.TabID != "" || task.Worktree.RootPane != "" {
		t.Fatalf("task worktree live fields = %+v, want cleared", task.Worktree)
	}
	if task.Worktree.Path != path || task.Worktree.Branch != branch || task.Worktree.CreatedAt == "" {
		t.Fatalf("task worktree history = %+v, want path, branch, and created_at", task.Worktree)
	}
	if methods := server.Methods(); !reflect.DeepEqual(methods, []string{"workspace.get", "agent.list", "workspace.close"}) {
		t.Fatalf("methods = %v, want get, agent.list, then close", methods)
	}
}

func gitBranchExists(repo, branch string) (bool, error) {
	return gitx.BranchExists(context.Background(), repo, branch)
}

func TestRemoveStatusGuard(t *testing.T) {
	repo := worktreetest.NewTestRepo(t)
	home := worktreetest.NewTestHome(t, "repo-live")
	path := filepath.Join(home, "projects", "repo-live", "worktrees", "1-live-demo")
	branch := "task/1-live-demo"
	worktreetest.AddLinkedWorktree(t, repo, path, branch, "main")
	task := worktreetest.WorktreeTask(path, branch)
	task.Status = "created"
	worktreetest.WriteProject(t, home, worktreetest.Project(repo))
	worktreetest.WriteTask(t, home, task)
	server := worktreetest.StartFakeHerdr(t, func(request worktreetest.FakeRequest) worktreetest.FakeReply {
		return worktreetest.FakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
	})

	envelope := Run(context.Background(), home, Options{Task: "1-live-demo"})
	if envelope.Outcome != "failed" || fmt.Sprint(envelope.Error) != "task is created: complete, verify, or cancel it first" {
		t.Fatalf("Remove() = %+v, want the status guard", envelope)
	}
	if methods := server.Methods(); len(methods) != 0 {
		t.Fatalf("methods = %v, want no request", methods)
	}
	if _, err := os.Lstat(path); err != nil {
		t.Fatalf("worktree path removed: %v", err)
	}
	if got := worktreetest.ReadTask(t, home, "1-live-demo"); !reflect.DeepEqual(got, task) {
		t.Fatalf("task changed to %+v, want %+v", got, task)
	}
}

func TestRemoveNoWorktreeGuard(t *testing.T) {
	repo := worktreetest.NewTestRepo(t)
	home := worktreetest.NewTestHome(t, "repo-live")
	task := worktreetest.Task()
	task.Status = "verified"
	worktreetest.WriteProject(t, home, worktreetest.Project(repo))
	worktreetest.WriteTask(t, home, task)
	server := worktreetest.StartFakeHerdr(t, func(request worktreetest.FakeRequest) worktreetest.FakeReply {
		return worktreetest.FakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
	})

	envelope := Run(context.Background(), home, Options{Task: "1-live-demo"})
	if envelope.Outcome != "failed" || fmt.Sprint(envelope.Error) != "task has no worktree" {
		t.Fatalf("Remove() = %+v, want the worktree guard", envelope)
	}
	if methods := server.Methods(); len(methods) != 0 {
		t.Fatalf("methods = %v, want no request", methods)
	}
}

func TestRemoveDirtyGuard(t *testing.T) {
	repo := worktreetest.NewTestRepo(t)
	home := worktreetest.NewTestHome(t, "repo-live")
	path := filepath.Join(home, "projects", "repo-live", "worktrees", "1-live-demo")
	branch := "task/1-live-demo"
	worktreetest.AddLinkedWorktree(t, repo, path, branch, "main")
	if err := os.WriteFile(filepath.Join(path, "dirty.txt"), []byte("dirty\n"), 0o644); err != nil {
		t.Fatalf("write dirty: %v", err)
	}
	worktreetest.WriteProject(t, home, worktreetest.Project(repo))
	worktreetest.WriteTask(t, home, worktreetest.WorktreeTask(path, branch))
	server := worktreetest.StartFakeHerdr(t, func(request worktreetest.FakeRequest) worktreetest.FakeReply {
		if request.Method == "workspace.get" {
			return worktreetest.FakeReply{Result: worktreetest.WorkspaceInfoReply(t, worktreetest.LinkedWorkspace("w2", path, repo))}
		}
		return worktreetest.FakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
	})

	envelope := Run(context.Background(), home, Options{Task: "1-live-demo"})
	if envelope.Outcome != "failed" || fmt.Sprint(envelope.Error) != "worktree is dirty" {
		t.Fatalf("Remove() = %+v, want the dirty guard", envelope)
	}
	if methods := server.Methods(); !reflect.DeepEqual(methods, []string{"workspace.get"}) {
		t.Fatalf("methods = %v, want only workspace.get", methods)
	}
	if _, err := os.Lstat(path); err != nil {
		t.Fatalf("worktree path removed: %v", err)
	}
}

func TestRemoveMergedGuard(t *testing.T) {
	repo := worktreetest.NewTestRepo(t)
	home := worktreetest.NewTestHome(t, "repo-live")
	path := filepath.Join(home, "projects", "repo-live", "worktrees", "1-live-demo")
	branch := "task/1-live-demo"
	worktreetest.AddLinkedWorktree(t, repo, path, branch, "main")
	if err := os.WriteFile(filepath.Join(path, "work.txt"), []byte("work\n"), 0o644); err != nil {
		t.Fatalf("write work: %v", err)
	}
	worktreetest.RunGit(t, path, "add", "work.txt")
	worktreetest.RunGit(t, path, "commit", "-q", "-m", "work")
	worktreetest.WriteProject(t, home, worktreetest.Project(repo))
	worktreetest.WriteTask(t, home, worktreetest.WorktreeTask(path, branch))
	server := worktreetest.StartFakeHerdr(t, func(request worktreetest.FakeRequest) worktreetest.FakeReply {
		if request.Method == "workspace.get" {
			return worktreetest.FakeReply{Result: worktreetest.WorkspaceInfoReply(t, worktreetest.LinkedWorkspace("w2", path, repo))}
		}
		return worktreetest.FakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
	})

	envelope := Run(context.Background(), home, Options{Task: "1-live-demo"})
	if envelope.Outcome != "failed" || fmt.Sprint(envelope.Error) != "branch not merged into main" {
		t.Fatalf("Remove() = %+v, want the merged guard", envelope)
	}
	if methods := server.Methods(); !reflect.DeepEqual(methods, []string{"workspace.get"}) {
		t.Fatalf("methods = %v, want only workspace.get", methods)
	}
	if _, err := os.Lstat(path); err != nil {
		t.Fatalf("worktree path removed: %v", err)
	}
}

func TestRemoveForceSkipsDirtyAndMergedGuards(t *testing.T) {
	repo := worktreetest.NewTestRepo(t)
	home := worktreetest.NewTestHome(t, "repo-live")
	path := filepath.Join(home, "projects", "repo-live", "worktrees", "1-live-demo")
	branch := "task/1-live-demo"
	worktreetest.AddLinkedWorktree(t, repo, path, branch, "main")
	if err := os.WriteFile(filepath.Join(path, "work.txt"), []byte("work\n"), 0o644); err != nil {
		t.Fatalf("write work: %v", err)
	}
	worktreetest.RunGit(t, path, "add", "work.txt")
	worktreetest.RunGit(t, path, "commit", "-q", "-m", "work")
	if err := os.WriteFile(filepath.Join(path, "dirty.txt"), []byte("dirty\n"), 0o644); err != nil {
		t.Fatalf("write dirty: %v", err)
	}
	worktreetest.WriteProject(t, home, worktreetest.Project(repo))
	worktreetest.WriteTask(t, home, worktreetest.WorktreeTask(path, branch))
	server := worktreetest.StartFakeHerdr(t, func(request worktreetest.FakeRequest) worktreetest.FakeReply {
		switch request.Method {
		case "workspace.get":
			return worktreetest.FakeReply{Result: worktreetest.WorkspaceInfoReply(t, worktreetest.LinkedWorkspace("w2", path, repo))}
		case "agent.list":
			return worktreetest.FakeReply{Result: `{"type":"agent_list","agents":[]}`}
		case "workspace.close":
			return worktreetest.FakeReply{Result: `{"type":"ok"}`}
		}
		return worktreetest.FakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
	})

	envelope := Run(context.Background(), home, Options{Task: "1-live-demo", Force: true})
	if envelope.Outcome != "ok" {
		t.Fatalf("Remove() = %+v, want ok with force", envelope)
	}
	if _, err := os.Lstat(path); !os.IsNotExist(err) {
		t.Fatalf("forced worktree still exists: %v", err)
	}
	exists, err := gitBranchExists(repo, branch)
	if err != nil || !exists {
		t.Fatalf("branch %s exists = %v, %v, want true", branch, exists, err)
	}
	if methods := server.Methods(); !reflect.DeepEqual(methods, []string{"workspace.get", "agent.list", "workspace.close"}) {
		t.Fatalf("methods = %v, want get, agent.list, then close", methods)
	}
}

func TestRemoveClosedWorkspaceSkipsClose(t *testing.T) {
	repo := worktreetest.NewTestRepo(t)
	home := worktreetest.NewTestHome(t, "repo-live")
	path := filepath.Join(home, "projects", "repo-live", "worktrees", "1-live-demo")
	branch := "task/1-live-demo"
	worktreetest.AddLinkedWorktree(t, repo, path, branch, "main")
	worktreetest.WriteProject(t, home, worktreetest.Project(repo))
	worktreetest.WriteTask(t, home, worktreetest.WorktreeTask(path, branch))
	server := worktreetest.StartFakeHerdr(t, func(request worktreetest.FakeRequest) worktreetest.FakeReply {
		if request.Method == "workspace.get" {
			return worktreetest.FakeReply{Error: `{"code":"workspace_not_found","message":"workspace w2 not found"}`}
		}
		if request.Method == "agent.list" {
			return worktreetest.FakeReply{Result: `{"type":"agent_list","agents":[]}`}
		}
		return worktreetest.FakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
	})

	envelope := Run(context.Background(), home, Options{Task: "1-live-demo"})
	if envelope.Outcome != "ok" {
		t.Fatalf("Remove() = %+v, want ok", envelope)
	}
	if worktreetest.HasEffect(envelope, "workspace", "closed", "w2") {
		t.Fatalf("effects = %+v, want no close effect", envelope.Effects)
	}
	if !worktreetest.HasEffect(envelope, "worktree", "removed", path) {
		t.Fatalf("effects = %+v, want worktree removed", envelope.Effects)
	}
	if _, err := os.Lstat(path); !os.IsNotExist(err) {
		t.Fatalf("worktree still exists: %v", err)
	}
	if methods := server.Methods(); !reflect.DeepEqual(methods, []string{"workspace.get", "agent.list"}) {
		t.Fatalf("methods = %v, want workspace.get and agent.list", methods)
	}
}

func TestRemoveIdentityMismatchRefusesWithForce(t *testing.T) {
	cases := []struct {
		name    string
		want    string
		prepare func(t *testing.T, home string) (repo, path string, task records.Task, handler func(worktreetest.FakeRequest) worktreetest.FakeReply)
	}{
		{
			name: "git is_linked_worktree",
			want: "worktree mismatch: is_linked_worktree",
			prepare: func(t *testing.T, home string) (string, string, records.Task, func(worktreetest.FakeRequest) worktreetest.FakeReply) {
				repo := worktreetest.NewTestRepo(t)
				path := filepath.Join(home, "projects", "repo-live", "worktrees", "1-live-demo")
				if err := os.MkdirAll(path, 0o755); err != nil {
					t.Fatalf("mkdir primary checkout: %v", err)
				}
				worktreetest.RunGit(t, path, "init", "-q", "-b", "main")
				worktreetest.RunGit(t, path, "config", "user.name", "Hive Test")
				worktreetest.RunGit(t, path, "config", "user.email", "test@example.com")
				worktreetest.RunGit(t, path, "commit", "-q", "--allow-empty", "-m", "init")
				return repo, path, worktreetest.WorktreeTask(path, "task/1-live-demo"), nil
			},
		},
		{
			name: "git repo_root",
			want: "worktree mismatch: repo_root",
			prepare: func(t *testing.T, home string) (string, string, records.Task, func(worktreetest.FakeRequest) worktreetest.FakeReply) {
				registered := worktreetest.NewTestRepo(t)
				other := worktreetest.NewTestRepo(t)
				path := filepath.Join(home, "projects", "repo-live", "worktrees", "1-live-demo")
				worktreetest.AddLinkedWorktree(t, other, path, "task/1-live-demo", "main")
				return registered, path, worktreetest.WorktreeTask(path, "task/1-live-demo"), nil
			},
		},
		{
			name: "git branch",
			want: "worktree mismatch: branch",
			prepare: func(t *testing.T, home string) (string, string, records.Task, func(worktreetest.FakeRequest) worktreetest.FakeReply) {
				repo := worktreetest.NewTestRepo(t)
				path := filepath.Join(home, "projects", "repo-live", "worktrees", "1-live-demo")
				worktreetest.AddLinkedWorktree(t, repo, path, "task/1-live-demo", "main")
				return repo, path, worktreetest.WorktreeTask(path, "task/other"), nil
			},
		},
		{
			name: "herdr is_linked_worktree",
			want: "worktree mismatch: is_linked_worktree",
			prepare: func(t *testing.T, home string) (string, string, records.Task, func(worktreetest.FakeRequest) worktreetest.FakeReply) {
				repo := worktreetest.NewTestRepo(t)
				path := filepath.Join(home, "projects", "repo-live", "worktrees", "1-live-demo")
				worktreetest.AddLinkedWorktree(t, repo, path, "task/1-live-demo", "main")
				workspace := worktreetest.LinkedWorkspace("w2", path, repo)
				workspace.Worktree.IsLinkedWorktree = false
				return repo, path, worktreetest.WorktreeTask(path, "task/1-live-demo"), func(request worktreetest.FakeRequest) worktreetest.FakeReply {
					if request.Method == "workspace.get" {
						return worktreetest.FakeReply{Result: worktreetest.WorkspaceInfoReply(t, workspace)}
					}
					return worktreetest.FakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
				}
			},
		},
		{
			name: "herdr checkout_path",
			want: "worktree mismatch: checkout_path",
			prepare: func(t *testing.T, home string) (string, string, records.Task, func(worktreetest.FakeRequest) worktreetest.FakeReply) {
				repo := worktreetest.NewTestRepo(t)
				path := filepath.Join(home, "projects", "repo-live", "worktrees", "1-live-demo")
				worktreetest.AddLinkedWorktree(t, repo, path, "task/1-live-demo", "main")
				workspace := worktreetest.LinkedWorkspace("w2", filepath.Join(t.TempDir(), "other"), repo)
				return repo, path, worktreetest.WorktreeTask(path, "task/1-live-demo"), func(request worktreetest.FakeRequest) worktreetest.FakeReply {
					if request.Method == "workspace.get" {
						return worktreetest.FakeReply{Result: worktreetest.WorkspaceInfoReply(t, workspace)}
					}
					return worktreetest.FakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
				}
			},
		},
		{
			name: "herdr repo_root",
			want: "worktree mismatch: repo_root",
			prepare: func(t *testing.T, home string) (string, string, records.Task, func(worktreetest.FakeRequest) worktreetest.FakeReply) {
				repo := worktreetest.NewTestRepo(t)
				path := filepath.Join(home, "projects", "repo-live", "worktrees", "1-live-demo")
				worktreetest.AddLinkedWorktree(t, repo, path, "task/1-live-demo", "main")
				workspace := worktreetest.LinkedWorkspace("w2", path, repo)
				workspace.Worktree.RepoRoot = t.TempDir()
				return repo, path, worktreetest.WorktreeTask(path, "task/1-live-demo"), func(request worktreetest.FakeRequest) worktreetest.FakeReply {
					if request.Method == "workspace.get" {
						return worktreetest.FakeReply{Result: worktreetest.WorkspaceInfoReply(t, workspace)}
					}
					return worktreetest.FakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
				}
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			home := worktreetest.NewTestHome(t, "repo-live")
			repo, path, task, handler := tc.prepare(t, home)
			worktreetest.WriteProject(t, home, worktreetest.Project(repo))
			worktreetest.WriteTask(t, home, task)
			server := worktreetest.StartFakeHerdr(t, func(request worktreetest.FakeRequest) worktreetest.FakeReply {
				if handler != nil {
					return handler(request)
				}
				return worktreetest.FakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
			})

			envelope := Run(context.Background(), home, Options{Task: "1-live-demo", Force: true})
			if envelope.Outcome != "failed" {
				t.Fatalf("Remove() = %+v, want failed", envelope)
			}
			if got := fmt.Sprint(envelope.Error); !strings.Contains(got, tc.want) {
				t.Fatalf("error = %q, want %q", got, tc.want)
			}
			if _, found := server.RequestFor("workspace.close"); found {
				t.Fatalf("methods = %v, want no close", server.Methods())
			}
			if _, err := os.Lstat(path); err != nil {
				t.Fatalf("worktree path removed after mismatch: %v", err)
			}
			if got := worktreetest.ReadTask(t, home, "1-live-demo"); got.Worktree == nil || got.Worktree.RemovedAt != "" {
				t.Fatalf("task worktree = %+v, want unchanged", got.Worktree)
			}
		})
	}
}

func TestRemoveGitFailureAfterCloseIsPartial(t *testing.T) {
	repo := worktreetest.NewTestRepo(t)
	home := worktreetest.NewTestHome(t, "repo-live")
	path := filepath.Join(home, "projects", "repo-live", "worktrees", "1-live-demo")
	branch := "task/1-live-demo"
	worktreetest.AddLinkedWorktree(t, repo, path, branch, "main")
	worktreetest.WriteProject(t, home, worktreetest.Project(repo))
	worktreetest.WriteTask(t, home, worktreetest.WorktreeTask(path, branch))
	worktreetest.StartFakeHerdr(t, func(request worktreetest.FakeRequest) worktreetest.FakeReply {
		switch request.Method {
		case "workspace.get":
			return worktreetest.FakeReply{Result: worktreetest.WorkspaceInfoReply(t, worktreetest.LinkedWorkspace("w2", path, repo))}
		case "agent.list":
			return worktreetest.FakeReply{Result: `{"type":"agent_list","agents":[]}`}
		case "workspace.close":
			if err := os.Remove(filepath.Join(path, ".git")); err != nil {
				t.Fatalf("remove worktree marker: %v", err)
			}
			return worktreetest.FakeReply{Result: `{"type":"ok"}`}
		}
		return worktreetest.FakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
	})

	envelope := Run(context.Background(), home, Options{Task: "1-live-demo"})
	if envelope.Outcome != "partial" || envelope.Phase != "worktree" {
		t.Fatalf("Remove() = %+v, want partial worktree", envelope)
	}
	if !worktreetest.HasEffect(envelope, "workspace", "closed", "w2") {
		t.Fatalf("effects = %+v, want workspace closed", envelope.Effects)
	}
	if worktreetest.HasEffect(envelope, "worktree", "removed", path) {
		t.Fatalf("effects = %+v, want no worktree removed effect", envelope.Effects)
	}
	if got := fmt.Sprint(envelope.Error); !strings.Contains(got, "fatal") {
		t.Fatalf("error = %q, want the Git output", got)
	}
	if task := worktreetest.ReadTask(t, home, "1-live-demo"); task.Worktree.RemovedAt != "" {
		t.Fatalf("task worktree = %+v, want no removal record", task.Worktree)
	}
}

func TestRemoveUnknownCloseOutcomeIsPartial(t *testing.T) {
	repo := worktreetest.NewTestRepo(t)
	home := worktreetest.NewTestHome(t, "repo-live")
	path := filepath.Join(home, "projects", "repo-live", "worktrees", "1-live-demo")
	branch := "task/1-live-demo"
	worktreetest.AddLinkedWorktree(t, repo, path, branch, "main")
	worktreetest.WriteProject(t, home, worktreetest.Project(repo))
	worktreetest.WriteTask(t, home, worktreetest.WorktreeTask(path, branch))
	worktreetest.StartFakeHerdr(t, func(request worktreetest.FakeRequest) worktreetest.FakeReply {
		switch request.Method {
		case "workspace.get":
			return worktreetest.FakeReply{Result: worktreetest.WorkspaceInfoReply(t, worktreetest.LinkedWorkspace("w2", path, repo))}
		case "agent.list":
			return worktreetest.FakeReply{Result: `{"type":"agent_list","agents":[]}`}
		case "workspace.close":
			return worktreetest.FakeReply{Close: true}
		}
		return worktreetest.FakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
	})

	envelope := Run(context.Background(), home, Options{Task: "1-live-demo"})
	if envelope.Outcome != "partial" || envelope.Phase != "workspace" {
		t.Fatalf("Remove() = %+v, want partial workspace", envelope)
	}
	if len(envelope.Effects) != 0 {
		t.Fatalf("effects = %+v, want no claimed effect", envelope.Effects)
	}
	if _, err := os.Lstat(path); err != nil {
		t.Fatalf("worktree removed after unknown close: %v", err)
	}
	if task := worktreetest.ReadTask(t, home, "1-live-demo"); task.Worktree.RemovedAt != "" {
		t.Fatalf("task worktree = %+v, want no removal record", task.Worktree)
	}
}

func TestRemoveGroupRootRefused(t *testing.T) {
	repo := worktreetest.NewTestRepo(t)
	home := worktreetest.NewTestHome(t, "repo-live")
	path := filepath.Join(home, "projects", "repo-live", "worktrees", "1-live-demo")
	branch := "task/1-live-demo"
	worktreetest.AddLinkedWorktree(t, repo, path, branch, "main")
	worktreetest.WriteProject(t, home, worktreetest.Project(repo))
	worktreetest.WriteTask(t, home, worktreetest.WorktreeTask(path, branch))
	worktreetest.StartFakeHerdr(t, func(request worktreetest.FakeRequest) worktreetest.FakeReply {
		switch request.Method {
		case "workspace.get":
			return worktreetest.FakeReply{Result: worktreetest.WorkspaceInfoReply(t, worktreetest.LinkedWorkspace("w2", path, repo))}
		case "agent.list":
			return worktreetest.FakeReply{Result: `{"type":"agent_list","agents":[]}`}
		case "workspace.close":
			return worktreetest.FakeReply{Error: `{"code":"workspace_group_close_required","message":"workspace has linked worktree workspaces; use --group (close_group=true in the API) to close the group"}`}
		}
		return worktreetest.FakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
	})

	envelope := Run(context.Background(), home, Options{Task: "1-live-demo"})
	if envelope.Outcome != "failed" || fmt.Sprint(envelope.Error) != "workspace is a group root: close it by hand" {
		t.Fatalf("Remove() = %+v, want the group root error", envelope)
	}
	if _, err := os.Lstat(path); err != nil {
		t.Fatalf("worktree removed after group root error: %v", err)
	}
}

func TestRemovePreservesConcurrentTaskChange(t *testing.T) {
	repo := worktreetest.NewTestRepo(t)
	home := worktreetest.NewTestHome(t, "repo-live")
	path := filepath.Join(home, "projects", "repo-live", "worktrees", "1-live-demo")
	branch := "task/1-live-demo"
	worktreetest.AddLinkedWorktree(t, repo, path, branch, "main")
	worktreetest.WriteProject(t, home, worktreetest.Project(repo))
	worktreetest.WriteTask(t, home, worktreetest.WorktreeTask(path, branch))
	arrived := make(chan struct{})
	release := make(chan struct{})
	worktreetest.StartFakeHerdr(t, func(request worktreetest.FakeRequest) worktreetest.FakeReply {
		switch request.Method {
		case "workspace.get":
			close(arrived)
			return worktreetest.FakeReply{Result: worktreetest.WorkspaceInfoReply(t, worktreetest.LinkedWorkspace("w2", path, repo)), Hold: release}
		case "agent.list":
			return worktreetest.FakeReply{Result: `{"type":"agent_list","agents":[]}`}
		case "workspace.close":
			return worktreetest.FakeReply{Result: `{"type":"ok"}`}
		}
		return worktreetest.FakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
	})

	done := make(chan output.Envelope, 1)
	go func() { done <- Run(context.Background(), home, Options{Task: "1-live-demo"}) }()
	select {
	case <-arrived:
	case <-time.After(2 * time.Second):
		t.Fatal("remove sent no workspace.get request")
	}
	changed := worktreetest.WorktreeTask(path, branch)
	changed.Status = "cancelled"
	changed.Title = "changed title"
	changed.ActingAgent = "worker-b"
	changed.CancelReason = "user stopped it"
	changed.Verification = &records.Verification{Agent: "verifier-b", Result: "pass", Reason: "checked", At: "2026-10-18T02:00:00Z"}
	worktreetest.WriteTask(t, home, changed)
	close(release)
	envelope := <-done
	if envelope.Outcome != "ok" {
		t.Fatalf("Remove() = %+v, want ok", envelope)
	}
	task := worktreetest.ReadTask(t, home, "1-live-demo")
	if task.Status != "cancelled" || task.Title != "changed title" || task.ActingAgent != "worker-b" || task.CancelReason != "user stopped it" {
		t.Fatalf("task = %+v, want the concurrent fields", task)
	}
	if task.Verification == nil || task.Verification.Agent != "verifier-b" {
		t.Fatalf("task verification = %+v, want the concurrent value", task.Verification)
	}
	if task.Worktree == nil || task.Worktree.RemovedAt == "" || task.Worktree.WorkspaceID != "" {
		t.Fatalf("task worktree = %+v, want removed_at and cleared fields", task.Worktree)
	}
}
