package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/Harrison-Blair/hive/internal/lib/cellgrid"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/tui/component"
)

// checkBoard compares the demo view of the fixture of board with its golden
// grid, cell for cell (decision 42).
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

// AC2: the Tasks view matches the Work board.
func TestWorkTasksMatchesWorkBoard(t *testing.T) {
	checkBoard(t, "Work")
}

// AC1, AC2: below 120 columns the details open as an overlay, as the Compact
// board shows.
func TestWorkTasksMatchesCompactBoard(t *testing.T) {
	checkBoard(t, "Compact")
}

// requireFits fails when view is not width×height cells or has a broken
// border (decision 46).
func requireFits(t *testing.T, view string, width, height int) {
	t.Helper()
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
}

// AC2: decision 46. At 79 columns the TUI shows the too-small notice. At 80
// and 119 columns the Tasks view shows the list, and the overlay after
// enter. From 120 columns the list and the details sit side by side. Each
// size is checked at its height and at one row more.
func TestWorkTasksFitsEachSize(t *testing.T) {
	for _, size := range []demoSize{{79, 24}, {80, 24}, {80, 25}, {119, 40}, {119, 41}, {120, 40}, {120, 41}} {
		t.Run(fmt.Sprintf("%dx%d", size.Width, size.Height), func(t *testing.T) {
			m := newTestDemoModel(t, "Work.fixture.json", demoSize{})
			m, _ = raise(t, m, tea.WindowSizeMsg{Width: size.Width, Height: size.Height})
			view := ansi.Strip(m.View().Content)
			if size.Width < minimumWidth {
				if !strings.Contains(view, "terminal too small") {
					t.Fatalf("view at %d columns is not the too-small notice:\n%s", size.Width, view)
				}
				return
			}
			requireFits(t, view, size.Width, size.Height)
			if !strings.Contains(view, "▶ 3-tui-restart") {
				t.Fatalf("selected task is not visible:\n%s", view)
			}
			if size.Width >= component.LongFormWidth {
				if !strings.Contains(view, "╭─ 3-tui-restart ─") {
					t.Fatalf("details are not beside the list:\n%s", view)
				}
				return
			}
			if strings.Contains(view, "ACCEPTANCE") {
				t.Fatalf("details show before enter below 120 columns:\n%s", view)
			}
			m, _ = raise(t, m, key(tea.KeyEnter, "enter"))
			view = ansi.Strip(m.View().Content)
			requireFits(t, view, size.Width, size.Height)
			if !strings.Contains(view, "╭─ 3-tui-restart · esc back") || !strings.Contains(view, "list › details") {
				t.Fatalf("enter did not open the details overlay:\n%s", view)
			}
		})
	}
}

// demoWithTasks gives the Work demo model at width×height with its tasks
// replaced.
func demoWithTasks(t *testing.T, width, height int, tasks ...records.Task) *Model {
	t.Helper()
	m := newTestDemoModel(t, "Work.fixture.json", demoSize{})
	m, _ = raise(t, m, tea.WindowSizeMsg{Width: width, Height: height})
	items := make([]taskItem, 0, len(tasks))
	for _, task := range tasks {
		items = append(items, taskItem{record: task})
	}
	m.tasks.accept(tasksData{items: items})
	m.invalidateView()
	return m
}

// AC2: with no task, the list is empty and the details say so.
func TestWorkTasksEmptyData(t *testing.T) {
	for _, width := range []int{80, 120} {
		m := demoWithTasks(t, width, 24)
		view := ansi.Strip(m.View().Content)
		requireFits(t, view, width, 24)
		if !strings.Contains(view, "tasks · 0") {
			t.Fatalf("empty list at %d columns:\n%s", width, view)
		}
		if width >= component.LongFormWidth && !strings.Contains(view, "no task selected") {
			t.Fatalf("empty details at %d columns:\n%s", width, view)
		}
		m, _ = raise(t, m, key('j', "j"))
		m, _ = raise(t, m, key('b', "b"))
		requireFits(t, ansi.Strip(m.View().Content), width, 24)
	}
}

// AC2: text longer than its column is cut with … and keeps the frame.
func TestWorkTasksCutsLongText(t *testing.T) {
	long := records.Task{ID: strings.Repeat("long-id-", 8), Title: strings.Repeat("a long title ", 20), Status: "assigned", DependsOn: []string{strings.Repeat("dep-", 30)}}
	for _, size := range []demoSize{{80, 24}, {120, 40}} {
		m := demoWithTasks(t, size.Width, size.Height, long)
		if size.Width < component.LongFormWidth {
			m, _ = raise(t, m, key(tea.KeyEnter, "enter"))
		}
		view := ansi.Strip(m.View().Content)
		requireFits(t, view, size.Width, size.Height)
		if !strings.Contains(view, "▶ long-id-long-id-long-id-… assigned") && size.Width >= component.LongFormWidth {
			t.Fatalf("long ID is not cut to its column:\n%s", view)
		}
		if !strings.Contains(view, "…") {
			t.Fatalf("long text is not cut with …:\n%s", view)
		}
	}
}

