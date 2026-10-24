package inbox

import (
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"

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
		line := fmt.Sprintf("%d\t%s\t%s\t%s\t%s", record.Seq, renderValue(record.Kind), renderValue(record.Agent), renderValue(record.CreatedAt), renderValue(record.Text))
		if record.Path != "" {
			line += "\t" + renderValue(record.Path)
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n") + "\n"
}

func renderValue(value string) string {
	for _, r := range value {
		if unicode.IsControl(r) {
			return strconv.Quote(value)
		}
	}
	return value
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

// Consumed returns the sequence IDs of completed consume effects.
func Consumed(envelope output.Envelope) []uint64 {
	ids := make([]uint64, 0, len(envelope.Effects))
	for _, effect := range envelope.Effects {
		if effect.Resource != "inbox" || effect.Action != "consumed" {
			continue
		}
		id := strings.TrimSuffix(filepath.Base(effect.Path), ".json")
		seq, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			continue
		}
		ids = append(ids, seq)
	}
	return ids
}

// WriteReport tells stderr to run the command again with --all.
func WriteReport(stderr io.Writer, inbox string, ids []uint64) error {
	for _, id := range ids {
		if _, err := fmt.Fprintf(stderr, "consumed %d\n", id); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintf(stderr, "run: hive inbox --inbox %s --all\n", inbox)
	return err
}
