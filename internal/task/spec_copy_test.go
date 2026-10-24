package task

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
)

// writeFile writes one fixture file and its parent folders.
func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// readFileOrEmpty returns the body of one file and whether it exists.
func readFileOrEmpty(t *testing.T, path string) (string, bool) {
	t.Helper()
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return "", false
	}
	if err != nil {
		t.Fatal(err)
	}
	return string(data), true
}

func hasTaskEffect(envelope output.Envelope, resource, action, path string) bool {
	for _, effect := range envelope.Effects {
		if effect.Resource == resource && effect.Action == action && effect.Path == path {
			return true
		}
	}
	return false
}

// AC2: the README check reads the home. The project repository of the
// fixture does not exist, so a check on the base branch cannot pass.
func TestCreateFeatureChecksHomeREADME(t *testing.T) {
	home := newHome(t)
	registerProject(t, home, "one")
	readme := filepath.Join(home, ".hive", "spec", "f", "README.md")

	envelope := Create(home, CreateOptions{Project: "one", Title: "Work", Feature: "f"})
	if envelope.Outcome != "failed" || envelope.Phase != "validate" || !strings.Contains(fmt.Sprint(envelope.Error), readme) {
		t.Fatalf("Create() without home README = %#v; want a validate failure that names %s", envelope, readme)
	}

	writeFile(t, readme, "# F\n")
	task := createOK(t, home, CreateOptions{Project: "one", Title: "Work", Feature: "f"})
	if task.Feature != "f" {
		t.Fatalf("task = %#v; want feature f", task)
	}
}

// AC2: the brief of a feature task goes to the spec folder of the home.
func TestCreateFeatureWritesHomeBrief(t *testing.T) {
	home := newHome(t)
	registerProject(t, home, "one")
	writeFile(t, filepath.Join(home, ".hive", "spec", "f", "README.md"), "# F\n")
	const brief = "# Brief\n"

	envelope := Create(home, CreateOptions{Project: "one", Title: "Work", Feature: "f", BriefSet: true, BriefPath: "-", Input: strings.NewReader(brief)})
	if envelope.Outcome != "ok" {
		t.Fatalf("Create() = %#v; want ok", envelope)
	}
	task := envelope.Result.(records.Task)
	relative := ".hive/spec/f/tasks/" + task.ID + "/brief.md"
	if got, ok := readFileOrEmpty(t, filepath.Join(home, filepath.FromSlash(relative))); !ok || got != brief {
		t.Fatalf("home feature brief = %q, %v; want %q", got, ok, brief)
	}
	if !hasTaskEffect(envelope, "brief", "created", relative) {
		t.Fatalf("effects = %+v; want the created home feature brief %s", envelope.Effects, relative)
	}
}

// AC2: a spec task needs no README, and its brief also goes to the home.
func TestCreateSpecTaskWritesHomeBriefWithoutREADME(t *testing.T) {
	home := newHome(t)
	registerProject(t, home, "one")

	task := createOK(t, home, CreateOptions{Project: "one", Title: "Spec", Feature: "f", Spec: true})
	path := filepath.Join(home, ".hive", "spec", "f", "tasks", task.ID, "brief.md")
	if got, ok := readFileOrEmpty(t, path); !ok || got != "Spec" {
		t.Fatalf("home feature brief = %q, %v; want the title", got, ok)
	}
}

