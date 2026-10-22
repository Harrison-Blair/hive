package tui

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/charmbracelet/x/ansi"
)

func TestGoalsOpenTasksForSelectedWave(t *testing.T) {
	home := goalsQueueFixture(t)
	writeRecord(t, home, []string{"goals"}, "g1-demo", records.Goal{
		ID: "g1-demo", Project: "demo", Title: "Goal one", Grant: "Goal grant", Status: "active",
		Waves: []records.GoalWave{
			{ID: "done", Title: "Done wave", DoneWhen: "local", Tasks: []string{"4-free"}, Status: "done"},
			{ID: "A", Title: "Wave alpha", DoneWhen: "dev", Tasks: []string{"1-alpha", "2-beta", "3-gamma"}, Status: "open"},
		},
	})
	m := newTestModel(t, home, fixedClock(t, time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)))
	m, cmd := raise(t, m, key('3', "3"))
	m = deliverAgentCommands(t, m, cmd)
	m, cmd = raise(t, m, key('t', "t"))
	m = deliverAgentCommands(t, m, cmd)

	if m.screenIndex != 4 || m.tasks.tab != tasksAll || m.tasks.selectedID != "1-alpha" {
		t.Fatalf("Goals did not open the selected wave tasks at the first task: screen=%d tab=%d task=%q", m.screenIndex, m.tasks.tab, m.tasks.selectedID)
	}
	view := ansi.Strip(m.currentScreen().View(160, 30))
	if !strings.Contains(view, "▶ 1-alpha") {
		t.Fatalf("Tasks screen did not show the first wave task selected:\n%s", view)
	}
}

func TestGoalsOpenTasksResetsProjectFilter(t *testing.T) {
	home := goalsQueueFixture(t)
	writeRecord(t, home, []string{"projects"}, "other", records.Project{ID: "other", Path: "/repo/other", Base: "dev"})
	writeRecord(t, home, []string{"tasks"}, "8-other", records.Task{ID: "8-other", Project: "other", Title: "Other task", Status: "created", DependsOn: []string{}})
	m := newTestModel(t, home, fixedClock(t, time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)))
	m, cmd := openWorkView(t, m, 2)
	m = deliverAgentCommands(t, m, cmd)
	m, _ = raise(t, m, key('p', "p"))
	m, _ = raise(t, m, key('p', "p"))
	if got := m.tasks.projectFilter(); got != "other" {
		t.Fatalf("Tasks filter setup failed: got %q, want other", got)
	}

	// Decision 22: [ selects the previous view, so [ [ gives Goals.
	for range 2 {
		m, cmd = raise(t, m, key('[', "["))
		m = deliverAgentCommands(t, m, cmd)
	}
	m, cmd = raise(t, m, key('t', "t"))
	m = deliverAgentCommands(t, m, cmd)

	if m.tasks.projectIndex != 0 || m.tasks.selectedID != "1-alpha" {
		t.Fatalf("Goals did not reset the Tasks project filter and select the first wave task: project=%q task=%q", m.tasks.projectFilter(), m.tasks.selectedID)
	}
	view := ansi.Strip(m.currentScreen().View(160, 30))
	if !strings.Contains(view, "project: all") || !strings.Contains(view, "▶ 1-alpha") {
		t.Fatalf("Tasks screen did not show the selected first wave task with all projects:\n%s", view)
	}
}

