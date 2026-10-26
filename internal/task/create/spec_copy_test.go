package create

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/task/tasktest"
)

// AC2: the README check reads the home. The project repository of the
// fixture does not exist, so a check on the base branch cannot pass.
func TestCreateFeatureChecksHomeREADME(t *testing.T) {
	home := tasktest.NewHome(t)
	tasktest.RegisterProject(t, home, "one")
	readme := filepath.Join(home, ".hive", "spec", "f", "README.md")

	envelope := Run(home, Options{Project: "one", Title: "Work", Feature: "f"})
	if envelope.Outcome != "failed" || envelope.Phase != "validate" || !strings.Contains(fmt.Sprint(envelope.Error), readme) {
		t.Fatalf("Create() without home README = %#v; want a validate failure that names %s", envelope, readme)
	}

	tasktest.WriteFile(t, readme, "# F\n")
	task := createOK(t, home, Options{Project: "one", Title: "Work", Feature: "f"})
	if task.Feature != "f" {
		t.Fatalf("task = %#v; want feature f", task)
	}
}

// AC2: the brief of a feature task goes to the spec folder of the home.
func TestCreateFeatureWritesHomeBrief(t *testing.T) {
	home := tasktest.NewHome(t)
	tasktest.RegisterProject(t, home, "one")
	tasktest.WriteFile(t, filepath.Join(home, ".hive", "spec", "f", "README.md"), "# F\n")
	const brief = "# Brief\n"

	envelope := Run(home, Options{Project: "one", Title: "Work", Feature: "f", BriefSet: true, BriefPath: "-", Input: strings.NewReader(brief)})
	if envelope.Outcome != "ok" {
		t.Fatalf("Create() = %#v; want ok", envelope)
	}
	task := envelope.Result.(records.Task)
	relative := ".hive/spec/f/tasks/" + task.ID + "/brief.md"
	if got, ok := tasktest.ReadFileOrEmpty(t, filepath.Join(home, filepath.FromSlash(relative))); !ok || got != brief {
		t.Fatalf("home feature brief = %q, %v; want %q", got, ok, brief)
	}
	if !tasktest.HasTaskEffect(envelope, "brief", "created", relative) {
		t.Fatalf("effects = %+v; want the created home feature brief %s", envelope.Effects, relative)
	}
}

// AC2: a spec task needs no README, and its brief also goes to the home.
func TestCreateSpecTaskWritesHomeBriefWithoutREADME(t *testing.T) {
	home := tasktest.NewHome(t)
	tasktest.RegisterProject(t, home, "one")

	task := createOK(t, home, Options{Project: "one", Title: "Spec", Feature: "f", Spec: true})
	path := filepath.Join(home, ".hive", "spec", "f", "tasks", task.ID, "brief.md")
	if got, ok := tasktest.ReadFileOrEmpty(t, path); !ok || got != "Spec" {
		t.Fatalf("home feature brief = %q, %v; want the title", got, ok)
	}
}
