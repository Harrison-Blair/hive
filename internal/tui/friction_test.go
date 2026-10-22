package tui

import (
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/cellgrid"
)

// checkBoard compares the whole demo view of board with its golden grid
// (decision 42).
func checkBoard(t *testing.T, board string) {
	t.Helper()
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

// AC2: each Friction board matches its golden grid cell for cell.
func TestFrictionMatchesItsBoards(t *testing.T) {
	for _, board := range []string{"Friction", "FrictionResolved", "CompactFriction"} {
		t.Run(board, func(t *testing.T) { checkBoard(t, board) })
	}
}
