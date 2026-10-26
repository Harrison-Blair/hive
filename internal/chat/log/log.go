// Package log renders the chat history for the chat log command.
package log

import (
	"fmt"
	"strings"

	chatlib "github.com/Harrison-Blair/hive/internal/chat"
	"github.com/Harrison-Blair/hive/internal/lib/output"
)

// Text renders each message and the count of malformed lines.
func Text(envelope output.Envelope) string {
	result, ok := envelope.Result.(chatlib.LogResult)
	if !ok {
		return ""
	}
	var text strings.Builder
	for _, entry := range result.Entries {
		fmt.Fprintf(&text, "%s %s", entry.ID, entry.From)
		if entry.Delivery != "" {
			fmt.Fprintf(&text, " [%s]", entry.Delivery)
		}
		if entry.ReplyTo != "" {
			fmt.Fprintf(&text, " -> %s", entry.ReplyTo)
		}
		fmt.Fprintf(&text, ": %s\n", entry.Text)
	}
	if result.BadLines != 0 {
		fmt.Fprintf(&text, "skipped %d bad chat lines\n", result.BadLines)
	}
	return text.String()
}
