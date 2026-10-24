package worktree

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
)

func writeSpecFile(t *testing.T, root, relative, body string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(relative))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func sha256Text(body string) string {
	sum := sha256.Sum256([]byte(body))
	return hex.EncodeToString(sum[:])
}

// createFeatureWorktree runs one worktree create for a feature task against a
// fake herdr that makes a real linked worktree.
func createFeatureWorktree(t *testing.T, repo, home string, task records.Task, logger *slog.Logger) (output.Envelope, string) {
	t.Helper()
	worktreePath := filepath.Join(home, "projects", task.Project, "worktrees", task.ID)
	branch := "task/" + task.ID
	var listCalls atomic.Int32
	startFakeHerdr(t, func(request fakeRequest) fakeReply {
		switch request.Method {
		case "workspace.list":
			if listCalls.Add(1) == 1 {
				return fakeReply{Result: workspaceListReply(t)}
			}
			return fakeReply{Result: workspaceListReply(t, linkedWorkspace("w2", worktreePath, repo))}
		case "worktree.create":
			runGit(t, repo, "worktree", "add", "-q", "-b", branch, worktreePath, "main")
			return fakeReply{Result: createReply(t, "w2", "w2:t1", "w2:p1", worktreePath, branch, repo)}
		}
		return fakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
	})
	return Create(context.Background(), home, CreateOptions{Task: task.ID, Logger: logger}), worktreePath
}

func featureFixture(t *testing.T) (string, string, records.Task) {
	t.Helper()
	repo := newTestRepo(t)
	home := newTestHome(t, "repo-live")
	task := testTask()
	task.Feature = "f"
	writeProject(t, home, testProject(repo))
	writeTask(t, home, task)
	writeSpecFile(t, home, ".hive/state/briefs/"+task.ID+".md", "# Brief\n")
	return repo, home, task
}

// AC3: worktree create copies the feature folder of the home, except each
// task folder of another task, and records the SHA-256 hash of each file.
func TestCreateCopiesFeatureSpecInAndRecordsHashes(t *testing.T) {
	repo, home, task := featureFixture(t)
	copied := map[string]string{
		".hive/spec/f/README.md":                           "# F\n",
		".hive/spec/f/screens/chat.md":                     "chat rules\n",
		".hive/spec/f/design/chat.grid":                    "grid\n",
		".hive/spec/f/tasks/" + task.ID + "/brief.md":      "# Brief\n",
		".hive/spec/f/tasks/" + task.ID + "/notes/plan.md": "plan\n",
	}
	for relative, body := range copied {
		writeSpecFile(t, home, relative, body)
	}
	writeSpecFile(t, home, ".hive/spec/f/tasks/2-other/evidence/V1.md", "other\n")
	writeSpecFile(t, home, ".hive/spec/f/tasks/2-other/brief.md", "other brief\n")
	writeSpecFile(t, home, ".hive/spec/g/README.md", "# G\n")
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug}))

	envelope, worktreePath := createFeatureWorktree(t, repo, home, task, logger)
	if envelope.Outcome != "ok" {
		t.Fatalf("Create() = %+v, want ok", envelope)
	}
	for relative, body := range copied {
		got, err := os.ReadFile(filepath.Join(worktreePath, filepath.FromSlash(relative)))
		if err != nil || string(got) != body {
			t.Errorf("worktree %s = %q, %v; want %q", relative, got, err, body)
		}
	}
	for _, relative := range []string{".hive/spec/f/tasks/2-other", ".hive/spec/g"} {
		if _, err := os.Lstat(filepath.Join(worktreePath, filepath.FromSlash(relative))); !os.IsNotExist(err) {
			t.Errorf("worktree has %s (%v); want no copy", relative, err)
		}
	}

	data, err := os.ReadFile(filepath.Join(home, ".hive", "state", "spec-copies", task.ID+".json"))
	if err != nil {
		t.Fatalf("read the copy-in record: %v", err)
	}
	var record struct {
		Task    string `json:"task"`
		Feature string `json:"feature"`
		Files   []struct {
			Path   string `json:"path"`
			SHA256 string `json:"sha256"`
		} `json:"files"`
	}
	if err := json.Unmarshal(data, &record); err != nil {
		t.Fatal(err)
	}
	if record.Task != task.ID || record.Feature != "f" {
		t.Fatalf("record = %+v, want task %s and feature f", record, task.ID)
	}
	got := map[string]string{}
	for _, file := range record.Files {
		got[file.Path] = file.SHA256
	}
	want := map[string]string{}
	for relative, body := range copied {
		want[relative] = sha256Text(body)
	}
	if len(got) != len(want) {
		t.Fatalf("recorded files = %v, want %v", got, want)
	}
	for path, hash := range want {
		if got[path] != hash {
			t.Errorf("recorded hash of %s = %q, want %q", path, got[path], hash)
		}
	}
	if !hasEffect(envelope, "spec", "copied", worktreePath+"/.hive/spec/f/") {
		t.Fatalf("effects = %+v, want the copied spec folder", envelope.Effects)
	}

	var count float64 = -1
	for _, line := range strings.Split(strings.TrimSpace(logs.String()), "\n") {
		var event map[string]any
		if json.Unmarshal([]byte(line), &event) == nil && event["msg"] == "worktree spec copy-in" {
			count, _ = event["files"].(float64)
		}
	}
	if int(count) != len(copied) {
		t.Fatalf("copy-in log count = %v, want %d; log:\n%s", count, len(copied), logs.String())
	}
}

