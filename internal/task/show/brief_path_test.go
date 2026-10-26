package show

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Harrison-Blair/hive/internal/task/create"
	"github.com/Harrison-Blair/hive/internal/task/tasktest"
)

func TestShowRejectsLinkedBriefAncestor(t *testing.T) {
	home := tasktest.NewHome(t)
	tasktest.RegisterProject(t, home, "one")
	task := createOK(t, home, create.Options{Project: "one", Title: "Second step"})

	briefs := tasktest.BriefAncestorPath(home)
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

	envelope := Run(home, task.ID, Options{})
	if envelope.Outcome != "failed" || envelope.Error == nil {
		t.Fatalf("Show() = %#v; want a failed outcome", envelope)
	}
	if _, ok := envelope.Result.(Result); ok {
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
