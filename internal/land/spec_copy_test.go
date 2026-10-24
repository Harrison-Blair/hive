package land

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

func writeSpec(t *testing.T, root, relative, body string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(relative))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readSpec(t *testing.T, root, relative string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relative)))
	if err != nil {
		t.Fatalf("read %s: %v", relative, err)
	}
	return string(data)
}

func hashText(body string) string {
	sum := sha256.Sum256([]byte(body))
	return hex.EncodeToString(sum[:])
}

// writeCopyIn writes the copy-in record that worktree create writes.
func writeCopyIn(t *testing.T, home, taskID string, files map[string]string) {
	t.Helper()
	type file struct {
		Path   string `json:"path"`
		SHA256 string `json:"sha256"`
	}
	record := struct {
		Task    string `json:"task"`
		Feature string `json:"feature"`
		Files   []file `json:"files"`
	}{Task: taskID, Feature: "f"}
	for path, body := range files {
		record.Files = append(record.Files, file{Path: path, SHA256: hashText(body)})
	}
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.WriteIn([]string{"spec-copies"}, taskID, record); err != nil {
		t.Fatal(err)
	}
}

func hasSpecEffect(envelope output.Envelope, action, path string) bool {
	for _, effect := range envelope.Effects {
		if effect.Resource == "spec" && effect.Action == action && effect.Path == path {
			return true
		}
	}
	return false
}

// specLandFixture makes a home, a repository with a task branch, and a
// worktree folder with a copy of the feature spec. The copy-in record holds
// the home files as they were at copy-in.
func specLandFixture(t *testing.T) (string, string, string, records.Task) {
	t.Helper()
	repo := newRepo(t)
	addTaskBranch(t, repo, "task/1-a", "a.txt", "a\n")
	home := newHome(t)
	worktree := t.TempDir()
	task := testTask("1-a", "task/1-a")
	task.Feature = "f"
	task.Worktree.Path = worktree
	writeProject(t, home, testProject(repo))
	writeTask(t, home, task)
	atCopyIn := map[string]string{
		".hive/spec/f/README.md":     "readme v0\n",
		".hive/spec/f/same.md":       "same\n",
		".hive/spec/f/deleted.md":    "deleted\n",
		".hive/spec/f/other-home.md": "other v0\n",
	}
	for path, body := range atCopyIn {
		writeSpec(t, home, path, body)
		if path != ".hive/spec/f/deleted.md" {
			writeSpec(t, worktree, path, body)
		}
	}
	writeCopyIn(t, home, task.ID, atCopyIn)
	return repo, home, worktree, task
}

// AC5: land copies each changed or new spec file outside tasks/ to the home.
// A file that the task deleted stays in the home, and the report names it.
func TestLandCopiesChangedSpecFilesOut(t *testing.T) {
	_, home, worktree, task := specLandFixture(t)
	writeSpec(t, worktree, ".hive/spec/f/README.md", "readme v1\n")
	writeSpec(t, worktree, ".hive/spec/f/screens/new.md", "new screen\n")
	writeSpec(t, worktree, ".hive/spec/f/tasks/1-a/evidence/V1.md", "evidence\n")
	// The home changed a file that the task did not change.
	writeSpec(t, home, ".hive/spec/f/other-home.md", "other v1 from the home\n")
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug}))

	envelope := Land(context.Background(), home, Options{Task: task.ID, Mode: "merge", Grant: "demo", Logger: logger})
	if envelope.Outcome != "ok" || envelope.Error != nil {
		t.Fatalf("Land() = %+v, want ok", envelope)
	}
	if got := readSpec(t, home, ".hive/spec/f/README.md"); got != "readme v1\n" {
		t.Errorf("home README = %q, want the task change", got)
	}
	if got := readSpec(t, home, ".hive/spec/f/screens/new.md"); got != "new screen\n" {
		t.Errorf("home new screen = %q, want the new file", got)
	}
	if got := readSpec(t, home, ".hive/spec/f/other-home.md"); got != "other v1 from the home\n" {
		t.Errorf("home other-home.md = %q, want the home change kept", got)
	}
	if got := readSpec(t, home, ".hive/spec/f/deleted.md"); got != "deleted\n" {
		t.Errorf("home deleted.md = %q, want the home file kept", got)
	}
	if _, err := os.Lstat(filepath.Join(home, ".hive", "spec", "f", "tasks", "1-a", "evidence", "V1.md")); !os.IsNotExist(err) {
		t.Errorf("land copied a file below tasks/ (%v); want no copy", err)
	}
	for action, path := range map[string]string{
		"copied": ".hive/spec/f/README.md",
		"kept":   ".hive/spec/f/deleted.md",
	} {
		if !hasSpecEffect(envelope, action, path) {
			t.Errorf("effects = %+v, want %s %s", envelope.Effects, action, path)
		}
	}
	if !hasSpecEffect(envelope, "copied", ".hive/spec/f/screens/new.md") {
		t.Errorf("effects = %+v, want the copied new file", envelope.Effects)
	}
	if hasSpecEffect(envelope, "copied", ".hive/spec/f/same.md") || hasSpecEffect(envelope, "copied", ".hive/spec/f/other-home.md") {
		t.Errorf("effects = %+v, want no copy of an unchanged file", envelope.Effects)
	}
	if !strings.Contains(logs.String(), `"msg":"land spec copy"`) || !strings.Contains(logs.String(), ".hive/spec/f/README.md") {
		t.Errorf("log = %s, want a spec copy event for the README", logs.String())
	}
}

