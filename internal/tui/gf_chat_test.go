package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/Harrison-Blair/hive/internal/lib/cellgrid"
)

// gfDesignDir holds the golden-frames boards, their grids, and their
// fixtures.
const gfDesignDir = "../../.hive/spec/golden-frames/design"

// gfChatBoards are the 14 Chat screen boards of the golden frames (AC8).
var gfChatBoards = []string{
	"GfChat", "GfChatEmpty", "GfChatEvents", "GfChatTall", "GfChatMarkdown",
	"GfChatNoCoordinator", "GfChatNotice", "GfChatAsk", "GfChatAskInput",
	"GfChatAskMulti", "GfChatAskPreview", "GfChatAskStopped", "GfChatAskReview",
	"GfChatAskMany",
}

// newGfDemoModel gives the demo model of the fixture of a golden-frames
// board.
func newGfDemoModel(t *testing.T, board string) *Model {
	t.Helper()
	loaded, err := loadDemoFixture(filepath.Join(gfDesignDir, board+".fixture.json"))
	if err != nil {
		t.Fatal(err)
	}
	m := newDemoModel(t.TempDir(), &demoState{fixture: loaded})
	t.Cleanup(m.Stop)
	return m
}

// readGfGrid reads the golden grid of a golden-frames board.
func readGfGrid(t *testing.T, board string) cellgrid.Grid {
	t.Helper()
	file, err := os.Open(filepath.Join(gfDesignDir, board+".grid"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = file.Close() }()
	grid, err := cellgrid.ReadGrid(file)
	if err != nil {
		t.Fatal(err)
	}
	return grid
}

// AC8, tui-v2 decisions 42 and 44: each Chat board of the golden frames
// matches the View of its fixture cell for cell, with no mask.
func TestGfChatBoardsMatch(t *testing.T) {
	for _, board := range gfChatBoards {
		t.Run(board, func(t *testing.T) {
			want := readGfGrid(t, board)
			m := newGfDemoModel(t, board)
			got := renderGrid(t, m, want.Width, want.Height)
			diffs, err := cellgrid.Compare(want, got, nil)
			if err != nil {
				t.Fatal(err)
			}
			if len(diffs) != 0 {
				shown := diffs[:min(len(diffs), 12)]
				t.Fatalf("%d cells differ from %s.grid:\n%s\nrendered:\n%s", len(diffs), board, cellgrid.FormatDiffs(shown), m.View().Content)
			}
		})
	}
}

// AC9 and tui-v2 decision 46: each Chat board state at 79×24 shows the
// too-small notice. At 80×24, 119×40, and 120×40 it fills the pane, and no
// border breaks.
func TestGfChatBoardsKeepTheFrameAtEachBreakpoint(t *testing.T) {
	for _, size := range []demoSize{{79, 24}, {80, 24}, {119, 40}, {120, 40}} {
		for _, board := range gfChatBoards {
			t.Run(fmt.Sprintf("%dx%d/%s", size.Width, size.Height, board), func(t *testing.T) {
				m := newGfDemoModel(t, board)
				m, _ = raise(t, m, tea.WindowSizeMsg{Width: size.Width, Height: size.Height})
				if size.Width < minimumWidth {
					requireView(t, m, "terminal too small")
					return
				}
				requireFits(t, m, size.Width, size.Height)
			})
		}
	}
}
