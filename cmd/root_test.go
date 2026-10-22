package cmd

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
)

func TestMain(m *testing.M) {
	for _, key := range []string{"HERDR_ENV", "HERDR_PANE_ID", "HERDR_SOCKET_PATH", "HIVE_NO_STARTUP"} {
		if err := os.Unsetenv(key); err != nil {
			panic(err)
		}
	}
	configDir, err := os.MkdirTemp("", "hive-cmd-test-config-")
	if err != nil {
		panic(err)
	}
	if err := os.Setenv("XDG_CONFIG_HOME", configDir); err != nil {
		panic(err)
	}
	code := m.Run()
	_ = os.RemoveAll(configDir)
	os.Exit(code)
}

func TestHomeFlagIsUnknown(t *testing.T) {
	t.Setenv("HIVE_HOME", newRootTestHome(t))
	var stdout, stderr bytes.Buffer
	status := executeWithStreams([]string{"--" + "home", "/tmp/ignored", "version"}, strings.NewReader(""), &stdout, &stderr)
	if status != 2 {
		t.Fatalf("status = %d; want 2; stdout %q, stderr %q", status, stdout.String(), stderr.String())
	}
	if !strings.Contains(stderr.String(), "unknown flag: --"+"home") {
		t.Fatalf("stderr = %q; want unknown flag error", stderr.String())
	}
}

func TestVersionSubcommandPrintsVersion(t *testing.T) {
	var out bytes.Buffer
	if err := ExecuteWithArgs([]string{"version"}, strings.NewReader(""), &out); err != nil {
		t.Fatalf("version: %v", err)
	}
	if got := out.String(); !strings.HasPrefix(got, "hive dev (commit unknown, go") || !strings.HasSuffix(got, ")\n") || strings.Count(got, "\n") != 1 {
		t.Fatalf("output = %q, want one version line", got)
	}
}

func newRootTestHome(t *testing.T) string {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, ".hive"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".hive", "README.md"), []byte("home\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HIVE_HOME", home)
	return home
}

func TestBareHiveRunsTUIUntilQuit(t *testing.T) {
	newRootTestHome(t)
	var out bytes.Buffer
	status := executeWithStreams(nil, strings.NewReader("\x03\x03"), &out, &out)
	if status != 0 {
		t.Fatalf("status = %d; want 0", status)
	}
	// A Bubble Tea program hides the cursor on start.
	if !strings.Contains(out.String(), "\x1b[?25l") {
		t.Fatalf("output %q shows no TUI ran", out.String())
	}
	if !strings.Contains(out.String(), " hive") {
		t.Fatalf("output %q does not show the TUI header", out.String())
	}
	if strings.Contains(out.String(), `"outcome"`) {
		t.Fatalf("successful TUI wrote an envelope: %q", out.String())
	}
}

func TestBareHiveWithoutHerdrShowsStartupNotice(t *testing.T) {
	newRootTestHome(t)
	t.Setenv("HERDR_ENV", "")
	t.Setenv("HERDR_PANE_ID", "")
	var out bytes.Buffer
	status := executeWithStreams(nil, strings.NewReader("\x03\x03"), &out, &out)
	if status != 0 {
		t.Fatalf("status = %d; want 0", status)
	}
	if !strings.Contains(out.String(), "not in herdr: no coordinator or supervisor started") {
		t.Fatalf("output %q lacks the no-herdr startup notice", out.String())
	}
}

func TestBareJSONRefusesBeforeTUI(t *testing.T) {
	newRootTestHome(t)
	var out bytes.Buffer
	status := executeWithStreams([]string{"--json"}, strings.NewReader(""), &out, &out)
	if status != 1 {
		t.Fatalf("status = %d; want 1", status)
	}
	got := out.String()
	if strings.Contains(got, "\x1b[?25l") {
		t.Fatalf("JSON refusal started the TUI: %q", got)
	}
	for _, want := range []string{`"outcome":"failed"`, `"phase":"validate"`, `"error":"TUI requires text output"`} {
		if !strings.Contains(got, want) {
			t.Fatalf("output %q does not contain %s", got, want)
		}
	}
	if strings.Count(got, "\n") != 1 {
		t.Fatalf("output %q is not one envelope line", got)
	}
}

