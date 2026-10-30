package log

import (
	"context"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

func guardClock(year int, month time.Month, day int) func() time.Time {
	return func() time.Time { return time.Date(year, month, day, 12, 0, 0, 0, time.Local) }
}

func guardRecord(clock func() time.Time, message string) slog.Record {
	return slog.NewRecord(clock(), slog.LevelDebug, message, 0)
}

func guardHandle(writer *Writer, clock func() time.Time, message string) error {
	return writer.Handle(context.Background(), guardRecord(clock, message))
}

func guardDirEntries(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	return names
}

func TestStaticLogPathRejectsLinkedLogsDirectory(t *testing.T) {
	home := t.TempDir()
	external := t.TempDir()
	if err := os.Symlink(external, filepath.Join(home, "logs")); err != nil {
		t.Fatal(err)
	}
	writer, err := NewWriter(home, "cli", guardClock(2026, time.January, 2))
	if err == nil {
		if writer != nil {
			if handleErr := guardHandle(writer, guardClock(2026, time.January, 2), "must not write"); handleErr != nil {
				t.Logf("old writer write error: %v", handleErr)
			}
			_ = writer.Close()
		}
		t.Fatalf("NewWriter accepted a linked logs directory; external entries = %v", guardDirEntries(t, external))
	}
	if !strings.Contains(err.Error(), "not a native directory") {
		t.Fatalf("error = %q; want a native directory rejection", err)
	}
	if entries := guardDirEntries(t, external); len(entries) != 0 {
		t.Fatalf("rejected linked logs directory changed external target: %v", entries)
	}
}

func TestStaticLogPathRejectsNonRegularDailyPaths(t *testing.T) {
	validTarget := filepath.Join(t.TempDir(), "target.jsonl")
	sentinel := []byte("synthetic target bytes\n")
	if err := os.WriteFile(validTarget, sentinel, 0o644); err != nil {
		t.Fatal(err)
	}
	danglingTarget := filepath.Join(t.TempDir(), "missing.jsonl")

	cases := []struct {
		name     string
		makePath func(t *testing.T, path string)
		write    bool
		target   string
		dangling bool
	}{
		{
			name: "valid symlink",
			makePath: func(t *testing.T, path string) {
				if err := os.Symlink(validTarget, path); err != nil {
					t.Fatal(err)
				}
			},
			write:  true,
			target: validTarget,
		},
		{
			name: "dangling symlink",
			makePath: func(t *testing.T, path string) {
				if err := os.Symlink(danglingTarget, path); err != nil {
					t.Fatal(err)
				}
			},
			write:    true,
			target:   danglingTarget,
			dangling: true,
		},
		{
			name: "fifo",
			makePath: func(t *testing.T, path string) {
				if err := syscall.Mkfifo(path, 0o600); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "directory",
			makePath: func(t *testing.T, path string) {
				if err := os.Mkdir(path, 0o755); err != nil {
					t.Fatal(err)
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			home := t.TempDir()
			logs := filepath.Join(home, "logs")
			if err := os.MkdirAll(logs, 0o755); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(logs, "2026-01-02.jsonl")
			tc.makePath(t, path)
			clock := guardClock(2026, time.January, 2)
			writer, err := NewWriter(home, "cli", clock)
			if err == nil {
				if tc.write && writer != nil {
					if handleErr := guardHandle(writer, clock, "must not write"); handleErr != nil {
						t.Logf("old writer write error: %v", handleErr)
					}
				}
				if writer != nil {
					_ = writer.Close()
				}
				t.Fatalf("NewWriter accepted a nonregular daily %s path", tc.name)
			}
			if !strings.Contains(err.Error(), "not a regular file") {
				t.Fatalf("error = %q; want a regular file rejection", err)
			}
			if tc.dangling {
				if _, statErr := os.Lstat(tc.target); !os.IsNotExist(statErr) {
					t.Fatalf("rejected dangling link created target: stat error %v", statErr)
				}
				return
			}
			if tc.target != "" {
				data, readErr := os.ReadFile(tc.target)
				if readErr != nil {
					t.Fatal(readErr)
				}
				if string(data) != string(sentinel) {
					t.Fatalf("rejected link changed target to %q", data)
				}
			}
		})
	}
}

func TestStaticLogPathRejectsDailySocket(t *testing.T) {
	home := t.TempDir()
	logs := filepath.Join(home, "logs")
	if err := os.MkdirAll(logs, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(logs, "2026-01-02.jsonl")
	listener, err := net.Listen("unix", path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = listener.Close() }()
	writer, err := NewWriter(home, "cli", guardClock(2026, time.January, 2))
	if err == nil {
		if writer != nil {
			_ = writer.Close()
		}
		t.Fatal("NewWriter accepted a socket daily path")
	}
	if !strings.Contains(err.Error(), "not a regular file") {
		t.Fatalf("error = %q; want a regular file rejection", err)
	}
}

func TestStaticLogPathRejectsDailyFIFOWithoutBlocking(t *testing.T) {
	home := t.TempDir()
	logs := filepath.Join(home, "logs")
	if err := os.MkdirAll(logs, 0o755); err != nil {
		t.Fatal(err)
	}
	fifo := filepath.Join(logs, "2026-01-02.jsonl")
	if err := syscall.Mkfifo(fifo, 0o600); err != nil {
		t.Fatal(err)
	}
	clock := guardClock(2026, time.January, 2)
	writer, err := NewWriter(home, "cli", clock)
	if err == nil {
		if writer != nil {
			_ = writer.Close()
		}
		t.Fatal("NewWriter accepted a FIFO daily path without an error")
	}
	if !strings.Contains(err.Error(), "not a regular file") {
		t.Fatalf("error = %q; want a regular file rejection", err)
	}
}

func TestStaticLogPathRejectsRotationToFIFOWithoutBlocking(t *testing.T) {
	home := t.TempDir()
	current := time.Date(2026, time.January, 2, 12, 0, 0, 0, time.Local)
	clock := func() time.Time { return current }
	writer, err := NewWriter(home, "cli", clock)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = writer.Close() }()
	if err := guardHandle(writer, clock, "day one"); err != nil {
		t.Fatal(err)
	}
	logs := filepath.Join(home, "logs")
	fifo := filepath.Join(logs, "2026-01-03.jsonl")
	if err := syscall.Mkfifo(fifo, 0o600); err != nil {
		t.Fatal(err)
	}
	current = time.Date(2026, time.January, 3, 12, 0, 0, 0, time.Local)
	done := make(chan error, 1)
	go func() {
		done <- guardHandle(writer, clock, "day two")
	}()
	select {
	case handleErr := <-done:
		if handleErr == nil {
			t.Fatal("Handle accepted a FIFO rotation path")
		}
		if !strings.Contains(handleErr.Error(), "not a regular file") {
			t.Fatalf("error = %q; want a regular file rejection", handleErr)
		}
	case <-time.After(2 * time.Second):
		// A cooperating reader releases only this owned FIFO. No process gets a signal.
		fd, openErr := syscall.Open(fifo, syscall.O_RDONLY|syscall.O_NONBLOCK, 0)
		if openErr != nil {
			t.Fatalf("old Handle blocked on FIFO and the owned reader failed: %v", openErr)
		}
		select {
		case settleErr := <-done:
			if settleErr != nil {
				t.Logf("old Handle returned after the owned reader attached: %v", settleErr)
			}
		case <-time.After(time.Second):
			t.Fatal("old Handle did not settle after the owned reader attached")
		}
		_ = syscall.Close(fd)
		t.Fatal("Handle blocked more than 2 seconds on a static FIFO rotation path")
	}
	if err := os.Remove(fifo); err != nil {
		t.Fatal(err)
	}
	if err := guardHandle(writer, clock, "after removal"); err != nil {
		t.Fatalf("retry after FIFO removal: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(logs, "2026-01-03.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "after removal") {
		t.Fatalf("daily file after retry = %q", data)
	}
}

func TestStaticLogPathRejectsRotationToDailyLinkAndRetries(t *testing.T) {
	home := t.TempDir()
	current := time.Date(2026, time.January, 2, 12, 0, 0, 0, time.Local)
	clock := func() time.Time { return current }
	writer, err := NewWriter(home, "cli", clock)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = writer.Close() }()
	if err := guardHandle(writer, clock, "day one"); err != nil {
		t.Fatal(err)
	}
	dayOnePath := filepath.Join(home, "logs", "2026-01-02.jsonl")
	dayOneBefore, err := os.ReadFile(dayOnePath)
	if err != nil {
		t.Fatal(err)
	}
	external := filepath.Join(t.TempDir(), "target.jsonl")
	sentinel := []byte("synthetic target bytes\n")
	if err := os.WriteFile(external, sentinel, 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(home, "logs", "2026-01-03.jsonl")
	if err := os.Symlink(external, link); err != nil {
		t.Fatal(err)
	}
	current = time.Date(2026, time.January, 3, 12, 0, 0, 0, time.Local)
	if handleErr := guardHandle(writer, clock, "day two"); handleErr == nil {
		t.Fatal("Handle accepted a symlink rotation path and wrote through it")
	} else if !strings.Contains(handleErr.Error(), "not a regular file") {
		t.Fatalf("error = %q; want a regular file rejection", handleErr)
	}
	data, err := os.ReadFile(external)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(sentinel) {
		t.Fatalf("rejected rotation changed external target to %q", data)
	}
	dayOneAfter, err := os.ReadFile(dayOnePath)
	if err != nil {
		t.Fatal(err)
	}
	if string(dayOneAfter) != string(dayOneBefore) {
		t.Fatalf("failed rotation changed the current daily file: %q", dayOneAfter)
	}
	if err := os.Remove(link); err != nil {
		t.Fatal(err)
	}
	if err := guardHandle(writer, clock, "after removal"); err != nil {
		t.Fatalf("retry after link removal: %v", err)
	}
	dayTwoData, err := os.ReadFile(filepath.Join(home, "logs", "2026-01-03.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(dayTwoData), "after removal") {
		t.Fatalf("daily file after retry = %q", dayTwoData)
	}
	if data, err := os.ReadFile(external); err != nil {
		t.Fatal(err)
	} else if string(data) != string(sentinel) {
		t.Fatalf("retry wrote to the rejected external target: %q", data)
	}
}

func TestStaticLogPathRejectsLinkedLogsDirectoryOnRotation(t *testing.T) {
	home := t.TempDir()
	current := time.Date(2026, time.January, 2, 12, 0, 0, 0, time.Local)
	clock := func() time.Time { return current }
	writer, err := NewWriter(home, "cli", clock)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = writer.Close() }()
	if err := guardHandle(writer, clock, "day one"); err != nil {
		t.Fatal(err)
	}
	logs := filepath.Join(home, "logs")
	realLogs := filepath.Join(home, "logs-real")
	if err := os.Rename(logs, realLogs); err != nil {
		t.Fatal(err)
	}
	external := t.TempDir()
	if err := os.Symlink(external, logs); err != nil {
		t.Fatal(err)
	}
	current = time.Date(2026, time.January, 3, 12, 0, 0, 0, time.Local)
	if handleErr := guardHandle(writer, clock, "day two"); handleErr == nil {
		t.Fatal("Handle accepted a linked logs directory on rotation")
	} else if !strings.Contains(handleErr.Error(), "not a native directory") {
		t.Fatalf("error = %q; want a native directory rejection", handleErr)
	}
	if entries := guardDirEntries(t, external); len(entries) != 0 {
		t.Fatalf("rejected linked logs directory changed external target: %v", entries)
	}
	if err := os.Remove(logs); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(realLogs, logs); err != nil {
		t.Fatal(err)
	}
	if err := guardHandle(writer, clock, "after restore"); err != nil {
		t.Fatalf("retry after logs directory restore: %v", err)
	}
	if _, err := os.Stat(filepath.Join(logs, "2026-01-03.jsonl")); err != nil {
		t.Fatalf("retry did not make the valid next daily file: %v", err)
	}
}

func TestStaticLogPathAllowsValidCreationAppendAndWholeHomeAlias(t *testing.T) {
	realHome := t.TempDir()
	parent := t.TempDir()
	alias := filepath.Join(parent, "alias")
	if err := os.Symlink(realHome, alias); err != nil {
		t.Fatal(err)
	}
	dayOne := guardClock(2026, time.January, 2)
	writer, err := NewWriter(alias, "cli", dayOne)
	if err != nil {
		t.Fatal(err)
	}
	if err := guardHandle(writer, dayOne, "first"); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(realHome, "logs", "2026-01-02.jsonl")
	first, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(first), "first") {
		t.Fatalf("first daily file = %q", first)
	}
	info, err := os.Stat(path)
	if err != nil || info.Mode().Perm() != 0o644 {
		t.Fatalf("daily file mode = %v, %v; want 0644", info, err)
	}
	logsInfo, err := os.Stat(filepath.Join(realHome, "logs"))
	if err != nil || logsInfo.Mode().Perm() != 0o755 {
		t.Fatalf("logs directory mode = %v, %v; want 0755", logsInfo, err)
	}
	writer, err = NewWriter(alias, "cli", dayOne)
	if err != nil {
		t.Fatal(err)
	}
	if err := guardHandle(writer, dayOne, "second"); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	appended, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(appended), string(first)) || strings.Count(string(appended), "\n") != 2 {
		t.Fatalf("appended daily file = %q", appended)
	}
	aliasInfo, err := os.Lstat(alias)
	if err != nil || aliasInfo.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("whole-home alias changed: %v, %v", aliasInfo, err)
	}
}
