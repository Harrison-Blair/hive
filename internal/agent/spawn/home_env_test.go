package spawn

import (
	"os/exec"
	"testing"

	"github.com/Harrison-Blair/hive/internal/agent"
	"github.com/Harrison-Blair/hive/internal/agent/agenttest"
)

// TestSpawnGivesTheAgentTheHome checks that each placement starts the harness
// with HIVE_HOME set to the home of the caller, so the hive commands of the
// agent use the same home.
func TestSpawnGivesTheAgentTheHome(t *testing.T) {
	for _, test := range []struct {
		name string
		// setup registers the project and opens workspaces, and returns the
		// caller pane, the project, and the worktree of the spawn.
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
	} {
		t.Run(test.name, func(t *testing.T) {
			fake := agenttest.NewHerdr(t)
			home := agenttest.Home(t)
			options := base("kid")
			options.CallerPane, options.Project, options.Worktree = test.setup(t, fake, home)
			if envelope := spawnRun(t, home, options); envelope.Outcome != "ok" {
				t.Fatalf("envelope = %+v", envelope)
			}
			record, err := agent.Load(home, "kid")
			if err != nil {
				t.Fatal(err)
			}
			pane, _ := fake.Pane(record.Pane)
			if got := pane.StartEnv["HIVE_HOME"]; got != home {
				t.Fatalf("HIVE_HOME of the agent = %q; want the home %q (start env %v)", got, home, pane.StartEnv)
			}
		})
	}
}

func TestExportHomeKeepsThePathInAShell(t *testing.T) {
	for _, home := range []string{"/home/u/.hive", "/tmp/it's here/$HOME `x` \\n"} {
		out, err := exec.Command("sh", "-c", exportHome(home)+`; printf %s "$HIVE_HOME"`).Output()
		if err != nil || string(out) != home {
			t.Fatalf("shell HIVE_HOME = %q, %v; want %q", out, err, home)
		}
	}
}