// AC3: task edit --brief writes the brief to the home and to the worktree.
func TestEditBriefWritesHomeFeatureBrief(t *testing.T) {
	home := newHome(t)
	registerProject(t, home, "one")
	worktree := t.TempDir()
	task := records.Task{
		ID: "1-seed", Project: "one", Title: "Seed", Feature: "feature-a",
		ControllingRole: "manager-one", Status: "assigned", DependsOn: []string{},
		CreatedAt: "2026-01-01T00:00:00Z", UpdatedAt: "2026-01-01T00:00:00Z",
		Worktree: &records.Worktree{Path: worktree},
	}
	seedTask(t, home, task)
	seedAgent(t, home, records.Agent{
		Key: "manager-one", Name: "manager-one", Role: "manager", Project: "one",
		Status: records.AgentLive, PaneID: "w1:manager", Generation: 1,
		StartedAt: "2026-01-01T00:00:00Z", UpdatedAt: "2026-01-01T00:00:00Z",
	})
	seedClaim(t, home, records.Claim{Key: "manager-one", Agent: "manager-one", Generation: 1})
	writeFile(t, filepath.Join(home, ".hive", "state", "briefs", task.ID+".md"), "old brief")

	const updated = "updated brief\n"
	envelope := Edit(home, EditOptions{Task: task.ID, Agent: "manager-one", BriefSet: true, BriefPath: "-", Input: strings.NewReader(updated)})
	if envelope.Outcome != "ok" {
		t.Fatalf("Edit() = %+v; want ok", envelope)
	}
	relative := ".hive/spec/feature-a/tasks/1-seed/brief.md"
	if got, ok := readFileOrEmpty(t, filepath.Join(home, filepath.FromSlash(relative))); !ok || got != updated {
		t.Fatalf("home feature brief = %q, %v; want %q", got, ok, updated)
	}
	if got, ok := readFileOrEmpty(t, filepath.Join(worktree, filepath.FromSlash(relative))); !ok || got != updated {
		t.Fatalf("worktree feature brief = %q, %v; want %q", got, ok, updated)
	}
	if !hasTaskEffect(envelope, "brief", "updated", relative) {
		t.Fatalf("effects = %+v; want the updated home feature brief", envelope.Effects)
	}
}

// AC4: task complete copies the evidence folder of the worktree to the home.
// It copies no other file of the feature folder.
func TestCompleteCopiesEvidenceToHome(t *testing.T) {
	home, task := newSeed(t, "assigned")
	worktree := t.TempDir()
	task.Feature = "f"
	task.Worktree = &records.Worktree{Path: worktree, Branch: "task/" + task.ID}
	seedTask(t, home, task)
	evidence := ".hive/spec/f/tasks/1-seed/evidence/"
	writeFile(t, filepath.Join(worktree, filepath.FromSlash(evidence), "V1.md"), "red and green\n")
	writeFile(t, filepath.Join(worktree, filepath.FromSlash(evidence), "screens", "chat.png"), "png\n")
	writeFile(t, filepath.Join(worktree, ".hive", "spec", "f", "README.md"), "worktree README\n")

	envelope := Complete(home, TransitionOptions{Task: task.ID, Agent: "worker-1"})
	if envelope.Outcome != "ok" || envelope.Error != nil {
		t.Fatalf("Complete() = %#v; want ok", envelope)
	}
	for name, want := range map[string]string{"V1.md": "red and green\n", "screens/chat.png": "png\n"} {
		relative := evidence + name
		if got, ok := readFileOrEmpty(t, filepath.Join(home, filepath.FromSlash(relative))); !ok || got != want {
			t.Errorf("home %s = %q, %v; want %q", relative, got, ok, want)
		}
		if !hasTaskEffect(envelope, "evidence", "copied", relative) {
			t.Errorf("effects = %+v; want copied %s", envelope.Effects, relative)
		}
	}
	if _, ok := readFileOrEmpty(t, filepath.Join(home, ".hive", "spec", "f", "README.md")); ok {
		t.Fatalf("complete copied the feature README; want the evidence folder only")
	}
}

