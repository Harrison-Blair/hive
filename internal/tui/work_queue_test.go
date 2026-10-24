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

// AC2: the demo view of the WorkQueue fixture matches its golden grid cell
// for cell.
func TestWorkQueueMatchesItsBoard(t *testing.T) {
	if skipRetired(t, "WorkQueue") {
		return
	}
	want := readBoardGrid(t, "WorkQueue")
	m := newTestDemoModel(t, "WorkQueue.fixture.json", demoSize{})
	got := renderGrid(t, m, want.Width, want.Height)
	diffs, err := cellgrid.Compare(want, got, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(diffs) != 0 {
		t.Fatalf("%d cells differ from WorkQueue.grid:\n%s", len(diffs), cellgrid.FormatDiffs(diffs))
	}
}

// queueModel opens Work › Queue at width×height on the goalsQueueFixture
// home, in normal mode.
func queueModel(t *testing.T, home string, width, height int) *Model {
	t.Helper()
	m := newTestModel(t, home, fixedClock(t, time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)))
	m, _ = raise(t, m, tea.WindowSizeMsg{Width: width, Height: height})
	m, cmd := openWorkView(t, m, 1)
	return deliverAgentCommands(t, m, cmd)
}

// AC1, AC3: the record loaders give the grant groups, the dependency tree,
// and the stops of the board.
func TestQueueShowsGrantGroupsTreesAndStops(t *testing.T) {
	home := goalsQueueFixture(t)
	writeRecord(t, home, []string{"agents"}, "w82", records.Agent{Key: "w82", Name: "w82", Role: "worker", Status: records.AgentUnresolved})
	writeRecord(t, home, []string{"tasks"}, "82-goals-queue-screens", records.Task{
		ID: "82-goals-queue-screens", Project: "demo", Title: "Long ID", Status: "assigned", ActingAgent: "w82", DependsOn: []string{},
	})
	m := queueModel(t, home, 120, 40)
	requireView(t, m,
		"queue · demo · 5 tasks · 2 stops",
		"│ GRANT  Goal grant ",
		"│ ▶ 1-alpha                             assigned   │",
		"│   └▶ 2-beta                           created    │",
		"│      ! stop: choose a color ",
		"│     └▶ 3-gamma                        created    │",
		"│ NO GRANT ",
		"│   4-free                              created    │",
		"│   82-goals-queue-screens              assigned   │",
		"│   ! stop: worker unresolved ",
	)
	view := ansi.Strip(m.View().Content)
	if strings.Index(view, "NO GRANT") < strings.Index(view, "3-gamma") {
		t.Fatalf("NO GRANT is not after the grant group:\n%s", view)
	}
}

// A landed task is in the list with the word landed in the done token.
func TestQueueShowsLandedTasks(t *testing.T) {
	home := goalsQueueFixture(t)
	writeRecord(t, home, []string{"tasks"}, "4-free", records.Task{
		ID: "4-free", Project: "demo", Status: "verified", DependsOn: []string{},
		Delivery: &records.Delivery{Grant: "Delivery grant", Commit: "abc", Pushed: true},
	})
	m := queueModel(t, home, 120, 40)
	requireView(t, m, "│ GRANT  Delivery grant ", "│ ▶ 4-free                              landed     │")
	if got := statusToken(t, m, "4-free"); got != cellgrid.Done {
		t.Fatalf("landed token = %v, want done", got)
	}
}

// statusToken gives the token of the status cell of the row of task id.
func statusToken(t *testing.T, m *Model, id string) cellgrid.Token {
	t.Helper()
	grid := renderGrid(t, m, m.width, m.height)
	for row := range grid.Height {
		line := ""
		for col := range grid.Width {
			line += string(grid.At(row, col).Rune)
		}
		if strings.Contains(line, "│ ▶ "+id+" ") || strings.Contains(line, "│   "+id+" ") {
			return grid.At(row, 40).Token
		}
	}
	t.Fatalf("no row for %s", id)
	return 0
}

