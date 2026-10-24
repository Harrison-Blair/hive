package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/Harrison-Blair/hive/internal/lib/cellgrid"
)

// gfLogsBoards are the 4 Logs screen boards of the golden frames.
var gfLogsBoards = []string{"GfLogs", "GfLogsEmpty", "GfLogsRaw", "GfLogsError"}

// AC3, tui-v2 decisions 42 and 44: each Logs board matches the View of its
// fixture cell for cell, with no mask.
func TestGfLogsBoardsMatch(t *testing.T) {
	for _, board := range gfLogsBoards {
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

// AC3 and tui-v2 decision 46: each Logs board state at 79×24 shows the
// too-small notice. At 80×24, 119×40, and 120×40 it fills the pane, no
// border breaks, the title row is on row 0, the boxes close on the row above
// the blank row and the status rows, and the selected event shows in the
// list or in the overlay.
func TestGfLogsBoardsKeepTheFrameAtEachBreakpoint(t *testing.T) {
	for _, size := range []demoSize{{79, 24}, {80, 24}, {119, 40}, {120, 40}} {
		for _, board := range gfLogsBoards {
			t.Run(fmt.Sprintf("%dx%d/%s", size.Width, size.Height, board), func(t *testing.T) {
				m := newGfDemoModel(t, board)
				m, _ = raise(t, m, tea.WindowSizeMsg{Width: size.Width, Height: size.Height})
				if size.Width < minimumWidth {
					requireView(t, m, "terminal too small")
					return
				}
				rows := strings.Split(requireFits(t, m, size.Width, size.Height), "\n")
				if !strings.HasPrefix(rows[0], " logs · ") {
					t.Fatalf("row 0 = %q; want the title row", rows[0])
				}
				status := len(m.statusRows(size.Width, size.Height))
				bottom := len(rows) - status - 2
				if !strings.HasPrefix(rows[bottom], "╰") || strings.TrimSpace(rows[bottom+1]) != "" {
					t.Fatalf("row %d = %q, row %d = %q; want the box bottom, then a blank row:\n%s", bottom, rows[bottom], bottom+1, rows[bottom+1], strings.Join(rows, "\n"))
				}
				if row, ok := m.selectedLogRow(); ok {
					// Below 120 columns the panel focus is the overlay
					// of the selected event.
					want := "▶ " + logTime(row.event, m.now().Location())
					if m.detailsFocused() && size.Width < 120 {
						want = "event " + logTime(row.event, m.now().Location()) + " · "
					}
					if !strings.Contains(strings.Join(rows, "\n"), want) {
						t.Fatalf("selected event %q does not show:\n%s", want, strings.Join(rows, "\n"))
					}
				}
			})
		}
	}
}