func TestBareHiveNoStartupMakesNoTabCalls(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		env  string
	}{
		{name: "flag", args: []string{"--no-startup"}},
		{name: "variable", env: "1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			newRootTestHome(t)
			methods := newRootFakeHerdr(t)
			t.Setenv("HERDR_ENV", "1")
			t.Setenv("HERDR_PANE_ID", "w1:p1")
			t.Setenv("HIVE_NO_STARTUP", tc.env)
			var out bytes.Buffer
			status := executeWithStreams(tc.args, strings.NewReader("\x03\x03"), &out, &out)
			if status != 0 {
				t.Fatalf("status = %d; want 0; output %q", status, out.String())
			}
			if !strings.Contains(out.String(), " hive") {
				t.Fatalf("output %q does not show the TUI header", out.String())
			}
			for _, method := range methods() {
				if method == "tab.create" || method == "tab.rename" || method == "pane.send_input" {
					t.Fatalf("herdr methods = %v; want no tab or input call", methods())
				}
			}
		})
	}
}

func TestBareHiveOtherNoStartupValueRunsStartup(t *testing.T) {
	newRootTestHome(t)
	methods := newRootFakeHerdr(t)
	t.Setenv("HERDR_ENV", "1")
	t.Setenv("HERDR_PANE_ID", "w1:p1")
	t.Setenv("HIVE_NO_STARTUP", "true")
	var out bytes.Buffer
	_ = executeWithStreams(nil, strings.NewReader("\x03\x03"), &out, &out)
	renamed := false
	for _, method := range methods() {
		if method == "tab.rename" {
			renamed = true
		}
	}
	if !renamed {
		t.Fatalf("herdr methods = %v; want startup to rename the tab", methods())
	}
}

// newRootFakeHerdr serves the herdr socket and returns the methods it got.
// It answers ping, pane.get, and tab.rename, and refuses each other method.
func newRootFakeHerdr(t *testing.T) func() []string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "herdr.sock")
	listener, err := net.Listen("unix", path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	t.Setenv("HERDR_SOCKET_PATH", path)
	var mu sync.Mutex
	methods := []string{}
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go func() {
				defer func() { _ = conn.Close() }()
				line, err := bufio.NewReader(conn).ReadBytes('\n')
				if err != nil {
					return
				}
				var request struct {
					ID     string `json:"id"`
					Method string `json:"method"`
				}
				if err := json.Unmarshal(line, &request); err != nil {
					return
				}
				mu.Lock()
				methods = append(methods, request.Method)
				mu.Unlock()
				switch request.Method {
				case "ping":
					_, _ = fmt.Fprintf(conn, `{"id":%q,"result":{"type":"pong","version":"0.9.5","protocol":22}}`+"\n", request.ID)
				case "pane.get":
					_, _ = fmt.Fprintf(conn, `{"id":%q,"result":{"type":"pane_info","pane":{"pane_id":"w1:p1","terminal_id":"term_1","workspace_id":"w1","tab_id":"w1:t1","focused":true,"agent_status":"unknown","revision":1,"cwd":"/tmp"}}}`+"\n", request.ID)
				case "tab.rename":
					_, _ = fmt.Fprintf(conn, `{"id":%q,"result":{"type":"tab_info","tab":{"tab_id":"w1:t1","workspace_id":"w1","number":1,"label":"hive","focused":true,"pane_count":1,"agent_status":"unknown"}}}`+"\n", request.ID)
				default:
					_, _ = fmt.Fprintf(conn, `{"id":%q,"error":{"code":"refused","message":"fake herdr refuses %s"}}`+"\n", request.ID, request.Method)
				}
			}()
		}
	}()
	return func() []string {
		mu.Lock()
		defer mu.Unlock()
		return append([]string(nil), methods...)
	}
}
