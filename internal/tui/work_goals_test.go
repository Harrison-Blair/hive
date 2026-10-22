package tui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/Harrison-Blair/hive/internal/lib/cellgrid"
	"github.com/Harrison-Blair/hive/internal/lib/records"
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

// AC2: the Goals view matches its board cell for cell.
func TestWorkGoalsMatchesItsBoard(t *testing.T) {
	checkBoard(t, "WorkGoals")
}

// goalsModel opens the Goals view of a scratch home at width×height in
// normal mode (decision 38).
func goalsModel(t *testing.T, home string, width, height int) *Model {
	t.Helper()
	m := newTestModel(t, home, fixedClock(t, time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)))
	m, cmd := raise(t, m, key('3', "3"))
	m = deliverAgentCommands(t, m, cmd)
	m, _ = raise(t, m, tea.WindowSizeMsg{Width: width, Height: height})
	return m
}

// AC1, AC3: normal mode loads the goal records into the list and the
// details. Each state has a glyph or a word in plain text.
func TestGoalsDetailsShowWavesTasksAndStops(t *testing.T) {
	m := goalsModel(t, goalsQueueFixture(t), 120, 40)
	requireView(t, m,
		"goals · 1",
		"▶ g1-demo                 active    0/1",
		"─ g1-demo ─",
		"title     Goal one",
		"status    active",
		"grant     Goal grant",
		"manager   manager-demo",
		"A    Wave alpha                 dev        0/3     ○ open",
		"A   1-alpha                       ▶ assigned",
		"A   2-beta                        ✋ created",
		"A   3-gamma                       · created",
		"NEEDS YOU",
		"2-beta  choose a color",
		"o goal design   t open task in Tasks",
	)
}

// AC1: the list puts active goals first, then planned goals, then other
// goals, then done goals. Only a pushed delivery counts as landed, and a
// landed task shows its delivery.
func TestGoalsSortAndCountOnlyPushedDeliveries(t *testing.T) {
	home := goalsQueueFixture(t)
	for _, goal := range []records.Goal{
		{ID: "g0-done", Project: "demo", Title: "Done", Status: "done"},
		{ID: "g3-planned", Project: "demo", Title: "Planned", Status: "planned"},
		{ID: "g4-paused", Project: "demo", Title: "Paused", Status: "paused"},
	} {
		writeRecord(t, home, []string{"goals"}, goal.ID, goal)
	}
	writeRecord(t, home, []string{"tasks"}, "1-alpha", records.Task{
		ID: "1-alpha", Project: "demo", Title: "Alpha", Status: "verified", DependsOn: []string{},
		Delivery: &records.Delivery{Mode: "merge", Commit: "abc1234567890", Pushed: true},
	})
	writeRecord(t, home, []string{"tasks"}, "2-beta", records.Task{
		ID: "2-beta", Project: "demo", Title: "Beta", Status: "verified", DependsOn: []string{},
		Delivery: &records.Delivery{Mode: "merge", Commit: "def456", Pushed: false},
	})
	m := goalsModel(t, home, 120, 40)
	view := ansi.Strip(m.View().Content)
	last := -1
	for _, id := range []string{"g1-demo", "g3-planned", "g4-paused", "g0-done"} {
		index := strings.Index(view, " "+id+" ")
		if index <= last {
			t.Fatalf("goal %s is not in the order active, planned, other, done:\n%s", id, view)
		}
		last = index
	}
	requireView(t, m, "dev        1/3", "1-alpha                       merge abc12345", "2-beta                        ▶ verified")
}

