package agent

import (
	"bufio"
	"encoding/json"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/process"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

type fakeReply struct {
	result string
	error  string
	close  bool
}

type fakeRequest struct {
	ID     string
	Method string
	Params string
}

type fakeHerdr struct {
	t        *testing.T
	listener net.Listener
	path     string
	mu       sync.Mutex
	replies  map[string][]fakeReply
	requests []fakeRequest
	before   func(method string, params json.RawMessage)
	closed   bool
}

func newFakeHerdr(t *testing.T) *fakeHerdr {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "herdr.sock")
	listener, err := net.Listen("unix", path)
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	fake := &fakeHerdr{t: t, listener: listener, path: path, replies: map[string][]fakeReply{}}
	t.Setenv("HERDR_SOCKET_PATH", path)
	go fake.serve()
	t.Cleanup(fake.close)
	return fake
}

func (f *fakeHerdr) close() {
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

func (f *fakeHerdr) result(method, result string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.replies[method] = append(f.replies[method], fakeReply{result: result})
}

func (f *fakeHerdr) failure(method, code, message string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.replies[method] = append(f.replies[method], fakeReply{error: `{"code":` + jsonQuote(code) + `,"message":` + jsonQuote(message) + `}`})
}

func (f *fakeHerdr) drop(method string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.replies[method] = append(f.replies[method], fakeReply{close: true})
}

func (f *fakeHerdr) setBefore(fn func(method string, params json.RawMessage)) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.before = fn
}

func (f *fakeHerdr) serve() {
	for {
		conn, err := f.listener.Accept()
		if err != nil {
			return
		}
		go f.handle(conn)
	}
}

func (f *fakeHerdr) handle(conn net.Conn) {
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
	f.requests = append(f.requests, fakeRequest{ID: request.ID, Method: request.Method, Params: string(request.Params)})
	before := f.before
	var reply fakeReply
	found := false
	if request.Method == "ping" {
		reply = fakeReply{result: `{"type":"pong","version":"0.9.5","protocol":22}`}
		found = true
	} else if list := f.replies[request.Method]; len(list) > 0 {
		reply = list[0]
		f.replies[request.Method] = list[1:]
		found = true
	}
	f.mu.Unlock()
	if before != nil {
		before(request.Method, request.Params)
	}
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

func (f *fakeHerdr) recorded(method string) []fakeRequest {
	f.mu.Lock()
	defer f.mu.Unlock()
	found := []fakeRequest{}
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

func testHome(t *testing.T) string {
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

func writeRecord(t *testing.T, home string, record records.Agent) {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatalf("store open: %v", err)
	}
	if err := registry.WriteIn([]string{"agents"}, record.Key, record); err != nil {
		t.Fatalf("write agent: %v", err)
	}
}

func writeClaimRecord(t *testing.T, home string, claim records.Claim) {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatalf("store open: %v", err)
	}
	if err := registry.WriteIn([]string{"claims"}, claim.Key, claim); err != nil {
		t.Fatalf("write claim: %v", err)
	}
}

func writeProject(t *testing.T, home, id string) {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatalf("store open: %v", err)
	}
	record := records.Project{ID: id, Path: "/tmp/" + id, Base: "dev", RegisteredAt: "2026-01-01T00:00:00Z"}
	if err := registry.WriteIn([]string{"projects"}, id, record); err != nil {
		t.Fatalf("write project: %v", err)
	}
}

func writeTask(t *testing.T, home, project, task string) {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatalf("store open: %v", err)
	}
	if err := registry.WriteIn([]string{"tasks"}, task, records.Task{ID: task, Project: project, Status: "ready"}); err != nil {
		t.Fatalf("write task: %v", err)
	}
}

type sleepProcess struct {
	command *exec.Cmd
	pid     int64
	start   uint64
}

func startSleep(t *testing.T) sleepProcess {
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
	return sleepProcess{command: command, pid: pid, start: start}
}

func agentJSON(paneID, terminalID, workspaceID, tabID, status, name string) string {
	return `{"pane_id":"` + paneID + `","terminal_id":"` + terminalID + `","workspace_id":"` + workspaceID + `","tab_id":"` + tabID + `","agent_status":"` + status + `","focused":false,"revision":2,"name":"` + name + `"}`
}

func agentJSONSession(paneID, terminalID, workspaceID, tabID, status, name, source, kind, value string) string {
	return `{"pane_id":"` + paneID + `","terminal_id":"` + terminalID + `","workspace_id":"` + workspaceID + `","tab_id":"` + tabID + `","agent_status":"` + status + `","focused":false,"revision":2,"name":"` + name + `","agent_session":{"source":"` + source + `","agent":"claude","kind":"` + kind + `","value":"` + value + `"}}`
}

func paneInfoJSON(paneID, terminalID, workspaceID, tabID, cwd string) string {
	return `{"pane_id":"` + paneID + `","terminal_id":"` + terminalID + `","workspace_id":"` + workspaceID + `","tab_id":"` + tabID + `","focused":false,"agent_status":"unknown","revision":1,"cwd":"` + cwd + `"}`
}

func scriptReadyShell(fake *fakeHerdr, paneID, terminalID, workspaceID, tabID, cwd string) {
	fake.result("pane.get", `{"type":"pane_info","pane":`+paneInfoJSON(paneID, terminalID, workspaceID, tabID, cwd)+`}`)
	fake.result("pane.process_info", `{"type":"pane_process_info","process_info":{"pane_id":"`+paneID+`","shell_pid":1,"foreground_process_group_id":1,"foreground_processes":[`+processEntry(1, "bash", "/usr/bin/bash")+`]}}`)
}

func processInfoJSON(paneID string, entries ...string) string {
	body := `{"pane_id":"` + paneID + `","shell_pid":null,"foreground_process_group_id":null,"foreground_processes":[`
	for index, entry := range entries {
		if index > 0 {
			body += ","
		}
		body += entry
	}
	return body + `]}`
}

func processEntry(pid int64, name string, argv ...string) string {
	argvJSON, _ := json.Marshal(argv)
	return `{"pid":` + jsonNumber(pid) + `,"name":"` + name + `","argv":` + string(argvJSON) + `,"cmdline":"` + name + `","cwd":"/tmp/x"}`
}

func jsonNumber(value int64) string {
	data, _ := json.Marshal(value)
	return string(data)
}

func workspaceJSON(id string, paneCount int) string {
	return `{"workspace_id":"` + id + `","number":1,"label":"x","focused":false,"pane_count":` + jsonNumber(int64(paneCount)) + `,"tab_count":1,"active_tab_id":"` + id + `:t1","agent_status":"unknown"}`
}
