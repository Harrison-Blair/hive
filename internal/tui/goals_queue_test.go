package tui

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/charmbracelet/x/ansi"
)

func TestGoalsScreenRendersWaveProgressAndUnassignedTasks(t *testing.T) {
	home := goalsQueueFixture(t)
	m := newTestModel(t, home, fixedClock(t, time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)))
	m, cmd := raise(t, m, key('3', "3"))
	m = deliverAgentCommands(t, m, cmd)

	view := ansi.Strip(m.currentScreen().View(200, 40))
	for _, line := range []string{
		"demo · Goal one · done when: landed on dev",
		"dispatch ✓ ─▶ tasks 0/3 ▶ ─▶ ◎ landed on dev ─▶ PR open ─▶ merged to main ─▶ release published",
		"▶ 1-alpha  ✋ 2-beta  · 3-gamma",
		"unassigned",
		"demo: · 4-free  ▶ 82-goals-queue-screens",
	} {
		if !strings.Contains(view, line) {
			t.Fatalf("Goals view does not contain rendered line %q:\n%s", line, view)
		}
	}
}

func TestQueueScreenShowsTransitiveStopsAndProjectFilter(t *testing.T) {
	home := goalsQueueFixture(t)
	writeRecord(t, home, []string{"projects"}, "zeta", records.Project{ID: "zeta", Path: "/repo/zeta", Base: "dev"})
	writeRecord(t, home, []string{"tasks"}, "7-zed", records.Task{ID: "7-zed", Project: "zeta", Title: "Zed", Status: "created", DependsOn: []string{}})

	m := newTestModel(t, home, fixedClock(t, time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)))
	m, cmd := raise(t, m, key('4', "4"))
	m = deliverAgentCommands(t, m, cmd)
	view := ansi.Strip(m.currentScreen().View(200, 40))
	for _, line := range []string{
		"demo · Goal one · grant: Goal grant",
		"now  1-alpha  —  ▶ assigned",
		"1  2-beta  1-alpha lands, then you  ✋ you",
		"3  3-gamma  you  ┄ waits",
		"zeta",
		"7-zed",
	} {
		if !strings.Contains(view, line) {
			t.Fatalf("Queue view does not contain rendered line %q:\n%s", line, view)
		}
	}
	if !strings.Contains(m.currentScreen().KeyHints(), "order comes from the dependencies") {
		t.Fatalf("Queue footer omits the dependency-order hint: %q", m.currentScreen().KeyHints())
	}
	if !strings.Contains(view, "1-alpha lands") {
		t.Fatalf("Queue did not order the dependency before its descendants:\n%s", view)
	}

	m, _ = raise(t, m, key('s', "s"))
	view = ansi.Strip(m.currentScreen().View(200, 40))
	if strings.Contains(view, "now · 1-alpha") || strings.Contains(view, "7-zed") {
		t.Fatalf("stop-only view retained rows without a stop:\n%s", view)
	}
	for _, line := range []string{"2-beta", "3-gamma", "✋ you", "┄ waits"} {
		if !strings.Contains(view, line) {
			t.Fatalf("stop-only view omits %q:\n%s", line, view)
		}
	}

	m, _ = raise(t, m, key('s', "s"))
	m, _ = raise(t, m, key('p', "p"))
	m, _ = raise(t, m, key('p', "p"))
	view = ansi.Strip(m.currentScreen().View(200, 40))
	if !strings.Contains(view, "zeta") || strings.Contains(view, "2-beta") {
		t.Fatalf("project filter did not cycle to zeta:\n%s", view)
	}
}

func TestGoalsRefreshKeepsGoalSelectionAndQueueProjectFilter(t *testing.T) {
	home := goalsQueueFixture(t)
	writeRecord(t, home, []string{"goals"}, "g2-other", records.Goal{
		ID: "g2-other", Project: "demo", Title: "Goal two", Grant: "Grant two", Status: "active",
		Waves: []records.GoalWave{},
	})
	writeRecord(t, home, []string{"projects"}, "zeta", records.Project{ID: "zeta", Path: "/repo/zeta", Base: "dev"})
	writeRecord(t, home, []string{"tasks"}, "7-zed", records.Task{ID: "7-zed", Project: "zeta", Title: "Zed", Status: "created", DependsOn: []string{}})
	clock := fixedClock(t, time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC))
	m := newTestModel(t, home, clock)
	m, cmd := raise(t, m, key('3', "3"))
	m = deliverAgentCommands(t, m, cmd)
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyDown})
	writeRecord(t, home, []string{"goals"}, "g2-other", records.Goal{
		ID: "g2-other", Project: "demo", Title: "Goal two updated", Grant: "Grant two", Status: "active",
		Waves: []records.GoalWave{},
	})
	m = refreshGoalsQueue(t, m)
	view := ansi.Strip(m.currentScreen().View(200, 40))
	if !strings.Contains(view, "▶ demo · Goal two") {
		t.Fatalf("Goals refresh lost its selected goal:\n%s", view)
	}

	m, cmd = raise(t, m, key('4', "4"))
	m = deliverAgentCommands(t, m, cmd)
	m, _ = raise(t, m, key('p', "p"))
	writeRecord(t, home, []string{"tasks"}, "7-zed", records.Task{ID: "7-zed", Project: "zeta", Title: "Zed updated", Status: "created", DependsOn: []string{}})
	m = refreshGoalsQueue(t, m)
	view = ansi.Strip(m.currentScreen().View(200, 40))
	if !strings.Contains(view, "demo · Goal one") || strings.Contains(view, "zeta ·") {
		t.Fatalf("Queue refresh lost its project filter:\n%s", view)
	}
}

