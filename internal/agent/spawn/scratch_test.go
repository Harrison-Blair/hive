package spawn

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"

	"github.com/Harrison-Blair/hive/internal/agent"
	"github.com/Harrison-Blair/hive/internal/agent/agenttest"
	"github.com/Harrison-Blair/hive/internal/lib/output"
)

// namedDir makes the directory name in a new temporary directory.
func namedDir(t *testing.T, name string) string {
	t.Helper()
	parent, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(parent, name)
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}

// namedRepo makes a Git repository with one commit in the directory name.
func namedRepo(t *testing.T, name string) string {
	t.Helper()
	dir := namedDir(t, name)
	for _, args := range [][]string{
		{"init", "-q", "-b", "main"},
		{"-c", "user.name=t", "-c", "user.email=t@t", "commit", "-q", "--allow-empty", "-m", "init"},
	} {
		if out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
	return dir
}

// TestSpawnGivesTheAgentTheScratchPlace checks that each placement makes the
// scratch place of the project where the agent starts and sets HIVE_TMP to
// it. The agent of --branch starts in a linked checkout and gets the name of
// the repository.
func TestSpawnGivesTheAgentTheScratchPlace(t *testing.T) {
	for _, test := range []struct {
		name    string
		project string
		setup   func(t *testing.T, fake *agenttest.Herdr, options *Options)
	}{
		{"no placement flag", "caller-dir", func(t *testing.T, fake *agenttest.Herdr, options *Options) {
			_, pane := fake.AddWorkspace("mine", "", false)
			options.CallerPane, options.Cwd = pane.ID, namedDir(t, "Caller Dir")
		}},
		{"--branch", "steno-repo", func(t *testing.T, fake *agenttest.Herdr, options *Options) {
			options.Cwd, options.Branch = namedRepo(t, "Steno Repo"), "feat-x"
		}},
		{"--directory", "some-dir", func(t *testing.T, fake *agenttest.Herdr, options *Options) {
			options.Directory = namedDir(t, "Some Dir")
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			fake := agenttest.NewHerdr(t)
			home := agenttest.Home(t)
			options := base("kid")
			test.setup(t, fake, &options)
			envelope := spawnRun(t, home, options)
			if envelope.Outcome != "ok" {
				t.Fatalf("envelope = %+v", envelope)
			}
			record, err := agent.Load(home, "kid")
			if err != nil {
				t.Fatal(err)
			}
			pane, _ := fake.Pane(record.Pane)
			want := filepath.Join(home, "tmp", test.project)
			if got := pane.StartEnv["HIVE_TMP"]; got != want {
				t.Fatalf("HIVE_TMP of the agent = %q; want %q (start env %v)", got, want, pane.StartEnv)
			}
			if info, err := os.Stat(want); err != nil || !info.IsDir() {
				t.Fatalf("scratch place %s: %v; want a directory", want, err)
			}
			effect := output.Effect{Resource: "directory", Path: want, Action: "created"}
			if !slices.Contains(envelope.Effects, effect) {
				t.Fatalf("effects = %+v; want %+v", envelope.Effects, effect)
			}
		})
	}
}

// TestSpawnFailsWhenTheScratchPlaceCannotBeMade checks that spawn fails with
// the cause and makes no herdr request and no record when the tmp directory
// of the home is a file.
func TestSpawnFailsWhenTheScratchPlaceCannotBeMade(t *testing.T) {
	fake := agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	if err := os.WriteFile(filepath.Join(home, "tmp"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	options := base("kid")
	options.Directory = namedDir(t, "some-dir")
	envelope := spawnRun(t, home, options)
	if envelope.Outcome != "failed" || envelope.Phase != "scratch" {
		t.Fatalf("envelope = %+v; want failed in phase scratch", envelope)
	}
	if count := fake.RequestCount(); count != 0 {
		t.Fatalf("herdr requests = %v; want none", fake.Methods())
	}
	if _, err := agent.Load(home, "kid"); err == nil {
		t.Fatal("record of kid exists; want none")
	}
}
