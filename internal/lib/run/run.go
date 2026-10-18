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
		return nil, diagnostic("run", name, err)
	}
	stderrRead, stderrWrite, err := os.Pipe()
	if err != nil {
		closeAll(stdoutRead, stdoutWrite)
		return nil, diagnostic("run", name, err)
	}
	cmd.Stdout = stdoutWrite
	cmd.Stderr = stderrWrite

	if err := cmd.Start(); err != nil {
		closeAll(stdoutRead, stdoutWrite, stderrRead, stderrWrite)
		logProcessEnd(logger, command, -1, started, err)
		return nil, diagnostic("start", name, err)
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

	stopGroup := func() {
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}

	var waitErr error
	stopped := false
	select {
	case waitErr = <-waitCh:
	case <-runCtx.Done():
		stopped = true
		stopGroup()
		waitErr = <-waitCh
	}

	stdout, stderr, stoppedDuringDrain := awaitStreams(runCtx, stopGroup, stdoutCh, stderrCh, stdoutRead, stderrRead)
	_ = stdoutRead.Close()
	_ = stderrRead.Close()
	if stoppedDuringDrain {
		stopped = true
	}

	if stopped {
		runErr := runResultError(runCtx)
		logProcessEnd(logger, command, -1, started, runErr)
		return nil, runErr
	}

	exitCode := 0
	if waitErr != nil {
		var exitError *exec.ExitError
		if !errors.As(waitErr, &exitError) {
			logProcessEnd(logger, command, -1, started, waitErr)
			return nil, diagnostic("wait", name, waitErr)
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

func awaitStreams(ctx context.Context, stop func(), stdoutCh, stderrCh <-chan streamResult, stdoutRead, stderrRead *os.File) (streamResult, streamResult, bool) {
	var stdout, stderr streamResult
	stdoutPending, stderrPending := true, true
	stopped := false
	for stdoutPending || stderrPending {
		var outCh, errCh <-chan streamResult
		if stdoutPending {
			outCh = stdoutCh
		}
		if stderrPending {
			errCh = stderrCh
		}
		if stopped {
			select {
			case result := <-outCh:
				stdout, stdoutPending = result, false
			case result := <-errCh:
				stderr, stderrPending = result, false
			}
			continue
		}
		select {
		case result := <-outCh:
			stdout, stdoutPending = result, false
		case result := <-errCh:
			stderr, stderrPending = result, false
		case <-ctx.Done():
			stopped = true
			stop()
			_ = stdoutRead.Close()
			_ = stderrRead.Close()
		}
	}
	return stdout, stderr, stopped
}

func runResultError(ctx context.Context) error {
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return ErrTimeout
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return ErrTimeout
}

func diagnostic(verb, name string, err error) error {
	return errors.New(oneLine(fmt.Sprintf("%s %s: %s", verb, name, err)))
}

func oneLine(text string) string {
	return strings.Map(func(value rune) rune {
		if value < 0x20 || value == 0x7f {
			return '?'
		}
		return value
	}, text)
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
		attributes = append(attributes, "error", oneLine(runErr.Error()))
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
