package tui

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/Harrison-Blair/hive/internal/doctor"
	"github.com/Harrison-Blair/hive/internal/logs"
)

const tickInterval = 2 * time.Second

// tickMsg refreshes the header and the current screen.
type tickMsg struct{}

type headerLoadMsg struct {
	state headerState
	err   error
}

type screenLoadMsg struct {
	request    dataRequest
	events     []logs.Event
	agents     agentsData
	goalsQueue goalsQueueData
	chat       chatScreenData
	health     doctor.Result
	err        error
}

type agentsPeekMsg struct {
	paneID string
	text   string
	err    error
}

type attachMsg struct{ err error }

type quitNoticeExpiredMsg struct{ at time.Time }

type logWriteErrMsg struct{ err error }

func (m *Model) tick() tea.Cmd {
	return tea.Tick(tickInterval, func(time.Time) tea.Msg { return tickMsg{} })
}
