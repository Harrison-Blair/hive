package tui

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/charmbracelet/x/ansi"
)

func TestTasksScreenShowsEvidencePanelAndDefinitions(t *testing.T) {
	home := writeTasksFixture(t)
	clock := fixedClock(t, time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC))
	m := newTestModel(t, home, clock)
	m.screenIndex = 4
	cmd := m.startCurrentScreenLoad()
	m, _ = raise(t, m, cmd())
	for range 2 {
		// Decision 6: tab moves the region. The test gives tab to the current Tasks screen.
		m, _ = pressScreenKey(t, m, key(0, "tab"))
	}

	view := ansi.Strip(m.currentScreen().View(200, 50))
	for _, text := range []string{
		"task details",
		"┤ evidence ├",
		"79-tui-shell · TUI shell",
		"brief: tasks/79-tui-shell/brief.md",
		"V1 fail · V1 failed: missing shell load · 2026-10-20T10:00:00Z",
		"landed = delivery commit set and pushed · verified = verification passed",
	} {
		if !strings.Contains(view, text) {
			t.Errorf("Tasks view does not contain %q:\n%s", text, view)
		}
	}
}

func TestTasksScreenVOpensVerifiedEvidence(t *testing.T) {
	home := writeTasksFixture(t)
	clock := fixedClock(t, time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC))
	m := newTestModel(t, home, clock)
	m.screenIndex = 4
	cmd := m.startCurrentScreenLoad()
	m, _ = raise(t, m, cmd())
	for range 2 {
		// Decision 6: tab moves the region. The test gives tab to the current Tasks screen.
		m, _ = pressScreenKey(t, m, key(0, "tab"))
	}

	m, cmd = raise(t, m, key('v', "v"))
	if cmd == nil {
		t.Fatal("v did not request VERIFIED.md")
	}
	m = deliverAgentCommands(t, m, cmd)
	if view := ansi.Strip(m.currentScreen().View(120, 36)); !strings.Contains(view, "# VERIFIED") {
		t.Fatalf("v did not open VERIFIED.md:\n%s", view)
	}
}

func TestTasksScreenDOpensSelectedTaskDependencyGraph(t *testing.T) {
	home := writeTasksFixture(t)
	writeRecord(t, home, []string{"tasks"}, "83-downstream-task", records.Task{
		ID: "83-downstream-task", Project: "hive", Title: "Downstream task", Status: "assigned", DependsOn: []string{"79-tui-shell"},
	})
	clock := fixedClock(t, time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC))
	m := newTestModel(t, home, clock)
	m.screenIndex = 4
	cmd := m.startCurrentScreenLoad()
	m, _ = raise(t, m, cmd())
	for range 2 {
		// Decision 6: tab moves the region. The test gives tab to the current Tasks screen.
		m, _ = pressScreenKey(t, m, key(0, "tab"))
	}

	m, cmd = raise(t, m, key('d', "d"))
	if cmd == nil {
		t.Fatal("d did not log opening the dependency graph")
	}
	m = deliverAgentCommands(t, m, cmd)
	view := ansi.Strip(m.currentScreen().View(120, 36))
	for _, text := range []string{
		"dependency graph · 79-tui-shell",
		"79-tui-shell ← 81-agents-screen (needs)",
		"79-tui-shell → 83-downstream-task (blocks)",
	} {
		if !strings.Contains(view, text) {
			t.Errorf("dependency graph does not contain %q:\n%s", text, view)
		}
	}
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyEscape})
	if view := ansi.Strip(m.currentScreen().View(120, 36)); strings.Contains(view, "dependency graph") {
		t.Fatalf("escape did not close the dependency graph:\n%s", view)
	}
}
