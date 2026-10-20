package tui

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/process"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

func TestStartupCreatesCoordinatorAndSupervisorTabs(t *testing.T) {
	homePath := newTestHome(t)
	configDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configDir)
	if err := os.MkdirAll(filepath.Join(configDir, "hive"), 0o755); err != nil {
		t.Fatal(err)
	}
	config := "home = " + strconvQuote(homePath) + "\n[coordinator]\nharness = \"pi\"\nmodel = \"cheap\"\n"
	if err := os.WriteFile(filepath.Join(configDir, "hive", "config.toml"), []byte(config), 0o600); err != nil {
		t.Fatal(err)
	}
	server := newStartupHerdr(t)
	t.Setenv("HERDR_ENV", "1")
	t.Setenv("HERDR_PANE_ID", "w1:p1")
	var gotCoordinator bool
	starter := func(_ context.Context, gotHome, paneID, harnessKind, model string, _ *slog.Logger) error {
		if got := requestsFor(server.requests, "pane.send_input"); len(got) != 1 {
			t.Errorf("coordinator started before supervisor: pane.send_input requests = %d; want 1", len(got))
		}
		gotCoordinator = true
		if gotHome != homePath || paneID != "w1:p2" || harnessKind != "pi" || model != "cheap" {
			t.Errorf("coordinator start = %q, %q, %q, %q", gotHome, paneID, harnessKind, model)
		}
		return nil
	}

	_, err := setupSessionWith(homePath, nil, starter)
	if err != nil {
		t.Fatalf("setupSessionWith(): %v", err)
	}
	if !gotCoordinator {
		t.Fatal("coordinator starter was not called")
	}
	requireStartupLabels(t, server.requests, "coordinator", "supervisor")
	renames := requestsFor(server.requests, "tab.rename")
	if len(renames) != 1 || !strings.Contains(renames[0].Params, `"tab_id":"w1:t1"`) || !strings.Contains(renames[0].Params, `"label":"hive"`) {
		t.Fatalf("tab.rename requests = %#v; want the current tab labeled hive", renames)
	}
	for _, request := range requestsFor(server.requests, "tab.create") {
		var params struct {
			Cwd string            `json:"cwd"`
			Env map[string]string `json:"env"`
		}
		if err := json.Unmarshal([]byte(request.Params), &params); err != nil {
			t.Fatalf("decode tab.create params: %v", err)
		}
		if params.Cwd != homePath || params.Env["HIVE_HOME"] != homePath || params.Env["XDG_CONFIG_HOME"] != configDir {
			t.Errorf("tab.create cwd/env = %q/%#v; want home %q and config %q", params.Cwd, params.Env, homePath, configDir)
		}
	}
	inputs := requestsFor(server.requests, "pane.send_input")
	if len(inputs) != 1 || !strings.Contains(inputs[0].Params, "hive supervise") {
		t.Fatalf("pane.send_input requests = %#v; want one hive supervise command", inputs)
	}
}

func TestStartupKeepsLiveCoordinatorAndFreshSupervisor(t *testing.T) {
	homePath := newTestHome(t)
	writeStartupCoordinator(t, homePath)
	writeStartupSupervisor(t, homePath, int64(os.Getpid()), currentProcessStart(t))
	server := newStartupHerdr(t)
	t.Setenv("HERDR_ENV", "1")
	t.Setenv("HERDR_PANE_ID", "w1:p1")
	starter := func(context.Context, string, string, string, string, *slog.Logger) error {
		t.Fatal("startup attempted to replace the live coordinator")
		return nil
	}

	_, err := setupSessionWith(homePath, nil, starter)
	if err != nil {
		t.Fatalf("setupSessionWith(): %v", err)
	}
	if got := requestsFor(server.requests, "tab.create"); len(got) != 0 {
		t.Fatalf("tab.create requests = %#v; want none", got)
	}
	if got := requestsFor(server.requests, "pane.send_input"); len(got) != 0 {
		t.Fatalf("pane.send_input requests = %#v; want none", got)
	}
}

