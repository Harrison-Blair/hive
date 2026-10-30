package spawn

import (
	"os/exec"
	"slices"
	"testing"

	"github.com/Harrison-Blair/hive/internal/agent"
	"github.com/Harrison-Blair/hive/internal/agent/agenttest"
)

// placements are the four places of a spawn. Each setup registers the
// project and opens workspaces, and returns the caller pane, the project, and
// the worktree of the spawn.
var placements = []struct {
	name  string
	setup func(t *testing.T, fake *agenttest.Herdr, home string) (caller, project, worktree string)
}{
	{"new tab in the caller workspace", func(t *testing.T, fake *agenttest.Herdr, home string) (string, string, string) {
		_, pane := fake.AddWorkspace("mine", "", false)
		return pane.ID, "", ""
	}},
	{"new tab in the checkout workspace", func(t *testing.T, fake *agenttest.Herdr, home string) (string, string, string) {
		repo := agenttest.GitRepo(t)
		agenttest.RegisterProject(t, home, "p1", repo, "main")
		fake.AddWorkspace("any", repo, false)
		return "", "p1", ""
	}},
	{"root pane of a new checkout workspace", func(t *testing.T, fake *agenttest.Herdr, home string) (string, string, string) {
		agenttest.RegisterProject(t, home, "p1", agenttest.GitRepo(t), "main")
		return "", "p1", ""
	}},
	{"root pane of a new worktree", func(t *testing.T, fake *agenttest.Herdr, home string) (string, string, string) {
		repo := agenttest.GitRepo(t)
		agenttest.RegisterProject(t, home, "p1", repo, "main")
		fake.AddWorkspace("any", repo, false)
		return "", "p1", "feat-x"
	}},
}

// spawnIn spawns the agent kid with harness in the placement test and returns
// its pane.
func spawnIn(t *testing.T, test int, harness string) (agenttest.Pane, string) {
	t.Helper()
	fake := agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	options := base("kid")
	options.Harness = harness
	options.CallerPane, options.Project, options.Worktree = placements[test].setup(t, fake, home)
	if envelope := spawnRun(t, home, options); envelope.Outcome != "ok" {
		t.Fatalf("envelope = %+v", envelope)
	}
	record, err := agent.Load(home, "kid")
	if err != nil {
		t.Fatal(err)
	}
	pane, _ := fake.Pane(record.Pane)
	return pane, home
}

// TestSpawnGivesTheAgentTheHome checks that each placement starts the harness
// with HIVE_HOME set to the home of the caller, so the hive commands of the
// agent use the same home.
func TestSpawnGivesTheAgentTheHome(t *testing.T) {
	for i, test := range placements {
		t.Run(test.name, func(t *testing.T) {
			pane, home := spawnIn(t, i, "pi")
			if got := pane.StartEnv["HIVE_HOME"]; got != home {
				t.Fatalf("HIVE_HOME of the agent = %q; want the home %q (start env %v)", got, home, pane.StartEnv)
			}
		})
	}
}

// TestSpawnGivesClaudeTheTrustEnv checks that each placement, the root pane
// of a new workspace or worktree too, starts Claude Code with
// CLAUDE_CODE_SANDBOXED=1, so Claude shows no trust dialog.
func TestSpawnGivesClaudeTheTrustEnv(t *testing.T) {
	for i, test := range placements {
		t.Run(test.name, func(t *testing.T) {
			pane, home := spawnIn(t, i, "claude")
			if got := pane.StartEnv["CLAUDE_CODE_SANDBOXED"]; got != "1" {
				t.Fatalf("CLAUDE_CODE_SANDBOXED of the agent = %q; want 1 (start env %v)", got, pane.StartEnv)
			}
			if got := pane.StartEnv["HIVE_HOME"]; got != home {
				t.Fatalf("HIVE_HOME of the agent = %q; want the home %q (start env %v)", got, home, pane.StartEnv)
			}
		})
	}
}

// TestSpawnLetsClaudeAcceptMessages checks that each placement starts Claude
// Code with the setting that accepts the messages of hive agent send.
func TestSpawnLetsClaudeAcceptMessages(t *testing.T) {
	for i, test := range placements {
		t.Run(test.name, func(t *testing.T) {
			pane, _ := spawnIn(t, i, "claude")
			index := slices.Index(pane.Args, "--settings")
			if index < 0 || index+1 >= len(pane.Args) || pane.Args[index+1] != `{"crossSessionInbound":"accept"}` {
				t.Fatalf("args of the agent = %q; want --settings with crossSessionInbound accept", pane.Args)
			}
		})
	}
}

// TestSpawnGivesPiTheTrustArgs checks that each placement starts Pi with
// --approve.
func TestSpawnGivesPiTheTrustArgs(t *testing.T) {
	for i, test := range placements {
		t.Run(test.name, func(t *testing.T) {
			pane, _ := spawnIn(t, i, "pi")
			if !slices.Contains(pane.Args, "--approve") {
				t.Fatalf("args of the agent = %q; want --approve", pane.Args)
			}
		})
	}
}

func TestExportEnvKeepsEachValueInAShell(t *testing.T) {
	for _, home := range []string{"/home/u/.hive", "/tmp/it's here/$HOME `x` \\n"} {
		env := map[string]string{"HIVE_HOME": home, "CLAUDE_CODE_SANDBOXED": "1"}
		out, err := exec.Command("sh", "-c", exportEnv(env)+`; printf '%s|%s' "$HIVE_HOME" "$CLAUDE_CODE_SANDBOXED"`).Output()
		if want := home + "|1"; err != nil || string(out) != want {
			t.Fatalf("shell HIVE_HOME|CLAUDE_CODE_SANDBOXED = %q, %v; want %q", out, err, want)
		}
	}
}
