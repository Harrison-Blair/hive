package doctor

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	hivelog "github.com/Harrison-Blair/hive/internal/lib/log"
)

func newHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, ".hive"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".hive", "README.md"), []byte("# home\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return home
}

func startHerdr(t *testing.T, protocol int) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "herdr.sock")
	listener, err := net.Listen("unix", path)
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan struct{})
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go func() {
				defer func() { _ = conn.Close() }()
				var request struct {
					ID     string `json:"id"`
					Method string `json:"method"`
				}
				if err := json.NewDecoder(conn).Decode(&request); err != nil {
					return
				}
				if request.Method != "ping" {
					_, _ = fmt.Fprintf(conn, `{"id":%q,"result":{"type":"ok"}}`+"\n", request.ID)
					return
				}
				_, _ = fmt.Fprintf(conn, `{"id":%q,"result":{"type":"pong","version":"0.9.5","protocol":%d}}`+"\n", request.ID, protocol)
			}()
		}
	}()
	t.Cleanup(func() {
		close(done)
		_ = listener.Close()
	})
	return path
}

func pathWithGitOnly(t *testing.T) string {
	t.Helper()
	realGit, err := exec.LookPath("git")
	if err != nil {
		t.Fatalf("git is not installed: %v", err)
	}
	dir := t.TempDir()
	if err := os.Symlink(realGit, filepath.Join(dir, "git")); err != nil {
		t.Fatal(err)
	}
	return dir
}

func findCheck(t *testing.T, result Result, name string) Check {
	t.Helper()
	for _, check := range result.Checks {
		if check.Name == name {
			return check
		}
	}
	t.Fatalf("check %q is absent from %v", name, result.Checks)
	return Check{}
}

func TestRunChecksWithFakeHerdr(t *testing.T) {
	home := newHome(t)
	t.Setenv("HERDR_SOCKET_PATH", startHerdr(t, 22))
	t.Setenv("HERDR_ENV", "1")
	t.Setenv("HERDR_PANE_ID", "w1:p1")
	t.Setenv("PATH", pathWithGitOnly(t))

	result := Run(context.Background(), home, nil)
	if result.Failed() || result.ExitCode() != 0 {
		t.Fatalf("result = %+v, want no failed required check", result)
	}
	socket := findCheck(t, result, "herdr socket")
	if socket.Status != StatusOK || !socket.Required {
		t.Fatalf("herdr socket = %+v, want a required ok", socket)
	}
	if !strings.Contains(socket.Detail, "protocol 22") || !strings.Contains(socket.Detail, os.Getenv("HERDR_SOCKET_PATH")) {
		t.Fatalf("herdr socket detail = %q, want the path and protocol 22", socket.Detail)
	}
	if pane := findCheck(t, result, "herdr pane"); pane.Status != StatusOK || !strings.Contains(pane.Detail, "w1:p1") {
		t.Fatalf("herdr pane = %+v, want ok with the pane", pane)
	}
	if homeCheck := findCheck(t, result, "home"); homeCheck.Status != StatusOK || !homeCheck.Required {
		t.Fatalf("home = %+v, want a required ok", homeCheck)
	}
	if storeCheck := findCheck(t, result, "state store"); storeCheck.Status != StatusOK || !storeCheck.Required {
		t.Fatalf("state store = %+v, want a required ok", storeCheck)
	}
	if gitCheck := findCheck(t, result, "git"); gitCheck.Status != StatusOK || !gitCheck.Required {
		t.Fatalf("git = %+v, want a required ok", gitCheck)
	}
	if supervisor := findCheck(t, result, "supervisor"); supervisor.Status != StatusWarn || supervisor.Required {
		t.Fatalf("supervisor = %+v, want an optional warn", supervisor)
	}
	for _, name := range []string{"claude", "codex", "opencode", "pi", "cursor-agent"} {
		check := findCheck(t, result, "harness "+name)
		if check.Status != StatusWarn || check.Required {
			t.Fatalf("harness %s = %+v, want an optional warn", name, check)
		}
	}
	if gh := findCheck(t, result, "gh"); gh.Status != StatusWarn || gh.Required {
		t.Fatalf("gh = %+v, want an optional warn", gh)
	}
	if unresolved := findCheck(t, result, "unresolved agents"); unresolved.Status != StatusOK || unresolved.Required || unresolved.Detail != "0" {
		t.Fatalf("unresolved agents = %+v, want an optional ok", unresolved)
	}
	if len(result.Checks) != 14 {
		t.Fatalf("checks = %d, want 14", len(result.Checks))
	}
}

func TestRunWithoutHerdrSocket(t *testing.T) {
	home := newHome(t)
	t.Setenv("HERDR_SOCKET_PATH", "")
	t.Setenv("HERDR_ENV", "")
	t.Setenv("HERDR_PANE_ID", "")
	t.Setenv("PATH", pathWithGitOnly(t))
	result := Run(context.Background(), home, nil)
	check := findCheck(t, result, "herdr socket")
	if check.Status != StatusFail || !check.Required {
		t.Fatalf("herdr socket = %+v, want a required fail", check)
	}
	if check.Detail != "HERDR_SOCKET_PATH is not set" {
		t.Fatalf("detail = %q, want the unset message", check.Detail)
	}
	if !result.Failed() || result.ExitCode() != 1 {
		t.Fatalf("result = %+v, want exit 1", result)
	}
}

