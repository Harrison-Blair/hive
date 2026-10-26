package remove

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/Harrison-Blair/hive/internal/lib/store"
	"github.com/Harrison-Blair/hive/internal/worktree/worktreetest"
)

func TestFixRemoveSelectedKeyMustMatchLoadedTask(t *testing.T) {
	repo := worktreetest.NewTestRepo(t)
	home := worktreetest.NewTestHome(t, "repo-live")
	path := worktreetest.PermittedPath(home, "repo-live", "1-live-demo")
	branch := "task/1-live-demo"
	worktreetest.AddLinkedWorktree(t, repo, path, branch, "main")
	worktreetest.WriteProject(t, home, worktreetest.Project(repo))
	mismatched := worktreetest.WorktreeTask(path, branch)
	mismatched.ID = "2-other"
	mismatched.Worktree.WorkspaceID = ""
	worktreetest.WriteTaskAt(t, home, "1-live-demo", mismatched)
	other := worktreetest.WorktreeTask("/unrelated/history", "task/2-other")
	other.ID = "2-other"
	worktreetest.WriteTask(t, home, other)
	server := worktreetest.StartFakeHerdr(t, func(request worktreetest.FakeRequest) worktreetest.FakeReply {
		return worktreetest.FakeReply{Error: `{"code":"invalid_request","message":"no request allowed"}`}
	})

	envelope := Run(context.Background(), home, Options{Task: "1-live-demo"})
	if envelope.Outcome != "failed" {
		t.Fatalf("Remove() = %+v, want failed", envelope)
	}
	if methods := server.Methods(); len(methods) != 0 {
		t.Fatalf("methods = %v, want no request", methods)
	}
	if _, err := os.Lstat(path); err != nil {
		t.Fatalf("worktree removed for a mismatched record: %v", err)
	}
	if got := worktreetest.ReadTask(t, home, "2-other"); !reflect.DeepEqual(got, other) {
		t.Fatalf("different record changed to %+v, want %+v", got, other)
	}
}

func TestFixRemoveSelectedKeyMustMatchLoadedProject(t *testing.T) {
	repo := worktreetest.NewTestRepo(t)
	home := worktreetest.NewTestHome(t, "repo-live")
	path := worktreetest.PermittedPath(home, "repo-live", "1-live-demo")
	branch := "task/1-live-demo"
	worktreetest.AddLinkedWorktree(t, repo, path, branch, "main")
	project := worktreetest.Project(repo)
	project.ID = "other-project"
	worktreetest.WriteProjectAt(t, home, "repo-live", project)
	task := worktreetest.WorktreeTask(path, branch)
	task.Worktree.WorkspaceID = ""
	worktreetest.WriteTask(t, home, task)
	server := worktreetest.StartFakeHerdr(t, func(request worktreetest.FakeRequest) worktreetest.FakeReply {
		return worktreetest.FakeReply{Error: `{"code":"invalid_request","message":"no request allowed"}`}
	})

	envelope := Run(context.Background(), home, Options{Task: "1-live-demo"})
	if envelope.Outcome != "failed" {
		t.Fatalf("Remove() = %+v, want failed", envelope)
	}
	if methods := server.Methods(); len(methods) != 0 {
		t.Fatalf("methods = %v, want no request", methods)
	}
	if _, err := os.Lstat(path); err != nil {
		t.Fatalf("worktree removed for a mismatched project record: %v", err)
	}
}