func TestGoalsSortsActiveBeforeDoneAndCountsOnlyPushedDeliveries(t *testing.T) {
	home := goalsQueueFixture(t)
	writeRecord(t, home, []string{"goals"}, "g0-done-goal-with-a-long-slug", records.Goal{
		ID: "g0-done-goal-with-a-long-slug", Project: "demo", Title: "Done goal", Status: "done", Waves: []records.GoalWave{},
	})
	writeRecord(t, home, []string{"tasks"}, "1-alpha", records.Task{
		ID: "1-alpha", Project: "demo", Title: "Alpha", Status: "assigned", DependsOn: []string{},
		Delivery: &records.Delivery{Commit: "abc123", Pushed: true},
	})
	writeRecord(t, home, []string{"tasks"}, "2-beta", records.Task{
		ID: "2-beta", Project: "demo", Title: "Beta", Status: "created", NeedsUser: "choose a color", DependsOn: []string{"1-alpha"},
		Delivery: &records.Delivery{Commit: "def456", Pushed: false},
	})
	m := newTestModel(t, home, fixedClock(t, time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)))
	m, cmd := raise(t, m, key('3', "3"))
	m = deliverAgentCommands(t, m, cmd)
	view := ansi.Strip(m.currentScreen().View(200, 40))
	active := strings.Index(view, "demo · Goal one")
	done := strings.Index(view, "demo · Done goal")
	if active < 0 || done <= active {
		t.Fatalf("active goal did not precede the dimmed done goal:\n%s", view)
	}
	if !strings.Contains(view, "tasks 1/3 ▶") {
		t.Fatalf("Goals counted an unpushed commit as landed:\n%s", view)
	}
}

func TestGoalsAndQueuePanelsFitStandardFrame(t *testing.T) {
	home := goalsQueueFixture(t)
	m := newTestModel(t, home, fixedClock(t, time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)))
	for _, keyText := range []string{"3", "4"} {
		m, cmd := raise(t, m, key(rune(keyText[0]), keyText))
		m = deliverAgentCommands(t, m, cmd)
		content := m.currentScreen().View(56, 14)
		if keyText == "3" && !strings.Contains(ansi.Strip(content), "Goal one") {
			t.Fatalf("Goals frame omitted its goal panel:\n%s", ansi.Strip(content))
		}
		if keyText == "4" && (!strings.Contains(ansi.Strip(content), "1-alpha") || !strings.Contains(m.currentScreen().KeyHints(), "order comes from the dependencies")) {
			t.Fatalf("Queue frame omitted its task rows or footer hint:\n%s", ansi.Strip(content))
		}
		for _, line := range strings.Split(ansi.Strip(content), "\n") {
			if got := lipgloss.Width(line); got > 52 {
				t.Fatalf("%s screen line width %d exceeds its 52-cell content area: %q", m.currentScreenName(), got, line)
			}
		}
	}
}

func refreshGoalsQueue(t *testing.T, model *Model) *Model {
	t.Helper()
	return runIdleTick(t, model)
}

func TestGoalsQueueProgramUsesStringReader(t *testing.T) {
	home := goalsQueueFixture(t)
	var out strings.Builder
	if err := RunHome(home, strings.NewReader("3\x03\x03"), &out); err != nil {
		t.Fatalf("RunHome: %v", err)
	}
	if !strings.Contains(out.String(), "▶ 3 Goals") || !strings.Contains(out.String(), "unassigned") {
		t.Fatalf("program output did not render the Goals screen:\n%s", out.String())
	}
}

func goalsQueueFixture(t *testing.T) string {
	t.Helper()
	home := newTestHome(t)
	writeRecord(t, home, []string{"projects"}, "demo", records.Project{ID: "demo", Path: "/repo/demo", Base: "dev"})
	writeRecord(t, home, []string{"goals"}, "g1-demo", records.Goal{
		ID: "g1-demo", Project: "demo", Title: "Goal one", Grant: "Goal grant", Status: "active",
		Waves: []records.GoalWave{{ID: "A", Title: "Wave alpha", DoneWhen: "dev", Tasks: []string{"1-alpha", "2-beta", "3-gamma"}, Status: "open"}},
	})
	writeRecord(t, home, []string{"tasks"}, "1-alpha", records.Task{
		ID: "1-alpha", Project: "demo", Title: "Alpha", Status: "assigned", DependsOn: []string{},
	})
	writeRecord(t, home, []string{"tasks"}, "2-beta", records.Task{
		ID: "2-beta", Project: "demo", Title: "Beta", Status: "created", NeedsUser: "choose a color", DependsOn: []string{"1-alpha"},
	})
	writeRecord(t, home, []string{"tasks"}, "3-gamma", records.Task{
		ID: "3-gamma", Project: "demo", Title: "Gamma", Status: "created", DependsOn: []string{"2-beta"},
	})
	writeRecord(t, home, []string{"tasks"}, "4-free", records.Task{
		ID: "4-free", Project: "demo", Title: "Free", Status: "created", DependsOn: []string{},
	})
	writeRecord(t, home, []string{"tasks"}, "82-goals-queue-screens", records.Task{
		ID: "82-goals-queue-screens", Project: "demo", Title: "Long ID", Status: "assigned", DependsOn: []string{},
	})
	return home
}
