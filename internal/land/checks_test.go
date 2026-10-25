package land

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// newLandHerdr serves the herdr ping and pane.get calls of one pane in a
// workspace. It is an external dependency of the user rule of land.
func newLandHerdr(t *testing.T, workspaceID string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "herdr.sock")
	listener, err := net.Listen("unix", path)
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("HERDR_SOCKET_PATH", path)
	t.Cleanup(func() {
		_ = listener.Close()
		_ = os.Remove(path)
	})
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go func(conn net.Conn) {
				defer func() { _ = conn.Close() }()
				line, err := bufio.NewReader(conn).ReadBytes('\n')
				if err != nil {
					return
				}
				var request struct {
					ID     string          `json:"id"`
					Method string          `json:"method"`
					Params json.RawMessage `json:"params"`
				}
				if err := json.Unmarshal(line, &request); err != nil {
					t.Errorf("decode herdr request: %v", err)
					return
				}
				var result string
				switch request.Method {
				case "ping":
					result = `{"type":"pong","version":"0.9.5","protocol":22}`
				case "pane.get":
					var params struct {
						PaneID string `json:"pane_id"`
					}
					if err := json.Unmarshal(request.Params, &params); err != nil {
						t.Errorf("decode pane.get params: %v", err)
						return
					}
					result = fmt.Sprintf(`{"type":"pane_info","pane":{"pane_id":%q,"terminal_id":"term_user","workspace_id":%q,"tab_id":"t1","focused":false,"agent_status":"idle","revision":1}}`, params.PaneID, workspaceID)
				default:
					t.Errorf("unexpected herdr method %q", request.Method)
					return
				}
				_, _ = fmt.Fprintf(conn, `{"id":%q,"result":%s}`+"\n", request.ID, result)
			}(conn)
		}
	}()
}

func writeLandAgent(t *testing.T, home string, record records.Agent) {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.WriteIn([]string{"agents"}, record.Key, record); err != nil {
		t.Fatal(err)
	}
}

func writeLandClaim(t *testing.T, home string, claim records.Claim) {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.WriteIn([]string{"claims"}, claim.Key, claim); err != nil {
		t.Fatal(err)
	}
}

// seedLandCallers writes the owner, the coordinator, and one bound builder.
func seedLandCallers(t *testing.T, home string) {
	t.Helper()
	writeLandAgent(t, home, records.Agent{
		Key: "owner-repo-land-1", Name: "owner-repo-land", Profile: records.ProfilePlanner,
		Owns: "repo-land", Project: "repo-land", Status: records.AgentLive,
		PaneID: "w0:p1", WorkspaceID: "w0", Generation: 1,
	})
	writeLandClaim(t, home, records.Claim{Key: "owner-repo-land", Agent: "owner-repo-land-1", Generation: 1})
	writeLandAgent(t, home, records.Agent{
		Key: "coordinator-g1", Name: "coordinator", Kind: records.AgentKindCoordinator,
		Profile: records.ProfileCoordinator, Status: records.AgentLive,
		PaneID: "w0:p2", WorkspaceID: "w0", Generation: 1,
	})
	writeLandClaim(t, home, records.Claim{Key: "coordinator", Agent: "coordinator-g1", Generation: 1})
	writeLandAgent(t, home, records.Agent{
		Key: "builder-1", Name: "builder-1", Profile: records.ProfileBuilder,
		Project: "repo-land", Task: "1-a", Status: records.AgentLive,
		PaneID: "w2:p1", WorkspaceID: "w2", Generation: 1,
	})
}

// seedLandRepo makes a home and a repository with one verified task branch.
func seedLandRepo(t *testing.T) (string, string) {
	t.Helper()
	repo := newRepo(t)
	bare := newBare(t)
	runGit(t, repo, "remote", "add", "origin", bare)
	runGit(t, repo, "push", "-q", "-u", "origin", "main")
	addTaskBranch(t, repo, "task/1-a", "a.txt", "a\n")
	home := newHome(t)
	writeProject(t, home, testProject(repo))
	writeTask(t, home, testTask("1-a", "task/1-a"))
	return home, repo
}

// TestLandRefusesWithoutGrant proves the exact D8 grant refusal and that it
// has no effect.
func TestLandRefusesWithoutGrant(t *testing.T) {
	home, repo := seedLandRepo(t)
	tip := strings.TrimSpace(runGit(t, repo, "rev-parse", "main"))
	envelope := Land(context.Background(), home, Options{Task: "1-a", Mode: "merge"})
	if envelope.Outcome != "failed" || envelope.Error != "land refused: pass --grant" {
		t.Fatalf("Land() = %+v; want the exact grant refusal", envelope)
	}
	if got := strings.TrimSpace(runGit(t, repo, "rev-parse", "main")); got != tip {
		t.Fatalf("main = %s; want unchanged tip %s", got, tip)
	}
	if got := loadTask(t, home, "1-a"); got.Delivery != nil {
		t.Fatalf("delivery = %+v; want no delivery effect", got.Delivery)
	}
}

