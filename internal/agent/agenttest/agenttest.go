// Package agenttest holds the fake herdr server and the fixture helpers that
// the tests of several agent operation packages share. Only tests import it.
package agenttest

import (
	"bufio"
	"encoding/json"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"unicode/utf8"

	"github.com/Harrison-Blair/hive/internal/lib/process"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

type fakeReply struct {
	result string
	error  string
	close  bool
}

// FakeRequest is one request that the fake herdr server received.
type FakeRequest struct {
	ID     string
	Method string
	Params string
}

// FakeHerdr serves scripted herdr replies on a test socket.
type FakeHerdr struct {
	t        *testing.T
	listener net.Listener
	path     string
	mu       sync.Mutex
	replies  map[string][]fakeReply
	requests []FakeRequest
	closed   bool
}

// NewFakeHerdr starts a fake herdr server and points HERDR_SOCKET_PATH at it.
func NewFakeHerdr(t *testing.T) *FakeHerdr {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "herdr.sock")
	listener, err := net.Listen("unix", path)
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	fake := &FakeHerdr{t: t, listener: listener, path: path, replies: map[string][]fakeReply{}}
	t.Setenv("HERDR_SOCKET_PATH", path)
	go fake.serve()
	t.Cleanup(fake.close)
	return fake
}

func (f *FakeHerdr) close() {
	f.mu.Lock()
	if f.closed {
		f.mu.Unlock()
		return
	}
	f.closed = true
	f.mu.Unlock()
	_ = f.listener.Close()
	_ = os.Remove(f.path)
}

// Result queues one result reply for method.
func (f *FakeHerdr) Result(method, result string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.replies[method] = append(f.replies[method], fakeReply{result: result})
}

// Failure queues one error reply for method.
func (f *FakeHerdr) Failure(method, code, message string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.replies[method] = append(f.replies[method], fakeReply{error: `{"code":` + jsonQuote(code) + `,"message":` + jsonQuote(message) + `}`})
}

// Drop queues one closed connection for method.
func (f *FakeHerdr) Drop(method string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.replies[method] = append(f.replies[method], fakeReply{close: true})
}

func (f *FakeHerdr) serve() {
	for {
		conn, err := f.listener.Accept()
		if err != nil {
			return
		}
		go f.handle(conn)
	}
}

func (f *FakeHerdr) handle(conn net.Conn) {
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
	f.requests = append(f.requests, FakeRequest{ID: request.ID, Method: request.Method, Params: string(request.Params)})
	var reply fakeReply
	found := false
	if request.Method == "ping" {
		reply = fakeReply{result: `{"type":"pong","version":"0.9.5","protocol":22}`}
		found = true
	} else if list := f.replies[request.Method]; len(list) > 0 {
		reply = list[0]
		f.replies[request.Method] = list[1:]
		found = true
	} else if request.Method == "tab.rename" {
		var params struct {
			TabID string `json:"tab_id"`
			Label string `json:"label"`
		}
		if err := json.Unmarshal(request.Params, &params); err != nil {
			f.t.Errorf("decode tab.rename params: %v", err)
			return
		}
		workspaceID := strings.Split(params.TabID, ":")[0]
		reply = fakeReply{result: `{"type":"tab_info","tab":{"tab_id":"` + params.TabID + `","workspace_id":"` + workspaceID + `","number":1,"label":"` + params.Label + `","focused":false,"pane_count":1,"agent_status":"unknown"}}`}
		found = true
	}
	f.mu.Unlock()
	if !found {
		f.t.Errorf("unexpected method %q", request.Method)
		return
	}
	if reply.close {
		return
	}
	body := `{"id":` + jsonQuote(request.ID)
	if reply.error != "" {
		body += `,"error":` + reply.error + "}\n"
	} else {
		body += `,"result":` + reply.result + "}\n"
	}
	_, _ = conn.Write([]byte(body))
}

