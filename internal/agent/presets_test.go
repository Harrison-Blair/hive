package agent

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/records"
)

// readPresets returns the presets document of a home.
func readPresets(t *testing.T, home string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(home, "overrides", "presets.md"))
	if err != nil {
		t.Fatalf("presets file: %v", err)
	}
	return string(data)
}

func TestPresetsFirstWrite(t *testing.T) {
	home := testHome(t)
	if message := writePresets(home, nil); message != "" {
		t.Fatalf("writePresets = %q; want success", message)
	}
	text := readPresets(t, home)
	for _, heading := range []string{"## strong", "## decent", "## small"} {
		if !strings.Contains(text, heading) {
			t.Errorf("presets file does not hold %q:\n%s", heading, text)
		}
	}
	if text != presetsTemplate {
		t.Fatalf("presets file differs from the template:\n%s", text)
	}
}

func TestPresetsExistingFileKeepsEachByte(t *testing.T) {
	home := testHome(t)
	dir := filepath.Join(home, "overrides")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "presets.md"), []byte("my presets\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if message := writePresets(home, nil); message != "" {
		t.Fatalf("writePresets = %q; want success", message)
	}
	if got := readPresets(t, home); got != "my presets\n" {
		t.Fatalf("presets file = %q; want the changed bytes", got)
	}
}

func TestPresetsConcurrentFirstWrites(t *testing.T) {
	home := testHome(t)
	results := make([]string, 2)
	var wait sync.WaitGroup
	for index := range results {
		wait.Add(1)
		go func(index int) {
			defer wait.Done()
			results[index] = writePresets(home, nil)
		}(index)
	}
	wait.Wait()
	for index, message := range results {
		if message != "" {
			t.Errorf("writePresets %d = %q; want success", index, message)
		}
	}
	if got := readPresets(t, home); got != presetsTemplate {
		t.Fatalf("presets file differs from the template")
	}
	entries, err := os.ReadDir(filepath.Join(home, "overrides"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("overrides holds %d entries; want 1", len(entries))
	}
}

func TestPresetsFailedFirstWriteIsNotSuccess(t *testing.T) {
	home := testHome(t)
	blocked := filepath.Join(home, "blocked")
	if err := os.WriteFile(blocked, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if message := createPresetsFile(filepath.Join(blocked, "overrides"), filepath.Join(blocked, "overrides", "presets.md")); message == "" {
		t.Fatalf("createPresetsFile succeeded below a regular file")
	}
	if message := writePresets(home, nil); message != "" {
		t.Fatalf("writePresets = %q; want success", message)
	}
	if got := readPresets(t, home); got != presetsTemplate {
		t.Fatalf("presets file differs from the template")
	}
}

func TestPresetsPartialTempFileIsNotSuccess(t *testing.T) {
	home := testHome(t)
	dir := filepath.Join(home, "overrides")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".presets.tmp-partial"), []byte("partial"), 0o644); err != nil {
		t.Fatal(err)
	}
	if message := writePresets(home, nil); message != "" {
		t.Fatalf("writePresets = %q; want success", message)
	}
	if got := readPresets(t, home); got != presetsTemplate {
		t.Fatalf("presets file differs from the template")
	}
}

func TestPresetsStaticPath(t *testing.T) {
	cases := []struct {
		name    string
		prepare func(t *testing.T, home, outside string)
		want    string
	}{
		{
			name: "overrides directory is a symbolic link",
			prepare: func(t *testing.T, home, outside string) {
				if err := os.Symlink(outside, filepath.Join(home, "overrides")); err != nil {
					t.Fatal(err)
				}
			},
			want: "symbolic link",
		},
		{
			name: "overrides directory is a regular file",
			prepare: func(t *testing.T, home, outside string) {
				if err := os.WriteFile(filepath.Join(home, "overrides"), []byte("x"), 0o644); err != nil {
					t.Fatal(err)
				}
			},
			want: "not a directory",
		},
		{
			name: "destination is a symbolic link",
			prepare: func(t *testing.T, home, outside string) {
				dir := filepath.Join(home, "overrides")
				if err := os.MkdirAll(dir, 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(filepath.Join(outside, "target.md"), filepath.Join(dir, "presets.md")); err != nil {
					t.Fatal(err)
				}
			},
			want: "symbolic link",
		},
		{
			name: "destination is a directory",
			prepare: func(t *testing.T, home, outside string) {
				if err := os.MkdirAll(filepath.Join(home, "overrides", "presets.md"), 0o755); err != nil {
					t.Fatal(err)
				}
			},
			want: "not a regular file",
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			home := testHome(t)
			outside := t.TempDir()
			test.prepare(t, home, outside)
			message := writePresets(home, nil)
			if message == "" {
				t.Fatalf("writePresets succeeded; want a refusal")
			}
			if !strings.Contains(message, test.want) {
				t.Fatalf("error = %q; want %q", message, test.want)
			}
			if _, err := os.Stat(filepath.Join(outside, "presets.md")); !os.IsNotExist(err) {
				t.Fatalf("wrote the presets file outside the home")
			}
		})
	}
}

func TestPresetsWrittenAtManagerSpawn(t *testing.T) {
	home := testHome(t)
	dir := projectDirectory(t, home, "demo")
	writeProjectWorkspaceRecord(t, home, "demo", nil)
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	fake.result("worktree.list", managerWorktreeListJSON(""))
	fake.result("workspace.list", managerWorkspaceListJSON())
	fake.result("workspace.create", managerWorkspaceCreateJSON("w6", "hive:demo", dir, "w6:p1"))
	scriptManagerAgent(t, fake, sleep, "w6:p1", "w6", "manager-demo-g1", dir)
	envelope := Spawn(context.Background(), home, managerSpawnOptions("demo"))
	if envelope.Outcome != "ok" || envelope.Error != nil {
		t.Fatalf("envelope = %#v; want ok", envelope)
	}
	text := readPresets(t, home)
	for _, heading := range []string{"## strong", "## decent", "## small"} {
		if !strings.Contains(text, heading) {
			t.Errorf("presets file does not hold %q:\n%s", heading, text)
		}
	}
}

func TestProjectInstructionsOrder(t *testing.T) {
	home := testHome(t)
	project := records.Project{ID: "demo", Path: "/tmp/demo", Base: "dev", RegisteredAt: "2026-01-01T00:00:00Z"}
	if message := writeProjectInstructions(home, project, nil); message != "" {
		t.Fatalf("writeProjectInstructions = %q; want success", message)
	}
	data, err := os.ReadFile(filepath.Join(home, "projects", "demo", "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	lines := strings.SplitN(text, "\n", 2)
	if !strings.HasPrefix(lines[0], "You are the planner of project demo.") {
		t.Fatalf("first line = %q; want the generated line", lines[0])
	}
	roleGate := strings.Index(text, "## Role gate")
	profilePath := strings.Index(text, filepath.Join(home, "profiles", "planner.md"))
	delegate := strings.Index(text, "Delegate each source change to one builder.")
	if roleGate < 0 || profilePath < 0 || delegate < 0 {
		t.Fatalf("rendered text misses a part:\n%s", text)
	}
	if roleGate >= profilePath || profilePath >= delegate {
		t.Fatalf("rendered order is role gate %d, profile path %d, delegate %d", roleGate, profilePath, delegate)
	}
}
