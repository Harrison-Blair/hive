package run

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	hivelog "github.com/Harrison-Blair/hive/internal/lib/log"
)

func TestCommandHonorsCallerDeadline(t *testing.T) {
	t.Run("honors the caller deadline over the default", func(t *testing.T) {
		original := commandDeadline
		commandDeadline = 100 * time.Millisecond
		t.Cleanup(func() { commandDeadline = original })
		ctx, cancel := context.WithTimeout(context.Background(), 600*time.Millisecond)
		defer cancel()
		started := time.Now()
		_, err := Command(ctx, t.TempDir(), "sleep", "5")
		elapsed := time.Since(started)
		if !errors.Is(err, ErrTimeout) {
			t.Fatalf("Command() error = %v, want ErrTimeout", err)
		}
		if elapsed < 400*time.Millisecond {
			t.Fatalf("Command() returned after %v, want the caller deadline and not the 100ms default", elapsed)
		}
	})

	t.Run("honors an earlier caller cancellation", func(t *testing.T) {
		original := commandDeadline
		commandDeadline = time.Second
		t.Cleanup(func() { commandDeadline = original })
		ctx, cancel := context.WithCancel(context.Background())
		timer := time.AfterFunc(100*time.Millisecond, cancel)
		defer timer.Stop()
		defer cancel()
		started := time.Now()
		_, err := Command(ctx, t.TempDir(), "sleep", "5")
		elapsed := time.Since(started)
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Command() error = %v, want context.Canceled", err)
		}
		if errors.Is(err, ErrTimeout) {
			t.Fatalf("Command() error = %v, want a cancellation and not a timeout", err)
		}
		if elapsed > 2*time.Second {
			t.Fatalf("Command() returned after %v, want two seconds or less", elapsed)
		}
	})
}

func TestCommandTimeout(t *testing.T) {
	dir := t.TempDir()
	pidFile := filepath.Join(dir, "parent.pid")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	started := time.Now()
	_, err := Command(ctx, dir, "sh", "-c", `echo $$ > "$1"; sleep 5`, "sh", pidFile)
	elapsed := time.Since(started)
	if !errors.Is(err, ErrTimeout) {
		t.Fatalf("Command() error = %v, want ErrTimeout", err)
	}
	if elapsed > 2*time.Second {
		t.Fatalf("Command() returned after %v, want two seconds or less", elapsed)
	}
	assertGone(t, readPID(t, pidFile))
}

func TestCommandTimeoutKillsChildThatHoldsPipes(t *testing.T) {
	dir := t.TempDir()
	childFile := filepath.Join(dir, "child.pid")
	parentFile := filepath.Join(dir, "parent.pid")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	started := time.Now()
	_, err := Command(ctx, dir, "sh", "-c", `sleep 30 & echo $! > "$1"; echo $$ > "$2"; sleep 5`, "sh", childFile, parentFile)
	elapsed := time.Since(started)
	if !errors.Is(err, ErrTimeout) {
		t.Fatalf("Command() error = %v, want ErrTimeout", err)
	}
	if elapsed > 2*time.Second {
		t.Fatalf("Command() returned after %v, want two seconds or less", elapsed)
	}
	assertGone(t, readPID(t, childFile))
	assertGone(t, readPID(t, parentFile))
}

func TestCommandOutputCap(t *testing.T) {
	cases := []struct {
		size      int
		truncated bool
	}{
		{size: 1048575, truncated: false},
		{size: 1048576, truncated: false},
		{size: 1048577, truncated: true},
		{size: 2 * 1048576, truncated: true},
	}
	for _, testCase := range cases {
		t.Run(fmt.Sprintf("%d", testCase.size), func(t *testing.T) {
			result, err := Command(context.Background(), "", "sh", "-c", fmt.Sprintf("head -c %d /dev/zero", testCase.size))
			if err != nil {
				t.Fatalf("Command() error = %v", err)
			}
			want := testCase.size
			if testCase.truncated {
				want = 1048576
			}
			if len(result.Stdout) != want {
				t.Fatalf("stdout = %d bytes, want %d", len(result.Stdout), want)
			}
			if result.Truncated != testCase.truncated {
				t.Fatalf("Truncated = %v, want %v", result.Truncated, testCase.truncated)
			}
		})
	}
}

