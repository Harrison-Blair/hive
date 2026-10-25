package chat_test

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/Harrison-Blair/hive/cmd"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

type chatRequest struct {
	ID     string          `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

type fakeChatHerdr struct {
	listener    net.Listener
	mu          sync.Mutex
	paneGets    int
	prompts     int
	workspaceID string
}

func newFakeChatHerdr(t *testing.T, workspaceID string) *fakeChatHerdr {
	t.Helper()
	path := filepath.Join(t.TempDir(), "herdr.sock")
	listener, err := net.Listen("unix", path)
	if err != nil {
		t.Fatal(err)
	}
	fake := &fakeChatHerdr{listener: listener, workspaceID: workspaceID}
	t.Setenv("HERDR_SOCKET_PATH", path)
	go fake.serve()
	t.Cleanup(func() { _ = listener.Close(); _ = os.Remove(path) })
	return fake
}

func (f *fakeChatHerdr) serve() {
	for {
		conn, err := f.listener.Accept()
		if err != nil {
			return
		}
		go f.handle(conn)
	}
}

func (f *fakeChatHerdr) handle(conn net.Conn) {
	defer func() { _ = conn.Close() }()
	line, err := bufio.NewReader(conn).ReadBytes('\n')
	if err != nil {
		return
	}
	var request chatRequest
	if err := json.Unmarshal(line, &request); err != nil {
		return
	}
	var result string
	switch request.Method {
	case "ping":
		result = `{"type":"pong","version":"0.9.5","protocol":22}`
	case "pane.get":
		var params struct {
			PaneID string `json:"pane_id"`
		}
		if err := json.Unmarshal(request.Params, &params); err != nil {
			return
		}
		f.mu.Lock()
		f.paneGets++
		workspace := f.workspaceID
		f.mu.Unlock()
		result = fmt.Sprintf(`{"type":"pane_info","pane":{"pane_id":%q,"terminal_id":"term_user","workspace_id":%q,"tab_id":"%s:t1","focused":false,"agent_status":"idle","revision":1}}`, params.PaneID, workspace, workspace)
	case "agent.prompt":
		f.mu.Lock()
		f.prompts++
		f.mu.Unlock()
		result = `{"type":"agent_prompted","agent":{"terminal_id":"term_a","agent_status":"idle","workspace_id":"w1","tab_id":"w1:t1","pane_id":"w1:p1","focused":false,"revision":1}}`
	default:
		return
	}
	_, _ = fmt.Fprintf(conn, `{"id":%q,"result":%s}`+"\n", request.ID, result)
}

func (f *fakeChatHerdr) counts() (paneGets, prompts int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.paneGets, f.prompts
}

func TestChatSendRefusesProjectsWorkingDirectories(t *testing.T) {
	for _, testCase := range []struct {
		name   string
		linked bool
	}{
		{name: "project directory"},
		{name: "projects symlink", linked: true},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			home := newChatCommandHome(t)
			fake := newFakeChatHerdr(t, "w1")
			projectDir := filepath.Join(home, "projects", "demo", "worktrees", "task")
			if testCase.linked {
				targetRoot := t.TempDir()
				projectDir = filepath.Join(targetRoot, "demo", "worktrees", "task")
				if err := os.MkdirAll(projectDir, 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(targetRoot, filepath.Join(home, "projects")); err != nil {
					t.Fatal(err)
				}
			} else if err := os.MkdirAll(projectDir, 0o755); err != nil {
				t.Fatal(err)
			}
			runChatCommandFrom(t, projectDir, home, "")

			var out bytes.Buffer
			err := cmd.ExecuteWithArgs([]string{"chat", "send", "hello"}, strings.NewReader(""), &out)
			if err == nil {
				t.Fatal("chat send succeeded; want a rule 2 refusal")
			}
			if !strings.Contains(out.String(), "chat send is for the user only") || !strings.Contains(out.String(), "rule 2") {
				t.Fatalf("output = %q; want a rule 2 refusal", out.String())
			}
			assertNoChatHistory(t, home)
			if paneGets, prompts := fake.counts(); paneGets != 0 || prompts != 0 {
				t.Fatalf("herdr calls = pane.get %d, agent.prompt %d; want none", paneGets, prompts)
			}
		})
	}
}

func TestChatSendAllowsUserPaneAtHomeRoot(t *testing.T) {
	home := newChatCommandHome(t)
	fake := newFakeChatHerdr(t, "w1")
	runChatCommandFrom(t, home, home, "w1:p2")

	var out bytes.Buffer
	if err := cmd.ExecuteWithArgs([]string{"chat", "send", "hello"}, strings.NewReader(""), &out); err != nil {
		t.Fatalf("chat send: %v: %s", err, out.String())
	}
	if !strings.Contains(out.String(), "chat c1 sent") {
		t.Fatalf("output = %q; want a sent message", out.String())
	}
	files, err := filepath.Glob(filepath.Join(home, ".hive", "state", "chat", "*.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Fatalf("chat history files = %v; want one", files)
	}
	if paneGets, prompts := fake.counts(); paneGets != 1 || prompts != 1 {
		t.Fatalf("herdr calls = pane.get %d, agent.prompt %d; want one each", paneGets, prompts)
	}
}

func newChatCommandHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, ".hive"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".hive", "README.md"), []byte("home\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	coordinator := records.Agent{
		Key: "coordinator-g1", Name: "coordinator", Role: "coordinator", Kind: records.AgentKindCoordinator, Profile: records.ProfileCoordinator, PaneID: "w1:p1",
		WorkspaceID: "w1", Status: records.AgentLive, Generation: 1,
	}
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.WriteIn([]string{"agents"}, coordinator.Key, coordinator); err != nil {
		t.Fatal(err)
	}
	if err := registry.WriteIn([]string{"claims"}, "coordinator", records.Claim{Key: "coordinator", Agent: coordinator.Key, Generation: 1}); err != nil {
		t.Fatal(err)
	}
	return home
}

func runChatCommandFrom(t *testing.T, cwd, home, paneID string) {
	t.Helper()
	previous, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(cwd); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(previous) })
	t.Setenv("HIVE_HOME", home)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("HERDR_PANE_ID", paneID)
}

func assertNoChatHistory(t *testing.T, home string) {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(home, ".hive", "state", "chat", "*.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 0 {
		t.Fatalf("chat history files = %v; want none", files)
	}
}
