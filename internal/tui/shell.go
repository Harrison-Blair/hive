package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const (
	colorBackground = "#0e0f13"
	colorPanel      = "#12141a"
	colorBorder     = "#3d3a33"
	colorText       = "#e8e4da"
	colorDim        = "#a39d90"
	colorAccent     = "#f5b942"
	colorWorking    = "#6ee69a"
	colorDone       = "#7ccf6a"
	colorInfo       = "#6fb7ff"
)

var (
	headerStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color(colorText)).Background(lipgloss.Color(colorBackground)).Bold(true)
	workingStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color(colorWorking))
	doneStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color(colorDone))
	infoStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color(colorInfo))
	panelStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color(colorText)).Background(lipgloss.Color(colorPanel)).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(colorBorder)).Padding(0, 1)
	titleStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color(colorAccent)).Bold(true)
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(colorAccent)).Bold(true)
	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color(colorDim))
	footerStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color(colorDim)).Background(lipgloss.Color(colorBackground))
)

type dataRequest uint8

const (
	noDataRequest dataRequest = iota
	logsDataRequest
	agentsDataRequest
)

type screen interface {
	Update(tea.Msg) tea.Cmd
	View(width, height int) string
	KeyHints() string
	DataRequest() dataRequest
}

type placeholderScreen struct{ name string }

func (s placeholderScreen) Update(tea.Msg) tea.Cmd   { return nil }
func (s placeholderScreen) View(_, _ int) string     { return s.name + " · next wave" }
func (s placeholderScreen) KeyHints() string         { return "1–8 screens · ctrl+c twice to quit" }
func (s placeholderScreen) DataRequest() dataRequest { return noDataRequest }

type chatScreen struct{ placeholderScreen }
type goalsScreen struct{ placeholderScreen }
type queueScreen struct{ placeholderScreen }
type tasksScreen struct{ placeholderScreen }
type frictionScreen struct{ placeholderScreen }
type healthScreen struct{ placeholderScreen }

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

func (s logsScreen) View(_, _ int) string {
	body := s.model.bodyView()
	if s.model.mode == editing {
		body += "\n" + s.model.input.View()
	}
	return body
}

func (s logsScreen) KeyHints() string {
	return "1–8 screens · / filter · l level · f follow · g/G scroll · ctrl+c×2 quit"
}

func (s logsScreen) DataRequest() dataRequest { return logsDataRequest }

func (m *Model) makeScreens() []screen {
	return []screen{
		chatScreen{placeholderScreen{name: "Chat"}},
		agentsScreen{model: m},
		goalsScreen{placeholderScreen{name: "Goals"}},
		queueScreen{placeholderScreen{name: "Queue"}},
		tasksScreen{placeholderScreen{name: "Tasks"}},
		frictionScreen{placeholderScreen{name: "Friction"}},
		healthScreen{placeholderScreen{name: "Health"}},
		logsScreen{model: m},
	}
}

func (m *Model) currentScreen() screen { return m.screens[m.screenIndex] }

func (m *Model) currentScreenName() string {
	return [...]string{"Chat", "Agents", "Goals", "Queue", "Tasks", "Friction", "Health", "Logs"}[m.screenIndex]
}

func (m *Model) shellView() string {
	width := m.width
	if width < 1 {
		width = defaultWidth
	}
	header := headerStyle.Width(width).Render(m.headerLine())
	headerRows := lipgloss.Height(header)
	bodyHeight := m.height - headerRows - statusRows - 5
	if bodyHeight < 1 {
		bodyHeight = 1
	}
	sidebarWidth := 24
	if width >= 100 {
		sidebarWidth = 30
	}
	if sidebarWidth > width-12 {
		sidebarWidth = width - 12
	}
	if sidebarWidth < 12 {
		sidebarWidth = 12
	}
	mainWidth := width - sidebarWidth
	if mainWidth < 1 {
		mainWidth = 1
	}
	left := m.renderSidebar(sidebarWidth, bodyHeight)
	body := m.renderPanel(m.currentScreenName(), m.currentScreen().View(mainWidth, bodyHeight), mainWidth, bodyHeight)
	panels := lipgloss.JoinHorizontal(lipgloss.Top, left, body)
	footer := m.footerView(width)
	return lipgloss.JoinVertical(lipgloss.Left, header, panels, footer)
}

func (m *Model) renderSidebar(width, height int) string {
	lines := []string{titleStyle.Render("SCREENS")}
	names := [...]string{"Chat", "Agents", "Goals", "Queue", "Tasks", "Friction", "Health", "Logs"}
	for index, name := range names {
		line := "  " + string(rune('1'+index)) + " " + name
		if index == m.screenIndex {
			line = selectedStyle.Render("▶ " + string(rune('1'+index)) + " " + name)
		}
		lines = append(lines, line)
	}
	return panelStyle.Width(width).Height(height).Render(strings.Join(lines, "\n"))
}

func (m *Model) renderPanel(title, content string, width, height int) string {
	body := titleStyle.Render(title)
	if content != "" {
		body += "\n" + content
	}
	return panelStyle.Width(width).Height(height).Render(body)
}

func (m *Model) footerView(width int) string {
	notice := ""
	status := ""
	if m.screenIndex == 7 {
		status = m.statusLine()
	}
	if m.quitNotice {
		notice = "press ctrl+c again to quit"
	}
	if err := m.footerError(); err != "" {
		if notice == "" {
			notice = "error: " + err
		} else {
			status = "error: " + err
		}
	}
	hints := m.currentScreen().KeyHints()
	notice = fitLine(notice, width)
	status = fitLine(status, width)
	hints = fitLine(hints, width)
	return lipgloss.JoinVertical(lipgloss.Left,
		footerStyle.Width(width).Render(notice),
		footerStyle.Width(width).Render(status),
		footerStyle.Width(width).Render(hints),
	)
}

func (m *Model) footerError() string {
	for _, text := range []string{m.attachErr, m.headerErr, m.screenErr, m.logErr} {
		if text != "" {
			return text
		}
	}
	return ""
}

func fitLine(text string, width int) string {
	if width < 1 || lipgloss.Width(text) <= width {
		return text
	}
	var result strings.Builder
	used := 0
	for _, r := range text {
		runeWidth := lipgloss.Width(string(r))
		if used+runeWidth > width-1 {
			break
		}
		result.WriteRune(r)
		used += runeWidth
	}
	result.WriteRune('…')
	return result.String()
}