// TestLandCallerRules proves the owner, coordinator, other agent, unknown
// caller, and builder pane rules of D8 with their exact texts.
func TestLandCallerRules(t *testing.T) {
	cases := []struct {
		name    string
		options Options
		want    string
	}{
		{"owner", Options{Agent: "owner-repo-land"}, ""},
		{"coordinator", Options{Agent: "coordinator"}, ""},
		{"other agent", Options{Agent: "builder-1"}, "land refused: builder-1 is not the owner of project repo-land or the coordinator"},
		{"unknown caller", Options{Agent: "ghost"}, "caller unknown: pass --agent"},
		{"builder pane", Options{Pane: "w2:p1"}, "land refused: builder-1 is not the owner of project repo-land or the coordinator"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			home, repo := seedLandRepo(t)
			seedLandCallers(t, home)
			tip := strings.TrimSpace(runGit(t, repo, "rev-parse", "main"))
			options := test.options
			options.Task = "1-a"
			options.Mode = "merge"
			options.Grant = "test"
			envelope := Land(context.Background(), home, options)
			if test.want == "" {
				if envelope.Outcome != "ok" || envelope.Error != nil {
					t.Fatalf("Land() = %+v; want ok", envelope)
				}
				if got := loadTask(t, home, "1-a"); got.Delivery == nil || !got.Delivery.Pushed {
					t.Fatalf("delivery = %+v; want a pushed delivery", got.Delivery)
				}
				return
			}
			if envelope.Outcome != "failed" || envelope.Error != test.want {
				t.Fatalf("Land() = %+v; want %q", envelope, test.want)
			}
			if got := strings.TrimSpace(runGit(t, repo, "rev-parse", "main")); got != tip {
				t.Fatalf("main = %s; want unchanged tip %s", got, tip)
			}
			if got := loadTask(t, home, "1-a"); got.Delivery != nil {
				t.Fatalf("delivery = %+v; want no delivery effect", got.Delivery)
			}
		})
	}
}

// TestLandAllowsTheUserShell proves AC5 for a caller with no agent record and
// no pane.
func TestLandAllowsTheUserShell(t *testing.T) {
	home, _ := seedLandRepo(t)
	envelope := Land(context.Background(), home, Options{Task: "1-a", Mode: "merge", Grant: "test"})
	if envelope.Outcome != "ok" || envelope.Error != nil {
		t.Fatalf("Land() = %+v; want ok for the user shell", envelope)
	}
}

// TestLandUserRuleRefusesBuilderWorkspace proves AC5: a pane in the workspace
// of a live builder fails with the rule 3 text.
func TestLandUserRuleRefusesBuilderWorkspace(t *testing.T) {
	home, repo := seedLandRepo(t)
	writeLandAgent(t, home, records.Agent{
		Key: "builder-1", Name: "builder-1", Profile: records.ProfileBuilder,
		Project: "repo-land", Task: "1-a", Status: records.AgentLive,
		PaneID: "w2:p1", WorkspaceID: "w2", Generation: 1,
	})
	newLandHerdr(t, "w2")
	tip := strings.TrimSpace(runGit(t, repo, "rev-parse", "main"))
	envelope := Land(context.Background(), home, Options{Task: "1-a", Mode: "merge", Grant: "test", Pane: "w2:p9"})
	want := "land refused: the caller is not the user: chat send is for the user only: only the user can send chat; rule 3 (caller pane shares the workspace of the live agent builder-1)"
	if envelope.Outcome != "failed" || envelope.Error != want {
		t.Fatalf("Land() = %+v; want %q", envelope, want)
	}
	if got := strings.TrimSpace(runGit(t, repo, "rev-parse", "main")); got != tip {
		t.Fatalf("main = %s; want unchanged tip %s", got, tip)
	}
	if got := loadTask(t, home, "1-a"); got.Delivery != nil {
		t.Fatalf("delivery = %+v; want no delivery effect", got.Delivery)
	}
}

// TestLandUserRuleRefusesWorktreeDirectory proves that the user rule of land
// uses the chat working directory rule.
func TestLandUserRuleRefusesWorktreeDirectory(t *testing.T) {
	home, _ := seedLandRepo(t)
	workingDir := filepath.Join(home, "projects", "repo-land")
	if err := os.MkdirAll(workingDir, 0o755); err != nil {
		t.Fatal(err)
	}
	envelope := Land(context.Background(), home, Options{Task: "1-a", Mode: "merge", Grant: "test", WorkingDir: workingDir})
	want := "land refused: the caller is not the user: chat send is for the user only: only the user can send chat; rule 2 (working directory is below projects/)"
	if envelope.Outcome != "failed" || envelope.Error != want {
		t.Fatalf("Land() = %+v; want %q", envelope, want)
	}
}
