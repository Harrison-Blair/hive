package tui

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/Harrison-Blair/hive/internal/lib/process"
	"github.com/Harrison-Blair/hive/internal/lib/records"
)

// statusDurationDay is the clock of the duration render tests.
var statusDurationDay = time.Date(2026, 10, 23, 14, 22, 0, 0, time.UTC)

// TestStatusDurationFormatsWholeSecondsAndTotalHours covers the literal
// second, minute, hour, day, and three-digit-hour boundaries of task 183.
func TestStatusDurationFormatsWholeSecondsAndTotalHours(t *testing.T) {
	at := statusDurationDay
	cases := []struct {
		name    string
		elapsed time.Duration
		want    string
	}{
		{"zero", 0, "00:00:00"},
		{"fifty-nine seconds", 59 * time.Second, "00:00:59"},
		{"one minute", time.Minute, "00:01:00"},
		{"one hour", time.Hour, "01:00:00"},
		{"one day", 24 * time.Hour, "24:00:00"},
		{"twenty-five hours", 25 * time.Hour, "25:00:00"},
		{"three digit hours", 100 * time.Hour, "100:00:00"},
		{"mixed", time.Hour + 2*time.Minute + 3*time.Second, "01:02:03"},
		{"last hour", 999*time.Hour + 59*time.Minute + 59*time.Second, "999:59:59"},
		{"sub-second is whole seconds", 5*time.Second + 999*time.Millisecond, "00:00:05"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := coordinatorStatusAge(at.Add(-tc.elapsed).Format(time.RFC3339Nano), at)
			if got != tc.want {
				t.Fatalf("duration = %q, want %q", got, tc.want)
			}
		})
	}
}

// TestStatusDurationShowsOnEachAffectedScreen checks the rendered text of
// every included duration field.
func TestStatusDurationShowsOnEachAffectedScreen(t *testing.T) {
	t.Run("Chat welcome and footer", func(t *testing.T) {
		m, _ := newFrameModel(t, 120, 40)
		m.header.coordinator = "working"
		m.header.coordinatorChangedAt = statusDurationDay.Add(-2 * time.Minute).Format(time.RFC3339)
		m.invalidateView()
		requireView(t, m, "● working 00:02:00")
		if got := strings.Count(ansi.Strip(m.View().Content), "● working 00:02:00"); got != 2 {
			t.Fatalf("the welcome and footer show the duration %d times, want 2", got)
		}
	})

	t.Run("Agents card and details", func(t *testing.T) {
		home, clock := statusDurationAgentsHome(t, statusDurationDay)
		name := "worker-with-a-very-long-name"
		m := openAgents(t, home, clock)
		requireView(t, m, "working · 00:02:00")
		m.graph.selected = name
		m.invalidateView()
		m, _ = raise(t, m, key(tea.KeyEnter, ""))
		requireView(t, m, "state     working · 00:02:00 · changed 14:20:00")
	})

	t.Run("Work card and task popup", func(t *testing.T) {
		m := statusDurationWorkModel(t)
		requireView(t, m, "worker-w… · 00:12:00")
		m = pressWork(t, m, key(tea.KeyEnter, ""))
		requireView(t, m, "worker     worker-with-a-very-long-name · working · 00:12:00")
	})

	t.Run("Health row", func(t *testing.T) {
		m := newGfDemoModel(t, "GfHealth")
		requireView(t, m, "working 00:02:05", "live 01:02:05", "idle 00:06:05")
	})
}

