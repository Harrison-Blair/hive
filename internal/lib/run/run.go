// Package run runs one external process with bounded time and output.
package run

import (
	"context"
	"errors"
	"log/slog"
)

// ErrTimeout reports that the deadline stopped the process group.
var ErrTimeout = errors.New("process timeout")

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

// Command runs one process in its own process group.
func Command(ctx context.Context, dir, name string, args ...string) (*Result, error) {
	return &Result{}, nil
}

// readStream reads one output stream up to a limit.
func readStream(reader interface{ Read([]byte) (int, error) }, limit int) ([]byte, bool, error) {
	return nil, false, nil
}
