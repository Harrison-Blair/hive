package resolve_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/cmd"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

func startFakeHerdr(t *testing.T, handler func(method, params string) (string, string)) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "herdr.sock")
	listener, err := net.Listen("unix", path)
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go func() {
				defer func() { _ = conn.Close() }()
				var request struct {
					ID     string          `json:"id"`
					Method string          `json:"method"`
					Params json.RawMessage `json:"params"`
				}
				if err := json.NewDecoder(conn).Decode(&request); err != nil {
					return
				}
				if request.Method == "ping" {
					_, _ = fmt.Fprintf(conn, `{"id":%q,"result":{"type":"pong","version":"0.9.5","protocol":22}}`+"\n", request.ID)
					return
				}
				result, errText := handler(request.Method, string(request.Params))
				if errText != "" {
					_, _ = fmt.Fprintf(conn, `{"id":%q,"error":%s}`+"\n", request.ID, errText)
					return
				}
				_, _ = fmt.Fprintf(conn, `{"id":%q,"result":%s}`+"\n", request.ID, result)
			}()
		}
	}()
	t.Cleanup(func() { _ = listener.Close() })
	t.Setenv("HERDR_SOCKET_PATH", path)
}

func newResolveCommandHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, ".hive"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".hive", "README.md"), []byte("home\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return home
}

func seedResolveCommandRecord(t *testing.T, home, key, paneID string) records.Agent {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatalf("store open: %v", err)
	}
	record := records.Agent{
		Key: key, Name: key, Profile: records.ProfileBuilder, Project: "one", Task: "1-task",
		Harness: "claude", Status: records.AgentUnresolved, Spawned: true, Generation: 1,
		PaneID: paneID, WorkspaceID: "w1", TabID: "w1:t1", TerminalID: "term_a", Cwd: "/tmp/old",
		StartedAt: "2026-01-01T00:00:00Z", UpdatedAt: "2026-01-01T00:00:00Z",
	}
	if err := registry.WriteIn([]string{"agents"}, key, record); err != nil {
		t.Fatalf("write agent: %v", err)
	}
	return record
}

func runResolveCommand(t *testing.T, home string, args ...string) (string, error) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("HIVE_HOME", home)
	var out bytes.Buffer
	err := cmd.ExecuteWithArgs(args, strings.NewReader(""), &out)
	return out.String(), err
}

func TestResolveCommand(t *testing.T) {
	home := newResolveCommandHome(t)
	seedResolveCommandRecord(t, home, "worker-live", "w1:p1")
	seedResolveCommandRecord(t, home, "worker-exited", "w1:p2")
	startFakeHerdr(t, func(method, _ string) (string, string) {
		if method == "agent.get" {
			return `{"type":"agent_info","agent":{"pane_id":"w2:p5","terminal_id":"term_a","workspace_id":"w2","tab_id":"w2:t3","agent_status":"idle","focused":false,"revision":2,"name":"worker-live","agent":"claude","cwd":"/tmp/old"}}`, ""
		}
		return "", `{"code":"unknown_method","message":"unexpected method"}`
	})

	listOut, err := runResolveCommand(t, home, "agent", "list")
	if err != nil {
		t.Fatalf("agent list: %v: %s", err, listOut)
	}
	wantList := "worker-exited\t1\tbuilder\tunresolved\tone\t1-task\tw1:p2\nworker-live\t1\tbuilder\tunresolved\tone\t1-task\tw1:p1\n"
	if listOut != wantList {
		t.Fatalf("agent list = %q, want %q", listOut, wantList)
	}

	stopOut, stopErr := runResolveCommand(t, home, "agent", "stop", "--agent", "worker-live")
	if stopErr == nil || stopErr.Error() != "hive exited with status 1" {
		t.Fatalf("agent stop error = %v, want status 1", stopErr)
	}
	if stopOut != "hive: agent unresolved: run hive agent resolve\n" {
		t.Fatalf("agent stop = %q, want the unresolved refusal", stopOut)
	}

	doctorOut, doctorErr := runResolveCommand(t, home, "doctor")
	if doctorErr != nil {
		t.Fatalf("doctor: %v: %s", doctorErr, doctorOut)
	}
	if !strings.Contains(doctorOut, "warn unresolved agents: 2\n") {
		t.Fatalf("doctor = %q, want the unresolved agents warning", doctorOut)
	}

	liveOut, liveErr := runResolveCommand(t, home, "agent", "resolve", "--agent", "worker-live", "--as", "live")
	if liveErr != nil {
		t.Fatalf("agent resolve live: %v: %s", liveErr, liveOut)
	}
	wantLive := "resolved agent .hive/state/agents/worker-live.json\nagent worker-live live\n"
	if liveOut != wantLive {
		t.Fatalf("agent resolve live = %q, want %q", liveOut, wantLive)
	}
	live := readResolveCommandRecord(t, home, "worker-live")
	if live.Status != records.AgentLive || live.PaneID != "w2:p5" || live.WorkspaceID != "w2" || live.TabID != "w2:t3" || live.TerminalID != "term_a" {
		t.Fatalf("live record = %+v, want the live ids", live)
	}

	exitedOut, exitedErr := runResolveCommand(t, home, "agent", "resolve", "--agent", "worker-exited", "--as", "exited")
	if exitedErr != nil {
		t.Fatalf("agent resolve exited: %v: %s", exitedErr, exitedOut)
	}
	wantExited := "resolved agent .hive/state/agents/worker-exited.json\nagent worker-exited exited\n"
	if exitedOut != wantExited {
		t.Fatalf("agent resolve exited = %q, want %q", exitedOut, wantExited)
	}
	if exited := readResolveCommandRecord(t, home, "worker-exited"); exited.Status != records.AgentExited || exited.PaneID != "w1:p2" {
		t.Fatalf("exited record = %+v, want exited and the old pane", exited)
	}
}

func readResolveCommandRecord(t *testing.T, home, key string) records.Agent {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatalf("store open: %v", err)
	}
	var record records.Agent
	if err := registry.ReadIn([]string{"agents"}, key, &record); err != nil {
		t.Fatalf("read agent: %v", err)
	}
	return record
}