// TestStatusDurationClockAdvanceRefreshesUnchangedRecords drives the
// existing tick and Update path. The records stay the same and only the clock
// moves.
func TestStatusDurationClockAdvanceRefreshesUnchangedRecords(t *testing.T) {
	t.Run("Chat welcome and footer", func(t *testing.T) {
		base := statusDurationDay
		m := newGfDemoModel(t, "GfChatEmpty")
		requireView(t, m, "working 00:02:00")
		m.now = func() time.Time { return base.Add(7 * time.Second) }
		m, _ = raise(t, m, tickMsg{})
		if got := strings.Count(ansi.Strip(m.View().Content), "● working 00:02:07"); got != 2 {
			t.Fatalf("the welcome and footer did not refresh: %d matches, want 2", got)
		}
	})

	t.Run("Agents card and details", func(t *testing.T) {
		home, clock := statusDurationAgentsHome(t, statusDurationDay)
		name := "worker-with-a-very-long-name"
		m := openAgents(t, home, clock)
		requireView(t, m, "working · 00:02:00")
		m.graph.selected = name
		m.invalidateView()
		clock.set(statusDurationDay.Add(7 * time.Second))
		m, _ = raise(t, m, tickMsg{})
		requireView(t, m, "working · 00:02:07")
		m, _ = raise(t, m, key(tea.KeyEnter, ""))
		requireView(t, m, "state     working · 00:02:07 · changed 14:20:00")
		clock.set(statusDurationDay.Add(12 * time.Second))
		m, _ = raise(t, m, tickMsg{})
		requireView(t, m, "state     working · 00:02:12 · changed 14:20:00")
	})

	t.Run("Work card and task popup", func(t *testing.T) {
		m := statusDurationWorkModel(t)
		requireView(t, m, "worker-w… · 00:12:00")
		m.now = func() time.Time { return workTestNow.Add(7 * time.Second) }
		m, _ = raise(t, m, tickMsg{})
		requireView(t, m, "worker-w… · 00:12:07")
		m = pressWork(t, m, key(tea.KeyEnter, ""))
		requireView(t, m, "worker     worker-with-a-very-long-name · working · 00:12:07")
		m.now = func() time.Time { return workTestNow.Add(12 * time.Second) }
		m, _ = raise(t, m, tickMsg{})
		requireView(t, m, "worker     worker-with-a-very-long-name · working · 00:12:12")
	})

	t.Run("Health row", func(t *testing.T) {
		m := newGfDemoModel(t, "GfHealth")
		requireView(t, m, "working 00:02:05")
		m.now = func() time.Time {
			return time.Date(2026, 10, 23, 14, 22, 12, 0, time.UTC)
		}
		m, _ = raise(t, m, tickMsg{})
		requireView(t, m, "working 00:02:12", "live 01:02:12", "idle 00:06:12")
	})
}

// TestStatusDurationStateChangeResetsFromTheNewTimestamp follows a header
// load and the next tick after it.
func TestStatusDurationStateChangeResetsFromTheNewTimestamp(t *testing.T) {
	m, _ := newFrameModel(t, 120, 40)
	start := statusDurationDay
	m.now = func() time.Time { return start }
	m.header.coordinator = "working"
	m.header.coordinatorChangedAt = start.Format(time.RFC3339)
	m.invalidateView()
	requireView(t, m, "● working 00:00:00")

	m.now = func() time.Time { return start.Add(5 * time.Second) }
	m, _ = raise(t, m, tickMsg{})
	requireView(t, m, "● working 00:00:05")

	changed := m.header
	changed.coordinatorChangedAt = start.Add(5 * time.Second).Format(time.RFC3339)
	m, _ = raise(t, m, headerLoadMsg{state: changed})
	requireView(t, m, "● working 00:00:00")

	m.now = func() time.Time { return start.Add(8 * time.Second) }
	m, _ = raise(t, m, tickMsg{})
	requireView(t, m, "● working 00:00:03")
}

