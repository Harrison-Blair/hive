package tui

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
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
