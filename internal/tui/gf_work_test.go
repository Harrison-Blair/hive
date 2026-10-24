package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/Harrison-Blair/hive/internal/lib/cellgrid"
)

// gfWorkBoards are the 8 Work screen boards of the golden frames.
var gfWorkBoards = []string{
	"GfWork", "GfWorkEmpty", "GfWorkNoGoal", "GfWorkScroll",
	"GfWorkStops", "GfWorkCancelled", "GfWorkTask", "GfWorkDeps",
}

// gfWorkStatusLines is the status row of the Work boards: state, screen,
// and clock. Ask a27 decides if this row is the default of each screen
// except Chat. Until the answer, the test sets the row in the config.
var gfWorkStatusLines = [][]string{{"state", "screen", "clock"}}

// newGfWorkModel gives the demo model of a Work board with the status row
// of the boards.
func newGfWorkModel(t *testing.T, board string) *Model {
	t.Helper()
	m := newGfDemoModel(t, board)
	m.tuiConfig.StatusLines = gfWorkStatusLines
	m.invalidateView()
	return m
}

// newBoardDemoModel gives the demo model of a board: a golden-frames board
// when its name starts with Gf, else a tui-v2 board.
func newBoardDemoModel(t *testing.T, board string) *Model {
	t.Helper()
	if strings.HasPrefix(board, "Gf") {
		return newGfDemoModel(t, board)
	}
	return newTestDemoModel(t, board+".fixture.json", demoSize{})
}

// Each Work board of the golden frames matches the View of its fixture cell
// for cell, with no mask.
func TestGfWorkBoardsMatch(t *testing.T) {
	for _, board := range gfWorkBoards {
		t.Run(board, func(t *testing.T) {
			want := readGfGrid(t, board)
			m := newGfWorkModel(t, board)
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

// Each Work board state at 79×24 shows the too-small notice. At 80×24,
// 119×40, and 120×40 it fills the pane and no border breaks. The selected
// card shows, and an open popup stays open after the resize.
func TestGfWorkBoardsKeepTheFrameAtEachBreakpoint(t *testing.T) {
	for _, size := range []demoSize{{79, 24}, {80, 24}, {119, 40}, {120, 40}} {
		for _, board := range gfWorkBoards {
			t.Run(fmt.Sprintf("%dx%d/%s", size.Width, size.Height, board), func(t *testing.T) {
				m := newGfWorkModel(t, board)
				popup, selected := m.tasks.popup, m.tasks.selectedID
				m, _ = raise(t, m, tea.WindowSizeMsg{Width: size.Width, Height: size.Height})
				if size.Width < minimumWidth {
					requireView(t, m, "terminal too small")
					return
				}
				switch popup {
				case workPopupTask:
					requireWorkFits(t, m, size.Width, size.Height, workTaskPopupWidth, workTaskPopupHeight)
					requireView(t, m, selected+" · esc closes")
				case workPopupDeps:
					requireWorkFits(t, m, size.Width, size.Height, workDepsPopupWidth, workDepsPopupHeight)
					requireView(t, m, selected+" · dependencies · esc closes")
				default:
					requireWorkFits(t, m, size.Width, size.Height, 0, 0)
					if selected != "" {
						if got := selectedCard(t, m); got != selected {
							t.Errorf("selected card %q, want %q", got, selected)
						}
					}
				}
				if m.tasks.popup != popup {
					t.Errorf("popup %q after the resize, want %q", m.tasks.popup, popup)
				}
			})
		}
	}
}
