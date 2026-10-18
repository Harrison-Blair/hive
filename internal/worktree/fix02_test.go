package worktree

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
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

// fix02TaskRecord returns the raw stored task record for byte checks.
func fix02TaskRecord(t *testing.T, home, taskID string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(home, ".hive", "state", "tasks", taskID+".json"))
	if err != nil {
		t.Fatalf("read task record: %v", err)
	}
	return string(data)
}

func TestFix02CreateLatestRepositoryWorkspaceConflict(t *testing.T) {
	repo := newTestRepo(t)
	home := newTestHome(t, "repo-live")
	project := testProject(repo)
	writeProject(t, home, project)
	writeTask(t, home, testTask())
	path := filepath.Join(home, "projects", "repo-live", "worktrees", "1-live-demo")
	branch := "task/1-live-demo"
	lists := 0
	startFakeHerdr(t, func(request fakeRequest) fakeReply {
		switch request.Method {
		case "workspace.list":
			lists++
			if lists == 1 {
				return fakeReply{Result: workspaceListReply(t)}
			}
			latest := readProject(t, home, "repo-live")
			latest.RepoWorkspaceID = "w99"
			fixWriteProjectAt(t, home, "repo-live", latest)
			return fakeReply{Result: workspaceListReply(t, linkedWorkspace("w2", path, repo), repositoryWorkspace("w7", repo))}
		case "worktree.create":
			addLinkedWorktree(t, repo, path, branch, "main")
			return fakeReply{Result: createReply(t, "w2", "w2:t1", "w2:p1", path, branch, repo)}
		}
		return fakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
	})

	envelope := Create(context.Background(), home, CreateOptions{Task: "1-live-demo"})
	if envelope.Outcome != "partial" || envelope.Phase != "project" {
		t.Fatalf("Create() = %+v, want partial project", envelope)
	}
	for _, want := range []struct{ resource, action, path string }{
		{"workspace", "created", "w2"},
		{"worktree", "created", path},
		{"branch", "created", branch},
		{"workspace", "opened", "w7"},
	} {
		if !hasEffect(envelope, want.resource, want.action, want.path) {
			t.Fatalf("effects = %+v, want %s %s %s", envelope.Effects, want.resource, want.action, want.path)
		}
	}
	result, ok := envelope.Result.(CreateResult)
	if !ok || result.WorkspaceID != "w2" || result.RepoWorkspaceID != "w7" {
		t.Fatalf("result = %+v, want w2 and observed w7", envelope.Result)
	}
	got := readProject(t, home, "repo-live")
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
	task := readTask(t, home, "1-live-demo")
	if task.Worktree == nil || task.Worktree.WorkspaceID != "w2" || task.Worktree.Path != path || task.Worktree.Branch != branch {
		t.Fatalf("task worktree = %+v, want the successful task-write prefix", task.Worktree)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("external worktree lost: %v", err)
	}
	if got := strings.TrimSpace(runGit(t, path, "branch", "--show-current")); got != branch {
		t.Fatalf("branch = %q, want %q", got, branch)
	}
}

func TestFix02CreateLatestRepositoryOwnerConflict(t *testing.T) {
	repo := newTestRepo(t)
	home := newTestHome(t, "repo-live")
	project := testProject(repo)
	writeProject(t, home, project)
	writeTask(t, home, testTask())
	other := newTestRepo(t)
	path := filepath.Join(home, "projects", "repo-live", "worktrees", "1-live-demo")
	branch := "task/1-live-demo"
	lists := 0
	startFakeHerdr(t, func(request fakeRequest) fakeReply {
		switch request.Method {
		case "workspace.list":
			lists++
			if lists == 1 {
				return fakeReply{Result: workspaceListReply(t)}
			}
			return fakeReply{Result: workspaceListReply(t, linkedWorkspace("w2", path, repo), repositoryWorkspace("w7", repo))}
		case "worktree.create":
			addLinkedWorktree(t, repo, path, branch, "main")
			latest := readProject(t, home, "repo-live")
			latest.Path = other
			latest.Base = "release"
			fixWriteProjectAt(t, home, "repo-live", latest)
			return fakeReply{Result: createReply(t, "w2", "w2:t1", "w2:p1", path, branch, repo)}
		}
		return fakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
	})

	envelope := Create(context.Background(), home, CreateOptions{Task: "1-live-demo"})
	if envelope.Outcome != "partial" {
		t.Fatalf("Create() = %+v, want partial", envelope)
	}
	for _, want := range []struct{ resource, action, path string }{
		{"workspace", "created", "w2"},
		{"worktree", "created", path},
		{"branch", "created", branch},
		{"workspace", "opened", "w7"},
	} {
		if !hasEffect(envelope, want.resource, want.action, want.path) {
			t.Fatalf("effects = %+v, want %s %s %s", envelope.Effects, want.resource, want.action, want.path)
		}
	}
	result, ok := envelope.Result.(CreateResult)
	if !ok || result.RepoWorkspaceID != "w7" {
		t.Fatalf("result = %+v, want the observed w7", envelope.Result)
	}
	got := readProject(t, home, "repo-live")
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
	task := readTask(t, home, "1-live-demo")
	if task.Worktree == nil || task.Worktree.WorkspaceID != "w2" || task.Worktree.Path != path {
		t.Fatalf("task worktree = %+v, want the successful task-write prefix", task.Worktree)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("external worktree lost: %v", err)
	}
}

