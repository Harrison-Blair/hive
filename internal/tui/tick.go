package tui

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/Harrison-Blair/hive/internal/logs"
)

const tickInterval = 500 * time.Millisecond

// tickMsg refreshes the header and the viewport.
type tickMsg struct{}

// eventMsg carries one event from the follower.
type eventMsg struct {
	event logs.Event
	ok    bool
}

// followErrMsg reports the end of the follower.
type followErrMsg struct {
	err error
}

func (m *Model) tick() tea.Cmd {
	return tea.Tick(tickInterval, func(time.Time) tea.Msg { return tickMsg{} })
}

func (m *Model) waitEvent() tea.Cmd {
	events := m.eventCh
	return func() tea.Msg {
		if events == nil {
			return eventMsg{}
		}
		event, ok := <-events
		return eventMsg{event: event, ok: ok}
	}
}

func (m *Model) waitFollow() tea.Cmd {
	done := m.followDone
	return func() tea.Msg {
		if done == nil {
			return followErrMsg{}
		}
		<-done
		return followErrMsg{err: m.followErr}
	}
}
