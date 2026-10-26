package create

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/Harrison-Blair/hive/internal/worktree/worktreetest"
)

// fix02ProjectRecord returns the raw stored project record for byte checks.
func fix02ProjectRecord(t *testing.T, home, projectID string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(home, ".hive", "state", "projects", projectID+".json"))
	if err != nil {
		t.Fatalf("read project record: %v", err)
	}
	return string(data)
}

func TestFix02CreateLatestRepositoryWorkspaceConflict(t *testing.T) {
	repo := worktreetest.NewTestRepo(t)
	home := worktreetest.NewTestHome(t, "repo-live")
	project := worktreetest.Project(repo)
	worktreetest.WriteProject(t, home, project)
	worktreetest.WriteTask(t, home, worktreetest.Task())
	path := filepath.Join(home, "projects", "repo-live", "worktrees", "1-live-demo")
	branch := "task/1-live-demo"
	lists := 0
	worktreetest.StartFakeHerdr(t, func(request worktreetest.FakeRequest) worktreetest.FakeReply {
		switch request.Method {
		case "workspace.list":
			lists++
			if lists == 1 {
				return worktreetest.FakeReply{Result: worktreetest.WorkspaceListReply(t)}
			}
			latest := worktreetest.ReadProject(t, home, "repo-live")
			latest.RepoWorkspaceID = "w99"
			worktreetest.WriteProjectAt(t, home, "repo-live", latest)
			return worktreetest.FakeReply{Result: worktreetest.WorkspaceListReply(t, worktreetest.LinkedWorkspace("w2", path, repo), worktreetest.RepositoryWorkspace("w7", repo))}
		case "worktree.create":
			worktreetest.AddLinkedWorktree(t, repo, path, branch, "main")
			return worktreetest.FakeReply{Result: worktreetest.CreateReply(t, "w2", "w2:t1", "w2:p1", path, branch, repo)}
		}
		return worktreetest.FakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
	})

	envelope := Run(context.Background(), home, Options{Task: "1-live-demo"})
	if envelope.Outcome != "partial" || envelope.Phase != "project" {
		t.Fatalf("Create() = %+v, want partial project", envelope)
	}
	for _, want := range []struct{ resource, action, path string }{
		{"workspace", "created", "w2"},
		{"worktree", "created", path},
		{"branch", "created", branch},
		{"workspace", "opened", "w7"},
	} {
		if !worktreetest.HasEffect(envelope, want.resource, want.action, want.path) {
			t.Fatalf("effects = %+v, want %s %s %s", envelope.Effects, want.resource, want.action, want.path)
		}
	}
	result, ok := envelope.Result.(Result)
	if !ok || result.WorkspaceID != "w2" || result.RepoWorkspaceID != "w7" {
		t.Fatalf("result = %+v, want w2 and observed w7", envelope.Result)
	}
	got := worktreetest.ReadProject(t, home, "repo-live")
	if got.RepoWorkspaceID != "w99" {
		t.Fatalf("repo_workspace_id = %q, want the current w99", got.RepoWorkspaceID)
	}
	if got.Path != project.Path || got.Base != project.Base || got.RegisteredAt != project.RegisteredAt {
		t.Fatalf("project = %+v, want unchanged path, base, and registration", got)
	}
	raw := fix02ProjectRecord(t, home, "repo-live")
	if !strings.Contains(raw, `"repo_workspace_id": "w99"`) || strings.Contains(raw, `"repo_workspace_id": "w7"`) {
		t.Fatalf("stored project bytes = %s, want the current w99", raw)
	}
	task := worktreetest.ReadTask(t, home, "1-live-demo")
	if task.Worktree == nil || task.Worktree.WorkspaceID != "w2" || task.Worktree.Path != path || task.Worktree.Branch != branch {
		t.Fatalf("task worktree = %+v, want the successful task-write prefix", task.Worktree)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("external worktree lost: %v", err)
	}
	if got := strings.TrimSpace(worktreetest.RunGit(t, path, "branch", "--show-current")); got != branch {
		t.Fatalf("branch = %q, want %q", got, branch)
	}
}