func TestRunWithMissingSocket(t *testing.T) {
	home := newHome(t)
	path := filepath.Join(t.TempDir(), "none.sock")
	t.Setenv("HERDR_SOCKET_PATH", path)
	t.Setenv("PATH", pathWithGitOnly(t))
	result := Run(context.Background(), home, nil)
	check := findCheck(t, result, "herdr socket")
	if check.Status != StatusFail || !strings.Contains(check.Detail, path) {
		t.Fatalf("herdr socket = %+v, want a fail with %s", check, path)
	}
	if !result.Failed() {
		t.Fatal("Failed() = false, want true")
	}
}

func TestRunPaneWarnWithoutHerdrEnv(t *testing.T) {
	home := newHome(t)
	t.Setenv("HERDR_SOCKET_PATH", startHerdr(t, 22))
	t.Setenv("HERDR_ENV", "")
	t.Setenv("HERDR_PANE_ID", "")
	t.Setenv("PATH", pathWithGitOnly(t))
	result := Run(context.Background(), home, nil)
	check := findCheck(t, result, "herdr pane")
	if check.Status != StatusWarn || check.Required {
		t.Fatalf("herdr pane = %+v, want an optional warn", check)
	}
	if !strings.Contains(check.Detail, "outside a herdr pane") {
		t.Fatalf("detail = %q, want outside a herdr pane", check.Detail)
	}
	if result.ExitCode() != 0 {
		t.Fatalf("ExitCode() = %d, want 0", result.ExitCode())
	}
}

func TestRunGHAbsentWarn(t *testing.T) {
	home := newHome(t)
	t.Setenv("HERDR_SOCKET_PATH", startHerdr(t, 22))
	t.Setenv("PATH", pathWithGitOnly(t))
	result := Run(context.Background(), home, nil)
	if check := findCheck(t, result, "gh"); check.Status != StatusWarn {
		t.Fatalf("gh = %+v, want an optional warn", check)
	}
	if result.ExitCode() != 0 {
		t.Fatalf("ExitCode() = %d, want 0", result.ExitCode())
	}
}