// Recorded gives each received request of method.
func (f *FakeHerdr) Recorded(method string) []FakeRequest {
	f.mu.Lock()
	defer f.mu.Unlock()
	found := []FakeRequest{}
	for _, request := range f.requests {
		if request.Method == method {
			found = append(found, request)
		}
	}
	return found
}

func jsonQuote(value string) string {
	data, _ := json.Marshal(value)
	return string(data)
}

// NewHome makes one fixture home with its marker.
func NewHome(t *testing.T) string {
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

// WriteRecord writes one agent record.
func WriteRecord(t *testing.T, home string, record records.Agent) {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatalf("store open: %v", err)
	}
	if err := registry.WriteIn([]string{"agents"}, record.Key, record); err != nil {
		t.Fatalf("write agent: %v", err)
	}
}

// WriteClaimRecord writes one claim record.
func WriteClaimRecord(t *testing.T, home string, claim records.Claim) {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatalf("store open: %v", err)
	}
	if err := registry.WriteIn([]string{"claims"}, claim.Key, claim); err != nil {
		t.Fatalf("write claim: %v", err)
	}
}

// SleepProcess is a live child process with its start identity.
type SleepProcess struct {
	command *exec.Cmd
	PID     int64
	Start   uint64
}

// StartSleep starts a sleep process that the test cleanup ends.
func StartSleep(t *testing.T) SleepProcess {
	t.Helper()
	command := exec.Command("sleep", "60")
	if err := command.Start(); err != nil {
		t.Fatalf("start sleep: %v", err)
	}
	pid := int64(command.Process.Pid)
	start, err := process.Start(pid)
	if err != nil {
		_ = command.Process.Kill()
		_, _ = command.Process.Wait()
		t.Fatalf("process start: %v", err)
	}
	cleanup := func() {
		_ = command.Process.Kill()
		_, _ = command.Process.Wait()
	}
	t.Cleanup(cleanup)
	return SleepProcess{command: command, PID: pid, Start: start}
}

// AgentJSON gives one herdr agent object.
func AgentJSON(paneID, terminalID, workspaceID, tabID, status, name string) string {
	return `{"pane_id":"` + paneID + `","terminal_id":"` + terminalID + `","workspace_id":"` + workspaceID + `","tab_id":"` + tabID + `","agent_status":"` + status + `","focused":false,"revision":2,"name":"` + name + `"}`
}

// PaneInfoJSON gives one herdr pane object.
func PaneInfoJSON(paneID, terminalID, workspaceID, tabID, cwd string) string {
	return `{"pane_id":"` + paneID + `","terminal_id":"` + terminalID + `","workspace_id":"` + workspaceID + `","tab_id":"` + tabID + `","focused":false,"agent_status":"unknown","revision":1,"cwd":"` + cwd + `"}`
}

// ProcessInfoJSON gives one process info object with no shell PID.
func ProcessInfoJSON(paneID string, entries ...string) string {
	body := `{"pane_id":"` + paneID + `","shell_pid":null,"foreground_process_group_id":null,"foreground_processes":[`
	for index, entry := range entries {
		if index > 0 {
			body += ","
		}
		body += entry
	}
	return body + `]}`
}

// ProcessEntry gives one foreground process object.
func ProcessEntry(pid int64, name string, argv ...string) string {
	argvJSON, _ := json.Marshal(argv)
	return `{"pid":` + jsonNumber(pid) + `,"name":"` + name + `","argv":` + string(argvJSON) + `,"cmdline":"` + name + `","cwd":"/tmp/x"}`
}

func jsonNumber(value int64) string {
	data, _ := json.Marshal(value)
	return string(data)
}

// ProcessInfoShellJSON gives one process info object with a shell PID.
func ProcessInfoShellJSON(paneID string, shellPID int64, entries ...string) string {
	body := `{"pane_id":"` + paneID + `","shell_pid":` + jsonNumber(shellPID) + `,"foreground_process_group_id":null,"foreground_processes":[`
	for index, entry := range entries {
		if index > 0 {
			body += ","
		}
		body += entry
	}
	return body + `]}`
}

