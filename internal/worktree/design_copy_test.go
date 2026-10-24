package worktree

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// AC1, AC3: worktree create copies the design folder of each other feature,
// with no other file of that feature, and reports the counts.
func TestCreateCopiesDesignFoldersOfOtherFeatures(t *testing.T) {
	repo, home, task := featureFixture(t)
	writeSpecFile(t, home, ".hive/spec/f/README.md", "# F\n")
	designs := map[string]string{
		".hive/spec/g/design/chat.grid":            "g grid\n",
		".hive/spec/g/design/fixtures/a/frame.txt": "g frame\n",
		".hive/spec/h/design/board.dc.html":        "h board\n",
	}
	for relative, body := range designs {
		writeSpecFile(t, home, relative, body)
	}
	excluded := []string{
		".hive/spec/g/README.md",
		".hive/spec/g/screens/chat.md",
		".hive/spec/g/tasks/9-g/brief.md",
		".hive/spec/g/tasks/9-g/evidence/V1.md",
		".hive/spec/h/notes.md",
		".hive/spec/k/README.md",
	}
	for _, relative := range excluded {
		writeSpecFile(t, home, relative, "excluded\n")
	}
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug}))

	envelope, worktreePath := createFeatureWorktree(t, repo, home, task, logger)
	if envelope.Outcome != "ok" {
		t.Fatalf("Create() = %+v, want ok", envelope)
	}
	for relative, body := range designs {
		got, err := os.ReadFile(filepath.Join(worktreePath, filepath.FromSlash(relative)))
		if err != nil || string(got) != body {
			t.Errorf("worktree %s = %q, %v; want %q", relative, got, err, body)
		}
	}
	for _, relative := range append(excluded, ".hive/spec/k") {
		if _, err := os.Lstat(filepath.Join(worktreePath, filepath.FromSlash(relative))); !os.IsNotExist(err) {
			t.Errorf("worktree has %s (%v); want no copy", relative, err)
		}
	}

	// AC2: the hash record holds only the feature of the task.
	record, found, err := ReadSpecCopy(home, task.ID)
	if err != nil || !found {
		t.Fatalf("ReadSpecCopy() = %v, %v", found, err)
	}
	for _, file := range record.Files {
		if !strings.HasPrefix(file.Path, ".hive/spec/f/") {
			t.Errorf("record has %s; want only files of feature f", file.Path)
		}
	}

	// AC3: text, JSON result, and log give the counts.
	if text := CreateText(envelope); !strings.Contains(text, "copied design 2 features 3 files\n") {
		t.Errorf("CreateText() = %q, want the line %q", text, "copied design 2 features 3 files")
	}
	data, err := json.Marshal(envelope.Result)
	if err != nil {
		t.Fatal(err)
	}
	var result struct {
		Design struct {
			Features *int `json:"features"`
			Files    *int `json:"files"`
		} `json:"design"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}
	if result.Design.Features == nil || *result.Design.Features != 2 || result.Design.Files == nil || *result.Design.Files != 3 {
		t.Errorf("result = %s, want design features 2 and files 3", data)
	}
	events := 0
	for _, line := range strings.Split(strings.TrimSpace(logs.String()), "\n") {
		var event map[string]any
		if json.Unmarshal([]byte(line), &event) != nil || event["msg"] != "worktree design copy-in" {
			continue
		}
		events++
		if event["task_id"] != task.ID || event["features"] != float64(2) || event["files"] != float64(3) {
			t.Errorf("design log event = %v, want task_id %s, features 2, files 3", event, task.ID)
		}
	}
	if events != 1 {
		t.Fatalf("design log events = %d, want 1; log:\n%s", events, logs.String())
	}
}

// AC1: a home with no .hive/spec/ copies nothing, and create does not fail.
func TestCreateWithoutSpecFolderCopiesNoDesign(t *testing.T) {
	repo, home, task := featureFixture(t)
	if err := os.RemoveAll(filepath.Join(home, ".hive", "spec")); err != nil {
		t.Fatal(err)
	}

	envelope, worktreePath := createFeatureWorktree(t, repo, home, task, nil)
	if envelope.Outcome != "ok" {
		t.Fatalf("Create() = %+v, want ok", envelope)
	}
	entries, err := os.ReadDir(filepath.Join(worktreePath, ".hive", "spec"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "f" {
		t.Fatalf("worktree .hive/spec has %v; want only the brief folder of f", entries)
	}
	data, err := json.Marshal(envelope.Result)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"design":{"features":0,"files":0}`) {
		t.Fatalf("result = %s, want design features 0 and files 0", data)
	}
}
