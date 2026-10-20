package spawn_test

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/Harrison-Blair/hive/cmd"
	"github.com/Harrison-Blair/hive/cmd/agent/register"
	"github.com/Harrison-Blair/hive/cmd/common"
)

type cliFake struct {
	t        *testing.T
	listener net.Listener
	mu       sync.Mutex
	replies  map[string][]string
	requests []string
}

func newCLIFake(t *testing.T, method string, replies ...string) *cliFake {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "herdr.sock")
	listener, err := net.Listen("unix", path)
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	fake := &cliFake{t: t, listener: listener, replies: map[string][]string{
		"pane.get":          {`{"type":"pane_info","pane":{"pane_id":"w1:p1","terminal_id":"term_a","workspace_id":"w1","tab_id":"w1:t1","focused":false,"agent_status":"unknown","revision":1,"cwd":"/tmp"}}`},
		"pane.process_info": {`{"type":"pane_process_info","process_info":{"pane_id":"w1:p1","shell_pid":1,"foreground_process_group_id":1,"foreground_processes":[{"pid":1,"name":"bash","argv":["/usr/bin/bash"],"cmdline":"/usr/bin/bash","cwd":"/tmp"}]}}`},
	}}
	fake.replies[method] = replies
	t.Setenv("HERDR_SOCKET_PATH", path)
	go fake.serve()
	t.Cleanup(func() { _ = listener.Close() })
	return fake
}

func (f *cliFake) serve() {
	for {
		conn, err := f.listener.Accept()
		if err != nil {
			return
		}
		go f.handle(conn)
	}
}

func (f *cliFake) handle(conn net.Conn) {
	defer func() { _ = conn.Close() }()
	line, err := bufio.NewReader(conn).ReadBytes('\n')
	if err != nil {
		return
	}
	var request struct {
		ID     string          `json:"id"`
		Method string          `json:"method"`
		Params json.RawMessage `json:"params"`
	}
	if err := json.Unmarshal(line, &request); err != nil {
		return
	}
	f.mu.Lock()
	f.requests = append(f.requests, request.Method+" "+string(request.Params))
	var reply string
	if request.Method == "ping" {
		reply = `{"type":"pong","version":"0.9.5","protocol":22}`
	} else if list := f.replies[request.Method]; len(list) > 0 {
		reply = list[0]
		f.replies[request.Method] = list[1:]
	}
	f.mu.Unlock()
	if reply == "" {
		return
	}
	body := `{"id":` + quote(request.ID)
	if strings.HasPrefix(reply, `{"error":`) {
		body += `,"error":` + strings.TrimSuffix(strings.TrimPrefix(reply, `{"error":`), `}`) + "}\n"
	} else {
		body += `,"result":` + reply + "}\n"
	}
	_, _ = conn.Write([]byte(body))
}

func (f *cliFake) requestText() string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return strings.Join(f.requests, "\n")
}

func quote(value string) string {
	data, _ := json.Marshal(value)
	return string(data)
}

func cliHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, ".hive"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".hive", "README.md"), []byte("fixture\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return home
}

func runCLI(t *testing.T, home string, args ...string) (int, string, string) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("HIVE_HOME", home)
	root := cmd.NewRootCmd()
	var stdout, stderr bytes.Buffer
	root.SetArgs(args)
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	err := root.Execute()
	status := 0
	if err != nil {
		var statusError *common.StatusError
		if errors.As(err, &statusError) {
			status = statusError.Status
		} else {
			t.Fatalf("execute error = %v", err)
		}
	}
	return status, stdout.String(), stderr.String()
}

func TestSpawnJSONOneObjectAndArgs(t *testing.T) {
	home := cliHome(t)
	fake := newCLIFake(t, "agent.start", `{"error":{"code":"invalid_agent_name","message":"name is too long"}}`)
	status, stdout, stderr := runCLI(t, home, "--json", "agent", "spawn", "--pane", "w1:p1", "--role", "coordinator", "--harness", "claude", "--model", "m", "--name", "t10-cli", "--arg", "--one", "--arg", "two")
	if status != 1 {
		t.Fatalf("status = %d; want 1 (stderr %q)", status, stderr)
	}
	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) != 1 {
		t.Fatalf("stdout lines = %d; want one JSON object: %q", len(lines), stdout)
	}
	var envelope struct {
		Outcome string `json:"outcome"`
		Phase   string `json:"phase"`
		Effects []any  `json:"effects"`
		Error   any    `json:"error"`
		Result  struct {
			Agent struct {
				Name    string `json:"name"`
				Key     string `json:"key"`
				Spawned bool   `json:"spawned"`
				Status  string `json:"status"`
			} `json:"agent"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(lines[0]), &envelope); err != nil {
		t.Fatalf("stdout is not one JSON object: %v: %q", err, lines[0])
	}
	if envelope.Outcome != "failed" || envelope.Phase != "start" {
		t.Fatalf("envelope = %#v", envelope)
	}
	if envelope.Result.Agent.Name != "t10-cli" || envelope.Result.Agent.Key != "t10-cli-g1" {
		t.Fatalf("agent name/key = %q/%q", envelope.Result.Agent.Name, envelope.Result.Agent.Key)
	}
	if !envelope.Result.Agent.Spawned {
		t.Fatalf("spawned = false; want true")
	}
	if envelope.Result.Agent.Status != "exited" {
		t.Fatalf("status = %q; want exited", envelope.Result.Agent.Status)
	}
	request := fake.requestText()
	if !strings.Contains(request, `"args":["--dangerously-skip-permissions","--model","m","--one","two"]`) {
		t.Fatalf("start request = %s; want repeated --arg values", request)
	}
}

func TestSpawnTextDefaultAndExit(t *testing.T) {
	home := cliHome(t)
	status, stdout, stderr := runCLI(t, home, "agent", "spawn", "--pane", "w1:p1", "--role", "coordinator", "--harness", "nope")
	if status != 1 {
		t.Fatalf("status = %d; want 1", status)
	}
	if stdout != "" {
		t.Fatalf("stdout = %q; want empty text on failure", stdout)
	}
	if !strings.Contains(stderr, "unknown harness: nope") {
		t.Fatalf("stderr = %q", stderr)
	}
}

func TestSpawnPartialExitIsNonzero(t *testing.T) {
	home := cliHome(t)
	newCLIFake(t, "agent.start", "")
	status, stdout, _ := runCLI(t, home, "--json", "agent", "spawn", "--pane", "w1:p1", "--role", "coordinator", "--harness", "claude")
	if status != 1 {
		t.Fatalf("status = %d; want 1 for partial", status)
	}
	if !strings.Contains(stdout, `"outcome":"partial"`) {
		t.Fatalf("stdout = %q; want partial", stdout)
	}
}

func TestSpawnWriterErrorIsObservable(t *testing.T) {
	home := cliHome(t)
	newCLIFake(t, "agent.start", `{"error":{"code":"invalid_agent_name","message":"bad"}}`)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("HIVE_HOME", home)
	root := cmd.NewRootCmd()
	root.SetArgs([]string{"--json", "agent", "spawn", "--pane", "w1:p1", "--role", "coordinator", "--harness", "claude"})
	root.SetOut(failingWriter{})
	root.SetErr(&bytes.Buffer{})
	err := root.Execute()
	var statusError *common.StatusError
	if !errors.As(err, &statusError) || statusError.Status != 1 {
		t.Fatalf("execute error = %v; want status 1", err)
	}
}

func TestRegisterHasNoSpawnedFlag(t *testing.T) {
	if register.New().Flags().Lookup("spawned") != nil {
		t.Fatalf("agent register has a --spawned flag")
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("write failed") }
