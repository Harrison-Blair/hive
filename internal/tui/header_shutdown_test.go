package tui

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/Harrison-Blair/hive/internal/lib/records"
)

// TestRunProgramCancelsAdmittedHeaderLoad holds the real header loader in the
// external doctor command of the supervisor binary check. Shutdown must cancel
// the admitted header work, so the command settles and runProgram returns
// without waiting for the doctor check bound.
func TestRunProgramCancelsAdmittedHeaderLoad(t *testing.T) {
	home := newTestHome(t)
	tools := t.TempDir()
	startedPath := filepath.Join(tools, "gh-started")
	pidPath := filepath.Join(tools, "gh-pid")
	script := "#!/bin/sh\n" +
		"if [ \"$1\" = \"--version\" ]; then\n" +
		"  echo 'gh version 0.0.0-t179-header'\n" +
		"  exit 0\n" +
		"fi\n" +
		"printf '%s\\n' \"$$\" > \"$T179_HEADER_GH_PID\"\n" +
		"printf 'started\\n' > \"$T179_HEADER_GH_STARTED\"\n" +
		"exec sleep 600\n"
	if err := os.WriteFile(filepath.Join(tools, "gh"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tools, "hive"), []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("T179_HEADER_GH_PID", pidPath)
	t.Setenv("T179_HEADER_GH_STARTED", startedPath)
	t.Setenv("PATH", tools+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("HERDR_SOCKET_PATH", "")

	writeRecord(t, home, []string{"claims"}, "supervisor", records.Claim{
		Key: "supervisor", Agent: "supervisor", Generation: 1,
		PID: int64(os.Getpid()), PIDStart: currentProcessStart(t),
	})

	model := New(home, nil)
	defer model.Stop()
	input, keys := io.Pipe()
	defer func() { _ = keys.Close() }()
	done := make(chan error, 1)
	go func() {
		done <- runProgram(model, input, io.Discard, defaultWidth, defaultHeight)
	}()

	waitForHeaderTestPath(t, startedPath)
	sentAt := time.Now()
	if _, err := fmt.Fprint(keys, "\x03\x03"); err != nil {
		t.Fatalf("send the quit keys: %v", err)
	}

	var runErr error
	select {
	case runErr = <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("runProgram did not return after the admitted header load was cancelled")
	}
	elapsed := time.Since(sentAt)
	if runErr != nil {
		t.Fatalf("runProgram: %v", runErr)
	}
	if elapsed >= 2*time.Second {
		t.Fatalf("runProgram returned after %s; shutdown did not cancel the admitted header external command", elapsed)
	}

	pidText := strings.TrimSpace(readLogFile(t, pidPath))
	var pid int
	if _, err := fmt.Sscan(pidText, &pid); err != nil || pid <= 0 {
		t.Fatalf("external command pid = %q: %v", pidText, err)
	}
	if err := syscall.Kill(pid, 0); err == nil {
		t.Fatalf("external header command %d is still running after runProgram returned", pid)
	} else if !errors.Is(err, syscall.ESRCH) {
		t.Fatalf("check external header command %d: %v", pid, err)
	}

	path := onlyLogFile(t, home)
	before := readLogFile(t, path)
	record := slog.NewRecord(time.Now(), slog.LevelDebug, "late header event", 0)
	if err := model.logger.Handler().Handle(context.Background(), record); err == nil {
		t.Fatal("the TUI writer accepted an event after runProgram returned")
	}
	if after := readLogFile(t, path); after != before {
		t.Fatalf("a late TUI event changed the log file:\nbefore:\n%s\nafter:\n%s", before, after)
	}
	if got := strings.Count(before, `"action":"quit request"`); got != 1 {
		t.Fatalf("accepted quit events = %d; want exactly 1", got)
	}

	accessed := false
	cmd := model.startLoad(func() tea.Msg {
		accessed = true
		return nil
	})
	if msg := cmd(); msg != nil {
		if _, ok := msg.(loaderStoppedMsg); !ok {
			t.Fatalf("a callback after return gave %T; want loaderStoppedMsg", msg)
		}
	} else {
		t.Fatal("a callback after return gave no message; want loaderStoppedMsg")
	}
	if accessed {
		t.Fatal("a loader began home access after runProgram returned")
	}
}

// TestReadHeaderCachesTheSupervisorBinaryCheck keeps the normal header result
// and the cache of the supervisor binary check.
func TestReadHeaderCachesTheSupervisorBinaryCheck(t *testing.T) {
	home := newTestHome(t)
	tools := t.TempDir()
	countPath := filepath.Join(tools, "auth-count")
	script := "#!/bin/sh\n" +
		"if [ \"$1\" = \"--version\" ]; then\n" +
		"  echo 'gh version 0.0.0-t179-header'\n" +
		"  exit 0\n" +
		"fi\n" +
		"printf 'auth\\n' >> \"$T179_HEADER_CACHE_COUNT\"\n" +
		"exit 0\n"
	if err := os.WriteFile(filepath.Join(tools, "gh"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tools, "hive"), []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("T179_HEADER_CACHE_COUNT", countPath)
	t.Setenv("PATH", tools+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("HERDR_SOCKET_PATH", "")
	writeRecord(t, home, []string{"claims"}, "supervisor", records.Claim{
		Key: "supervisor", Agent: "supervisor", Generation: 1,
		PID: int64(os.Getpid()), PIDStart: currentProcessStart(t),
	})

	state, err := readHeader(context.Background(), home)
	if err != nil {
		t.Fatalf("readHeader(): %v", err)
	}
	if state.supervisor != "fresh" {
		t.Fatalf("supervisor state = %q; want fresh", state.supervisor)
	}
	if got := authCount(t, countPath); got != 1 {
		t.Fatalf("supervisor binary checks = %d; want 1", got)
	}
	again, err := readHeader(context.Background(), home)
	if err != nil {
		t.Fatalf("second readHeader(): %v", err)
	}
	if again.supervisor != "fresh" || again.supervisorOld != state.supervisorOld {
		t.Fatalf("second header state = %#v; want the cached first result", again)
	}
	if got := authCount(t, countPath); got != 1 {
		t.Fatalf("supervisor binary checks after the second read = %d; want the cached 1", got)
	}
}

// waitForHeaderTestPath waits for an external helper to create path.
func waitForHeaderTestPath(t *testing.T, path string) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for {
		if _, err := os.Stat(path); err == nil {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("the external header command did not start: %s", path)
		}
		time.Sleep(time.Millisecond)
	}
}

// authCount counts the external authentication checks.
func authCount(t *testing.T, path string) int {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return strings.Count(string(data), "auth")
}