// AC2: the selected row has the marker ▶ in accent and bold, and its ID in
// accent and bold. Each other row keeps its ID in fg.
func TestGoalsSelectedRowIsAccentBold(t *testing.T) {
	m := newTestDemoModel(t, "WorkGoals.fixture.json", demoSize{})
	grid := renderGrid(t, m, 120, 40)
	for col, char := range []rune("▶ g1-tui-restart") {
		cell := grid.At(7, 2+col)
		if cell.Rune != char {
			t.Fatalf("row 7 col %d = %q; want %q", 2+col, cell.Rune, char)
		}
		if char != ' ' && (cell.Token != cellgrid.Accent || !cell.Bold) {
			t.Fatalf("row 7 col %d %q is %s bold %t; want accent bold", 2+col, cell.Rune, cell.Token, cell.Bold)
		}
	}
	if cell := grid.At(6, 4); cell.Token != cellgrid.FG || cell.Bold {
		t.Fatalf("unselected ID is %s bold %t; want fg", cell.Token, cell.Bold)
	}
}

// AC2 (decision 46): at each size the view fits the pane, no border
// breaks, and the selected goal is visible. At 79 columns the view is the
// too-small notice. Below 120 columns enter opens the details overlay.
func TestGoalsFitEachSize(t *testing.T) {
	long := strings.Repeat("very-long-goal-slug-", 6)
	cases := map[string]func(t *testing.T) string{
		"fixture": func(t *testing.T) string { return goalsQueueFixture(t) },
		"empty":   func(t *testing.T) string { return newTestHome(t) },
		"long": func(t *testing.T) string {
			home := goalsQueueFixture(t)
			for index := range 50 {
				id := fmt.Sprintf("g%d-%s", index+2, long)
				writeRecord(t, home, []string{"goals"}, fmt.Sprintf("g%d-x", index+2), records.Goal{
					ID: id, Project: "demo", Title: strings.Repeat("title ", 40), Grant: strings.Repeat("grant ", 40), Status: "active",
					Waves: []records.GoalWave{{ID: "A", Title: strings.Repeat("wave ", 20), DoneWhen: "dev", Tasks: []string{"1-alpha"}, Status: "open"}},
				})
			}
			return home
		},
	}
	sizes := []demoSize{{79, 24}, {80, 24}, {80, 25}, {119, 40}, {120, 40}, {120, 24}, {120, 25}}
	for name, makeHome := range cases {
		for _, size := range sizes {
			t.Run(fmt.Sprintf("%s/%dx%d", name, size.Width, size.Height), func(t *testing.T) {
				m := goalsModel(t, makeHome(t), size.Width, size.Height)
				if name == "long" {
					m, _ = raise(t, m, key('G', "G"))
				}
				selected := m.goalsQueue.selectedGoal
				for _, msgs := range [][]tea.Msg{nil, {key(0, "enter")}} {
					for _, msg := range msgs {
						m, _ = raise(t, m, msg)
					}
					view := ansi.Strip(m.View().Content)
					if size.Width < minimumWidth {
						if !strings.Contains(view, "terminal too small") {
							t.Fatalf("view at %d columns is not the too-small notice:\n%s", size.Width, view)
						}
						return
					}
					lines := strings.Split(view, "\n")
					if len(lines) != size.Height {
						t.Fatalf("view has %d lines, want %d", len(lines), size.Height)
					}
					for index, line := range lines {
						if got := ansi.StringWidth(line); got != size.Width {
							t.Fatalf("line %d has width %d, want %d: %q", index+1, got, size.Width, line)
						}
					}
					requireWholeBorders(t, view)
					if selected == "" {
						continue
					}
					shown := cutCell(selected, goalIDWidth-1)
					if m.detailsFocused() && size.Width < 120 {
						if !strings.Contains(view, "─ "+cutCell(selected, size.Width-6)) {
							t.Fatalf("overlay does not show the selected goal %s:\n%s", selected, view)
						}
					} else if !strings.Contains(view, "▶ "+shown) {
						t.Fatalf("selected goal %s is not visible:\n%s", selected, view)
					}
				}
			})
		}
	}
}