// AC1: the details of the selected task, with the STOP section of a stop.
func TestQueueDetailsShowTheSelectedTask(t *testing.T) {
	m := queueModel(t, goalsQueueFixture(t), 120, 40)
	requireView(t, m, "╮╭─ 1-alpha ─", "title     Alpha", "goal      g1-demo · wave A", "grant     Goal grant", "worker    none", "depends   none", "t open task in Tasks  s show stops only")
	rejectView(t, m, "STOP")
	m, _ = raise(t, m, key('j', "j"))
	requireView(t, m, "╮╭─ 2-beta ─", "depends   1-alpha", "STOP", "choose a color", "The task cannot move until the user decides.", "NEXT STEP", "Ask the coordinator in 2 Chat.")
}

// AC1, AC4: s shows stops only, and s again shows each task.
func TestQueueStopsOnly(t *testing.T) {
	m := queueModel(t, goalsQueueFixture(t), 120, 40)
	m, _ = raise(t, m, key('s', "s"))
	requireView(t, m, "│ ▶ 2-beta ", "! stop: choose a color", "s show all tasks")
	rejectView(t, m, "│   1-alpha", "3-gamma", "4-free", "└▶")
	m, _ = raise(t, m, key('s', "s"))
	requireView(t, m, "1-alpha", "│ ▶ └▶ 2-beta ", "4-free")
}

// With no stop, stops only shows "no stops", and no task shows "no tasks".
func TestQueueEmptyData(t *testing.T) {
	m := queueModel(t, newTestHome(t), 120, 40)
	requireView(t, m, "queue · 0 tasks · 0 stops", "│ no tasks ", "╭─ details ─", "no task selected")
	m, _ = raise(t, m, key('s', "s"))
	requireView(t, m, "│ no stops ")
}

// Each project has its groups. A project with no tasks shows "no tasks",
// and the title names no project.
func TestQueueShowsEachProject(t *testing.T) {
	home := goalsQueueFixture(t)
	writeRecord(t, home, []string{"projects"}, "empty", records.Project{ID: "empty", Path: "/repo/empty", Base: "dev"})
	writeRecord(t, home, []string{"projects"}, "zeta", records.Project{ID: "zeta", Path: "/repo/zeta", Base: "dev"})
	writeRecord(t, home, []string{"tasks"}, "7-zed", records.Task{ID: "7-zed", Project: "zeta", Title: "Zed", Status: "created", DependsOn: []string{}})
	m := queueModel(t, home, 120, 40)
	requireView(t, m, "queue · 6 tasks · 1 stop", "demo · GRANT  Goal grant", "demo · NO GRANT", "empty · no tasks", "zeta · NO GRANT", "7-zed")
}

// AC4: j, k, g, and G move the selection. t opens the selected task in
// Tasks. p does nothing.
func TestQueueKeys(t *testing.T) {
	m := queueModel(t, goalsQueueFixture(t), 120, 40)
	m, _ = raise(t, m, key('G', "G"))
	requireView(t, m, "▶ 82-goals-queue-screens")
	m, _ = raise(t, m, key('k', "k"))
	requireView(t, m, "▶ 4-free")
	m, _ = raise(t, m, key('g', "g"))
	requireView(t, m, "▶ 1-alpha")
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyDown})
	before := m.View().Content
	m, _ = raise(t, m, key('p', "p"))
	if m.View().Content != before {
		t.Fatal("p changed the Queue view")
	}
	m, cmd := raise(t, m, key('t', "t"))
	m = deliverAgentCommands(t, m, cmd)
	if m.viewName() != "Tasks" || m.tasks.selectedID != "2-beta" {
		t.Fatalf("t gave view %s task %q, want Tasks 2-beta", m.viewName(), m.tasks.selectedID)
	}
	requireView(t, m, "▶ 2-beta")
}

