package task

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/records"
)

// briefAncestorPath returns the computed brief directory of a home.
func briefAncestorPath(home string) string {
	return filepath.Join(home, ".hive", "state", "briefs")
}

func TestCreateRejectsLinkedBriefAncestor(t *testing.T) {
	home := newHome(t)
	registerProject(t, home, "one")
	external := t.TempDir()
	briefs := briefAncestorPath(home)
	if err := os.MkdirAll(filepath.Dir(briefs), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(external, briefs); err != nil {
		t.Fatal(err)
	}

	envelope := Create(home, CreateOptions{Project: "one", Title: "Seed"})
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
	home := newHome(t)
	registerProject(t, home, "one")
	briefs := briefAncestorPath(home)
	if err := os.MkdirAll(filepath.Dir(briefs), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(briefs, []byte("not a directory"), 0o644); err != nil {
		t.Fatal(err)
	}

	envelope := Create(home, CreateOptions{Project: "one", Title: "Seed"})
	if envelope.Outcome != "failed" || envelope.Error == nil {
		t.Fatalf("Create() = %#v; want a failed outcome", envelope)
	}
	if _, ok := counterValue(t, home); ok {
		t.Fatal("wrong-type brief ancestor changed the counter")
	}
	assertNoTaskRecord(t, home, "1-seed")
}

func TestShowRejectsLinkedBriefAncestor(t *testing.T) {
	home := newHome(t)
	registerProject(t, home, "one")
	task := createOK(t, home, CreateOptions{Project: "one", Title: "Second step"})

	briefs := briefAncestorPath(home)
	archived := filepath.Join(home, ".hive", "state", "briefs-original")
	if err := os.Rename(briefs, archived); err != nil {
		t.Fatal(err)
	}
	external := t.TempDir()
	const externalBrief = "external brief"
	if err := os.WriteFile(filepath.Join(external, task.ID+".md"), []byte(externalBrief), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(external, briefs); err != nil {
		t.Fatal(err)
	}

	envelope := Show(home, task.ID, ShowOptions{})
	if envelope.Outcome != "failed" || envelope.Error == nil {
		t.Fatalf("Show() = %#v; want a failed outcome", envelope)
	}
	if _, ok := envelope.Result.(ShowResult); ok {
		t.Fatalf("Show() result = %#v; want no brief from the external directory", envelope.Result)
	}
	data, err := os.ReadFile(filepath.Join(external, task.ID+".md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != externalBrief {
		t.Fatalf("external brief = %q; want %q", data, externalBrief)
	}
}

func TestCreateKeepsUserBriefBelowLinkedDirectory(t *testing.T) {
	home := newHome(t)
	registerProject(t, home, "one")
	target := t.TempDir()
	const userBrief = "user selected brief"
	if err := os.WriteFile(filepath.Join(target, "brief.md"), []byte(userBrief), 0o644); err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(t.TempDir(), "alias")
	if err := os.Symlink(target, alias); err != nil {
		t.Fatal(err)
	}

	task := createOK(t, home, CreateOptions{
		Project:   "one",
		Title:     "Step",
		BriefPath: filepath.Join(alias, "brief.md"),
		BriefSet:  true,
	})
	data, err := os.ReadFile(filepath.Join(briefAncestorPath(home), task.ID+".md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != userBrief {
		t.Fatalf("stored brief = %q; want the user-selected brief", data)
	}
}

func TestCreateAcceptsSymlinkedHomeRoot(t *testing.T) {
	home := newHome(t)
	registerProject(t, home, "one")
	alias := filepath.Join(filepath.Dir(home), "alias")
	if err := os.Symlink(home, alias); err != nil {
		t.Fatal(err)
	}

	envelope := Create(alias, CreateOptions{Project: "one", Title: "Step"})
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
