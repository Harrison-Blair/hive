// Package memory writes memory Markdown files and the memory index.
//
// write.go reads input, checks paths, and writes one memory under the store lock.
// index.go builds the index from memory files.
// frontmatter.go parses the memory frontmatter block.
package memory

import (
	"io"
	"log/slog"
	"time"

	"github.com/Harrison-Blair/hive/internal/lib/output"
)

// WriteOptions holds the values for one memory write.
type WriteOptions struct {
	Scope      string
	Name       string
	Input      io.Reader
	Deadline   time.Duration
	Logger     *slog.Logger
	IndexWrite func(path string, data []byte) error
}

// Write writes one memory and then the memory index.
func Write(home string, options WriteOptions) output.Envelope {
	return output.Envelope{Outcome: "failed", Phase: "empty", Effects: []output.Effect{}, Error: "empty", Result: nil}
}
