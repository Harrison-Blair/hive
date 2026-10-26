package tui

import (
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/cellgrid"
	"github.com/charmbracelet/x/ansi"
)

// AC3 of task 147 and tui-v2 decisions 42 and 44: each Agents board of the
// golden frames matches the View of its fixture cell for cell, with no mask.
// D11 replaces the role row by the profile row and groups a general agent
// with a project in that project. The test applies that decision to the
// expected grid and keeps the board file unchanged.
func TestGfAgentsBoardsMatch(t *testing.T) {
	for _, board := range gfAgentsBoards {
		t.Run(board, func(t *testing.T) {
			want := d11AgentsGrid(t, board, readGfGrid(t, board))
			m := newGfDemoModel(t, board)
			got := renderGrid(t, m, want.Width, want.Height)
			diffs, err := cellgrid.Compare(want, got, nil)
			if err != nil {
				t.Fatal(err)
			}
			if len(diffs) != 0 {
				shown := diffs[:min(len(diffs), 40)]
				t.Fatalf("%d cells differ from the D11 form of %s.grid:\n%s\nrendered:\n%s", len(diffs), board, cellgrid.FormatDiffs(shown), m.View().Content)
			}
		})
	}
}

// d11AgentsGrid gives the expected golden grid with the D11 changes. The
// older boards show the role row and put spec-check-1 in home because its
// parent card is gone. D11 shows the profile row and puts an agent in the
// pod of its project.
func d11AgentsGrid(t *testing.T, board string, grid cellgrid.Grid) cellgrid.Grid {
	t.Helper()
	switch board {
	case "GfAgents":
		// spec-check-1 moves from home to the hive pod: the home note goes
		// and the hive stopped count rises from 1 to 2.
		clearGridCells(&grid, 15, 83, 17, cellgrid.FG)
		setGridRune(&grid, 25, 11, '2')
	case "GfAgentsDetails":
		// The details row shows profile builder in place of role worker.
		clearGridCells(&grid, 10, 31, 58, cellgrid.FG)
		setGridText(&grid, 10, 31, " ", cellgrid.FG)
		setGridText(&grid, 10, 32, "profile", cellgrid.Dim)
		setGridText(&grid, 10, 39, "   ", cellgrid.FG)
		setGridText(&grid, 10, 42, "builder", cellgrid.FG)
		// The home note goes and the hive stopped count rises to 2.
		clearGridCells(&grid, 15, 91, 9, cellgrid.Border)
		setGridRune(&grid, 25, 11, '2')
	}
	return grid
}

// clearGridCells sets width cells from row and col to a blank with token.
func clearGridCells(grid *cellgrid.Grid, row, col, width int, token cellgrid.Token) {
	for index := range width {
		grid.Cells[row*grid.Width+col+index] = cellgrid.Cell{Rune: ' ', Token: token}
	}
}

// setGridRune changes one rune and keeps the token and the bold flag.
func setGridRune(grid *cellgrid.Grid, row, col int, r rune) {
	cell := &grid.Cells[row*grid.Width+col]
	cell.Rune = r
}

// setGridText writes text at row and col with token.
func setGridText(grid *cellgrid.Grid, row, col int, text string, token cellgrid.Token) {
	for index, r := range []rune(text) {
		grid.Cells[row*grid.Width+col+index] = cellgrid.Cell{Rune: r, Token: token}
	}
}

// The D11 details row shows profile builder, not role worker.
func TestGfAgentsDetailsShowsTheD11ProfileRow(t *testing.T) {
	m := newGfDemoModel(t, "GfAgentsDetails")
	m.invalidateView()
	rows := strings.Split(ansi.Strip(m.View().Content), "\n")
	if got := string([]rune(rows[10])[31:89]); got != " profile   builder"+strings.Repeat(" ", 40) {
		t.Fatalf("details row = %q; want the profile label of D11", got)
	}
	if strings.Contains(rows[10], "role") {
		t.Fatalf("details row shows a role label: %q", rows[10])
	}
}
