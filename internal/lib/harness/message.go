package harness

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"time"

	"github.com/Harrison-Blair/hive/internal/lib/herdr"
)

// writeTimeout bounds the write of one message line to a socket.
const writeTimeout = 5 * time.Second

// claudeInbound is the --settings value that makes Claude Code deliver each
// cross-session message without an approval dialog. A session that bypasses
// permission prompts holds a message from an unknown sender by default.
const claudeInbound = `{"crossSessionInbound":"accept"}`

// Running reports whether pane has a process of profile among its foreground
// processes. It does not use the agent status of herdr, which can be unknown
// for a live harness.
func Running(ctx context.Context, profile Profile, paneID string, logger *slog.Logger) (bool, error) {
	info, err := herdr.PaneProcessInfo(ctx, paneID, herdr.WithLogger(logger))
	if err != nil {
		return false, err
	}
	return len(processesOf(profile, info)) > 0, nil
}

// harnessProcess returns the process of profile in the foreground group of
// pane. It prefers the group leader, because the harness can start child
// processes with the same name.
func harnessProcess(ctx context.Context, profile Profile, paneID string, logger *slog.Logger) (herdr.Process, error) {
	info, err := herdr.PaneProcessInfo(ctx, paneID, herdr.WithLogger(logger))
	if err != nil {
		return herdr.Process{}, fmt.Errorf("process info of pane %s: %w", paneID, err)
	}
	found := processesOf(profile, info)
	if len(found) == 0 {
		return herdr.Process{}, fmt.Errorf("no %s process in pane %s", profile.Kind(), paneID)
	}
	for _, process := range found {
		if process.PID == info.ForegroundProcessGroupID {
			return process, nil
		}
	}
	return found[0], nil
}

// processesOf returns each foreground process of info whose name is a
// process name of profile.
func processesOf(profile Profile, info herdr.ProcessInfo) []herdr.Process {
	var found []herdr.Process
	for _, process := range info.ForegroundProcesses {
		if slices.Contains(profile.ProcessNames(), process.Name) {
			found = append(found, process)
		}
	}
	return found
}

// hasArgPair reports whether argv holds name followed by value.
func hasArgPair(argv []string, name, value string) bool {
	for index := 0; index+1 < len(argv); index++ {
		if argv[index] == name && argv[index+1] == value {
			return true
		}
	}
	return false
}

// socketPath returns the socket <pid>.sock of the process pid in
// $XDG_RUNTIME_DIR/sub, or else in the directory fallback.
func socketPath(sub string, pid int, fallback string) (string, error) {
	name := strconv.Itoa(pid) + ".sock"
	var paths []string
	if dir := os.Getenv("XDG_RUNTIME_DIR"); dir != "" {
		paths = append(paths, filepath.Join(dir, sub, name))
	}
	paths = append(paths, filepath.Join(fallback, name))
	for _, path := range paths {
		if info, err := os.Stat(path); err == nil && info.Mode().Type() == fs.ModeSocket {
			return path, nil
		}
	}
	return "", fmt.Errorf("no message socket at %s", paths[0])
}

// writeLine connects to the Unix socket path, writes line and a newline, and
// closes the connection.
func writeLine(ctx context.Context, path string, line []byte) error {
	ctx, cancel := context.WithTimeout(ctx, writeTimeout)
	defer cancel()
	var dialer net.Dialer
	conn, err := dialer.DialContext(ctx, "unix", path)
	if err != nil {
		return err
	}
	deadline, _ := ctx.Deadline()
	if err := conn.SetWriteDeadline(deadline); err != nil {
		return errors.Join(err, conn.Close())
	}
	_, err = conn.Write(append(line, '\n'))
	return errors.Join(err, conn.Close())
}
