package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/Harrison-Blair/hive/internal/lib/cellgrid"
)

// gfHelpBoards are the Help boards of the golden frames: help.md gives one.
var gfHelpBoards = []string{"GfHelp"}

// help.md state help, tui-v2 decisions 42 and 44: each Help board matches
// the View of its fixture cell for cell, with no mask.
func TestGfHelpBoardsMatch(t *testing.T) {
	for _, board := range gfHelpBoards {
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

// help.md rules 1 and 4, tui-v2 decision 46: each Help board at 79×24 shows
// the too-small notice. At 80×24, 119×40, and 120×40 the view fills the
// pane, no border breaks, the Help box is whole and at the center, and a
// body taller than the box has the thumb. Golden-frames decision 5: a resize
// keeps the overlay and its scroll position.
func TestGfHelpBoardsKeepTheBoxAtEachBreakpoint(t *testing.T) {
	for _, size := range []demoSize{{79, 24}, {80, 24}, {119, 40}, {120, 40}} {
		for _, board := range gfHelpBoards {
			t.Run(fmt.Sprintf("%dx%d/%s", size.Width, size.Height, board), func(t *testing.T) {
				m := newGfDemoModel(t, board)
				m, _ = raise(t, m, tea.WindowSizeMsg{Width: size.Width, Height: size.Height})
				if size.Width < minimumWidth {
					requireView(t, m, "terminal too small")
					return
				}
				rows := strings.Split(requireFits(t, m, size.Width, size.Height), "\n")
				top, bottom := helpBox(t, rows)
				_, left := cellOf(rows, "╭─ "+helpTitle)
				width := min(helpBoxWidth, size.Width)
				if margin := size.Width - width; left != margin/2 {
					t.Errorf("box at column %d; want %d at the center", left, margin/2)
				}
				if margin := size.Height - (bottom - top + 1); top != margin/2 {
					t.Errorf("box at row %d; want %d at the center", top, margin/2)
				}
				thumb := false
				for row := top; row <= bottom; row++ {
					first, last := ansi.Cut(rows[row], left, left+1), ansi.Cut(rows[row], left+width-1, left+width)
					edge := [2]string{"│", "│┃"}
					switch row {
					case top:
						edge = [2]string{"╭", "╮"}
					case bottom:
						edge = [2]string{"╰", "╯"}
					}
					if first != edge[0] || !strings.Contains(edge[1], last) {
						t.Fatalf("box row %d has the edges %q and %q:\n%s", row, first, last, strings.Join(rows, "\n"))
					}
					thumb = thumb || last == "┃"
				}
				if fits := len(m.helpLines())+2 <= size.Height; thumb == fits {
					t.Errorf("thumb %v with %d body lines in %d rows", thumb, len(m.helpLines()), size.Height)
				}
				m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyDown})
				offset := m.helpOffset
				m, _ = raise(t, m, tea.WindowSizeMsg{Width: size.Width + 1, Height: size.Height})
				if !m.helpOpen || m.helpOffset != offset {
					t.Errorf("resize gave help %v, offset %d; want open, offset %d", m.helpOpen, m.helpOffset, offset)
				}
			})
		}
	}
}