func TestFix02RemoveLatestTopologyConflict(t *testing.T) {
	repo := newTestRepo(t)
	home := newTestHome(t, "repo-live")
	path := filepath.Join(home, "projects", "repo-live", "worktrees", "1-live-demo")
	branch := "task/1-live-demo"
	addLinkedWorktree(t, repo, path, branch, "main")
	writeProject(t, home, testProject(repo))
	writeTask(t, home, worktreeTask(path, branch))
	startFakeHerdr(t, func(request fakeRequest) fakeReply {
		switch request.Method {
		case "workspace.get":
			return fakeReply{Result: workspaceInfoReply(t, linkedWorkspace("w2", path, repo))}
		case "workspace.close":
			latest := readTask(t, home, "1-live-demo")
			latest.Worktree.WorkspaceID = "w3"
			latest.Worktree.TabID = "w3:t2"
			latest.Worktree.RootPane = "w3:p4"
			fixWriteTaskAt(t, home, "1-live-demo", latest)
			return fakeReply{Result: `{"type":"ok"}`}
		}
		return fakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
	})

	envelope := Remove(context.Background(), home, RemoveOptions{Task: "1-live-demo"})
	if envelope.Outcome != "partial" || envelope.Phase != "task" {
		t.Fatalf("Remove() = %+v, want partial task", envelope)
	}
	if !hasEffect(envelope, "workspace", "closed", "w2") || !hasEffect(envelope, "worktree", "removed", path) {
		t.Fatalf("effects = %+v, want the actual w2 close and Git removal", envelope.Effects)
	}
	got := readTask(t, home, "1-live-demo")
	if got.Worktree == nil {
		t.Fatal("task worktree is nil, want the current topology")
	}
	if got.Worktree.WorkspaceID != "w3" || got.Worktree.TabID != "w3:t2" || got.Worktree.RootPane != "w3:p4" {
		t.Fatalf("worktree topology = %+v, want the current w3/w3:t2/w3:p4", got.Worktree)
	}
	if got.Worktree.Path != path || got.Worktree.Branch != branch || got.Worktree.RemovedAt != "" {
		t.Fatalf("worktree = %+v, want unchanged path, branch, and no removed time", got.Worktree)
	}
	raw := fix02TaskRecord(t, home, "1-live-demo")
	for _, want := range []string{`"workspace_id": "w3"`, `"tab_id": "w3:t2"`, `"root_pane": "w3:p4"`, `"removed_at": ""`} {
		if !strings.Contains(raw, want) {
			t.Fatalf("stored task bytes = %s, want %s", raw, want)
		}
	}
	if _, err := os.Lstat(path); !os.IsNotExist(err) {
		t.Fatalf("expected the real Git removal before the conflict, got %v", err)
	}
}

