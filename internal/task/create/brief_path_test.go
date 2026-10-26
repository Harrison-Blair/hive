package create

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/task/tasktest"
)

func TestCreateRejectsLinkedBriefAncestor(t *testing.T) {
	home := tasktest.NewHome(t)
	tasktest.RegisterProject(t, home, "one")
	external := t.TempDir()
	briefs := tasktest.BriefAncestorPath(home)
	if err := os.MkdirAll(filepath.Dir(briefs), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(external, briefs); err != nil {
		t.Fatal(err)
	}

	envelope := Run(home, Options{Project: "one", Title: "Seed"})
	if envelope.Outcome != "failed" || envelope.Error == nil {
		t.Fatalf("Create() = %#v; want a failed outcome", envelope)
	}
	if message, ok := envelope.Error.(string); !ok || !strings.Contains(message, "brief path invalid") {
		t.Fatalf("Create() error = %v; want a brief path error", envelope.Error)
	}
	entries, err := os.ReadDir(external)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("external directory holds %d entries; want none", len(entries))
	}
	if _, ok := counterValue(t, home); ok {
		t.Fatal("linked brief ancestor changed the counter")
	}
	assertNoTaskRecord(t, home, "1-seed")
}

func TestCreateRejectsWrongTypeBriefAncestor(t *testing.T) {
	home := tasktest.NewHome(t)
	tasktest.RegisterProject(t, home, "one")
	briefs := tasktest.BriefAncestorPath(home)
	if err := os.MkdirAll(filepath.Dir(briefs), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(briefs, []byte("not a directory"), 0o644); err != nil {
		t.Fatal(err)
	}

	envelope := Run(home, Options{Project: "one", Title: "Seed"})
	if envelope.Outcome != "failed" || envelope.Error == nil {
		t.Fatalf("Create() = %#v; want a failed outcome", envelope)
	}
	if _, ok := counterValue(t, home); ok {
		t.Fatal("wrong-type brief ancestor changed the counter")
	}
	assertNoTaskRecord(t, home, "1-seed")
}

func TestCreateKeepsUserBriefBelowLinkedDirectory(t *testing.T) {
	home := tasktest.NewHome(t)
	tasktest.RegisterProject(t, home, "one")
	target := t.TempDir()
	const userBrief = "user selected brief"
	if err := os.WriteFile(filepath.Join(target, "brief.md"), []byte(userBrief), 0o644); err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(t.TempDir(), "alias")
	if err := os.Symlink(target, alias); err != nil {
		t.Fatal(err)
	}

	task := createOK(t, home, Options{
		Project:   "one",
		Title:     "Step",
		BriefPath: filepath.Join(alias, "brief.md"),
		BriefSet:  true,
	})
	data, err := os.ReadFile(filepath.Join(tasktest.BriefAncestorPath(home), task.ID+".md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != userBrief {
		t.Fatalf("stored brief = %q; want the user-selected brief", data)
	}
}

func TestCreateAcceptsSymlinkedHomeRoot(t *testing.T) {
	home := tasktest.NewHome(t)
	tasktest.RegisterProject(t, home, "one")
	alias := filepath.Join(filepath.Dir(home), "alias")
	if err := os.Symlink(home, alias); err != nil {
		t.Fatal(err)
	}

	envelope := Run(alias, Options{Project: "one", Title: "Step"})
	if envelope.Outcome != "ok" || envelope.Error != nil {
		t.Fatalf("Create() with a symlinked home root = %#v; want ok", envelope)
	}
	task, ok := envelope.Result.(records.Task)
	if !ok {
		t.Fatalf("result = %#v; want a task record", envelope.Result)
	}
	if _, err := os.Stat(filepath.Join(home, ".hive", "state", "tasks", task.ID+".json")); err != nil {
		t.Fatalf("canonical task record: %v", err)
	}
}
