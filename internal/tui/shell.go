package tui

import (
	"strings"

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

type logsScreen struct{ model *Model }

func (s logsScreen) Update(msg tea.Msg) tea.Cmd {
	press, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil
	}
	if s.model.mode == editing {
		return s.model.updateEditing(press)
	}
	cmd := s.model.updateViewing(press)
	updated, viewportCmd := s.model.viewport.Update(press)
	s.model.viewport = updated
	return batch(cmd, viewportCmd)
}

func (s logsScreen) View(width, height int) string {
	m := s.model
	if width < 1 {
		width = 1
	}
	filterLines := wrapFilterLine(m.statusLine(), width-4-lipgloss.Width("▶ "))
	if len(filterLines) > 0 {
		levelFilter := "level: " + m.level
		filterLines[0] = selectedStyle.Render("▶ "+levelFilter) + strings.TrimPrefix(filterLines[0], levelFilter)
	}
	listWidth := width
	if m.logDetail && width >= 72 {
		listWidth = width * 2 / 3
	}
	m.viewport.SetWidth(max(1, listWidth-4))
	m.viewport.SetHeight(max(1, height-len(filterLines)-4))
	m.viewport.EnsureVisible(m.logSelection, 0, 1)
	left := "logs · " + m.logDate.Format("2006-01-02") + ".jsonl\n" + strings.Join(filterLines, "\n") + "\n" + m.bodyView()
	if m.mode == editing {
		left += "\n" + m.input.View()
	}
	if !m.logDetail {
		return left
	}
	if width < 72 {
		return left + "\n" + m.detailView(width, max(1, height/3))
	}
	detailWidth := width - listWidth - 1
	detail := panelStyle.Width(max(1, detailWidth-4)).Height(max(1, height-2)).Render(m.detailView(detailWidth, height-4))
	return lipgloss.JoinHorizontal(lipgloss.Top, left, detail)
}

func (s logsScreen) KeyHints() string {
	return "1–8 screens · l level · c component · p project · a agent · / search · r request · [ ] day · f follow · g/G scroll · enter JSON · esc back"
}

func (s logsScreen) DataRequest() dataRequest { return logsDataRequest }

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
