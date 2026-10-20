package tui

import (
	"context"
	"log/slog"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Harrison-Blair/hive/internal/doctor"
)

type healthScreen struct{ model *Model }

type healthState struct {
	result   doctor.Result
	selected string
}

func loadHealth(home string, logger *slog.Logger) doctor.Result {
	return doctor.Run(context.Background(), home, logger)
}

func (healthScreen) DataRequest() dataRequest { return healthDataRequest }

func (s healthScreen) KeyHints() string {
	return "↑/↓ j/k select · r re-check · ask in Chat · 1–8 screens · ctrl+c×2 quit"
}

func (s healthScreen) Update(msg tea.Msg) tea.Cmd {
	press, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil
	}
	switch press.String() {
	case "up", "k":
		if s.model.health.moveSelection(-1) {
			return s.model.logAction("health selection", slog.String("check", s.model.health.selected))
		}
	case "down", "j":
		if s.model.health.moveSelection(1) {
			return s.model.logAction("health selection", slog.String("check", s.model.health.selected))
		}
	case "r":
		s.model.screenErr = ""
		return batch(s.model.logAction("health refresh"), s.model.startCurrentScreenLoad())
	}
	return nil
}

func (s healthScreen) View(width, _ int) string {
	contentWidth := width - 4
	if contentWidth < 1 {
		contentWidth = 1
	}
	gap := 2
	columnWidth := (contentWidth - gap) / 2
	if columnWidth < 1 {
		columnWidth = 1
	}
	homeChecks, agentChecks := healthGroups(s.model.health.result)
	lines := []string{lipgloss.JoinHorizontal(lipgloss.Top,
		dimStyle.Width(columnWidth).Render("┤ home ├"),
		strings.Repeat(" ", gap),
		dimStyle.Width(columnWidth).Render("┤ agents ├"),
	)}
	rows := max(len(homeChecks), len(agentChecks))
	if rows == 0 {
		lines = append(lines, "loading checks...")
	} else {
		for index := 0; index < rows; index++ {
			homeLine, agentLine := "", ""
			if index < len(homeChecks) {
				check := homeChecks[index]
				homeLine = healthCheckLine(check, check.Name == s.model.health.selected)
			}
			if index < len(agentChecks) {
				check := agentChecks[index]
				agentLine = healthCheckLine(check, check.Name == s.model.health.selected)
			}
			homeLine = trimGraphLine(homeLine, columnWidth)
			agentLine = trimGraphLine(agentLine, columnWidth)
			padding := columnWidth - lipgloss.Width(homeLine)
			if padding < 0 {
				padding = 0
			}
			lines = append(lines, homeLine+strings.Repeat(" ", padding)+strings.Repeat(" ", gap)+agentLine)
		}
	}
	if selected, ok := s.model.health.selectedCheck(); ok {
		lines = append(lines, lipgloss.NewStyle().Width(contentWidth).Render("DETAIL "+selected.Name+": "+selected.Detail))
	} else {
		lines = append(lines, "DETAIL no check selected")
	}
	lines = append(lines, "capacity · later: capacity from qmeter. Not in this goal.")
	return strings.Join(lines, "\n")
}

func healthCheckLine(check doctor.Check, selected bool) string {
	mark := "?"
	style := dimStyle
	switch check.Status {
	case doctor.StatusOK:
		mark, style = "✓", workingStyle
	case doctor.StatusWarn:
		mark, style = "!", infoStyle
	case doctor.StatusFail:
		mark, style = "✕", titleStyle
	}
	name := check.Name
	if selected {
		name = selectedStyle.Render(name)
	}
	return style.Render(mark) + " " + name
}

func healthGroups(result doctor.Result) ([]doctor.Check, []doctor.Check) {
	homeChecks := make([]doctor.Check, 0, len(result.Checks))
	agentChecks := make([]doctor.Check, 0, len(result.Checks))
	for _, check := range result.Checks {
		if check.Name == "starting agents" || check.Name == "unresolved agents" ||
			check.Name == "stale worktrees" || check.Name == "unpushed task branches" || strings.HasPrefix(check.Name, "harness ") {
			agentChecks = append(agentChecks, check)
		} else {
			homeChecks = append(homeChecks, check)
		}
	}
	return homeChecks, agentChecks
}

func (s *healthState) accept(result doctor.Result) {
	selected := s.selected
	s.result = result
	for _, check := range result.Checks {
		if check.Name == selected {
			return
		}
	}
	s.selected = ""
	if len(result.Checks) != 0 {
		s.selected = result.Checks[0].Name
	}
}

func (s *healthState) selectedCheck() (doctor.Check, bool) {
	for _, check := range s.result.Checks {
		if check.Name == s.selected {
			return check, true
		}
	}
	return doctor.Check{}, false
}

func (s *healthState) moveSelection(delta int) bool {
	checks := s.result.Checks
	if len(checks) == 0 {
		return false
	}
	index := 0
	for i, check := range checks {
		if check.Name == s.selected {
			index = i
			break
		}
	}
	next := index + delta
	if next < 0 {
		next = 0
	}
	if next >= len(checks) {
		next = len(checks) - 1
	}
	if next == index {
		return false
	}
	s.selected = checks[next].Name
	return true
}