func TestRunHeldLock(t *testing.T) {
	home := newHome(t)
	state := filepath.Join(home, ".hive", "state")
	if err := os.MkdirAll(state, 0o755); err != nil {
		t.Fatal(err)
	}
	lock, err := os.OpenFile(filepath.Join(state, "lock"), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = lock.Close() }()
	if err := syscall.Flock(int(lock.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = syscall.Flock(int(lock.Fd()), syscall.LOCK_UN) }()

	t.Setenv("HERDR_SOCKET_PATH", startHerdr(t, 22))
	t.Setenv("PATH", pathWithGitOnly(t))
	started := time.Now()
	result := Run(context.Background(), home, nil)
	if elapsed := time.Since(started); elapsed > 6*time.Second {
		t.Fatalf("Run() took %v, want a bounded check", elapsed)
	}
	check := findCheck(t, result, "state store")
	if check.Status != StatusFail || check.Detail != "store lock timeout" {
		t.Fatalf("state store = %+v, want store lock timeout", check)
	}
	if !result.Failed() {
		t.Fatal("Failed() = false, want true")
	}
}

func TestRunLinkedRecord(t *testing.T) {
	home := newHome(t)
	collection := filepath.Join(home, ".hive", "state", "agents")
	if err := os.MkdirAll(collection, 0o755); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(t.TempDir(), "record.json")
	if err := os.WriteFile(target, []byte(`{"id":"linked"}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	linked := filepath.Join(collection, "linked.json")
	if err := os.Symlink(target, linked); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HERDR_SOCKET_PATH", startHerdr(t, 22))
	t.Setenv("PATH", pathWithGitOnly(t))
	result := Run(context.Background(), home, nil)
	check := findCheck(t, result, "state store")
	if check.Status != StatusFail || !strings.Contains(check.Detail, linked) {
		t.Fatalf("state store = %+v, want a fail with %s", check, linked)
	}
}

func TestRunLinkedCollection(t *testing.T) {
	home := newHome(t)
	state := filepath.Join(home, ".hive", "state")
	if err := os.MkdirAll(state, 0o755); err != nil {
		t.Fatal(err)
	}
	target := t.TempDir()
	if err := os.WriteFile(filepath.Join(target, "record.json"), []byte(`{"id":"x"}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	linked := filepath.Join(state, "agents")
	if err := os.Symlink(target, linked); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HERDR_SOCKET_PATH", startHerdr(t, 22))
	t.Setenv("PATH", pathWithGitOnly(t))
	result := Run(context.Background(), home, nil)
	check := findCheck(t, result, "state store")
	if check.Status != StatusFail || !strings.Contains(check.Detail, linked) {
		t.Fatalf("state store = %+v, want a fail with %s", check, linked)
	}
}

func TestRunLinkedStateRoot(t *testing.T) {
	home := newHome(t)
	linked := filepath.Join(home, ".hive", "state")
	if err := os.Symlink(t.TempDir(), linked); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HERDR_SOCKET_PATH", startHerdr(t, 22))
	t.Setenv("PATH", pathWithGitOnly(t))
	result := Run(context.Background(), home, nil)
	check := findCheck(t, result, "state store")
	if check.Status != StatusFail || !strings.Contains(check.Detail, linked) {
		t.Fatalf("state store = %+v, want a fail with %s", check, linked)
	}
}

func TestRunLinkedHiveAncestor(t *testing.T) {
	home := t.TempDir()
	target := t.TempDir()
	if err := os.WriteFile(filepath.Join(target, "README.md"), []byte("# home\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	linked := filepath.Join(home, ".hive")
	if err := os.Symlink(target, linked); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HERDR_SOCKET_PATH", startHerdr(t, 22))
	t.Setenv("PATH", pathWithGitOnly(t))
	result := Run(context.Background(), home, nil)
	check := findCheck(t, result, "state store")
	if check.Status != StatusFail || !strings.Contains(check.Detail, linked) {
		t.Fatalf("state store = %+v, want a fail with %s", check, linked)
	}
}

func TestRunHomeWithoutMarker(t *testing.T) {
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, ".hive"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HERDR_SOCKET_PATH", startHerdr(t, 22))
	t.Setenv("PATH", pathWithGitOnly(t))
	result := Run(context.Background(), home, nil)
	check := findCheck(t, result, "home")
	if check.Status != StatusFail || !check.Required {
		t.Fatalf("home = %+v, want a required fail", check)
	}
}

func TestRunStateStoreMakesNoRecord(t *testing.T) {
	home := newHome(t)
	t.Setenv("HERDR_SOCKET_PATH", startHerdr(t, 22))
	t.Setenv("PATH", pathWithGitOnly(t))
	result := Run(context.Background(), home, nil)
	if check := findCheck(t, result, "state store"); check.Status != StatusOK {
		t.Fatalf("state store = %+v, want ok", check)
	}
	entries, err := os.ReadDir(filepath.Join(home, ".hive", "state"))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".json") {
			t.Fatalf("state store check wrote %s", entry.Name())
		}
	}
}

func TestRender(t *testing.T) {
	result := Result{Checks: []Check{
		{Name: "herdr socket", Status: StatusOK, Required: true, Detail: "/tmp/s.sock protocol 22"},
		{Name: "gh", Status: StatusWarn, Detail: "not on PATH"},
	}}
	got := Render(result)
	want := "ok herdr socket: /tmp/s.sock protocol 22\nwarn gh: not on PATH\n"
	if got != want {
		t.Fatalf("Render() = %q, want %q", got, want)
	}
}

func TestResultJSONFields(t *testing.T) {
	result := Result{Checks: []Check{{Name: "home", Status: StatusFail, Required: true, Detail: "missing"}}}
	data, err := json.Marshal(result.Checks)
	if err != nil {
		t.Fatal(err)
	}
	var decoded []map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if len(decoded) != 1 {
		t.Fatalf("decoded = %v, want one check", decoded)
	}
	for _, key := range []string{"name", "status", "required", "detail"} {
		if _, ok := decoded[0][key]; !ok {
			t.Fatalf("check JSON has no %q: %s", key, data)
		}
	}
}

func readDoctorLogEvents(t *testing.T, home string) []map[string]any {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(home, ".hive", "logs"))
	if err != nil {
		t.Fatal(err)
	}
	var events []map[string]any
	for _, entry := range entries {
		data, err := os.ReadFile(filepath.Join(home, ".hive", "logs", entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
			if line == "" {
				continue
			}
			var event map[string]any
			if err := json.Unmarshal([]byte(line), &event); err != nil {
				t.Fatalf("log line %q: %v", line, err)
			}
			events = append(events, event)
		}
	}
	return events
}

func TestRunLogsProcessEvents(t *testing.T) {
	home := newHome(t)
	t.Setenv("HERDR_SOCKET_PATH", startHerdr(t, 22))
	t.Setenv("PATH", pathWithGitOnly(t))
	writer, err := hivelog.NewWriter(home, "doctor", time.Now)
	if err != nil {
		t.Fatal(err)
	}
	logger := slog.New(writer)
	result := Run(context.Background(), home, logger)
	if result.Failed() {
		t.Fatalf("result = %+v, want no failed required check", result)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	starts := 0
	ends := 0
	failedEnd := false
	commands := map[string]int{}
	for _, event := range readDoctorLogEvents(t, home) {
		switch event["msg"] {
		case "process start":
			starts++
			if command, ok := event["command"].(string); ok {
				commands[command]++
			}
		case "process end":
			ends++
			if _, ok := event["error"]; ok {
				failedEnd = true
			}
		}
	}
	if starts != 7 || ends != 7 {
		t.Fatalf("process events = %d starts and %d ends, want 7 and 7", starts, ends)
	}
	if commands["git --version"] != 1 {
		t.Fatalf("commands = %v, want one git --version", commands)
	}
	if !failedEnd {
		t.Fatal("log has no failed process end event")
	}
}