// AC4: ↓ ↑ j k g G move the selection. Each move shows the details of the
// new task from the top and closes an open file.
func TestWorkTasksMoveKeys(t *testing.T) {
	m := newTestDemoModel(t, "Work.fixture.json", demoSize{})
	for _, step := range []struct {
		press tea.KeyPressMsg
		want  string
	}{
		{tea.KeyPressMsg{Code: tea.KeyDown}, "2-stale-coordinator-startup"},
		{key('j', "j"), "1-cp5-end-to-end"},
		{key('j', "j"), "1-cp5-end-to-end"},
		{tea.KeyPressMsg{Code: tea.KeyUp}, "2-stale-coordinator-startup"},
		{key('g', "g"), "4-agents-equal-columns"},
		{key('G', "G"), "1-cp5-end-to-end"},
		{key('k', "k"), "2-stale-coordinator-startup"},
	} {
		m, _ = raise(t, m, step.press)
		if m.tasks.selectedID != step.want {
			t.Fatalf("%s selected %q, want %q", step.press.String(), m.tasks.selectedID, step.want)
		}
		requireView(t, m, "╭─ "+step.want)
	}
	m, _ = raise(t, m, key('d', "d"))
	requireView(t, m, "dependency graph · 2-stale-coordinator-startup", "3-tui-restart")
	m, _ = raise(t, m, key('j', "j"))
	rejectView(t, m, "dependency graph")
}

// AC1, AC4: b, v, and d open the brief, VERIFIED.md, and the dependencies of
// the selected task in its details. In demo mode the text comes from the
// fixture. esc closes the file first, and then leaves the details.
func TestWorkTasksFileKeys(t *testing.T) {
	m := newTestDemoModel(t, "Work.fixture.json", demoSize{})
	m, cmd := raise(t, m, key('b', "b"))
	deliverFileReads(m, cmd)
	requireView(t, m, "╭─ brief.md ─", "# Restart of the coordinator and the supervisor from the TUI")
	m, _ = raise(t, m, key(0, "esc"))
	requireView(t, m, "╭─ 3-tui-restart ─", "ACCEPTANCE")
	m, cmd = raise(t, m, key('v', "v"))
	deliverFileReads(m, cmd)
	requireView(t, m, "╭─ VERIFIED.md ─", "# VERIFIED 3-tui-restart")
	m, _ = raise(t, m, key('d', "d"))
	requireView(t, m, "dependency graph · 3-tui-restart", "3-tui-restart ← 2-stale-coordinator-startup · landed", "3-tui-restart → 4-agents-equal-columns · landed")

	// From 120 columns enter gives the focus to the details. In the details
	// esc closes the file, and a second esc gives the focus to the list.
	m, _ = raise(t, m, key(tea.KeyEnter, "enter"))
	if m.focusState() != "details" {
		t.Fatalf("enter gave focus %q, want details", m.focusState())
	}
	requireView(t, m, "[details] ↑↓ scroll · tab region · esc back to list · b brief · v verified · d deps · ? help")
	m, _ = raise(t, m, key(0, "esc"))
	rejectView(t, m, "dependency graph")
	if m.focusState() != "details" {
		t.Fatalf("esc with a file open gave focus %q, want details", m.focusState())
	}
	m, _ = raise(t, m, key(0, "esc"))
	if m.focusState() != "list" {
		t.Fatalf("second esc gave focus %q, want list", m.focusState())
	}
}

// AC4: decision 21. The Compact overlay takes its scroll keys and its action
// keys. A file key keeps the overlay. esc closes the file, then the overlay.
func TestWorkTasksOverlayKeys(t *testing.T) {
	m := newTestDemoModel(t, "Compact.fixture.json", demoSize{})
	requireView(t, m, "[details] ↑↓ scroll · esc back to list · b brief · v verified · d deps")
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyDown})
	if m.detailsOffset[tabWork] != 1 {
		t.Fatalf("↓ gave offset %d, want 1", m.detailsOffset[tabWork])
	}
	requireView(t, m, "│ goal      g1-tui-restart · wave B")
	rejectView(t, m, "│ status    landed")
	m, cmd := raise(t, m, key('b', "b"))
	deliverFileReads(m, cmd)
	requireView(t, m, "╭─ brief.md · esc back")
	m, _ = raise(t, m, key(0, "esc"))
	requireView(t, m, "╭─ 3-tui-restart · esc back")
	m, _ = raise(t, m, key(0, "esc"))
	requireView(t, m, "[list] ↑↓ move · enter details · [ ] view · ?")
	rejectView(t, m, "esc back")
}
