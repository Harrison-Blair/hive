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
	headerStyle   = lipgloss.NewStyle().Bold(true)
	workingStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color(colorWorking))
	doneStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color(colorDone))
	infoStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color(colorInfo))
	panelStyle    = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(colorBorder)).Padding(0, 1)
	titleStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color(colorAccent)).Bold(true)
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(colorAccent)).Bold(true)
	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color(colorDim))
)

type dataRequest uint8

const (
	chatDataRequest dataRequest = iota
	logsDataRequest
	agentsDataRequest
	goalsQueueDataRequest
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
		goalsScreen{model: m},
		queueScreen{model: m},
		tasksScreen{model: m},
		frictionScreen{model: m},
		healthScreen{model: m},
		logsScreen{model: m},
	}
}

func (m *Model) currentScreen() screen { return m.screens[m.screenIndex] }

func (m *Model) currentScreenName() string { return m.legacyName() }
