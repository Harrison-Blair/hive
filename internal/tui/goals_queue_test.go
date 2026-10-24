package tui

import (
	"testing"
	"time"

	"github.com/Harrison-Blair/hive/internal/lib/records"
)

func TestQueueHeaderShowsRunStopAndWaitCounts(t *testing.T) {
	home := goalsQueueFixture(t)
	m := newTestModel(t, home, fixedClock(t, time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)))
	m = deliverAgentCommands(t, m, m.startHeaderLoad())
	m, cmd := raise(t, m, ctrl('t'))
	m = deliverAgentCommands(t, m, cmd)

	// Decision 2: the header of each screen shows the agent states and
	// needs you. The header state keeps the Queue counts.
	if m.header.queueRun != 3 || m.header.queueStops != 1 || m.header.queueWaits != 1 {
		t.Fatalf("Queue counts = runs %d, stops %d, waits %d; want 3, 1, 1", m.header.queueRun, m.header.queueStops, m.header.queueWaits)
	}
	rejectView(t, m, "runs without you")
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