// AC7: worktree create has no allowlist check. A repository that ignores
// .hive/spec/ gets a worktree and an untracked, ignored spec copy.
func TestCreateFeatureInIgnoringRepoHasNoAllowlistStep(t *testing.T) {
	repo, home, task := featureFixture(t)
	if err := os.WriteFile(filepath.Join(repo, ".gitignore"), []byte("*\n!*/\n!*.md\n!.gitignore\n.hive/*\n!.hive/README.md\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, repo, "add", ".gitignore")
	runGit(t, repo, "commit", "-q", "-m", "ignore the spec")
	writeSpecFile(t, home, ".hive/spec/f/README.md", "# F\n")

	envelope, worktreePath := createFeatureWorktree(t, repo, home, task, nil)
	if envelope.Outcome != "ok" {
		t.Fatalf("Create() = %+v, want ok", envelope)
	}
	if got, want := effectResources(envelope), []string{"branch", "brief", "spec", "spec_copy", "workspace", "worktree"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("effect resources = %v, want %v", got, want)
	}
	if staged := strings.TrimSpace(runGit(t, worktreePath, "diff", "--cached", "--name-only")); staged != "" {
		t.Fatalf("staged paths = %q, want none", staged)
	}
	if status := strings.TrimSpace(runGit(t, worktreePath, "status", "--porcelain")); status != "" {
		t.Fatalf("git status = %q, want a clean worktree", status)
	}
	if _, err := os.Stat(filepath.Join(worktreePath, ".hive", "spec", "f", "README.md")); err != nil {
		t.Fatalf("worktree README: %v", err)
	}
}

// AC3: the record lists the paths in a stable order.
func TestSpecCopyRecordIsSorted(t *testing.T) {
	repo, home, task := featureFixture(t)
	for _, name := range []string{"z.md", "a.md", "m/b.md"} {
		writeSpecFile(t, home, ".hive/spec/f/"+name, name)
	}
	if envelope, _ := createFeatureWorktree(t, repo, home, task, nil); envelope.Outcome != "ok" {
		t.Fatalf("Create() = %+v, want ok", envelope)
	}
	record, found, err := ReadSpecCopy(home, task.ID)
	if err != nil || !found {
		t.Fatalf("ReadSpecCopy() = %v, %v", found, err)
	}
	paths := []string{}
	for _, file := range record.Files {
		paths = append(paths, file.Path)
	}
	if !sort.StringsAreSorted(paths) {
		t.Fatalf("paths = %v, want sorted", paths)
	}
}
