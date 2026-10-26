package remove

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/worktree/worktreetest"
)

// fix02TaskRecord returns the raw stored task record for byte checks.
func fix02TaskRecord(t *testing.T, home, taskID string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(home, ".hive", "state", "tasks", taskID+".json"))
	if err != nil {
		t.Fatalf("read task record: %v", err)
	}
	return string(data)
}

func TestFix02RemoveLatestTopologyConflict(t *testing.T) {
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
			latest := worktreetest.ReadTask(t, home, "1-live-demo")
			latest.Worktree.WorkspaceID = "w3"
			latest.Worktree.TabID = "w3:t2"
			latest.Worktree.RootPane = "w3:p4"
			worktreetest.WriteTaskAt(t, home, "1-live-demo", latest)
			return worktreetest.FakeReply{Result: `{"type":"ok"}`}
		}
		return worktreetest.FakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
	})

	envelope := Run(context.Background(), home, Options{Task: "1-live-demo"})
	if envelope.Outcome != "partial" || envelope.Phase != "task" {
		t.Fatalf("Remove() = %+v, want partial task", envelope)
	}
	if !worktreetest.HasEffect(envelope, "workspace", "closed", "w2") || !worktreetest.HasEffect(envelope, "worktree", "removed", path) {
		t.Fatalf("effects = %+v, want the actual w2 close and Git removal", envelope.Effects)
	}
	got := worktreetest.ReadTask(t, home, "1-live-demo")
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
	t.Run("remove", func(t *testing.T) {
		repo := worktreetest.NewTestRepo(t)
		home := worktreetest.NewTestHome(t, "repo-live")
		path := filepath.Join(home, "projects", "repo-live", "worktrees", "1-live-demo")
		branch := "task/1-live-demo"
		worktreetest.AddLinkedWorktree(t, repo, path, branch, "main")
		worktreetest.WriteProject(t, home, worktreetest.Project(repo))
		worktreetest.WriteTask(t, home, worktreetest.WorktreeTask(path, branch))
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

func TestFix02RemoveTopologyControls(t *testing.T) {
	for _, control := range []string{"unchanged", "cleared", "unrelated fields", "physical alias"} {
		t.Run(control, func(t *testing.T) {
			repo := worktreetest.NewTestRepo(t)
			home := worktreetest.NewTestHome(t, "repo-live")
			path := filepath.Join(home, "projects", "repo-live", "worktrees", "1-live-demo")
			branch := "task/1-live-demo"
			worktreetest.AddLinkedWorktree(t, repo, path, branch, "main")
			worktreetest.WriteProject(t, home, worktreetest.Project(repo))
			task := worktreetest.WorktreeTask(path, branch)
			if control == "physical alias" {
				alias := filepath.Join(home, "worktree-alias")
				if err := os.Symlink(path, alias); err != nil {
					t.Fatalf("symlink worktree: %v", err)
				}
				task.Worktree.Path = alias
			}
			worktreetest.WriteTask(t, home, task)
			worktreetest.StartFakeHerdr(t, func(request worktreetest.FakeRequest) worktreetest.FakeReply {
				switch request.Method {
				case "workspace.get":
					return worktreetest.FakeReply{Result: worktreetest.WorkspaceInfoReply(t, worktreetest.LinkedWorkspace("w2", path, repo))}
				case "agent.list":
					return worktreetest.FakeReply{Result: `{"type":"agent_list","agents":[]}`}
				case "workspace.close":
					latest := worktreetest.ReadTask(t, home, "1-live-demo")
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
					worktreetest.WriteTaskAt(t, home, "1-live-demo", latest)
					return worktreetest.FakeReply{Result: `{"type":"ok"}`}
				}
				return worktreetest.FakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
			})

			envelope := Run(context.Background(), home, Options{Task: "1-live-demo"})
			if envelope.Outcome != "ok" || envelope.Phase != "complete" {
				t.Fatalf("Remove() = %+v, want ok complete", envelope)
			}
			got := worktreetest.ReadTask(t, home, "1-live-demo")
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
