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
		}}, []string{"open · 1", "▶ very-long-entry-name-very-long-", "…", "tail"}},
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

// footerText gives the plain text of the footer row.
func footerText(m *Model) string {
	lines := strings.Split(ansi.Strip(m.View().Content), "\n")
	return strings.TrimRight(lines[len(lines)-1], " ")
}

// AC4: the keys of the Friction screen follow decisions 21 to 23, and the
// footer shows the key hints of each focus state.
func TestFrictionKeys(t *testing.T) {
	m := newTestDemoModel(t, "Friction.fixture.json", demoSize{})
	selected := func() string { entry, _ := m.friction.selected(); return entry.name() }
	if got := footerText(m); got != " [list] ↑↓ move · tab region · [ ] view · enter open · g/G top/end · ? help" {
		t.Fatalf("list footer = %q", got)
	}
	for _, step := range []struct {
		press tea.KeyPressMsg
		want  string
	}{
		{key('j', "j"), "ste-baseline-row-conflicts"},
		{tea.KeyPressMsg{Code: tea.KeyDown}, "verify-fail-lost-on-pass"},
		{key('k', "k"), "ste-baseline-row-conflicts"},
		{tea.KeyPressMsg{Code: tea.KeyUp}, "scratch-hive-starts-real-agents"},
		{key('G', "G"), "worker-gate-adds-golangci-lint"},
		{key('g', "g"), "agent-send-flag-means-target"},
		{tea.KeyPressMsg{Code: tea.KeyPgDown}, "worker-gate-adds-golangci-lint"},
		{tea.KeyPressMsg{Code: tea.KeyPgUp}, "agent-send-flag-means-target"},
		{key('a', "a"), "agent-send-flag-means-target"},
		{key('/', "/"), "agent-send-flag-means-target"},
	} {
		m, _ = raise(t, m, step.press)
		if got := selected(); got != step.want {
			t.Fatalf("key %q selected %q; want %q", step.press.String(), got, step.want)
		}
	}
	if m.focusState() != "list" {
		t.Fatalf("focus = %q after the list keys; want list", m.focusState())
	}
	m, _ = raise(t, m, key(']', "]"))
	if m.activeView() != 1 || selected() != "stale-coordinator-blocks-startup" {
		t.Fatalf("] gave view %d, selection %q; want Resolved, stale-coordinator-blocks-startup", m.activeView(), selected())
	}
	m, _ = raise(t, m, key('[', "["))
	if m.activeView() != 0 {
		t.Fatalf("[ gave view %d; want Open", m.activeView())
	}
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyTab})
	if got := footerText(m); got != " [details] ↑↓ scroll · tab region · esc back to list · ? help" {
		t.Fatalf("details footer = %q", got)
	}
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyDown})
	if selected() != "agent-send-flag-means-target" || m.detailsOffset[tabFriction] != 1 {
		t.Fatalf("↓ in the details gave selection %q, offset %d; want the same entry, offset 1", selected(), m.detailsOffset[tabFriction])
	}
	m, _ = raise(t, m, key('G', "G"))
	if view := ansi.Strip(m.View().Content); selected() != "agent-send-flag-means-target" || strings.Contains(view, "more lines") || !strings.Contains(view, "names the sender.") {
		t.Fatalf("G in the details gave selection %q and not the end of the details:\n%s", selected(), view)
	}
	m, _ = raise(t, m, key('g', "g"))
	if m.detailsOffset[tabFriction] != 0 {
		t.Fatalf("g in the details gave offset %d; want 0", m.detailsOffset[tabFriction])
	}
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyEscape})
	if m.focusState() != "list" {
		t.Fatalf("esc in the details gave focus %q; want list", m.focusState())
	}

	m, _ = raise(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	if got := footerText(m); got != " [list] ↑↓ move · enter details · [ ] view · ?" {
		t.Fatalf("compact list footer = %q", got)
	}
	m, _ = raise(t, m, key(0, "enter"))
	if got := footerText(m); got != " [details] ↑↓ scroll · esc back to list" {
		t.Fatalf("compact details footer = %q", got)
	}
}
