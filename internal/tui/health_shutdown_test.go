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
)

// TestRunProgramCancelsAdmittedHealthLoad holds the real Health loader in an
// external command that waits for cancellation. Shutdown must cancel the
// admitted Health work, so the command settles and runProgram returns without
// waiting for the doctor check bound.
func TestRunProgramCancelsAdmittedHealthLoad(t *testing.T) {
	home := newTestHome(t)
	tools := t.TempDir()
	startedPath := filepath.Join(tools, "gh-started")
	pidPath := filepath.Join(tools, "gh-pid")
	script := "#!/bin/sh\n" +
		"if [ \"$1\" = \"--version\" ]; then\n" +
		"  echo 'gh version 0.0.0-t179'\n" +
		"  exit 0\n" +
		"fi\n" +
		"printf '%s\\n' \"$$\" > \"$T179_GH_PID\"\n" +
		"printf 'started\\n' > \"$T179_GH_STARTED\"\n" +
		"exec sleep 600\n"
	if err := os.WriteFile(filepath.Join(tools, "gh"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("T179_GH_PID", pidPath)
	t.Setenv("T179_GH_STARTED", startedPath)
	t.Setenv("PATH", tools+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("HERDR_SOCKET_PATH", "")

	model := New(home, nil)
	defer model.Stop()
	model.screenIndex = 6

	input, keys := io.Pipe()
	defer func() { _ = keys.Close() }()
	done := make(chan error, 1)
	go func() {
		done <- runProgram(model, input, io.Discard, defaultWidth, defaultHeight)
	}()

	waitForTestPath(t, startedPath)
	sentAt := time.Now()
	if _, err := fmt.Fprint(keys, "\x03\x03"); err != nil {
		t.Fatalf("send the quit keys: %v", err)
	}

	var runErr error
	select {
	case runErr = <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("runProgram did not return after the admitted Health load was cancelled")
	}
	elapsed := time.Since(sentAt)
	if runErr != nil {
		t.Fatalf("runProgram: %v", runErr)
	}
	if elapsed >= 2*time.Second {
		t.Fatalf("runProgram returned after %s; shutdown did not cancel the admitted Health external command", elapsed)
	}

	pidText := strings.TrimSpace(readLogFile(t, pidPath))
	var pid int
	if _, err := fmt.Sscan(pidText, &pid); err != nil || pid <= 0 {
		t.Fatalf("external command pid = %q: %v", pidText, err)
	}
	if err := syscall.Kill(pid, 0); err == nil {
		t.Fatalf("external Health command %d is still running after runProgram returned", pid)
	} else if !errors.Is(err, syscall.ESRCH) {
		t.Fatalf("check external Health command %d: %v", pid, err)
	}

	path := onlyLogFile(t, home)
	before := readLogFile(t, path)
	record := slog.NewRecord(time.Now(), slog.LevelDebug, "late health event", 0)
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

// waitForTestPath waits for an external helper to create path.
func waitForTestPath(t *testing.T, path string) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for {
		if _, err := os.Stat(path); err == nil {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("the external Health command did not start: %s", path)
		}
		time.Sleep(time.Millisecond)
	}
}
