package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

// gfAgentsBoards are the 4 Agents boards of the golden frames.
var gfAgentsBoards = []string{"GfAgents", "GfAgentsPod", "GfAgentsDetails", "GfAgentsNone"}

// gfAgentsSelected gives the text of the selection of each Agents board that
// must stay on the screen at each size.
var gfAgentsSelected = map[string]string{
	"GfAgents":        "hive · wave B",
	"GfAgentsPod":     "t6-worker-1",
	"GfAgentsDetails": "t6-worker-1",
	"GfAgentsNone":    agentsNoneText,
}

// AC3 of task 147 and tui-v2 decision 46: each Agents board state at 79×24
// shows the too-small notice. At 80×24, 119×40, and 120×40 it fills the
// pane, no border breaks, the title row and the rule are rows 0 and 1, and
// the selection stays on the screen. From 120 columns the legend is the last
// body row, with one blank row below it (agents.md rule 7).
func TestGfAgentsBoardsKeepTheFrameAtEachBreakpoint(t *testing.T) {
	for _, size := range []demoSize{{79, 24}, {80, 24}, {119, 40}, {120, 40}} {
		for _, board := range gfAgentsBoards {
			t.Run(fmt.Sprintf("%dx%d/%s", size.Width, size.Height, board), func(t *testing.T) {
				m := newGfDemoModel(t, board)
				m, _ = raise(t, m, tea.WindowSizeMsg{Width: size.Width, Height: size.Height})
				view := ansi.Strip(m.View().Content)
				if size.Width < minimumWidth {
					requireView(t, m, "terminal too small")
					return
				}
				rows := strings.Split(view, "\n")
				if len(rows) != size.Height {
					t.Fatalf("view has %d rows, want %d:\n%s", len(rows), size.Height, view)
				}
				for index, row := range rows {
					if got := ansi.StringWidth(row); got != size.Width {
						t.Fatalf("row %d has width %d, want %d: %q", index, got, size.Width, row)
					}
				}
				// A lit line (decision 66) uses heavy glyphs on a light bus,
				// as ┏━◀━┹───╮. The light form of each heavy glyph gives
				// whole pairs again, so the check covers the lit line too.
				requireWholeAgentsBorders(t, gfLightGlyphs.Replace(view))
				if !strings.HasPrefix(rows[0], " agents ") || strings.Trim(rows[1], "─ ") != "" {
					t.Fatalf("rows 0 and 1 = %q, %q; want the title row and the rule", rows[0], rows[1])
				}
				if !strings.Contains(view, gfAgentsSelected[board]) {
					t.Fatalf("view does not show the selection %q:\n%s", gfAgentsSelected[board], view)
				}
				status := len(m.statusRows(size.Width, size.Height))
				legend := len(rows) - status - 2
				hasLegend := strings.HasPrefix(rows[legend], " ── spawn") && strings.TrimSpace(rows[legend+1]) == ""
				if wide := size.Width >= 120; hasLegend != wide {
					t.Fatalf("legend on row %d = %v, want %v:\n%s", legend, hasLegend, wide, view)
				}
			})
		}
	}
}

// gfLightGlyphs replaces each heavy box glyph with its light form.
var gfLightGlyphs = strings.NewReplacer(
	"┏", "╭", "┓", "╮", "┗", "╰", "┛", "╯", "━", "─", "┃", "│",
	"┹", "┴", "┺", "┴", "┻", "┴", "┷", "┴", "┸", "┴", "╃", "┼", "╅", "┼",
	"┰", "┬", "┯", "┬", "┱", "┬", "┲", "┬", "┳", "┬",
	"┨", "┤", "┥", "┤", "┫", "┤", "┠", "├", "┝", "├", "┣", "├", "┡", "├", "┟", "├",
)