// TestStatusDurationMissingInvalidAndFuture checks each existing unknown
// and future behavior.
func TestStatusDurationMissingInvalidAndFuture(t *testing.T) {
	t.Run("coordinator footer", func(t *testing.T) {
		for _, tc := range []struct {
			name, changed, want string
		}{
			{"missing", "", " ● working · "},
			{"invalid", "not-a-time", " ● working · "},
			{"future", statusDurationDay.Add(time.Hour).Format(time.RFC3339), " ● working 00:00:00 · "},
		} {
			t.Run(tc.name, func(t *testing.T) {
				m, _ := newFrameModel(t, 120, 40)
				m.header.coordinator = "working"
				m.header.coordinatorChangedAt = tc.changed
				m.invalidateView()
				if got := viewRow(m, 39); !strings.HasPrefix(got, tc.want) {
					t.Fatalf("status row = %q, want prefix %q", got, tc.want)
				}
			})
		}
	})

	t.Run("Agents card", func(t *testing.T) {
		for _, tc := range []struct {
			name, changed, want, reject string
		}{
			{"missing", "", "pi · working", "00:00:00"},
			{"invalid", "not-a-time", "pi · working", "00:00:00"},
			{"future", statusDurationDay.Add(time.Hour).Format(time.RFC3339), "working · 00:00:00", ""},
		} {
			t.Run(tc.name, func(t *testing.T) {
				m := newGfDemoModel(t, "GfAgents")
				for index := range m.agents.data.nodes {
					if m.agents.data.nodes[index].record.Key == "t5-worker-1" {
						m.agents.data.nodes[index].record.StatusChangedAt = tc.changed
					}
				}
				m.invalidateView()
				requireView(t, m, tc.want)
				if tc.reject != "" {
					rejectView(t, m, tc.reject)
				}
			})
		}
	})

	t.Run("Health row", func(t *testing.T) {
		for _, tc := range []struct {
			name, changed, want string
		}{
			{"missing", "", "working"},
			{"invalid", "not-a-time", "working"},
			{"future", statusDurationDay.Add(time.Hour).Format(time.RFC3339), "working 00:00:00"},
		} {
			t.Run(tc.name, func(t *testing.T) {
				record := records.Agent{Status: records.AgentLive, AgentStatus: "working", StatusChangedAt: tc.changed}
				word, _, _ := healthAgentState(record, statusDurationDay, 0)
				if word != tc.want {
					t.Fatalf("health state = %q, want %q", word, tc.want)
				}
			})
		}
	})

	t.Run("Work card", func(t *testing.T) {
		for _, tc := range []struct {
			name, changed, want, reject string
		}{
			{"missing", "", "worker-with-a-very-long-name", "00:00:00"},
			{"invalid", "not-a-time", "worker-with-a-very-long-name", "00:00:00"},
			{"future", workTestNow.Add(time.Hour).Format(time.RFC3339), "worker-w… · 00:00:00", ""},
		} {
			t.Run(tc.name, func(t *testing.T) {
				m := statusDurationWorkModel(t)
				agent := m.tasks.data.agents["worker-with-a-very-long-name"]
				agent.StatusChangedAt = tc.changed
				m.tasks.data.agents["worker-with-a-very-long-name"] = agent
				item, ok := m.tasks.data.item("5-e")
				if !ok {
					t.Fatal("task 5-e missing")
				}
				got, _ := m.tasks.cardLine(item, workTestNow)
				if !strings.Contains(got, tc.want) {
					t.Fatalf("card line = %q, want %q", got, tc.want)
				}
				if tc.reject != "" {
					if strings.Contains(got, tc.reject) {
						t.Fatalf("card line = %q, must not contain %q", got, tc.reject)
					}
				}
			})
		}
	})
}

// TestStatusDurationVisibilityStays checks that only the existing states
// show a duration.
func TestStatusDurationVisibilityStays(t *testing.T) {
	for _, tc := range []struct {
		state, prefix string
	}{
		{"idle", " ○ idle · "},
		{"blocked", " ▲ blocked · "},
	} {
		t.Run("coordinator "+tc.state, func(t *testing.T) {
			m, _ := newFrameModel(t, 120, 40)
			m.header.coordinator = tc.state
			m.header.coordinatorChangedAt = statusDurationDay.Add(-2 * time.Minute).Format(time.RFC3339)
			m.invalidateView()
			requireView(t, m, tc.prefix)
			rejectView(t, m, tc.state+" 00:02:00")
		})
	}

	t.Run("health non-live states", func(t *testing.T) {
		changed := statusDurationDay.Add(-2 * time.Minute).Format(time.RFC3339)
		for _, tc := range []struct {
			record records.Agent
			want   string
		}{
			{records.Agent{Status: records.AgentStarting, StatusChangedAt: changed}, "starting"},
			{records.Agent{Status: records.AgentStopped, StatusChangedAt: changed}, "stopped"},
			{records.Agent{Status: records.AgentExited, StatusChangedAt: changed}, "stopped"},
		} {
			word, _, _ := healthAgentState(tc.record, statusDurationDay, 0)
			if word != tc.want {
				t.Fatalf("health state = %q, want %q", word, tc.want)
			}
		}
	})
}