func TestGoalsOpenWaveDesign(t *testing.T) {
	tests := []struct {
		name       string
		features   []string
		files      map[string]string
		wantDesign string
	}{
		{
			name:       "one feature",
			features:   []string{"feature-one"},
			files:      map[string]string{"feature-one": "# Feature one design\n"},
			wantDesign: "# Feature one design",
		},
		{
			name:       "first feature wins",
			features:   []string{"feature-one", "feature-two"},
			files:      map[string]string{"feature-one": "# First feature design\n", "feature-two": "# Second feature design\n"},
			wantDesign: "# First feature design",
		},
		{name: "no feature"},
		{name: "missing file", features: []string{"missing-feature"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			home := goalsQueueFixture(t)
			for index, id := range []string{"1-alpha", "2-beta", "3-gamma"} {
				feature := ""
				if index < len(tt.features) {
					feature = tt.features[index]
				}
				task := records.Task{ID: id, Project: "demo", Title: id, Status: "created", Feature: feature}
				writeRecord(t, home, []string{"tasks"}, id, task)
			}
			for feature, content := range tt.files {
				writeText(t, filepath.Join(home, ".hive", "spec", feature, "README.md"), content)
			}
			m := newTestModel(t, home, fixedClock(t, time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)))
			m, cmd := raise(t, m, key('3', "3"))
			m = deliverAgentCommands(t, m, cmd)
			m, cmd = raise(t, m, key('o', "o"))
			m = deliverAgentCommands(t, m, cmd)

			if tt.wantDesign == "" {
				if m.screenIndex != 2 || m.tasks.fileOpen {
					t.Fatalf("missing wave design opened a file: screen=%d open=%t", m.screenIndex, m.tasks.fileOpen)
				}
				if !strings.Contains(ansi.Strip(m.View().Content), "no design for this wave") {
					t.Fatalf("Goals did not explain the missing wave design:\n%s", ansi.Strip(m.View().Content))
				}
				return
			}
			if m.screenIndex != 4 || !m.tasks.fileOpen || m.tasks.fileTitle != "README.md" || !strings.Contains(m.tasks.fileContent, tt.wantDesign) {
				t.Fatalf("Goals did not open the wave feature design: screen=%d open=%t title=%q content=%q", m.screenIndex, m.tasks.fileOpen, m.tasks.fileTitle, m.tasks.fileContent)
			}
			view := ansi.Strip(m.currentScreen().View(160, 30))
			if !strings.Contains(view, tt.wantDesign) {
				t.Fatalf("shared file viewer omitted the wave design %q:\n%s", tt.wantDesign, view)
			}
		})
	}
}

func TestQueueHeaderShowsRunStopAndWaitCounts(t *testing.T) {
	home := goalsQueueFixture(t)
	m := newTestModel(t, home, fixedClock(t, time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)))
	m = deliverAgentCommands(t, m, m.startHeaderLoad())
	m, cmd := openWorkView(t, m, 1)
	m = deliverAgentCommands(t, m, cmd)

	// Decision 2: the header of each screen shows the agent states and
	// needs you. The header state keeps the Queue counts for the Queue view.
	if m.header.queueRun != 3 || m.header.queueStops != 1 || m.header.queueWaits != 1 {
		t.Fatalf("Queue counts = runs %d, stops %d, waits %d; want 3, 1, 1", m.header.queueRun, m.header.queueStops, m.header.queueWaits)
	}
	rejectView(t, m, "runs without you")
}

func TestGoalsRefreshKeepsGoalSelection(t *testing.T) {
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
	if !strings.Contains(view, "▶ g2-other") {
		t.Fatalf("Goals refresh lost its selected goal:\n%s", view)
	}

}

func TestGoalsPanelsFitStandardFrame(t *testing.T) {
	home := goalsQueueFixture(t)
	m := newTestModel(t, home, fixedClock(t, time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)))
	// Decision 5: Goals is view 0 of the Work tab.
	for _, view := range []int{0} {
		m, cmd := openWorkView(t, m, view)
		m = deliverAgentCommands(t, m, cmd)
		content := m.currentScreen().View(56, 14)
		if view == 0 && !strings.Contains(ansi.Strip(content), "g1-demo") {
			t.Fatalf("Goals frame omitted its goal panel:\n%s", ansi.Strip(content))
		}
		// The Goals view draws its own panel, so it fills all 56 cells.
		limit := 52
		if view == 0 {
			limit = 56
		}
		for _, line := range strings.Split(ansi.Strip(content), "\n") {
			if got := lipgloss.Width(line); got > limit {
				t.Fatalf("%s screen line width %d exceeds its %d-cell content area: %q", m.currentScreenName(), got, limit, line)
			}
		}
	}
}

func refreshGoalsQueue(t *testing.T, model *Model) *Model {
	t.Helper()
	return runIdleTick(t, model)
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

// openWorkView selects the Work tab and its view. Decision 5 puts Goals,
// Queue, and Tasks in the views of the Work tab, and decision 22 gives ] to
// the next view. It gives the command of the last key.
func openWorkView(t *testing.T, m *Model, view int) (*Model, tea.Cmd) {
	t.Helper()
	m, cmd := raise(t, m, key('3', "3"))
	for range view {
		m = deliverAgentCommands(t, m, cmd)
		m, cmd = raise(t, m, key(']', "]"))
	}
	return m, cmd
}
