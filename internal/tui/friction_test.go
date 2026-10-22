package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/Harrison-Blair/hive/internal/lib/cellgrid"
	"github.com/charmbracelet/x/ansi"
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

// AC2: at each size of decision 46, at the minimum height and at one row
// more, the screen fits, and the selected entry stays visible after G. At 80
// and 119 columns enter opens the overlay.
func TestFrictionFitsEachSize(t *testing.T) {
	for _, size := range []demoSize{{79, 24}, {80, 24}, {80, 25}, {119, 40}, {119, 41}, {120, 40}, {120, 41}} {
		for _, board := range []string{"Friction", "FrictionResolved"} {
			t.Run(fmt.Sprintf("%s/%dx%d", board, size.Width, size.Height), func(t *testing.T) {
				m := newTestDemoModel(t, board+".fixture.json", demoSize{})
				m, _ = raise(t, m, tea.WindowSizeMsg{Width: size.Width, Height: size.Height})
				if size.Width < minimumWidth {
					requireView(t, m, "terminal too small")
					return
				}
				m, _ = raise(t, m, key('G', "G"))
				entries := m.friction.visibleEntries()
				last := entries[len(entries)-1].name()
				view := requireFits(t, m, size.Width, size.Height)
				// The list cuts a name longer than its column with …, so the
				// check reads the first 20 cells of the name.
				if !strings.Contains(view, "▶ "+last[:20]) {
					t.Fatalf("selected entry %q is not visible:\n%s", last, view)
				}
				if size.Width >= 120 {
					return
				}
				m, _ = raise(t, m, key(0, "enter"))
				view = requireFits(t, m, size.Width, size.Height)
				if m.focusState() != "details" || !strings.Contains(view, last+" · esc back") {
					t.Fatalf("enter gave focus %q; want the details overlay of %s:\n%s", m.focusState(), last, view)
				}
			})
		}
	}
}

// AC2: with no entries and with text longer than its column, the screen
// fits at 80×24 and 120×40.
func TestFrictionFitsEmptyAndLongData(t *testing.T) {
	long := strings.Repeat("very-long-entry-name-", 6) + "end.md"
	for _, tt := range []struct {
		name    string
		entries []frictionEntry
		shows   []string
	}{
		{"empty", nil, []string{"open · 0", "no entry selected"}},
		{"long", []frictionEntry{{
			file: long, path: "entries/" + long, title: strings.Repeat("A long title ", 12), status: "Open", opened: "2026-10-21", valid: true,
			sections: []frictionSection{{title: "Summary", body: []string{strings.Repeat("word ", 60) + "tail"}}},
		}}, []string{"open · 1", "▶ very-long-entry-name-very-long-entry-", "…", "tail"}},
	} {
		for _, size := range []demoSize{{80, 24}, {120, 40}} {
			t.Run(fmt.Sprintf("%s/%dx%d", tt.name, size.Width, size.Height), func(t *testing.T) {
				m := newTestDemoModel(t, "Friction.fixture.json", demoSize{})
				m, _ = raise(t, m, tea.WindowSizeMsg{Width: size.Width, Height: size.Height})
				m.friction.accept(frictionData{entries: tt.entries})
				if size.Width < 120 {
					m, _ = raise(t, m, key(0, "enter"))
					m, _ = raise(t, m, key(0, "esc"))
				}
				m.invalidateView()
				view := requireFits(t, m, size.Width, size.Height)
				if size.Width < 120 && len(tt.entries) > 0 {
					// Below 120 columns the details are an overlay.
					m, _ = raise(t, m, key(0, "enter"))
					view += requireFits(t, m, size.Width, size.Height)
				}
				for _, text := range tt.shows {
					if size.Width < 120 && text == "no entry selected" {
						continue
					}
					if !strings.Contains(view, text) {
						t.Errorf("view omits %q:\n%s", text, view)
					}
				}
			})
		}
	}
}