func TestCommandStderrCap(t *testing.T) {
	result, err := Command(context.Background(), "", "sh", "-c", "head -c 2097152 /dev/zero >&2")
	if err != nil {
		t.Fatalf("Command() error = %v", err)
	}
	if len(result.Stderr) != 1048576 {
		t.Fatalf("stderr = %d bytes, want 1048576", len(result.Stderr))
	}
	if !result.Truncated {
		t.Fatal("Truncated = false, want true")
	}
}

func TestCommandExitCodeIsNotAnError(t *testing.T) {
	result, err := Command(context.Background(), "", "false")
	if err != nil {
		t.Fatalf("Command() error = %v", err)
	}
	if result.ExitCode != 1 {
		t.Fatalf("ExitCode = %d, want 1", result.ExitCode)
	}
}

func TestCommandMissingBinary(t *testing.T) {
	if _, err := Command(context.Background(), "", "definitely-not-a-hive-binary"); err == nil {
		t.Fatal("Command() error = nil, want a start error")
	}
}

func TestCommandReadsEndOfFileFromStdin(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	result, err := Command(ctx, "", "cat")
	if err != nil {
		t.Fatalf("Command() error = %v", err)
	}
	if result.Stdout != "" || result.Stderr != "" {
		t.Fatalf("output = %q and %q, want empty", result.Stdout, result.Stderr)
	}
}

func TestCommandUsesDir(t *testing.T) {
	dir := t.TempDir()
	result, err := Command(context.Background(), dir, "pwd")
	if err != nil {
		t.Fatalf("Command() error = %v", err)
	}
	want, err := filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(result.Stdout); got != want {
		t.Fatalf("pwd = %q, want %q", got, want)
	}
}

func TestCommandLogsStartAndEnd(t *testing.T) {
	home := t.TempDir()
	writer, err := hivelog.NewWriter(home, "run", time.Now)
	if err != nil {
		t.Fatal(err)
	}
	logger := slog.New(writer)
	defer func() { _ = writer.Close() }()
	ctx := WithLogger(context.Background(), logger)
	if _, err := Command(ctx, "", "true"); err != nil {
		t.Fatalf("Command() error = %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(filepath.Join(home, ".hive", "logs"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("log files = %d, want one", len(entries))
	}
	data, err := os.ReadFile(filepath.Join(home, ".hive", "logs", entries[0].Name()))
	if err != nil {
		t.Fatal(err)
	}
	events := map[string]map[string]any{}
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		var event map[string]any
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			t.Fatalf("log line %q: %v", line, err)
		}
		message, _ := event["msg"].(string)
		events[message] = event
	}
	start, ok := events["process start"]
	if !ok {
		t.Fatalf("log events = %v, want process start", events)
	}
	end, ok := events["process end"]
	if !ok {
		t.Fatalf("log events = %v, want process end", events)
	}
	if start["command"] != "true" || end["command"] != "true" {
		t.Fatalf("commands = %v and %v, want true", start["command"], end["command"])
	}
	if end["exit_code"] != float64(0) {
		t.Fatalf("exit_code = %v, want 0", end["exit_code"])
	}
	if _, ok := end["duration_ms"]; !ok {
		t.Fatalf("end event has no duration_ms: %v", end)
	}
}

