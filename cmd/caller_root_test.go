package cmd

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func stateFiles(t *testing.T, home string) map[string]string {
	t.Helper()
	files := map[string]string{}
	root := filepath.Join(home, ".hive", "state")
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		files[path] = string(data)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return files
}

func TestCallerAgentOverrideInPaneFailsForCallerFlags(t *testing.T) {
	for _, args := range [][]string{
		{"task", "complete", "--task", "2-b"},
		{"task", "verify", "--task", "2-b"},
		{"task", "reopen", "--task", "2-b", "--reason", "r"},
		{"task", "edit", "--task", "2-b", "--needs-user", "x"},
		{"goal", "create", "--project", "hive", "--title", "T", "--slug", "t", "--grant", "g"},
		{"goal", "edit", "--goal", "g1-t", "--add-wave", "w"},
		{"chat", "send", "hello"},
		{"chat", "reply", "--to", "c1"},
	} {
		t.Run(strings.Join(args[:2], " "), func(t *testing.T) {
			home := testHome(t)
			seedInboxPaneAgents(t, home)
			before := stateFiles(t, home)
			t.Setenv("HERDR_PANE_ID", "w1:p1")
			var stdout, stderr bytes.Buffer
			status := executeWithStreams(append(args, "--agent", "agent-b"), strings.NewReader(""), &stdout, &stderr)
			want := "--agent agent-b is not the caller agent-a"
			if args[0] == "chat" && args[1] == "send" {
				want = "chat send is for the user only: type in the Chat screen of the hive TUI"
			}
			if status == 0 || !strings.Contains(stderr.String(), want) {
				t.Fatalf("status = %d, stdout = %q, stderr = %q; want %q", status, stdout.String(), stderr.String(), want)
			}
			after := stateFiles(t, home)
			if len(after) != len(before) {
				t.Fatalf("state files = %d; want %d unchanged", len(after), len(before))
			}
			for path, data := range before {
				if after[path] != data {
					t.Fatalf("%s changed", path)
				}
			}
		})
	}
}

func TestTaskCompleteAgentOverrideFromOtherPaneLeavesTask(t *testing.T) {
	home, taskID := seedFinishCLI(t, "assigned")
	before := stateFiles(t, home)
	t.Setenv("HERDR_PANE_ID", "w1:p2")
	var stdout, stderr bytes.Buffer
	status := executeWithStreams([]string{"task", "complete", "--task", taskID, "--agent", "worker-1"}, strings.NewReader(""), &stdout, &stderr)
	if status == 0 || !strings.Contains(stderr.String(), "--agent worker-1 is not the caller verifier-1") {
		t.Fatalf("status = %d, stdout = %q, stderr = %q; want the --agent refusal", status, stdout.String(), stderr.String())
	}
	after := stateFiles(t, home)
	for path, data := range before {
		if after[path] != data {
			t.Fatalf("%s changed", path)
		}
	}
	if len(after) != len(before) {
		t.Fatalf("state files = %d; want %d", len(after), len(before))
	}
}

func TestTaskCompleteOwnAgentInPaneSucceeds(t *testing.T) {
	_, taskID := seedFinishCLI(t, "assigned")
	t.Setenv("HERDR_PANE_ID", "w1:p1")
	var stdout, stderr bytes.Buffer
	if status := executeWithStreams([]string{"task", "complete", "--task", taskID, "--agent", "worker-1"}, strings.NewReader(""), &stdout, &stderr); status != 0 {
		t.Fatalf("status = %d, stderr = %q; want 0", status, stderr.String())
	}
}
