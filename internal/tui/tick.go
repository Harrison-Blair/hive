package tui

import (
	"time"

	tea "charm.land/bubbletea/v2"
)

const tickInterval = 500 * time.Millisecond

// tickMsg refreshes the header and the viewport. It is also the one consumer
// of the follower event channel, so events keep their file order.
type tickMsg struct{}

// followErrMsg reports the end of the follower.
type followErrMsg struct {
	err error
}

func (m *Model) tick() tea.Cmd {
	return tea.Tick(tickInterval, func(time.Time) tea.Msg { return tickMsg{} })
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