// TestStatusDurationCompleteAtNormalAndCompactWidths keeps the full time
// digits at each supported width, and elides other text.
func TestStatusDurationCompleteAtNormalAndCompactWidths(t *testing.T) {
	for _, size := range []struct{ width, height int }{{120, 40}, {119, 40}, {80, 24}} {
		t.Run(fmt.Sprintf("%dx%d", size.width, size.height), func(t *testing.T) {
			t.Run("Chat", func(t *testing.T) {
				m, _ := newFrameModel(t, size.width, size.height)
				m.header.coordinator = "working"
				m.header.coordinatorChangedAt = statusDurationDay.Add(-2 * time.Minute).Format(time.RFC3339)
				m.invalidateView()
				requireFits(t, m, size.width, size.height)
				requireView(t, m, "● working 00:02:00")
			})

			t.Run("Agents", func(t *testing.T) {
				home, clock := statusDurationAgentsHome(t, statusDurationDay)
				m := openAgents(t, home, clock)
				m, _ = raise(t, m, tea.WindowSizeMsg{Width: size.width, Height: size.height})
				requireFits(t, m, size.width, size.height)
				requireView(t, m, "working · 00:02:00")
			})

			t.Run("Work", func(t *testing.T) {
				m := statusDurationWorkModel(t)
				m, _ = raise(t, m, tea.WindowSizeMsg{Width: size.width, Height: size.height})
				requireWorkFits(t, m, size.width, size.height, 0, 0)
				requireView(t, m, "00:12:00")
			})

			t.Run("Health", func(t *testing.T) {
				m := newGfDemoModel(t, "GfHealth")
				record := records.Agent{
					Key: "agent-with-a-very-long-name", Name: "agent-with-a-very-long-name",
					Harness: "a-very-long-harness", Status: records.AgentLive, AgentStatus: "working",
					PaneID: "w1:p1", StatusChangedAt: statusDurationDay.Add(-2 * time.Minute).Format(time.RFC3339),
				}
				m.health.acceptAgents(agentsData{nodes: []agentNode{{record: record, current: true}}})
				m, _ = raise(t, m, tea.WindowSizeMsg{Width: size.width, Height: size.height})
				requireHealthFits(t, m, size.width, size.height)
				requireView(t, m, "working 00:02:05")
			})
		})
	}
}

// statusDurationAgentsHome gives one project, a manager, the coordinator, and
// one worker with a long name and harness.
func statusDurationAgentsHome(t *testing.T, now time.Time) (string, *testClock) {
	t.Helper()
	home := newTestHome(t)
	clock := fixedClock(t, now)
	pidStart, err := process.Start(int64(os.Getpid()))
	if err != nil {
		t.Fatal(err)
	}
	writeRecord(t, home, []string{"claims"}, "supervisor", records.Claim{Key: "supervisor", Agent: "supervisor", Generation: 1, PID: int64(os.Getpid()), PIDStart: pidStart})
	writeRecord(t, home, []string{"projects"}, "demo", records.Project{ID: "demo", Path: "/repo/demo", Base: "dev"})
	writeRecord(t, home, []string{"agents"}, "coordinator", records.Agent{Key: "coordinator", Name: "coordinator", Kind: records.AgentKindCoordinator, Harness: "claude", Status: records.AgentLive, AgentStatus: "working", StatusChangedAt: now.Add(-2 * time.Minute).Format(time.RFC3339)})
	writeAgentRecord(t, home, records.Agent{Key: "manager-demo", Name: "manager-demo", Profile: records.ProfilePlanner, Owns: "demo", Project: "demo", Harness: "claude", Status: records.AgentLive, AgentStatus: "idle", StatusChangedAt: now.Add(-3 * time.Minute).Format(time.RFC3339)})
	writeAgentRecord(t, home, records.Agent{Key: "worker-with-a-very-long-name", Name: "worker-with-a-very-long-name", Profile: records.ProfileBuilder, Project: "demo", Harness: "a-very-long-harness", Parent: "manager-demo", Status: records.AgentLive, AgentStatus: "working", StatusChangedAt: now.Add(-2 * time.Minute).Format(time.RFC3339)})
	return home, clock
}

// statusDurationWorkModel gives the Work board with a long worker name.
func statusDurationWorkModel(t *testing.T) *Model {
	t.Helper()
	data := workTestData()
	long := "worker-with-a-very-long-name"
	data.Agents[0] = records.Agent{Key: long, Name: long, Profile: records.ProfileBuilder, Status: records.AgentLive, AgentStatus: "working", StatusChangedAt: workTestNow.Add(-12 * time.Minute).Format(time.RFC3339)}
	for index := range data.Tasks {
		if data.Tasks[index].ID == "5-e" {
			data.Tasks[index].ActingAgent = long
		}
	}
	return newWorkModel(t, data, demoUIState{Selection: "5-e"}, 120, 40)
}
