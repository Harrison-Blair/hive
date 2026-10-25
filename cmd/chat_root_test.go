package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

func seedChatCommandAgent(t *testing.T, home string, record records.Agent) {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.WriteIn([]string{"agents"}, record.Key, record); err != nil {
		t.Fatal(err)
	}
}

func seedChatCommandCoordinator(t *testing.T, home string) {
	t.Helper()
	record := records.Agent{Key: "coordinator-g1", Name: "coordinator", Role: "coordinator", PaneID: "w1:p1", Status: records.AgentLive, Generation: 1}
	seedChatCommandAgent(t, home, record)
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.WriteIn([]string{"claims"}, "coordinator", records.Claim{Key: "coordinator", Agent: record.Key, Generation: 1}); err != nil {
		t.Fatal(err)
	}
}

func seedChatCommandWorker(t *testing.T, home string) {
	t.Helper()
	seedChatCommandAgent(t, home, records.Agent{Key: "worker-one", Name: "worker-one", Role: "worker", PaneID: "w2:p1", Status: records.AgentLive, Generation: 1})
}

func chatCommandFiles(t *testing.T, home string) []string {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(home, ".hive", "state", "chat", "*.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	return files
}

func TestChatSendWithoutCoordinatorWritesNoChatEntry(t *testing.T) {
	home := newRootTestHome(t)
	var stdout, stderr bytes.Buffer
	status := executeWithStreams([]string{"chat", "send", "hello"}, strings.NewReader(""), &stdout, &stderr)
	want := "hive: chat send is for the user only: type in the Chat screen of the hive TUI\n"
	if status != 1 || stderr.String() != want {
		t.Fatalf("status = %d, stderr = %q; want status 1 and %q", status, stderr.String(), want)
	}
	if files := chatCommandFiles(t, home); len(files) != 0 {
		t.Fatalf("chat files = %v; want none", files)
	}
}

func TestChatSendAsRegisteredAgentWritesNoChatEntry(t *testing.T) {
	for _, args := range [][]string{
		{"chat", "send", "--agent", "worker-one", "hello"},
		{"chat", "send", "--pane", "w2:p1", "hello"},
	} {
		t.Run(strings.Join(args, "-"), func(t *testing.T) {
			home := newRootTestHome(t)
			t.Setenv("HERDR_PANE_ID", "")
			seedChatCommandCoordinator(t, home)
			seedChatCommandWorker(t, home)
			var stdout, stderr bytes.Buffer
			status := executeWithStreams(args, strings.NewReader(""), &stdout, &stderr)
			want := "hive: chat send is for the user only: type in the Chat screen of the hive TUI\n"
			if status != 1 || stderr.String() != want {
				t.Fatalf("status = %d, stderr = %q; want status 1 and %q", status, stderr.String(), want)
			}
			if files := chatCommandFiles(t, home); len(files) != 0 {
				t.Fatalf("chat files = %v; want none", files)
			}
		})
	}
}

func TestChatReplyAsNonCoordinatorWritesNoChatEntry(t *testing.T) {
	home := newRootTestHome(t)
	t.Setenv("HERDR_PANE_ID", "")
	seedChatCommandCoordinator(t, home)
	seedChatCommandWorker(t, home)
	var stdout, stderr bytes.Buffer
	status := executeWithStreams([]string{"chat", "reply", "--agent", "worker-one"}, strings.NewReader("answer"), &stdout, &stderr)
	if status == 0 || !strings.Contains(stderr.String(), "coordinator") {
		t.Fatalf("status = %d, stderr = %q; want coordinator-role refusal", status, stderr.String())
	}
	if files := chatCommandFiles(t, home); len(files) != 0 {
		t.Fatalf("chat files = %v; want none", files)
	}
}

func TestAgentSendUserTagWritesNoChatEntry(t *testing.T) {
	home := newRootTestHome(t)
	seedChatCommandWorker(t, home)
	var stdout, stderr bytes.Buffer
	status := executeWithStreams([]string{"agent", "send", "--agent", "worker-one", "user: x"}, strings.NewReader(""), &stdout, &stderr)
	if status == 0 || !strings.Contains(stderr.String(), "user-tagged") {
		t.Fatalf("status = %d, stderr = %q; want user-tag refusal", status, stderr.String())
	}
	if files := chatCommandFiles(t, home); len(files) != 0 {
		t.Fatalf("chat files = %v; want none", files)
	}
	if _, err := os.Stat(filepath.Join(home, ".hive", "state", "chat")); !os.IsNotExist(err) {
		t.Fatalf("chat directory exists after refusal: %v", err)
	}
}
