package tui

import (
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/cellgrid"
)

// checkLogsBoard builds the model of board from its fixture with the demo
// loader, renders View at the board size, and compares each cell with the
// golden grid of the board (decision 42).
func checkLogsBoard(t *testing.T, board string) {
	t.Helper()
	if skipRetired(t, board) {
		return
	}
	want := readBoardGrid(t, board)
	m := newTestDemoModel(t, board+".fixture.json", demoSize{})
	got := renderGrid(t, m, want.Width, want.Height)
	diffs, err := cellgrid.Compare(want, got, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(diffs) != 0 {
		t.Fatalf("%d cells differ from %s.grid:\n%s", len(diffs), board, cellgrid.FormatDiffs(diffs))
	}
}

func TestLogsMatchesItsBoard(t *testing.T) {
	checkLogsBoard(t, "Logs")
}

func TestCompactLogsMatchesItsBoard(t *testing.T) {
	checkLogsBoard(t, "CompactLogs")
}