func TestFix02UnicodeDiagnosticBound(t *testing.T) {
	diagnostic := strings.Repeat("\u20ac", 3000)
	for _, remove := range []bool{false, true} {
		t.Run(map[bool]string{false: "create", true: "remove"}[remove], func(t *testing.T) {
			repo := newTestRepo(t)
			home := newTestHome(t, "repo-live")
			path := filepath.Join(home, "projects", "repo-live", "worktrees", "1-live-demo")
			branch := "task/1-live-demo"
			if remove {
				addLinkedWorktree(t, repo, path, branch, "main")
			}
			writeProject(t, home, testProject(repo))
			if remove {
				writeTask(t, home, worktreeTask(path, branch))
			} else {
				writeTask(t, home, testTask())
			}
			startFakeHerdr(t, func(request fakeRequest) fakeReply {
				switch request.Method {
				case "workspace.list":
					return fakeReply{Result: workspaceListReply(t)}
				case "workspace.get":
					return fakeReply{Result: workspaceInfoReply(t, linkedWorkspace("w2", path, repo))}
				case "worktree.create", "workspace.close":
					return fakeReply{Error: jsonText(t, map[string]any{"code": "invalid_request", "message": diagnostic})}
				}
				return fakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
			})

			var envelope output.Envelope
			if remove {
				envelope = Remove(context.Background(), home, RemoveOptions{Task: "1-live-demo"})
			} else {
				envelope = Create(context.Background(), home, CreateOptions{Task: "1-live-demo"})
			}
			if envelope.Outcome != "failed" {
				t.Fatalf("command = %+v, want failed", envelope)
			}
			outputError, ok := envelope.Error.(string)
			if !ok || len(outputError) != len(diagnostic) {
				t.Fatalf("output error = %v, want the separate full %d-byte diagnostic", envelope.Error, len(diagnostic))
			}
			found := false
			for _, event := range fixLogEvents(t, home) {
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
}

func TestFix02LogBoundKeepsValidUTF8(t *testing.T) {
	asciiExactly := strings.Repeat("a", 4096)
	if got, truncated := boundLogText(asciiExactly); got != asciiExactly || truncated {
		t.Fatalf("4096 ASCII bytes = %d truncated=%v, want the same text without truncation", len(got), truncated)
	}
	asciiOver := strings.Repeat("a", 4097)
	if got, truncated := boundLogText(asciiOver); got != strings.Repeat("a", 4096) || !truncated {
		t.Fatalf("4097 ASCII bytes = %d truncated=%v, want 4096 bytes and truncation", len(got), truncated)
	}
	multibyteExactly := strings.Repeat("\u20ac", 1365) + "A"
	if got, truncated := boundLogText(multibyteExactly); got != multibyteExactly || truncated || !utf8.ValidString(got) {
		t.Fatalf("4096-byte multibyte boundary = %d truncated=%v valid=%v, want unchanged valid text", len(got), truncated, utf8.ValidString(got))
	}
	multibyteOver := strings.Repeat("\u20ac", 1365) + "AB"
	if got, truncated := boundLogText(multibyteOver); got != strings.Repeat("\u20ac", 1365)+"A" || !truncated || !utf8.ValidString(got) {
		t.Fatalf("4097-byte multibyte boundary = %d truncated=%v valid=%v, want 4096 valid bytes", len(got), truncated, utf8.ValidString(got))
	}
	multibyteCut := strings.Repeat("\u20ac", 1366)
	if got, truncated := boundLogText(multibyteCut); got != strings.Repeat("\u20ac", 1365) || len(got) != 4095 || !truncated || !utf8.ValidString(got) {
		t.Fatalf("4098-byte multibyte input = %d truncated=%v valid=%v, want 4095 valid bytes", len(got), truncated, utf8.ValidString(got))
	}
}

func TestFix02CreateProjectMetadataControls(t *testing.T) {
	for _, control := range []string{"unchanged", "same desired", "physical alias", "unrelated field"} {
		t.Run(control, func(t *testing.T) {
			repo := newTestRepo(t)
			home := newTestHome(t, "repo-live")
			project := testProject(repo)
			writeProject(t, home, project)
			writeTask(t, home, testTask())
			path := filepath.Join(home, "projects", "repo-live", "worktrees", "1-live-demo")
			branch := "task/1-live-demo"
			alias := filepath.Join(t.TempDir(), "repo-alias")
			if err := os.Symlink(repo, alias); err != nil {
				t.Fatalf("symlink repository: %v", err)
			}
			lists := 0
			startFakeHerdr(t, func(request fakeRequest) fakeReply {
				switch request.Method {
				case "workspace.list":
					lists++
					if lists == 1 {
						return fakeReply{Result: workspaceListReply(t)}
					}
					latest := readProject(t, home, "repo-live")
					switch control {
					case "same desired":
						latest.RepoWorkspaceID = "w7"
					case "physical alias":
						latest.Path = alias
					case "unrelated field":
						latest.Base = "release"
					}
					fixWriteProjectAt(t, home, "repo-live", latest)
					return fakeReply{Result: workspaceListReply(t, linkedWorkspace("w2", path, repo), repositoryWorkspace("w7", repo))}
				case "worktree.create":
					addLinkedWorktree(t, repo, path, branch, "main")
					return fakeReply{Result: createReply(t, "w2", "w2:t1", "w2:p1", path, branch, repo)}
				}
				return fakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
			})

			envelope := Create(context.Background(), home, CreateOptions{Task: "1-live-demo"})
			if envelope.Outcome != "ok" || envelope.Phase != "complete" {
				t.Fatalf("Create() = %+v, want ok complete", envelope)
			}
			got := readProject(t, home, "repo-live")
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

func TestFix02RemoveTopologyControls(t *testing.T) {
	for _, control := range []string{"unchanged", "cleared", "unrelated fields", "physical alias"} {
		t.Run(control, func(t *testing.T) {
			repo := newTestRepo(t)
			home := newTestHome(t, "repo-live")
			path := filepath.Join(home, "projects", "repo-live", "worktrees", "1-live-demo")
			branch := "task/1-live-demo"
			addLinkedWorktree(t, repo, path, branch, "main")
			writeProject(t, home, testProject(repo))
			task := worktreeTask(path, branch)
			if control == "physical alias" {
				alias := filepath.Join(home, "worktree-alias")
				if err := os.Symlink(path, alias); err != nil {
					t.Fatalf("symlink worktree: %v", err)
				}
				task.Worktree.Path = alias
			}
			writeTask(t, home, task)
			startFakeHerdr(t, func(request fakeRequest) fakeReply {
				switch request.Method {
				case "workspace.get":
					return fakeReply{Result: workspaceInfoReply(t, linkedWorkspace("w2", path, repo))}
				case "workspace.close":
					latest := readTask(t, home, "1-live-demo")
					switch control {
					case "cleared":
						latest.Worktree.WorkspaceID = ""
						latest.Worktree.TabID = ""
						latest.Worktree.RootPane = ""
					case "unrelated fields":
						latest.Title = "concurrent title"
						latest.Status = "cancelled"
						latest.ActingAgent = "worker-b"
						latest.CancelReason = "user stopped it"
						latest.DependsOn = []string{"9-other"}
						latest.Verification = &records.Verification{Agent: "verifier-b", Result: "pass", Reason: "checked", At: "2026-10-18T02:00:00Z"}
					}
					fixWriteTaskAt(t, home, "1-live-demo", latest)
					return fakeReply{Result: `{"type":"ok"}`}
				}
				return fakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
			})

			envelope := Remove(context.Background(), home, RemoveOptions{Task: "1-live-demo"})
			if envelope.Outcome != "ok" || envelope.Phase != "complete" {
				t.Fatalf("Remove() = %+v, want ok complete", envelope)
			}
			got := readTask(t, home, "1-live-demo")
			if got.Worktree == nil || got.Worktree.RemovedAt == "" {
				t.Fatalf("worktree = %+v, want a removed time", got.Worktree)
			}
			if got.Worktree.WorkspaceID != "" || got.Worktree.TabID != "" || got.Worktree.RootPane != "" {
				t.Fatalf("worktree topology = %+v, want cleared live fields", got.Worktree)
			}
			switch control {
			case "unrelated fields":
				if got.Title != "concurrent title" || got.Status != "cancelled" || got.ActingAgent != "worker-b" || got.CancelReason != "user stopped it" {
					t.Fatalf("task = %+v, want the unrelated concurrent fields", got)
				}
				if !reflect.DeepEqual(got.DependsOn, []string{"9-other"}) || got.Verification == nil || got.Verification.Agent != "verifier-b" {
					t.Fatalf("task = %+v, want concurrent dependencies and verification", got)
				}
			}
			if _, err := os.Lstat(path); !os.IsNotExist(err) {
				t.Fatalf("physical worktree remains after remove: %v", err)
			}
		})
	}
}