// AC5: when the home file differs from its copy-in hash, land does not write
// it. The code merge stays, the other files copy, and the result names the
// path and both hashes.
func TestLandRefusesSpecConflict(t *testing.T) {
	repo, home, worktree, task := specLandFixture(t)
	writeSpec(t, worktree, ".hive/spec/f/README.md", "readme from task two\n")
	writeSpec(t, worktree, ".hive/spec/f/screens/new.md", "new screen\n")
	// The first land wrote the README of another task.
	writeSpec(t, home, ".hive/spec/f/README.md", "readme from task one\n")
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug}))

	envelope := Land(context.Background(), home, Options{Task: task.ID, Mode: "merge", Grant: "demo", Logger: logger})
	if envelope.Outcome != "partial" || envelope.Phase != "spec" {
		t.Fatalf("Land() = %+v, want partial spec", envelope)
	}
	message := fmt.Sprint(envelope.Error)
	for _, want := range []string{".hive/spec/f/README.md", hashText("readme v0\n"), hashText("readme from task one\n"), "by hand"} {
		if !strings.Contains(message, want) {
			t.Errorf("error = %q, want %q", message, want)
		}
	}
	if got := readSpec(t, home, ".hive/spec/f/README.md"); got != "readme from task one\n" {
		t.Fatalf("home README = %q, want it unchanged", got)
	}
	if got := readSpec(t, home, ".hive/spec/f/screens/new.md"); got != "new screen\n" {
		t.Fatalf("home new screen = %q, want the other file copied", got)
	}
	if !hasSpecEffect(envelope, "conflict", ".hive/spec/f/README.md") {
		t.Fatalf("effects = %+v, want the conflict", envelope.Effects)
	}
	result, ok := envelope.Result.(Result)
	if !ok || result.Commit == "" {
		t.Fatalf("result = %#v, want the merge commit", envelope.Result)
	}
	if got := strings.TrimSpace(runGit(t, repo, "rev-parse", "main")); got != result.Commit {
		t.Fatalf("main = %s, want the merge commit %s to stay", got, result.Commit)
	}
	if !strings.Contains(logs.String(), `"msg":"land spec conflict"`) {
		t.Fatalf("log = %s, want a spec conflict event", logs.String())
	}
}

// AC5: a second copy-out of the same files is not a conflict, because the
// home file is equal to the worktree file.
func TestSpecCopyOutRepeatIsNoConflict(t *testing.T) {
	_, home, worktree, task := specLandFixture(t)
	writeSpec(t, worktree, ".hive/spec/f/README.md", "readme v1\n")
	if _, conflicts, err := copySpecOut(home, task, nil); err != nil || len(conflicts) != 0 {
		t.Fatalf("first copySpecOut() = %v, %v; want no conflict", conflicts, err)
	}
	effects, conflicts, err := copySpecOut(home, task, nil)
	if err != nil || len(conflicts) != 0 {
		t.Fatalf("second copySpecOut() = %v, %v; want no conflict", conflicts, err)
	}
	for _, effect := range effects {
		if effect.Action == "copied" {
			t.Fatalf("second effects = %+v, want no copy", effects)
		}
	}
}