func TestStartupStartsSupervisorWhenClaimIsStale(t *testing.T) {
	homePath := newTestHome(t)
	writeStartupCoordinator(t, homePath)
	writeStartupSupervisor(t, homePath, 99999999, 1)
	server := newStartupHerdr(t)
	t.Setenv("HERDR_ENV", "1")
	t.Setenv("HERDR_PANE_ID", "w1:p1")
	starter := func(context.Context, string, string, string, string, *slog.Logger) error {
		t.Fatal("startup attempted to replace the live coordinator")
		return nil
	}

	_, err := setupSessionWith(homePath, nil, starter)
	if err != nil {
		t.Fatalf("setupSessionWith(): %v", err)
	}
	requireStartupLabels(t, server.requests, "supervisor")
	if got := requestsFor(server.requests, "pane.send_input"); len(got) != 1 || !strings.Contains(got[0].Params, "hive supervise") {
		t.Fatalf("pane.send_input requests = %#v; want one hive supervise command", got)
	}
}

func TestStartupSkipsHerdrWhenEnvironmentIsMissing(t *testing.T) {
	homePath := newTestHome(t)
	server := newStartupHerdr(t)
	t.Setenv("HERDR_ENV", "")
	t.Setenv("HERDR_PANE_ID", "w1:p1")

	notice, err := setupSessionWith(homePath, nil, nil)
	if err != nil {
		t.Fatalf("setupSessionWith(): %v", err)
	}
	if notice != "not in herdr: no coordinator or supervisor started" {
		t.Fatalf("notice = %q", notice)
	}
	if len(server.requests) != 0 {
		t.Fatalf("herdr requests = %#v; want none", server.requests)
	}
}

func TestStartupFailureShowsInFooterAndDoesNotStopTUI(t *testing.T) {
	homePath := newTestHome(t)
	server := newStartupHerdr(t)
	server.failMethod = "pane.get"
	t.Setenv("HERDR_ENV", "1")
	t.Setenv("HERDR_PANE_ID", "w1:p1")
	var output strings.Builder
	if err := RunHome(homePath, strings.NewReader("\x03\x03"), &output); err != nil {
		t.Fatalf("RunHome(): %v", err)
	}
	if !strings.Contains(output.String(), "error:") || !strings.Contains(output.String(), "current pane") {
		t.Fatalf("TUI output %q does not show the startup error", output.String())
	}
	if got := requestsFor(server.requests, "tab.create"); len(got) != 0 {
		t.Fatalf("tab.create requests = %#v; want none after pane lookup failed", got)
	}
}

func TestStartupShowsNoHerdrNoticeWhenPaneIDIsMissing(t *testing.T) {
	homePath := newTestHome(t)
	server := newStartupHerdr(t)
	t.Setenv("HERDR_ENV", "1")
	t.Setenv("HERDR_PANE_ID", "")

	notice, err := setupSessionWith(homePath, nil, nil)
	if err != nil {
		t.Fatalf("setupSessionWith(): %v", err)
	}
	if notice != "not in herdr: no coordinator or supervisor started" || len(server.requests) != 0 {
		t.Fatalf("notice/requests = %q / %#v", notice, server.requests)
	}
}

func writeStartupCoordinator(t *testing.T, homePath string) {
	t.Helper()
	registry, err := store.Open(homePath)
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.WriteIn([]string{"agents"}, "coordinator-g1", records.Agent{Key: "coordinator-g1", Name: "coordinator", Role: "coordinator", Status: records.AgentLive}); err != nil {
		t.Fatal(err)
	}
}

func writeStartupSupervisor(t *testing.T, homePath string, pid int64, start uint64) {
	t.Helper()
	registry, err := store.Open(homePath)
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.WriteIn([]string{"claims"}, "supervisor", records.Claim{Key: "supervisor", Agent: "supervisor-g1", Generation: 1, PID: pid, PIDStart: start}); err != nil {
		t.Fatal(err)
	}
}

func currentProcessStart(t *testing.T) uint64 {
	t.Helper()
	start, err := process.Start(int64(os.Getpid()))
	if err != nil {
		t.Fatal(err)
	}
	return start
}

func requireStartupLabels(t *testing.T, requests []startupRequest, labels ...string) {
	t.Helper()
	created := requestsFor(requests, "tab.create")
	if len(created) != len(labels) {
		t.Fatalf("tab.create requests = %#v; want %d", created, len(labels))
	}
	for index, label := range labels {
		if !strings.Contains(created[index].Params, `"label":"`+label+`"`) || !strings.Contains(created[index].Params, `"workspace_id":"w1"`) {
			t.Errorf("tab.create request %d = %s; want label %q in workspace w1", index, created[index].Params, label)
		}
	}
}

