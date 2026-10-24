package tui

import (
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/cellgrid"
	"github.com/charmbracelet/x/ansi"
)

// checkBoard compares the whole demo view of board with its golden grid
// (decision 42).
func checkBoard(t *testing.T, board string) {
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

// AC2: each Friction board matches its golden grid cell for cell.
func TestFrictionMatchesItsBoards(t *testing.T) {
	for _, board := range []string{"Friction", "FrictionResolved", "CompactFriction"} {
		t.Run(board, func(t *testing.T) { checkBoard(t, board) })
	}
}

// requireFits fails when a line of the view is not width cells wide, when
// the view is not height lines, or when a border breaks (decision 46).
func requireFits(t *testing.T, m *Model, width, height int) string {
	t.Helper()
	view := ansi.Strip(m.View().Content)
	lines := strings.Split(view, "\n")
	if len(lines) != height {
		t.Fatalf("view has %d lines, want %d:\n%s", len(lines), height, view)
	}
	for index, line := range lines {
		if got := ansi.StringWidth(line); got != width {
			t.Fatalf("line %d has width %d, want %d: %q", index+1, got, width, line)
		}
	}
	requireWholeBorders(t, view)
	return view
}
