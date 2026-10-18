package inbox

import (
	"fmt"
	"io"
	"strings"

	inboxlib "github.com/Harrison-Blair/hive/internal/lib/inbox"
	"github.com/Harrison-Blair/hive/internal/lib/output"
)

// ReadText renders the records of one read result.
func ReadText(envelope output.Envelope) string {
	result, ok := envelope.Result.(inboxlib.Result)
	if !ok || len(result.Records) == 0 {
		return ""
	}
	lines := make([]string, 0, len(result.Records))
	for _, record := range result.Records {
		lines = append(lines, fmt.Sprintf("%d\t%s\t%s\t%s\t%s", record.Seq, record.Kind, record.Agent, record.CreatedAt, record.Text))
	}
	return strings.Join(lines, "\n") + "\n"
}

// SendText renders the effects of one append.
func SendText(envelope output.Envelope) string {
	lines := make([]string, 0, len(envelope.Effects))
	for _, effect := range envelope.Effects {
		lines = append(lines, fmt.Sprintf("%s %s %s", effect.Action, effect.Resource, effect.Path))
	}
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}

// Consumed returns the sequence IDs that a read marked.
func Consumed(envelope output.Envelope) []uint64 {
	result, ok := envelope.Result.(inboxlib.Result)
	if !ok {
		return nil
	}
	ids := make([]uint64, 0, len(result.Records))
	for _, record := range result.Records {
		ids = append(ids, record.Seq)
	}
	return ids
}

// WriteReport tells stderr to run the command again with --all.
func WriteReport(stderr io.Writer, role string, ids []uint64) error {
	for _, id := range ids {
		if _, err := fmt.Fprintf(stderr, "consumed %d\n", id); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintf(stderr, "run: hive inbox --role %s --all\n", role)
	return err
}