func requestsFor(requests []startupRequest, method string) []startupRequest {
	found := []startupRequest{}
	for _, request := range requests {
		if request.Method == method {
			found = append(found, request)
		}
	}
	return found
}

type startupRequest struct {
	Method string
	Params string
}

type startupHerdr struct {
	t          *testing.T
	listener   net.Listener
	mu         sync.Mutex
	requests   []startupRequest
	tabCount   int
	failMethod string
}

func newStartupHerdr(t *testing.T) *startupHerdr {
	t.Helper()
	path := filepath.Join(t.TempDir(), "herdr.sock")
	listener, err := net.Listen("unix", path)
	if err != nil {
		t.Fatal(err)
	}
	server := &startupHerdr{t: t, listener: listener}
	t.Setenv("HERDR_SOCKET_PATH", path)
	go server.serve()
	t.Cleanup(func() { _ = listener.Close() })
	return server
}

func (s *startupHerdr) serve() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return
		}
		go s.handle(conn)
	}
}

func (s *startupHerdr) handle(conn net.Conn) {
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
		s.t.Errorf("decode herdr request: %v", err)
		return
	}
	s.mu.Lock()
	s.requests = append(s.requests, startupRequest{Method: request.Method, Params: string(request.Params)})
	fail := request.Method == s.failMethod
	result := s.reply(request.Method, request.Params)
	s.mu.Unlock()
	if fail {
		_, _ = fmt.Fprintf(conn, `{"id":%q,"error":{"code":"pane_not_found","message":"pane not found"}}`+"\n", request.ID)
		return
	}
	if result == "" {
		s.t.Errorf("unexpected herdr method %q", request.Method)
		return
	}
	_, _ = fmt.Fprintf(conn, `{"id":%q,"result":%s}`+"\n", request.ID, result)
}

func (s *startupHerdr) reply(method string, params json.RawMessage) string {
	switch method {
	case "ping":
		return `{"type":"pong","version":"0.9.5","protocol":22}`
	case "pane.get":
		var target struct {
			PaneID string `json:"pane_id"`
		}
		if err := json.Unmarshal(params, &target); err != nil {
			s.t.Errorf("decode pane.get: %v", err)
			return ""
		}
		return `{"type":"pane_info","pane":{"pane_id":"` + target.PaneID + `","terminal_id":"term_current","workspace_id":"w1","tab_id":"w1:t1","focused":true,"agent_status":"unknown","revision":1,"cwd":"/tmp/home"}}`
	case "tab.rename":
		var target struct {
			TabID string `json:"tab_id"`
			Label string `json:"label"`
		}
		if err := json.Unmarshal(params, &target); err != nil {
			s.t.Errorf("decode tab.rename: %v", err)
			return ""
		}
		return `{"type":"tab_info","tab":{"tab_id":"` + target.TabID + `","workspace_id":"w1","number":1,"label":"` + target.Label + `","focused":true,"pane_count":1,"agent_status":"unknown"}}`
	case "tab.create":
		var target struct {
			WorkspaceID string `json:"workspace_id"`
			Cwd         string `json:"cwd"`
			Label       string `json:"label"`
		}
		if err := json.Unmarshal(params, &target); err != nil {
			s.t.Errorf("decode tab.create: %v", err)
			return ""
		}
		s.tabCount++
		tabID := fmt.Sprintf("w1:t%d", s.tabCount+1)
		paneID := fmt.Sprintf("w1:p%d", s.tabCount+1)
		return `{"type":"tab_created","tab":{"tab_id":"` + tabID + `","workspace_id":"` + target.WorkspaceID + `","number":` + fmt.Sprint(s.tabCount+1) + `,"label":"` + target.Label + `","focused":false,"pane_count":1,"agent_status":"unknown"},"root_pane":{"pane_id":"` + paneID + `","terminal_id":"term_created","workspace_id":"` + target.WorkspaceID + `","tab_id":"` + tabID + `","focused":false,"agent_status":"unknown","revision":0,"cwd":"` + target.Cwd + `"}}`
	case "pane.send_input":
		return `{"type":"ok"}`
	default:
		return ""
	}
}

func strconvQuote(value string) string {
	data, _ := json.Marshal(value)
	return string(data)
}