func TestFix02CreateLatestRepositoryOwnerConflict(t *testing.T) {
	repo := worktreetest.NewTestRepo(t)
	home := worktreetest.NewTestHome(t, "repo-live")
	project := worktreetest.Project(repo)
	worktreetest.WriteProject(t, home, project)
	worktreetest.WriteTask(t, home, worktreetest.Task())
	other := worktreetest.NewTestRepo(t)
	path := filepath.Join(home, "projects", "repo-live", "worktrees", "1-live-demo")
	branch := "task/1-live-demo"
	lists := 0
	worktreetest.StartFakeHerdr(t, func(request worktreetest.FakeRequest) worktreetest.FakeReply {
		switch request.Method {
		case "workspace.list":
			lists++
			if lists == 1 {
				return worktreetest.FakeReply{Result: worktreetest.WorkspaceListReply(t)}
			}
			return worktreetest.FakeReply{Result: worktreetest.WorkspaceListReply(t, worktreetest.LinkedWorkspace("w2", path, repo), worktreetest.RepositoryWorkspace("w7", repo))}
		case "worktree.create":
			worktreetest.AddLinkedWorktree(t, repo, path, branch, "main")
			latest := worktreetest.ReadProject(t, home, "repo-live")
			latest.Path = other
			latest.Base = "release"
			worktreetest.WriteProjectAt(t, home, "repo-live", latest)
			return worktreetest.FakeReply{Result: worktreetest.CreateReply(t, "w2", "w2:t1", "w2:p1", path, branch, repo)}
		}
		return worktreetest.FakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
	})

	envelope := Run(context.Background(), home, Options{Task: "1-live-demo"})
	if envelope.Outcome != "partial" {
		t.Fatalf("Create() = %+v, want partial", envelope)
	}
	for _, want := range []struct{ resource, action, path string }{
		{"workspace", "created", "w2"},
		{"worktree", "created", path},
		{"branch", "created", branch},
		{"workspace", "opened", "w7"},
	} {
		if !worktreetest.HasEffect(envelope, want.resource, want.action, want.path) {
			t.Fatalf("effects = %+v, want %s %s %s", envelope.Effects, want.resource, want.action, want.path)
		}
	}
	result, ok := envelope.Result.(Result)
	if !ok || result.RepoWorkspaceID != "w7" {
		t.Fatalf("result = %+v, want the observed w7", envelope.Result)
	}
	got := worktreetest.ReadProject(t, home, "repo-live")
	if got.Path != other {
		t.Fatalf("project path = %q, want the current other repository %q", got.Path, other)
	}
	if got.RepoWorkspaceID != "" {
		t.Fatalf("repo_workspace_id = %q, want no old w7 on the changed owner", got.RepoWorkspaceID)
	}
	if got.Base != "release" || got.RegisteredAt != project.RegisteredAt {
		t.Fatalf("project = %+v, want unrelated current fields preserved", got)
	}
	raw := fix02ProjectRecord(t, home, "repo-live")
	if !strings.Contains(raw, `"path": "`+other+`"`) || strings.Contains(raw, `"repo_workspace_id"`) {
		t.Fatalf("stored project bytes = %s, want the changed owner without w7", raw)
	}
	task := worktreetest.ReadTask(t, home, "1-live-demo")
	if task.Worktree == nil || task.Worktree.WorkspaceID != "w2" || task.Worktree.Path != path {
		t.Fatalf("task worktree = %+v, want the successful task-write prefix", task.Worktree)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("external worktree lost: %v", err)
	}
}

