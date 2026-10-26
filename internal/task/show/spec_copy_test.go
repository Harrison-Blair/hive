package show

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Harrison-Blair/hive/internal/task/create"
	"github.com/Harrison-Blair/hive/internal/task/tasktest"
)

// AC2: task show checks the README and the brief in the home. The project
// repository of the fixture does not exist, so a check on the base branch
// always warns.
func TestShowChecksFeatureFilesInHome(t *testing.T) {
	home := tasktest.NewHome(t)
	tasktest.RegisterProject(t, home, "one")
	tasktest.WriteFile(t, filepath.Join(home, ".hive", "spec", "f", "README.md"), "# F\n")
	task := createOK(t, home, create.Options{Project: "one", Title: "Work", Feature: "f"})

	shown := Run(home, task.ID, Options{})
	result, ok := shown.Result.(Result)
	if shown.Outcome != "ok" || !ok || len(result.Warnings) != 0 {
		t.Fatalf("Show() = %#v; want no warnings with the home README and brief", shown)
	}

	if err := os.Remove(filepath.Join(home, ".hive", "spec", "f", "README.md")); err != nil {
		t.Fatal(err)
	}
	shown = Run(home, task.ID, Options{})
	result = shown.Result.(Result)
	if len(result.Warnings) != 1 || result.Warnings[0] != "linked file missing: .hive/spec/f/README.md" {
		t.Fatalf("warnings = %v; want the missing home README only", result.Warnings)
	}
}
