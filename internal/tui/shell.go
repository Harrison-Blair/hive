package tui

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const (
	colorBorder  = "#3d3a33"
	colorDim     = "#a39d90"
	colorAccent  = "#f5b942"
	colorWorking = "#6ee69a"
	colorDone    = "#7ccf6a"
	colorInfo    = "#6fb7ff"
)

var (
	headerStyle  = lipgloss.NewStyle().Bold(true)
	workingStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(colorWorking))
	doneStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color(colorDone))
	infoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color(colorInfo))
	titleStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color(colorAccent)).Bold(true)
	dimStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color(colorDim))
)

type dataRequest uint8

const (
	chatDataRequest dataRequest = iota
	logsDataRequest
	agentsDataRequest
	healthDataRequest
	tasksDataRequest
	frictionDataRequest
)

type screen interface {
	Update(tea.Msg) tea.Cmd
	View(width, height int) string
	KeyHints() string
	DataRequest() dataRequest
}

func (m *Model) makeScreens() []screen {
	return []screen{
		chatScreen{model: m},
		agentsScreen{model: m},
		// Slots 3 and 4 held the Queue and Tasks views. The Work board is
		// one screen. The slots stay, so the screen numbers do not change.
		workScreen{model: m},
		workScreen{model: m},
		workScreen{model: m},
		frictionScreen{model: m},
		healthScreen{model: m},
		logsScreen{model: m},
	}
}

func (m *Model) currentScreen() screen { return m.screens[m.screenIndex] }
