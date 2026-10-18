// Package handoff writes the Markdown handoff of one role.
package handoff

import (
	"io"
	"log/slog"
	"time"

	"github.com/Harrison-Blair/hive/internal/lib/output"
)

// WriteOptions holds the values for one handoff write.
type WriteOptions struct {
	Scope    string
	Input    io.Reader
	Deadline time.Duration
	Logger   *slog.Logger
}

// Write replaces one handoff file.
func Write(home string, options WriteOptions) output.Envelope {
	return output.Envelope{Outcome: "failed", Phase: "empty", Effects: []output.Effect{}, Error: "empty", Result: nil}
}