// AC4 (decisions 21 to 23): the move keys select a goal, and the footer
// shows the key hints of each focus state.
func TestGoalsKeysAndHints(t *testing.T) {
	home := goalsQueueFixture(t)
	for _, id := range []string{"g2-two", "g3-three"} {
		writeRecord(t, home, []string{"goals"}, id, records.Goal{ID: id, Project: "demo", Title: id, Status: "active"})
	}
	m := goalsModel(t, home, 120, 40)
	for _, step := range []struct {
		press tea.KeyPressMsg
		want  string
	}{
		{tea.KeyPressMsg{Code: tea.KeyDown}, "g2-two"},
		{key('j', "j"), "g3-three"},
		{key('j', "j"), "g3-three"},
		{tea.KeyPressMsg{Code: tea.KeyUp}, "g2-two"},
		{key('k', "k"), "g1-demo"},
		{key('G', "G"), "g3-three"},
		{key('g', "g"), "g1-demo"},
	} {
		m, _ = raise(t, m, step.press)
		if m.goalsQueue.selectedGoal != step.want {
			t.Fatalf("key %q selected %q; want %q", step.press.String(), m.goalsQueue.selectedGoal, step.want)
		}
		requireView(t, m, "─ "+step.want+" ─")
	}
	requireView(t, m, "[list] ↑↓ move · tab region · [ ] view · enter open · o design · t task · / filter · ? help")
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyTab})
	requireView(t, m, "[details] ↑↓ scroll · tab region · esc back to list · ? help")
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyTab})
	m, _ = raise(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	requireView(t, m, "[list] ↑↓ move · enter details · [ ] view · / filter · ?")
	m, _ = raise(t, m, key(0, "enter"))
	requireView(t, m, "[details] ↑↓ scroll · esc back to list", "─ g1-demo · esc back ─")
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyEscape})
	requireView(t, m, "[list]")
}

// AC4: o on a goal with no design shows the notice in the footer at each
// width, and t opens the first task of the current wave in Tasks.
func TestGoalsDesignNoticeAndTaskKey(t *testing.T) {
	m := goalsModel(t, goalsQueueFixture(t), 80, 24)
	m, _ = raise(t, m, key('o', "o"))
	if footer := footerRow(m); !strings.Contains(footer, "[list] no design for this wave") {
		t.Fatalf("footer = %q; want the design notice", footer)
	}
	m, _ = raise(t, m, key('j', "j"))
	rejectView(t, m, "no design for this wave")
	m, cmd := raise(t, m, key('t', "t"))
	m = deliverAgentCommands(t, m, cmd)
	if m.screenIndex != 4 || m.tasks.selectedID != "1-alpha" {
		t.Fatalf("t gave screen %d task %q; want Tasks and 1-alpha", m.screenIndex, m.tasks.selectedID)
	}
}

// footerRow gives the plain text of the last row of the view.
func footerRow(m *Model) string {
	lines := strings.Split(ansi.Strip(m.View().Content), "\n")
	return lines[len(lines)-1]
}

// AC3: refresh in normal mode reloads the goal records and keeps the
// selection.
func TestGoalsRefreshReloadsRecords(t *testing.T) {
	home := goalsQueueFixture(t)
	m := goalsModel(t, home, 120, 40)
	writeRecord(t, home, []string{"goals"}, "g1-demo", records.Goal{ID: "g1-demo", Project: "demo", Title: "Goal renamed", Grant: "Goal grant", Status: "done"})
	writeRecord(t, home, []string{"goals"}, "g0-new", records.Goal{ID: "g0-new", Project: "demo", Title: "New", Status: "active"})
	m = refreshGoalsQueue(t, m)
	requireView(t, m, "goals · 2", "▶ g1-demo                 done      0/0", "title     Goal renamed")
}

// AC3: the program on a string reader draws the Goals view in normal mode.
func TestGoalsProgramUsesStringReader(t *testing.T) {
	home := goalsQueueFixture(t)
	var out strings.Builder
	if err := RunHome(home, strings.NewReader("3\x03\x03"), &out); err != nil {
		t.Fatalf("RunHome: %v", err)
	}
	if !strings.Contains(out.String(), "[Goals]") || !strings.Contains(out.String(), "goals · ") {
		t.Fatalf("program output did not render the Goals view:\n%s", out.String())
	}
}
