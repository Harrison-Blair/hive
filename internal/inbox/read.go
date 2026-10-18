package inbox

import (
	"io"

	"github.com/Harrison-Blair/hive/internal/lib/inbox"
	"github.com/Harrison-Blair/hive/internal/lib/output"
)

// Read returns all records or consumes the unread records of one role.
func Read(home, role string, all bool) output.Envelope {
	return output.Envelope{Outcome: "failed", Phase: "stub", Effects: []output.Effect{}, Error: "read not implemented", Result: nil}
}

// Send appends one inbox record.
func Send(home string, options inbox.AppendOptions) output.Envelope {
	return output.Envelope{Outcome: "failed", Phase: "stub", Effects: []output.Effect{}, Error: "send not implemented", Result: nil}
}

// ReadText renders records from a read result.
func ReadText(envelope output.Envelope) string { return "" }

// SendText renders the effects of one append.
func SendText(envelope output.Envelope) string { return "" }

// Consumed returns the sequence IDs that a read marked.
func Consumed(envelope output.Envelope) []uint64 { return nil }

// WriteReport tells stderr to run the command again with --all.
func WriteReport(stderr io.Writer, role string, ids []uint64) error { return nil }