// AC4: task verify copies VERIFIED.md to the home. An ignored evidence file
// does not make the worktree dirty.
func TestVerifyCopiesVerifiedToHome(t *testing.T) {
	repo := newVerifyGitRepo(t, true, false, false, true)
	writeFile(t, filepath.Join(repo, ".gitignore"), ".hive/spec/\n")
	runVerifyGit(t, repo, "add", ".gitignore")
	runVerifyGit(t, repo, "commit", "-q", "-m", "ignore the spec")
	runVerifyGit(t, repo, "push", "-q", "origin", "task/verify")
	home, task := verifyWithWorktree(t, repo)
	task.Feature = "f"
	seedTask(t, home, task)
	evidence := ".hive/spec/f/tasks/1-seed/evidence/"
	writeFile(t, filepath.Join(repo, filepath.FromSlash(evidence), "VERIFIED.md"), "pass\n")
	writeFile(t, filepath.Join(repo, filepath.FromSlash(evidence), "V1.md"), "worker run\n")

	envelope := Verify(home, TransitionOptions{Task: task.ID, Agent: "verifier-1"})
	if envelope.Outcome != "ok" || envelope.Error != nil {
		t.Fatalf("Verify() = %#v; want ok with ignored evidence files", envelope)
	}
	relative := evidence + "VERIFIED.md"
	if got, ok := readFileOrEmpty(t, filepath.Join(home, filepath.FromSlash(relative))); !ok || got != "pass\n" {
		t.Fatalf("home VERIFIED.md = %q, %v; want the verifier file", got, ok)
	}
	if !hasTaskEffect(envelope, "evidence", "copied", relative) {
		t.Fatalf("effects = %+v; want copied %s", envelope.Effects, relative)
	}
	if _, ok := readFileOrEmpty(t, filepath.Join(home, filepath.FromSlash(evidence), "V1.md")); ok {
		t.Fatalf("verify copied V1.md; want VERIFIED.md only")
	}
}

// AC4: a failed copy-out keeps the transition and reports a partial result.
func TestCompleteEvidenceCopyFailureIsPartial(t *testing.T) {
	home, task := newSeed(t, "assigned")
	worktree := t.TempDir()
	task.Feature = "f"
	task.Worktree = &records.Worktree{Path: worktree, Branch: "task/" + task.ID}
	seedTask(t, home, task)
	writeFile(t, filepath.Join(worktree, ".hive", "spec", "f", "tasks", "1-seed", "evidence", "V1.md"), "run\n")
	// A file at the home folder path blocks the copy.
	writeFile(t, filepath.Join(home, ".hive", "spec", "f", "tasks", "1-seed"), "not a folder")

	envelope := Complete(home, TransitionOptions{Task: task.ID, Agent: "worker-1"})
	if envelope.Outcome != "partial" || envelope.Phase != "evidence" {
		t.Fatalf("Complete() = %#v; want partial evidence", envelope)
	}
	if got := readTask(t, home, task.ID); got.Status != "completed" {
		t.Fatalf("status = %q; want completed", got.Status)
	}
}

// AC2: task show checks the README and the brief in the home. The project
// repository of the fixture does not exist, so a check on the base branch
// always warns.
func TestShowChecksFeatureFilesInHome(t *testing.T) {
	home := newHome(t)
	registerProject(t, home, "one")
	writeFile(t, filepath.Join(home, ".hive", "spec", "f", "README.md"), "# F\n")
	task := createOK(t, home, CreateOptions{Project: "one", Title: "Work", Feature: "f"})

	shown := Show(home, task.ID, ShowOptions{})
	result, ok := shown.Result.(ShowResult)
	if shown.Outcome != "ok" || !ok || len(result.Warnings) != 0 {
		t.Fatalf("Show() = %#v; want no warnings with the home README and brief", shown)
	}

	if err := os.Remove(filepath.Join(home, ".hive", "spec", "f", "README.md")); err != nil {
		t.Fatal(err)
	}
	shown = Show(home, task.ID, ShowOptions{})
	result = shown.Result.(ShowResult)
	if len(result.Warnings) != 1 || result.Warnings[0] != "linked file missing: .hive/spec/f/README.md" {
		t.Fatalf("warnings = %v; want the missing home README only", result.Warnings)
	}
}
