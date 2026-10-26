package register_test

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
	"github.com/Harrison-Blair/hive/cmd/common"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

type registerRequest struct {
	Method string
	Params string
}

type registerHerdr struct {
	t        *testing.T
	listener net.Listener
	mu       sync.Mutex
	requests []registerRequest
	pane     string
	rename   string
}

func newRegisterHerdr(t *testing.T, pane, rename string) *registerHerdr {
	t.Helper()
	path := filepath.Join(t.TempDir(), "herdr.sock")
	listener, err := net.Listen("unix", path)
	if err != nil {
		t.Fatal(err)
	}
	fake := &registerHerdr{t: t, listener: listener, pane: pane, rename: rename}
	t.Setenv("HERDR_SOCKET_PATH", path)
	go fake.serve()
	t.Cleanup(func() { _ = listener.Close() })
	return fake
}

func (fake *registerHerdr) serve() {
	for {
		conn, err := fake.listener.Accept()
		if err != nil {
			return
		}
		go fake.handle(conn)
	}
}

func (fake *registerHerdr) handle(conn net.Conn) {
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
	fake.mu.Lock()
	fake.requests = append(fake.requests, registerRequest{Method: request.Method, Params: string(request.Params)})
	fake.mu.Unlock()
	var body string
	switch request.Method {
	case "ping":
		body = `{"type":"pong","version":"0.9.5","protocol":22}`
	case "pane.get":
		body = fake.pane
	case "tab.rename":
		if fake.rename != "" {
			body = `{"error":{"code":"tab_not_found","message":"tab disappeared"}}`
			break
		}
		var params struct {
			TabID string `json:"tab_id"`
			Label string `json:"label"`
		}
		if err := json.Unmarshal(request.Params, &params); err != nil {
			fake.t.Errorf("decode tab.rename params: %v", err)
			return
		}
		workspace := strings.Split(params.TabID, ":")[0]
		body = `{"type":"tab_info","tab":{"tab_id":"` + params.TabID + `","workspace_id":"` + workspace + `","number":1,"label":"` + params.Label + `","focused":false,"pane_count":1,"agent_status":"unknown"}}`
	default:
		fake.t.Errorf("unexpected herdr method %q", request.Method)
		return
	}
	if strings.HasPrefix(body, `{"error":`) {
		body = `{"id":` + quoteRegister(request.ID) + `,"error":` + strings.TrimSuffix(strings.TrimPrefix(body, `{"error":`), `}`) + "}\n"
	} else {
		body = `{"id":` + quoteRegister(request.ID) + `,"result":` + body + "}\n"
	}
	_, _ = conn.Write([]byte(body))
}

func (fake *registerHerdr) recorded(method string) []registerRequest {
	fake.mu.Lock()
	defer fake.mu.Unlock()
	var result []registerRequest
	for _, request := range fake.requests {
		if request.Method == method {
			result = append(result, request)
		}
	}
	return result
}

func quoteRegister(value string) string {
	data, _ := json.Marshal(value)
	return string(data)
}

func registerHome(t *testing.T) string {
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

// seedRegisterTask writes the project and the task of a worker registration.
func seedRegisterTask(t *testing.T, home string) {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.WriteIn([]string{"projects"}, "demo", records.Project{ID: "demo", Path: "/tmp/demo", Base: "dev", RegisteredAt: "2026-01-01T00:00:00Z"}); err != nil {
		t.Fatal(err)
	}
	if err := registry.WriteIn([]string{"tasks"}, "1-demo", records.Task{ID: "1-demo", Project: "demo", Title: "demo", Status: "assigned"}); err != nil {
		t.Fatal(err)
	}
}

func runRegister(t *testing.T, home string, args ...string) (int, string) {
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
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q", stderr.String())
	}
	return status, stdout.String()
}

