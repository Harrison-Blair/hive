package tui

import (
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/cellgrid"
)

// AC3 of task 147 and tui-v2 decisions 42 and 44: each Agents board of the
// golden frames matches the View of its fixture cell for cell, with no mask.
func TestGfAgentsBoardsMatch(t *testing.T) {
	for _, board := range gfAgentsBoards {
		t.Run(board, func(t *testing.T) {
			want := readGfGrid(t, board)
			m := newGfDemoModel(t, board)
			got := renderGrid(t, m, want.Width, want.Height)
			diffs, err := cellgrid.Compare(want, got, nil)
			if err != nil {
				t.Fatal(err)
			}
			if len(diffs) != 0 {
				shown := diffs[:min(len(diffs), 40)]
				t.Fatalf("%d cells differ from %s.grid:\n%s\nrendered:\n%s", len(diffs), board, cellgrid.FormatDiffs(shown), m.View().Content)
			}
		})
	}
}
