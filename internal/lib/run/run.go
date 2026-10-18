// Package run runs one external process with bounded time and output.
//
// run.go owns the process group, the deadline, the cap of each stream, the
// exit code, and the log events.
package run

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

// ErrTimeout reports that the deadline stopped the process group.
var ErrTimeout = errors.New("process timeout")

const (
	defaultDeadline = 30 * time.Second
	maxStream       = 1 << 20
	normalGrace     = 2 * time.Second
	timeoutGrace    = 200 * time.Millisecond
)

// Result is the bounded output of one process.
type Result struct {
	Stdout    string
	Stderr    string
	ExitCode  int
	Truncated bool
}

type loggerKey struct{}

// WithLogger gives the process runner the logger of the command.
func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

type streamResult struct {
	data      []byte
	truncated bool
	err       error
}

// Command runs one process in its own process group. It returns the bounded
// output and the exit code. A non-zero exit code is not an error.
func Command(ctx context.Context, dir, name string, args ...string) (*Result, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	runCtx, cancel := context.WithTimeout(ctx, defaultDeadline)
	defer cancel()
	logger := loggerFrom(ctx)
	command := commandLine(name, args)
	started := time.Now()
	if logger != nil {
		logger.Debug("process start", "command", command)
	}

	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	stdoutRead, stdoutWrite, err := os.Pipe()
	if err != nil {
		return nil, fmt.Errorf("run %s: %s", name, err)
	}
	stderrRead, stderrWrite, err := os.Pipe()
	if err != nil {
		_ = stdoutRead.Close()
		_ = stdoutWrite.Close()
		return nil, fmt.Errorf("run %s: %s", name, err)
	}
	cmd.Stdout = stdoutWrite
	cmd.Stderr = stderrWrite

	if err := cmd.Start(); err != nil {
		closeAll(stdoutRead, stdoutWrite, stderrRead, stderrWrite)
		logProcessEnd(logger, command, -1, started, err)
		return nil, fmt.Errorf("start %s: %s", name, err)
	}
	_ = stdoutWrite.Close()
	_ = stderrWrite.Close()

	stdoutCh := make(chan streamResult, 1)
	stderrCh := make(chan streamResult, 1)
	go func() {
		data, truncated, err := readStream(stdoutRead, maxStream)
		stdoutCh <- streamResult{data: data, truncated: truncated, err: err}
	}()
	go func() {
		data, truncated, err := readStream(stderrRead, maxStream)
		stderrCh <- streamResult{data: data, truncated: truncated, err: err}
	}()

	waitCh := make(chan error, 1)
	go func() {
		waitCh <- cmd.Wait()
	}()

	var waitErr error
	timedOut := false
	select {
	case waitErr = <-waitCh:
	case <-runCtx.Done():
		timedOut = true
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		waitErr = <-waitCh
	}

	grace := normalGrace
	if timedOut {
		grace = timeoutGrace
	}
	stdout := collect(stdoutCh, stdoutRead, grace)
	stderr := collect(stderrCh, stderrRead, grace)
	_ = stdoutRead.Close()
	_ = stderrRead.Close()

	if timedOut {
		logProcessEnd(logger, command, -1, started, ErrTimeout)
		if errors.Is(runCtx.Err(), context.DeadlineExceeded) {
			return nil, ErrTimeout
		}
		return nil, runCtx.Err()
	}

	exitCode := 0
	if waitErr != nil {
		var exitError *exec.ExitError
		if !errors.As(waitErr, &exitError) {
			logProcessEnd(logger, command, -1, started, waitErr)
			return nil, fmt.Errorf("wait %s: %s", name, waitErr)
		}
		exitCode = exitError.ExitCode()
	}
	logProcessEnd(logger, command, exitCode, started, nil)
	return &Result{
		Stdout:    string(stdout.data),
		Stderr:    string(stderr.data),
		ExitCode:  exitCode,
		Truncated: stdout.truncated || stderr.truncated,
	}, nil
}

func collect(ch <-chan streamResult, file *os.File, grace time.Duration) streamResult {
	timer := time.NewTimer(grace)
	defer timer.Stop()
	select {
	case result := <-ch:
		return result
	case <-timer.C:
		_ = file.Close()
		return <-ch
	}
}

func readStream(reader io.Reader, limit int) ([]byte, bool, error) {
	buffer := make([]byte, 32*1024)
	data := []byte{}
	truncated := false
	for {
		count, err := reader.Read(buffer)
		if count > 0 {
			remaining := limit - len(data)
			if remaining > count {
				remaining = count
			}
			if remaining > 0 {
				data = append(data, buffer[:remaining]...)
			}
			if count > remaining {
				truncated = true
			}
		}
		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, os.ErrClosed) {
				return data, truncated, nil
			}
			return data, truncated, err
		}
	}
}

func commandLine(name string, args []string) string {
	if len(args) == 0 {
		return name
	}
	return name + " " + strings.Join(args, " ")
}

func logProcessEnd(logger *slog.Logger, command string, exitCode int, started time.Time, runErr error) {
	if logger == nil {
		return
	}
	attributes := []any{
		"command", command,
		"exit_code", exitCode,
		"duration_ms", time.Since(started).Milliseconds(),
	}
	if runErr != nil {
		attributes = append(attributes, "error", runErr.Error())
	}
	logger.Debug("process end", attributes...)
}

func loggerFrom(ctx context.Context) *slog.Logger {
	if ctx == nil {
		return nil
	}
	logger, _ := ctx.Value(loggerKey{}).(*slog.Logger)
	return logger
}

func closeAll(files ...*os.File) {
	for _, file := range files {
		_ = file.Close()
	}
}
