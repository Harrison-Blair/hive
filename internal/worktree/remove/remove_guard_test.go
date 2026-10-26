package remove

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
	"github.com/Harrison-Blair/hive/internal/worktree/worktreetest"
)

func writeGuardAgent(t *testing.T, home string, record records.Agent) {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatalf("store open: %v", err)
	}
	if err := registry.WriteIn([]string{"agents"}, record.Key, record); err != nil {
		t.Fatalf("write agent: %v", err)
	}
}

func guardFixture(t *testing.T) (string, string, string) {
	t.Helper()
	repo := worktreetest.NewTestRepo(t)
	home := worktreetest.NewTestHome(t, "repo-live")
	path := filepath.Join(home, "projects", "repo-live", "worktrees", "1-live-demo")
	worktreetest.AddLinkedWorktree(t, repo, path, "task/1-live-demo", "main")
	worktreetest.WriteProject(t, home, worktreetest.Project(repo))
	task := worktreetest.WorktreeTask(path, "task/1-live-demo")
	task.Status = "verified"
	worktreetest.WriteTask(t, home, task)
	return repo, home, path
}

func guardHerdr(_ *testing.T, _, _ string) func(worktreetest.FakeRequest) worktreetest.FakeReply {
	return func(worktreetest.FakeRequest) worktreetest.FakeReply {
		return worktreetest.FakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
	}
}

func TestRemoveAgentGuard(t *testing.T) {
	cases := []struct {
		name string
		seed func(t *testing.T, home, path string)
		fake func(t *testing.T, repo, path string) func(worktreetest.FakeRequest) worktreetest.FakeReply
		want string
	}{
		{
			name: "record of the task",
			seed: func(t *testing.T, home, path string) {
				writeGuardAgent(t, home, records.Agent{Key: "worker-3", Name: "worker-3", Profile: records.ProfileBuilder, Task: "1-live-demo", Status: records.AgentLive, Cwd: "/tmp/elsewhere"})
			},
			fake: guardHerdr,
			want: "agent live: worker-3",
		},
		{
			name: "record with the workspace",
			seed: func(t *testing.T, home, path string) {
				writeGuardAgent(t, home, records.Agent{Key: "worker-4", Name: "worker-4", Profile: records.ProfileBuilder, Task: "9-other", WorkspaceID: "w2", Status: records.AgentUnresolved, Cwd: "/tmp/elsewhere"})
			},
			fake: guardHerdr,
			want: "agent unresolved: worker-4",
		},
		{
			name: "record with the path",
			seed: func(t *testing.T, home, path string) {
				writeGuardAgent(t, home, records.Agent{Key: "worker-5", Name: "worker-5", Profile: records.ProfileBuilder, Task: "9-other", Cwd: path, Status: records.AgentLive})
			},
			fake: guardHerdr,
			want: "agent live: worker-5",
		},
		{
			name: "live herdr agent in the workspace",
			seed: func(t *testing.T, home, path string) {},
			fake: func(t *testing.T, repo, path string) func(worktreetest.FakeRequest) worktreetest.FakeReply {
				return func(request worktreetest.FakeRequest) worktreetest.FakeReply {
					switch request.Method {
					case "workspace.get":
						return worktreetest.FakeReply{Result: worktreetest.WorkspaceInfoReply(t, worktreetest.LinkedWorkspace("w2", path, repo))}
					case "agent.list":
						return worktreetest.FakeReply{Result: `{"type":"agent_list","agents":[{"pane_id":"w2:p9","terminal_id":"term_x","workspace_id":"w2","tab_id":"w2:t1","agent_status":"idle","focused":false,"revision":1}]}`}
					}
					return worktreetest.FakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
				}
			},
			want: "agent in workspace: w2:p9",
		},
	}

	for _, tc := range cases {
		for _, force := range []bool{false, true} {
			name := fmt.Sprintf("%s force=%t", tc.name, force)
			t.Run(name, func(t *testing.T) {
				repo, home, path := guardFixture(t)
				tc.seed(t, home, path)
				worktreetest.StartFakeHerdr(t, tc.fake(t, repo, path))
				envelope := Run(context.Background(), home, Options{Task: "1-live-demo", Force: force})
				if envelope.Outcome != "failed" || fmt.Sprint(envelope.Error) != tc.want {
					t.Fatalf("Remove() = %+v, want %q", envelope, tc.want)
				}
				if _, err := os.Lstat(path); err != nil {
					t.Fatalf("worktree removed by the guard: %v", err)
				}
			})
		}
	}

	t.Run("stopped and exited records are allowed", func(t *testing.T) {
		repo, home, path := guardFixture(t)
		writeGuardAgent(t, home, records.Agent{Key: "worker-6", Name: "worker-6", Profile: records.ProfileBuilder, Task: "1-live-demo", Status: records.AgentStopped, Cwd: path})
		writeGuardAgent(t, home, records.Agent{Key: "worker-7", Name: "worker-7", Profile: records.ProfileBuilder, Task: "1-live-demo", Status: records.AgentExited, WorkspaceID: "w2"})
		worktreetest.StartFakeHerdr(t, func(request worktreetest.FakeRequest) worktreetest.FakeReply {
			switch request.Method {
			case "workspace.get":
				return worktreetest.FakeReply{Result: worktreetest.WorkspaceInfoReply(t, worktreetest.LinkedWorkspace("w2", path, repo))}
			case "workspace.close":
				return worktreetest.FakeReply{Result: `{"type":"ok"}`}
			case "agent.list":
				return worktreetest.FakeReply{Result: `{"type":"agent_list","agents":[]}`}
			}
			return worktreetest.FakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
		})
		envelope := Run(context.Background(), home, Options{Task: "1-live-demo"})
		if envelope.Outcome != "ok" {
			t.Fatalf("Remove() = %+v, want ok", envelope)
		}
		if _, err := os.Lstat(path); !os.IsNotExist(err) {
			t.Fatalf("worktree still exists: %v", err)
		}
	})
}