// LiveRecord gives one live spawned builder record in pane w1:p1.
func LiveRecord(key, name string) records.Agent {
	return records.Agent{
		Key:     key,
		Name:    name,
		Profile: records.ProfileBuilder,
		Project: "hive",
		Task:    "12-spawn-worker",
		Harness: "claude",
		PaneID:  "w1:p1",
		Status:  records.AgentLive,
		Spawned: true,
	}
}

// ReadAgent reads one agent record.
func ReadAgent(t *testing.T, home, key string) records.Agent {
	t.Helper()
	st, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	var record records.Agent
	if err := st.ReadIn([]string{"agents"}, key, &record); err != nil {
		t.Fatalf("read agent %s: %v", key, err)
	}
	return record
}

// ReadLog gives the text of each daily log file of the home.
func ReadLog(t *testing.T, home string) string {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(home, ".hive", "logs", "*.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	text := ""
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		text += string(data)
	}
	return text
}

// ToString gives a string value as is and encodes each other value.
func ToString(value any) string {
	if value == nil {
		return ""
	}
	if text, ok := value.(string); ok {
		return text
	}
	data, _ := json.Marshal(value)
	return string(data)
}

// WriteProjectPath writes one project record with its path.
func WriteProjectPath(t *testing.T, home, id, path string) {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	record := records.Project{ID: id, Path: path, Base: "dev", RegisteredAt: "2026-01-01T00:00:00Z"}
	if err := registry.WriteIn([]string{"projects"}, id, record); err != nil {
		t.Fatal(err)
	}
}

// WriteTaskRecord writes one task record.
func WriteTaskRecord(t *testing.T, home string, task records.Task) {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.WriteIn([]string{"tasks"}, task.ID, task); err != nil {
		t.Fatal(err)
	}
}

// LogEvents decodes each log event of the home.
func LogEvents(t *testing.T, home string) []map[string]any {
	t.Helper()
	events := []map[string]any{}
	for _, line := range strings.Split(strings.TrimSpace(ReadLog(t, home)), "\n") {
		if line == "" {
			continue
		}
		var event map[string]any
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			t.Fatalf("log line = %q: %v", line, err)
		}
		events = append(events, event)
	}
	return events
}

// FindLogEvent gives the first event with message.
func FindLogEvent(t *testing.T, events []map[string]any, message string) map[string]any {
	t.Helper()
	for _, event := range events {
		if event["msg"] == message {
			return event
		}
	}
	t.Fatalf("log has no %q event", message)
	return nil
}

// CheckBody checks that the event body is a bounded, truncated prefix.
func CheckBody(t *testing.T, event map[string]any, prefix string) {
	t.Helper()
	body, ok := event["body"].(string)
	if !ok {
		t.Fatalf("event body = %#v; want a string", event["body"])
	}
	if len(body) > 4096 {
		t.Fatalf("body bytes = %d; want at most 4096", len(body))
	}
	if !utf8.ValidString(body) {
		t.Fatalf("body is not valid UTF-8")
	}
	if !strings.HasPrefix(prefix, body) {
		t.Fatalf("body is not a prefix of the source text")
	}
	if event["truncated"] != true {
		t.Fatalf("truncated = %#v; want true", event["truncated"])
	}
}

// ResolveSeedRecord gives one unresolved spawned builder record.
func ResolveSeedRecord() records.Agent {
	return records.Agent{
		Key: "worker-3", Name: "worker-3", Profile: records.ProfileBuilder, Project: "one", Task: "1-task",
		Harness: "claude", Status: records.AgentUnresolved, Spawned: true, Generation: 1,
		PaneID: "w1:p1", WorkspaceID: "w1", TabID: "w1:t1", TerminalID: "term_a", Cwd: "/tmp/old",
		StartedAt: "2026-01-01T00:00:00Z", UpdatedAt: "2026-01-01T00:00:00Z",
	}
}