// AC4: from 120 columns enter gives the focus to the details, and the
// footer shows the hints of each focus state. Below 120 columns enter opens
// the overlay, and t still works in it.
func TestQueueFocusStates(t *testing.T) {
	m := queueModel(t, goalsQueueFixture(t), 120, 40)
	requireKeys(t, m, "list", "↑↓ move · ^] box · [ ] view · enter open · s stops only · t task · ? help")
	m, _ = raise(t, m, key(tea.KeyEnter, "enter"))
	requireKeys(t, m, "details", "↑↓ scroll · ^] box · esc back to list · ? help")
	m, _ = raise(t, m, key(tea.KeyEscape, "esc"))
	requireKeys(t, m, "list", "")

	m = queueModel(t, goalsQueueFixture(t), 80, 24)
	m, _ = raise(t, m, key(tea.KeyEnter, "enter"))
	requireView(t, m, "╭─ 1-alpha · esc back ─")
	requireKeys(t, m, "details", "↑↓ scroll · esc back to list")
	m, cmd := raise(t, m, key('t', "t"))
	m = deliverAgentCommands(t, m, cmd)
	if m.viewName() != "Tasks" || m.tasks.selectedID != "1-alpha" {
		t.Fatalf("t in the overlay gave view %s task %q", m.viewName(), m.tasks.selectedID)
	}
}

// AC3: refresh reads the records again and keeps the selection.
func TestQueueRefreshKeepsSelection(t *testing.T) {
	home := goalsQueueFixture(t)
	m := queueModel(t, home, 120, 40)
	m, _ = raise(t, m, key('j', "j"))
	writeRecord(t, home, []string{"tasks"}, "5-new", records.Task{ID: "5-new", Project: "demo", Title: "New", Status: "created", DependsOn: []string{}})
	m = refreshGoalsQueue(t, m)
	requireView(t, m, "▶ 2-beta", "5-new", "queue · demo · 6 tasks")
}

// AC2 (decision 46): at each size the lines fit the pane, no border
// breaks, and the selected task stays visible, also with many tasks and a
// long task ID.
func TestQueueFitsEachSize(t *testing.T) {
	home := goalsQueueFixture(t)
	long := "99-" + strings.Repeat("very-long-task-id-", 4)
	writeRecord(t, home, []string{"tasks"}, long, records.Task{ID: long, Project: "demo", Title: strings.Repeat("long title ", 20), Status: "created", DependsOn: []string{}})
	for i := range 30 {
		id := fmt.Sprintf("%d-more", 100+i)
		writeRecord(t, home, []string{"tasks"}, id, records.Task{ID: id, Project: "demo", Status: "created", DependsOn: []string{}})
	}
	for _, size := range []demoSize{{79, 24}, {80, 24}, {80, 25}, {119, 40}, {119, 41}, {120, 40}, {120, 41}} {
		t.Run(fmt.Sprintf("%dx%d", size.Width, size.Height), func(t *testing.T) {
			m := queueModel(t, home, size.Width, size.Height)
			m, _ = raise(t, m, key('G', "G"))
			view := ansi.Strip(m.View().Content)
			if size.Width < minimumWidth {
				if !strings.Contains(view, "terminal too small") {
					t.Fatalf("no too-small notice:\n%s", view)
				}
				return
			}
			for index, line := range strings.Split(view, "\n") {
				if got := ansi.StringWidth(line); got != size.Width {
					t.Fatalf("line %d has width %d, want %d: %q", index+1, got, size.Width, line)
				}
			}
			requireWholeBorders(t, view)
			if !strings.Contains(view, "▶ 129-more") {
				t.Fatalf("the selected last task is not visible:\n%s", view)
			}
			if !strings.Contains(view, "┃") {
				t.Fatalf("the list that scrolls has no scroll mark:\n%s", view)
			}
			m, _ = raise(t, m, key('g', "g"))
			if view := ansi.Strip(m.View().Content); !strings.Contains(view, "▶ 1-alpha") || !strings.Contains(view, " 99-very-long-task-id-very-lon… ") {
				t.Fatalf("the top of the list or the cut long task ID is not visible:\n%s", view)
			}
		})
	}
}