func TestFixRemoveRejectsLinkedRegisteredSource(t *testing.T) {
	for _, force := range []bool{false, true} {
		t.Run(fmt.Sprint(force), func(t *testing.T) {
			repo := worktreetest.NewTestRepo(t)
			home := worktreetest.NewTestHome(t, "repo-live")
			path := worktreetest.PermittedPath(home, "repo-live", "1-live-demo")
			branch := "task/1-live-demo"
			worktreetest.AddLinkedWorktree(t, repo, path, branch, "main")
			source := filepath.Join(t.TempDir(), "linked-source")
			worktreetest.AddLinkedWorktree(t, repo, source, "task/source", "main")
			project := worktreetest.Project(repo)
			project.Path = source
			worktreetest.WriteProject(t, home, project)
			task := worktreetest.WorktreeTask(path, branch)
			task.Worktree.WorkspaceID = ""
			worktreetest.WriteTask(t, home, task)
			server := worktreetest.StartFakeHerdr(t, func(request worktreetest.FakeRequest) worktreetest.FakeReply {
				return worktreetest.FakeReply{Error: `{"code":"invalid_request","message":"no request allowed"}`}
			})

			envelope := Run(context.Background(), home, Options{Task: "1-live-demo", Force: force})
			if envelope.Outcome != "failed" || fmt.Sprint(envelope.Error) != "repository is a linked worktree: register the primary checkout" {
				t.Fatalf("Remove() = %+v, want the registered-primary refusal", envelope)
			}
			if methods := server.Methods(); len(methods) != 0 {
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
			repo := worktreetest.NewTestRepo(t)
			home := worktreetest.NewTestHome(t, "repo-live")
			path := filepath.Join(t.TempDir(), "unrelated-checkout")
			branch := "user/precious"
			worktreetest.AddLinkedWorktree(t, repo, path, branch, "main")
			worktreetest.WriteProject(t, home, worktreetest.Project(repo))
			task := worktreetest.WorktreeTask(path, branch)
			task.Worktree.WorkspaceID = ""
			worktreetest.WriteTask(t, home, task)
			server := worktreetest.StartFakeHerdr(t, func(request worktreetest.FakeRequest) worktreetest.FakeReply {
				return worktreetest.FakeReply{Error: `{"code":"invalid_request","message":"no request allowed"}`}
			})

			envelope := Run(context.Background(), home, Options{Task: "1-live-demo", Force: force})
			if envelope.Outcome != "failed" || fmt.Sprint(envelope.Error) != "worktree mismatch: path" {
				t.Fatalf("Remove() = %+v, want the permitted-path refusal", envelope)
			}
			if methods := server.Methods(); len(methods) != 0 {
				t.Fatalf("methods = %v, want no request", methods)
			}
			if _, err := os.Lstat(path); err != nil {
				t.Fatalf("unrelated checkout removed: %v", err)
			}
		})

		t.Run(fmt.Sprintf("branch/%v", force), func(t *testing.T) {
			repo := worktreetest.NewTestRepo(t)
			home := worktreetest.NewTestHome(t, "repo-live")
			path := worktreetest.PermittedPath(home, "repo-live", "1-live-demo")
			branch := "user/precious"
			worktreetest.AddLinkedWorktree(t, repo, path, branch, "main")
			worktreetest.WriteProject(t, home, worktreetest.Project(repo))
			task := worktreetest.WorktreeTask(path, branch)
			task.Worktree.WorkspaceID = ""
			worktreetest.WriteTask(t, home, task)
			server := worktreetest.StartFakeHerdr(t, func(request worktreetest.FakeRequest) worktreetest.FakeReply {
				return worktreetest.FakeReply{Error: `{"code":"invalid_request","message":"no request allowed"}`}
			})

			envelope := Run(context.Background(), home, Options{Task: "1-live-demo", Force: force})
			if envelope.Outcome != "failed" || fmt.Sprint(envelope.Error) != "worktree mismatch: branch" {
				t.Fatalf("Remove() = %+v, want the task-branch refusal", envelope)
			}
			if methods := server.Methods(); len(methods) != 0 {
				t.Fatalf("methods = %v, want no request", methods)
			}
			if _, err := os.Lstat(path); err != nil {
				t.Fatalf("non-task branch checkout removed: %v", err)
			}
		})
	}
}

func TestFixRemoveExternalLogsCarryApplicableIDs(t *testing.T) {
	repo := worktreetest.NewTestRepo(t)
	home := worktreetest.NewTestHome(t, "repo-live")
	path := worktreetest.PermittedPath(home, "repo-live", "1-live-demo")
	branch := "task/1-live-demo"
	worktreetest.AddLinkedWorktree(t, repo, path, branch, "main")
	worktreetest.WriteProject(t, home, worktreetest.Project(repo))
	task := worktreetest.WorktreeTask(path, branch)
	worktreetest.WriteTask(t, home, task)
	worktreetest.StartFakeHerdr(t, func(request worktreetest.FakeRequest) worktreetest.FakeReply {
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
	if envelope.Outcome != "ok" {
		t.Fatalf("Remove() = %+v, want ok", envelope)
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

func TestFixHeldCreateAndRemoveReleaseStoreLock(t *testing.T) {
	t.Run("remove", func(t *testing.T) {
		repo := worktreetest.NewTestRepo(t)
		home := worktreetest.NewTestHome(t, "repo-live")
		path := worktreetest.PermittedPath(home, "repo-live", "1-live-demo")
		branch := "task/1-live-demo"
		worktreetest.AddLinkedWorktree(t, repo, path, branch, "main")
		worktreetest.WriteProject(t, home, worktreetest.Project(repo))
		task := worktreetest.WorktreeTask(path, branch)
		worktreetest.WriteTask(t, home, task)
		arrived := make(chan struct{})
		release := make(chan struct{})
		worktreetest.StartFakeHerdr(t, func(request worktreetest.FakeRequest) worktreetest.FakeReply {
			switch request.Method {
			case "workspace.get":
				close(arrived)
				return worktreetest.FakeReply{Result: worktreetest.WorkspaceInfoReply(t, worktreetest.LinkedWorkspace("w2", path, repo)), Hold: release}
			case "workspace.close":
				return worktreetest.FakeReply{Result: `{"type":"ok"}`}
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
			t.Fatal("remove sent no held request")
		}
		registry, err := store.Open(home)
		if err != nil {
			t.Fatalf("open store: %v", err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()
		writeErr := registry.WithLock(ctx, func(locked *store.LockedStore) error {
			changed := worktreetest.WorktreeTask(path, branch)
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
