package tui

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

// TestStatusDurationHealthKeepsHundredHoursAtEachWidth is the literal
// regression for the Health STATE column. The complete state word and all
// three-digit-hour time digits stay at each supported width.
func TestStatusDurationHealthKeepsHundredHoursAtEachWidth(t *testing.T) {
	for _, size := range []struct{ width, height int }{{120, 40}, {119, 40}, {80, 24}} {
		t.Run(fmt.Sprintf("%dx%d", size.width, size.height), func(t *testing.T) {
			fixture, err := loadDemoFixture(filepath.Join(gfDesignDir, "GfHealth.fixture.json"))
			if err != nil {
				t.Fatal(err)
			}
			changed := fixture.now.Add(-100 * time.Hour).Format(time.RFC3339)
			for index := range fixture.Data.Agents {
				fixture.Data.Agents[index].StatusChangedAt = changed
			}
			m := newDemoModel(t.TempDir(), &demoState{fixture: fixture, size: demoSize{Width: size.width, Height: size.height}})
			defer m.Stop()
			m.demoResize(size.width, size.height)
			m.invalidateView()
			requireHealthFits(t, m, size.width, size.height)
			requireView(t, m, "working 100:00:00 w58:p2")
		})
	}
}

// TestStatusDurationWorkDetailsKeepsHundredHoursWithLongNameAtEachWidth is
// the literal regression for the Work task popup worker field. A long worker
// name is elided before the complete state and three-digit-hour duration.
func TestStatusDurationWorkDetailsKeepsHundredHoursWithLongNameAtEachWidth(t *testing.T) {
	const (
		oldName = "worker-with-a-very-long-name"
		newName = "worker-with-a-name-that-is-sixty-characters-long-for-width-test"
		want    = "worker     worker-with-a-name-that-is-sixty-character… · working · 100:00:00"
	)
	for _, size := range []struct{ width, height int }{{120, 40}, {119, 40}, {80, 24}} {
		t.Run(fmt.Sprintf("%dx%d", size.width, size.height), func(t *testing.T) {
			m := statusDurationWorkModel(t)
			agent := m.tasks.data.agents[oldName]
			agent.Key, agent.Name = newName, newName
			agent.StatusChangedAt = workTestNow.Add(-100 * time.Hour).Format(time.RFC3339)
			delete(m.tasks.data.agents, oldName)
			m.tasks.data.agents[newName] = agent
			for index := range m.tasks.data.items {
				if m.tasks.data.items[index].record.ActingAgent == oldName {
					m.tasks.data.items[index].record.ActingAgent = newName
				}
			}
			m = pressWork(t, m, key(tea.KeyEnter, ""))
			m, _ = raise(t, m, tea.WindowSizeMsg{Width: size.width, Height: size.height})
			requireWorkFits(t, m, size.width, size.height, workTaskPopupWidth, workTaskPopupHeight)
			requireView(t, m, want)
		})
	}
}

// TestStatusDurationHealthStateBudgetFollowsDisplayedStateAtEachWidth is the
// literal regression for the Health STATE column budget. The budget follows
// the complete displayed state at the 99 to 100 and 999 to 1000 hour-digit
// boundaries, and the header and following PANE column stay aligned with it.
func TestStatusDurationHealthStateBudgetFollowsDisplayedStateAtEachWidth(t *testing.T) {
	boundaries := []struct {
		name      string
		low, high int
		lowState  string
		highState string
		headerGap string
	}{
		{"two to three hour digits", 99, 100, "working 99:00:00", "working 100:00:00", "STATE             PANE"},
		{"three to four hour digits", 999, 1000, "working 999:00:00", "working 1000:00:00", "STATE              PANE"},
	}
	for _, boundary := range boundaries {
		t.Run(boundary.name, func(t *testing.T) {
			for _, size := range []struct{ width, height int }{{120, 40}, {119, 40}, {80, 24}} {
				t.Run(fmt.Sprintf("%dx%d", size.width, size.height), func(t *testing.T) {
					fixture, err := loadDemoFixture(filepath.Join(gfDesignDir, "GfHealth.fixture.json"))
					if err != nil {
						t.Fatal(err)
					}
					low := fixture.now.Add(-time.Duration(boundary.low) * time.Hour).Format(time.RFC3339)
					high := fixture.now.Add(-time.Duration(boundary.high) * time.Hour).Format(time.RFC3339)
					for index := range fixture.Data.Agents {
						switch fixture.Data.Agents[index].Key {
						case "coordinator-g2":
							fixture.Data.Agents[index].StatusChangedAt = low
						case "supervisor-g4":
							fixture.Data.Agents[index].AgentStatus = "working"
							fixture.Data.Agents[index].StatusChangedAt = high
						}
					}
					m := newDemoModel(t.TempDir(), &demoState{fixture: fixture, size: demoSize{Width: size.width, Height: size.height}})
					defer m.Stop()
					m.demoResize(size.width, size.height)
					m.invalidateView()
					view := requireHealthFits(t, m, size.width, size.height)
					requireView(t, m, boundary.lowState, boundary.highState, boundary.headerGap)
					rejectView(t, m, strings.TrimSuffix(boundary.highState, "00")+"…")
					header := statusDurationLine(t, view, "AGENT")
					row := statusDurationLine(t, view, boundary.highState)
					if got, want := statusDurationCell(row, boundary.highState), statusDurationCell(header, "STATE"); got != want {
						t.Fatalf("state starts at cell %d; header STATE starts at %d:\n%s\n%s", got, want, header, row)
					}
					if got, want := statusDurationCell(row, "w58:p3"), statusDurationCell(header, "PANE"); got != want {
						t.Fatalf("pane starts at cell %d; header PANE starts at %d:\n%s\n%s", got, want, header, row)
					}
				})
			}
		})
	}
}

// TestStatusDurationWorkDetailsEmptyWorkerKeepsLabelAtEachWidth is the
// literal regression for the Work task popup default worker field. A task
// with no ActingAgent keeps the worker label and shows none.
func TestStatusDurationWorkDetailsEmptyWorkerKeepsLabelAtEachWidth(t *testing.T) {
	for _, size := range []struct{ width, height int }{{120, 40}, {119, 40}, {80, 24}} {
		t.Run(fmt.Sprintf("%dx%d", size.width, size.height), func(t *testing.T) {
			m := newWorkModel(t, workTestData(), demoUIState{Selection: "4-d"}, 120, 40)
			m = pressWork(t, m, key(tea.KeyEnter, ""))
			m, _ = raise(t, m, tea.WindowSizeMsg{Width: size.width, Height: size.height})
			requireWorkFits(t, m, size.width, size.height, workTaskPopupWidth, workTaskPopupHeight)
			requireView(t, m, "worker     none")
		})
	}
}

// statusDurationLine gives the first plain view line that contains want.
func statusDurationLine(t *testing.T, view, want string) string {
	t.Helper()
	for _, line := range strings.Split(view, "\n") {
		if strings.Contains(line, want) {
			return line
		}
	}
	t.Fatalf("no line contains %q:\n%s", want, view)
	return ""
}

// statusDurationCell gives the cell of want in a plain view line.
func statusDurationCell(line, want string) int {
	at := strings.Index(line, want)
	if at < 0 {
		return -1
	}
	return ansi.StringWidth(line[:at])
}