func readPID(t *testing.T, path string) int {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for {
		data, err := os.ReadFile(path)
		if err == nil {
			pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
			if err != nil {
				t.Fatalf("pid file %s: %v", path, err)
			}
			return pid
		}
		if time.Now().After(deadline) {
			t.Fatalf("pid file %s: %v", path, err)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func assertGone(t *testing.T, pid int) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for {
		process, err := os.FindProcess(pid)
		if err != nil {
			return
		}
		if err := process.Signal(syscall.Signal(0)); err != nil {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("process %d still exists after the deadline", pid)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func readRunLogEvents(t *testing.T, home string) []map[string]any {
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

func TestCommandDeadlineWithInheritedPipeChild(t *testing.T) {
	dir := t.TempDir()
	childFile := filepath.Join(dir, "child.pid")
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	started := time.Now()
	_, err := Command(ctx, dir, "sh", "-c", `sleep 30 & echo $! > "$1"; exit 0`, "sh", childFile)
	elapsed := time.Since(started)
	if !errors.Is(err, ErrTimeout) {
		t.Fatalf("Command() error = %v, want ErrTimeout", err)
	}
	if elapsed > 2*time.Second {
		t.Fatalf("Command() returned after %v, want two seconds or less", elapsed)
	}
	assertGone(t, readPID(t, childFile))
}

func TestCommandCancellationWithInheritedPipeChild(t *testing.T) {
	dir := t.TempDir()
	childFile := filepath.Join(dir, "child.pid")
	ctx, cancel := context.WithCancel(context.Background())
	timer := time.AfterFunc(100*time.Millisecond, cancel)
	defer timer.Stop()
	defer cancel()
	started := time.Now()
	_, err := Command(ctx, dir, "sh", "-c", `sleep 30 & echo $! > "$1"; exit 0`, "sh", childFile)
	elapsed := time.Since(started)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Command() error = %v, want context.Canceled", err)
	}
	if errors.Is(err, ErrTimeout) {
		t.Fatalf("Command() error = %v, want a cancellation and not a timeout", err)
	}
	if elapsed > 2*time.Second {
		t.Fatalf("Command() returned after %v, want two seconds or less", elapsed)
	}
	assertGone(t, readPID(t, childFile))
}

func hasControlByte(text string) bool {
	for _, value := range text {
		if value < 0x20 || value == 0x7f {
			return true
		}
	}
	return false
}

func TestCommandErrorStaysOneLine(t *testing.T) {
	cases := []struct {
		name   string
		dir    string
		binary string
	}{
		{name: "name newline", binary: "no\nsuch"},
		{name: "name return", binary: "no\rsuch"},
		{name: "dir newline", dir: filepath.Join(t.TempDir(), "no\nsuch"), binary: "true"},
		{name: "dir return", dir: filepath.Join(t.TempDir(), "no\rsuch"), binary: "true"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := Command(context.Background(), testCase.dir, testCase.binary)
			if err == nil {
				t.Fatal("Command() error = nil, want a start error")
			}
			if strings.ContainsAny(err.Error(), "\n\r") {
				t.Fatalf("error = %q, want one physical line", err)
			}
			if hasControlByte(err.Error()) {
				t.Fatalf("error = %q, want no control character", err)
			}
		})
	}
}

func TestCommandCancellationLogsReturnedError(t *testing.T) {
	home := t.TempDir()
	writer, err := hivelog.NewWriter(home, "run", time.Now)
	if err != nil {
		t.Fatal(err)
	}
	logger := slog.New(writer)
	ctx := WithLogger(context.Background(), logger)
	ctx, cancel := context.WithCancel(ctx)
	timer := time.AfterFunc(100*time.Millisecond, cancel)
	defer timer.Stop()
	defer cancel()
	_, commandErr := Command(ctx, "", "sleep", "5")
	if !errors.Is(commandErr, context.Canceled) {
		t.Fatalf("Command() error = %v, want context.Canceled", commandErr)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	var end map[string]any
	for _, event := range readRunLogEvents(t, home) {
		if event["msg"] == "process end" {
			end = event
		}
	}
	if end == nil {
		t.Fatal("log has no process end event")
	}
	if got, _ := end["error"].(string); got != commandErr.Error() {
		t.Fatalf("logged error = %q, want the returned error %q", got, commandErr.Error())
	}
}

func TestCommandTimeoutLogsReturnedError(t *testing.T) {
	home := t.TempDir()
	writer, err := hivelog.NewWriter(home, "run", time.Now)
	if err != nil {
		t.Fatal(err)
	}
	logger := slog.New(writer)
	ctx := WithLogger(context.Background(), logger)
	ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()
	_, commandErr := Command(ctx, "", "sleep", "5")
	if !errors.Is(commandErr, ErrTimeout) {
		t.Fatalf("Command() error = %v, want ErrTimeout", commandErr)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	var end map[string]any
	for _, event := range readRunLogEvents(t, home) {
		if event["msg"] == "process end" {
			end = event
		}
	}
	if end == nil {
		t.Fatal("log has no process end event")
	}
	if got, _ := end["error"].(string); got != commandErr.Error() {
		t.Fatalf("logged error = %q, want the returned error %q", got, commandErr.Error())
	}
}
