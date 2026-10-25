package tui

import (
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/Harrison-Blair/hive/internal/tui/component"
)

// markerTickMsg advances the working markers of the current view.
type markerTickMsg struct{}

// startMarkerTick starts the working-marker timer when one marker needs
// updates and no timer runs. The timer is separate from the data refresh.
func (m *Model) startMarkerTick() tea.Cmd {
	if m.markerTickRunning || !m.workingMarkerNeeded() {
		return nil
	}
	m.markerTickRunning = true
	return tea.Tick(component.WorkingFrameInterval, func(time.Time) tea.Msg { return markerTickMsg{} })
}

// advanceMarkerTick gives the next marker frame and stops the timer when no
// working marker remains.
func (m *Model) advanceMarkerTick() tea.Cmd {
	if !m.workingMarkerNeeded() {
		m.markerTickRunning = false
		return nil
	}
	m.markerTickRunning = true
	m.markerFrame++
	m.invalidateView()
	return tea.Tick(component.WorkingFrameInterval, func(time.Time) tea.Msg { return markerTickMsg{} })
}

// workingMarkerNeeded reports that the view has a working marker: the
// coordinator state row on each screen, an agent card on Agents, or a
// verified task card on Work.
func (m *Model) workingMarkerNeeded() bool {
	if m.header.coordinator == "working" {
		return true
	}
	switch m.activeTab() {
	case tabAgents:
		for _, node := range m.agents.data.nodes {
			if node.status == "working" {
				return true
			}
		}
	case tabWork:
		for _, item := range m.tasks.data.items {
			if m.tasks.taskWorkingMarker(item) {
				return true
			}
		}
	case tabHealth:
		for _, node := range m.health.agentList() {
			if node.status == "working" {
				return true
			}
		}
	}
	return false
}
