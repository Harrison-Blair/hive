package agent

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstructionsWritten(t *testing.T) {
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
	path := filepath.Join(home, "projects", "demo", "AGENTS.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("instruction file: %v", err)
	}
	text := string(data)
	if !strings.HasPrefix(text, "You are the planner of project demo.") {
		t.Fatalf("first line = %q; want the generated line", strings.SplitN(text, "\n", 2)[0])
	}
	for _, want := range []string{"demo", home, "/tmp/demo"} {
		if !strings.Contains(text, want) {
			t.Fatalf("file does not hold %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, "{{") {
		t.Fatalf("file holds an unfilled placeholder:\n%s", text)
	}
}

func TestInstructionsReplacedOnEachSpawn(t *testing.T) {
	home := testHome(t)
	dir := projectDirectory(t, home, "demo")
	writeProjectWorkspaceRecord(t, home, "demo", nil)
	path := filepath.Join(home, "projects", "demo", "AGENTS.md")
	if err := os.WriteFile(path, []byte("old text that hive replaces\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	sleep := startSleep(t)
	fake := newFakeHerdr(t)
	fake.result("worktree.list", managerWorktreeListJSON(""))
	fake.result("workspace.list", managerWorkspaceListJSON())
	fake.result("workspace.create", managerWorkspaceCreateJSON("w6", "hive:demo", dir, "w6:p1"))
	scriptManagerAgent(t, fake, sleep, "w6:p1", "w6", "manager-demo-g1", dir)
	if envelope := Spawn(context.Background(), home, managerSpawnOptions("demo")); envelope.Outcome != "ok" {
		t.Fatalf("envelope = %#v; want ok", envelope)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "old text") {
		t.Fatalf("file was not replaced:\n%s", data)
	}
}

func TestInstructionsStaticPath(t *testing.T) {
	cases := []struct {
		name    string
		prepare func(t *testing.T, home, outside string)
		want    string
	}{
		{
			name: "project directory is a symbolic link",
			prepare: func(t *testing.T, home, outside string) {
				if err := os.MkdirAll(filepath.Join(home, "projects"), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(outside, filepath.Join(home, "projects", "demo")); err != nil {
					t.Fatal(err)
				}
			},
			want: "symbolic link",
		},
		{
			name: "destination is a directory",
			prepare: func(t *testing.T, home, outside string) {
				if err := os.MkdirAll(filepath.Join(home, "projects", "demo", "AGENTS.md"), 0o755); err != nil {
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
			writeProjectWorkspaceRecord(t, home, "demo", nil)
			dir, err := filepath.EvalSymlinks(filepath.Join(home, "projects", "demo"))
			if err != nil {
				t.Fatal(err)
			}
			fake := newFakeHerdr(t)
			fake.result("worktree.list", managerWorktreeListJSON(""))
			fake.result("workspace.list", managerWorkspaceListJSON())
			fake.result("workspace.create", managerWorkspaceCreateJSON("w6", "hive:demo", dir, "w6:p1"))
			fake.result("workspace.close", `{"type":"ok"}`)
			envelope := Spawn(context.Background(), home, managerSpawnOptions("demo"))
			if envelope.Outcome != "failed" {
				t.Fatalf("envelope = %#v; want failed", envelope)
			}
			if !strings.Contains(toString(envelope.Error), test.want) {
				t.Fatalf("error = %v; want %q", envelope.Error, test.want)
			}
			if _, err := os.Stat(filepath.Join(outside, "AGENTS.md")); !os.IsNotExist(err) {
				t.Fatalf("wrote the instruction file outside the home")
			}
			if len(fake.recorded("agent.start")) != 0 {
				t.Fatalf("static path failure started an agent")
			}
		})
	}
}
