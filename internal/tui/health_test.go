package tui

import (
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/cellgrid"
)

// checkBoardGolden builds the demo model from the fixture of board and
// compares the whole view with the golden grid, cell by cell (decision 42).
func checkBoardGolden(t *testing.T, board string) {
	t.Helper()
	want := readBoardGrid(t, board)
	m := newTestDemoModel(t, board+".fixture.json", demoSize{})
	got := renderGrid(t, m, want.Width, want.Height)
	diffs, err := cellgrid.Compare(want, got, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(diffs) != 0 {
		t.Fatalf("%d cells differ from %s.grid:\n%s\nrendered:\n%s", len(diffs), board, cellgrid.FormatDiffs(diffs), m.View().Content)
	}
}

func TestHealthAgentsMatchesItsBoard(t *testing.T) {
	checkBoardGolden(t, "HealthAgents")
}
