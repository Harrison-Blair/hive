package inbox

import (
	"strings"
	"testing"

	inboxlib "github.com/Harrison-Blair/hive/internal/lib/inbox"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
)

func TestReadTextEscapesTerminalControls(t *testing.T) {
	envelope := output.Envelope{Result: inboxlib.Result{Records: []records.Inbox{{
		Seq:       1,
		Role:      "coordinator",
		Kind:      "message\nFORGED",
		Agent:     "worker\x1b[31m",
		CreatedAt: "today\nFORGED",
		Text:      "hello\x1b[2J\nFORGED",
	}}}}
	text := ReadText(envelope)
	if strings.ContainsAny(text, "\x1b\r") {
		t.Fatalf("ReadText() = %q; want no terminal control bytes", text)
	}
	body := strings.TrimSuffix(text, "\n")
	if strings.Contains(body, "\n") {
		t.Fatalf("ReadText() = %q; want one physical line", text)
	}
	for _, escaped := range []string{`\n`, `\x1b`} {
		if !strings.Contains(body, escaped) {
			t.Fatalf("ReadText() = %q; want escaped text %q", text, escaped)
		}
	}
}

func TestReadTextKeepsSafeTextUnchanged(t *testing.T) {
	envelope := output.Envelope{Result: inboxlib.Result{Records: []records.Inbox{{
		Seq:       10,
		Role:      "coordinator",
		Kind:      records.InboxMessage,
		Agent:     "worker-x",
		CreatedAt: "2026-01-01T00:00:00Z",
		Text:      "hello",
	}}}}
	want := "10\tmessage\tworker-x\t2026-01-01T00:00:00Z\thello\n"
	if got := ReadText(envelope); got != want {
		t.Fatalf("ReadText() = %q; want %q", got, want)
	}
}
