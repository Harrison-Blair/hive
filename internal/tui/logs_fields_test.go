package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"github.com/Harrison-Blair/hive/internal/lib/cellgrid"
)

// The WARN event of the GfLogs board, in the key order of the board.
const gfWarnEvent = `{"time":"2026-10-21T13:38:02.114Z","level":"WARN","msg":"verify fail","component":"task","command":"task verify","project_id":"hive","task_id":"2-stale-coordinator-startup","agent_id":"t2-verifier-1","request_id":"r-7f31c2","duration_ms":412,"items":{"V1":"pass","V2":"pass","V3":"pass","V4":"fail","V5":"pass"},"reason":"live restart in a scratch home left the old supervisor claim in place","truncated":false}`

// logs.md rule 4: label and value rows, a nested object indented by two
// cells with a blank row before and after it, a wrapped value at a word with
// a blank row after it. The value column starts 14 cells after the label.
func TestLogFieldLinesOfTheWarnEvent(t *testing.T) {
	want := []string{
		"time          2026-10-21T13:38:02.114Z",
		"level         WARN",
		"msg           verify fail",
		"component     task",
		"command       task verify",
		"project_id    hive",
		"task_id       2-stale-coordinator-startup",
		"agent_id      t2-verifier-1",
		"request_id    r-7f31c2",
		"duration_ms   412",
		"",
		"items",
		"  V1          pass",
		"  V2          pass",
		"  V3          pass",
		"  V4          fail",
		"  V5          pass",
		"",
		"reason        live restart in a scratch home left",
		"              the old supervisor claim in place",
		"",
		"truncated     false",
	}
	got := logFieldLines([]byte(gfWarnEvent), 36)
	if plain := stripLines(got); strings.Join(plain, "\n") != strings.Join(want, "\n") {
		t.Fatalf("field lines:\n%s\nwant:\n%s", strings.Join(plain, "\n"), strings.Join(want, "\n"))
	}
	// Labels are dim. WARN and fail are accent, other values fg.
	for _, tc := range []struct {
		line, col int
		token     cellgrid.Token
	}{{0, 0, cellgrid.Dim}, {0, 14, cellgrid.FG}, {1, 14, cellgrid.Accent}, {15, 14, cellgrid.Accent}, {14, 14, cellgrid.FG}} {
		g, err := cellgrid.Convert(got[tc.line], 50, 1)
		if err != nil {
			t.Fatal(err)
		}
		if tok := g.At(0, tc.col).Token; tok != tc.token {
			t.Errorf("line %d col %d token %v, want %v", tc.line, tc.col, tok, tc.token)
		}
	}
}

// logs.md rule 4: ERROR is error. A long value with no space wraps at the
// column width. A truncated event ends with "body cut at 4 KiB" in accent.
// The panel shows time, level, msg, and component first, then the other
// keys in the order of the event.
func TestLogFieldLinesOfATruncatedErrorEvent(t *testing.T) {
	raw := `{"body":"0123456789abcdefghij0123456789","component":"agent","level":"ERROR","msg":"spawn","time":"2026-10-21T14:12:40Z","truncated":true}`
	want := []string{
		"time          2026-10-21T14:12:40Z",
		"level         ERROR",
		"msg           spawn",
		"component     agent",
		"body          0123456789abcdefghij",
		"              0123456789",
		"",
		"truncated     true",
		"body cut at 4 KiB",
	}
	got := logFieldLines([]byte(raw), 20)
	if plain := stripLines(got); strings.Join(plain, "\n") != strings.Join(want, "\n") {
		t.Fatalf("field lines:\n%s\nwant:\n%s", strings.Join(plain, "\n"), strings.Join(want, "\n"))
	}
	for line, want := range map[int]cellgrid.Token{1: cellgrid.Error, 8: cellgrid.Accent} {
		col := 14
		if line == 8 {
			col = 0
		}
		g, err := cellgrid.Convert(got[line], 40, 1)
		if err != nil {
			t.Fatal(err)
		}
		if tok := g.At(0, col).Token; tok != want {
			t.Errorf("line %d token %v, want %v", line, tok, want)
		}
	}
}

func stripLines(lines []string) []string {
	plain := make([]string, len(lines))
	for i, line := range lines {
		plain[i] = strings.TrimRight(ansi.Strip(line), " ")
	}
	return plain
}