func TestFix02UnicodeDiagnosticBound(t *testing.T) {
	diagnostic := strings.Repeat("\u20ac", 3000)
	t.Run("create", func(t *testing.T) {
		repo := worktreetest.NewTestRepo(t)
		home := worktreetest.NewTestHome(t, "repo-live")
		path := filepath.Join(home, "projects", "repo-live", "worktrees", "1-live-demo")
		worktreetest.WriteProject(t, home, worktreetest.Project(repo))
		worktreetest.WriteTask(t, home, worktreetest.Task())
		worktreetest.StartFakeHerdr(t, func(request worktreetest.FakeRequest) worktreetest.FakeReply {
			switch request.Method {
			case "workspace.list":
				return worktreetest.FakeReply{Result: worktreetest.WorkspaceListReply(t)}
			case "workspace.get":
				return worktreetest.FakeReply{Result: worktreetest.WorkspaceInfoReply(t, worktreetest.LinkedWorkspace("w2", path, repo))}
			case "agent.list":
				return worktreetest.FakeReply{Result: `{"type":"agent_list","agents":[]}`}
			case "worktree.create", "workspace.close":
				return worktreetest.FakeReply{Error: worktreetest.JSONText(t, map[string]any{"code": "invalid_request", "message": diagnostic})}
			}
			return worktreetest.FakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
		})

		envelope := Run(context.Background(), home, Options{Task: "1-live-demo"})
		if envelope.Outcome != "failed" {
			t.Fatalf("command = %+v, want failed", envelope)
		}
		outputError, ok := envelope.Error.(string)
		if !ok || len(outputError) != len(diagnostic) {
			t.Fatalf("output error = %v, want the separate full %d-byte diagnostic", envelope.Error, len(diagnostic))
		}
		found := false
		for _, event := range worktreetest.LogEvents(t, home) {
			if event["msg"] != "worktree error" {
				continue
			}
			found = true
			text, _ := event["error"].(string)
			if len(text) > 4096 {
				t.Fatalf("decoded logged error bytes = %d, want at most 4096", len(text))
			}
			if !utf8.ValidString(text) {
				t.Fatalf("logged error is not valid UTF-8: %q", text)
			}
			if event["truncated"] != true {
				t.Fatalf("truncated = %v, want true", event["truncated"])
			}
		}
		if !found {
			t.Fatal("missing worktree error event")
		}
	})
}

func TestFix02CreateProjectMetadataControls(t *testing.T) {
	for _, control := range []string{"unchanged", "same desired", "physical alias", "unrelated field"} {
		t.Run(control, func(t *testing.T) {
			repo := worktreetest.NewTestRepo(t)
			home := worktreetest.NewTestHome(t, "repo-live")
			project := worktreetest.Project(repo)
			worktreetest.WriteProject(t, home, project)
			worktreetest.WriteTask(t, home, worktreetest.Task())
			path := filepath.Join(home, "projects", "repo-live", "worktrees", "1-live-demo")
			branch := "task/1-live-demo"
			alias := filepath.Join(t.TempDir(), "repo-alias")
			if err := os.Symlink(repo, alias); err != nil {
				t.Fatalf("symlink repository: %v", err)
			}
			lists := 0
			worktreetest.StartFakeHerdr(t, func(request worktreetest.FakeRequest) worktreetest.FakeReply {
				switch request.Method {
				case "workspace.list":
					lists++
					if lists == 1 {
						return worktreetest.FakeReply{Result: worktreetest.WorkspaceListReply(t)}
					}
					latest := worktreetest.ReadProject(t, home, "repo-live")
					switch control {
					case "same desired":
						latest.RepoWorkspaceID = "w7"
					case "physical alias":
						latest.Path = alias
					case "unrelated field":
						latest.Base = "release"
					}
					worktreetest.WriteProjectAt(t, home, "repo-live", latest)
					return worktreetest.FakeReply{Result: worktreetest.WorkspaceListReply(t, worktreetest.LinkedWorkspace("w2", path, repo), worktreetest.RepositoryWorkspace("w7", repo))}
				case "worktree.create":
					worktreetest.AddLinkedWorktree(t, repo, path, branch, "main")
					return worktreetest.FakeReply{Result: worktreetest.CreateReply(t, "w2", "w2:t1", "w2:p1", path, branch, repo)}
				}
				return worktreetest.FakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
			})

			envelope := Run(context.Background(), home, Options{Task: "1-live-demo"})
			if envelope.Outcome != "ok" || envelope.Phase != "complete" {
				t.Fatalf("Create() = %+v, want ok complete", envelope)
			}
			got := worktreetest.ReadProject(t, home, "repo-live")
			if got.RepoWorkspaceID != "w7" {
				t.Fatalf("repo_workspace_id = %q, want w7", got.RepoWorkspaceID)
			}
			switch control {
			case "physical alias":
				if got.Path != alias {
					t.Fatalf("project path = %q, want the current physical alias %q", got.Path, alias)
				}
			case "unrelated field":
				if got.Base != "release" {
					t.Fatalf("project base = %q, want the concurrent release", got.Base)
				}
			default:
				if got.Path != repo {
					t.Fatalf("project path = %q, want %q", got.Path, repo)
				}
			}
		})
	}
}