func TestRegisterCommandRenamesRoleAndProjectLabels(t *testing.T) {
	cases := []struct {
		name     string
		role     string
		paneID   string
		tabID    string
		label    string
		explicit bool
	}{
		{name: "worker uses explicit pane tab", role: "worker", paneID: "w2:p9", tabID: "w2:t9", label: "worker-demo-1", explicit: true},
		{name: "verifier uses environment pane tab", role: "verifier", paneID: "w3:p3", tabID: "w3:t3", label: "verifier-demo-1"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			home := registerHome(t)
			seedRegisterTask(t, home)
			pane := `{"type":"pane_info","pane":{"pane_id":"` + test.paneID + `","terminal_id":"term_a","workspace_id":"` + strings.Split(test.tabID, ":")[0] + `","tab_id":"` + test.tabID + `","focused":false,"agent_status":"unknown","revision":1,"cwd":"/tmp"}}`
			fake := newRegisterHerdr(t, pane, "")
			envPane := test.paneID
			if test.explicit {
				envPane = "w1:p1"
			}
			t.Setenv("HERDR_PANE_ID", envPane)
			t.Setenv("HERDR_WORKSPACE_ID", "w1")
			t.Setenv("HERDR_TAB_ID", "w1:t1")
			args := []string{"--json", "agent", "register", "--name", test.label, "--project", "demo", "--task", "1-demo"}
			if test.explicit {
				args = append(args, "--pane", test.paneID)
			}
			status, stdout := runRegister(t, home, args...)
			if status != 0 {
				t.Fatalf("status = %d; stdout = %q", status, stdout)
			}
			var envelope struct {
				Outcome string `json:"outcome"`
				Error   any    `json:"error"`
			}
			if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
				t.Fatalf("stdout is not JSON: %v: %q", err, stdout)
			}
			if envelope.Outcome != "ok" || envelope.Error != nil {
				t.Fatalf("envelope = %#v", envelope)
			}
			paneRequests := fake.recorded("pane.get")
			if len(paneRequests) != 1 || paneRequests[0].Params != `{"pane_id":"`+test.paneID+`"}` {
				t.Fatalf("pane.get = %#v; want pane %s", paneRequests, test.paneID)
			}
			renames := fake.recorded("tab.rename")
			if len(renames) != 1 || renames[0].Params != `{"tab_id":"`+test.tabID+`","label":"`+test.label+`"}` {
				t.Fatalf("tab.rename = %#v; want tab %s label %s", renames, test.tabID, test.label)
			}
		})
	}
}

func TestRegisterCommandRenameFailureIsAnEffectAndKeepsSuccess(t *testing.T) {
	home := registerHome(t)
	pane := `{"type":"pane_info","pane":{"pane_id":"w2:p9","terminal_id":"term_a","workspace_id":"w2","tab_id":"w2:t9","focused":false,"agent_status":"unknown","revision":1,"cwd":"/tmp"}}`
	seedRegisterTask(t, home)
	fake := newRegisterHerdr(t, pane, "tab disappeared")
	status, stdout := runRegister(t, home, "--json", "agent", "register", "--name", "worker-demo-1", "--project", "demo", "--task", "1-demo", "--pane", "w2:p9")
	if status != 0 {
		t.Fatalf("status = %d; want 0; stdout = %q", status, stdout)
	}
	var envelope struct {
		Outcome string `json:"outcome"`
		Effects []struct {
			Resource string `json:"resource"`
			Path     string `json:"path"`
			Action   string `json:"action"`
		} `json:"effects"`
		Error any `json:"error"`
	}
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatalf("stdout is not JSON: %v: %q", err, stdout)
	}
	if envelope.Outcome != "ok" || envelope.Error != nil {
		t.Fatalf("envelope = %#v; want successful registration", envelope)
	}
	foundFailureEffect := false
	for _, effect := range envelope.Effects {
		if effect.Resource == "tab" && effect.Path == "w2:t9" && effect.Action == "rename failed" {
			foundFailureEffect = true
		}
	}
	if !foundFailureEffect {
		t.Fatalf("effects = %#v; want failed tab rename", envelope.Effects)
	}
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	var record records.Agent
	if err := registry.ReadIn([]string{"agents"}, "worker-demo-1", &record); err != nil {
		t.Fatal(err)
	}
	if record.Status != records.AgentLive {
		t.Fatalf("agent status = %q; want live", record.Status)
	}
	if got := fake.recorded("tab.rename"); len(got) != 1 {
		t.Fatalf("tab.rename requests = %d; want 1", len(got))
	}
}
