package harness

import (
	"fmt"
	"strings"
)

// MaxStartLine is the most bytes of a start command line. Herdr agent.start
// types the line into the pane shell, and a shell in canonical mode can cut a
// line at 4095 bytes. See
// reference/dogfood/friction/entries/agent-start-line-cut-at-4095-bytes.md.
const MaxStartLine = 4000

// StartLineLength returns the bytes of the start command line of binary and
// args, with each argument in single quotes, as herdr types it. Each single
// quote in an argument counts as four bytes: it ends the quote, adds an
// escaped quote, and starts a new quote.
func StartLineLength(binary string, args []string) int {
	length := len(binary)
	for _, arg := range args {
		length += len(" ''") + len(arg) + 3*strings.Count(arg, "'")
	}
	return length
}

// CheckStartLine fails when the start command line of binary and args is
// longer than MaxStartLine.
func CheckStartLine(binary string, args []string) error {
	if length := StartLineLength(binary, args); length > MaxStartLine {
		return fmt.Errorf("the start command is %d bytes, more than the limit of %d bytes: use a shorter --model", length, MaxStartLine)
	}
	return nil
}
