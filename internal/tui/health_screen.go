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

func (s healthScreen) View(width, height int) string {
	contentWidth := max(1, width-4)
	gap := 2
	columnWidth := max(1, (contentWidth-gap)/2)
	homeChecks, agentChecks := healthGroups(s.model.health.result)
	homeLines := []string{dimStyle.Render("┤ home ├")}
	agentTitle := "┤ agents ├"
	if columnWidth-4 < lipgloss.Width(agentTitle) {
		agentTitle = "agents"
	}
	agentLines := []string{dimStyle.Render(agentTitle)}
	if len(homeChecks)+len(agentChecks) == 0 {
		loading := trimGraphLine("loading checks...", columnWidth-4)
		homeLines = append(homeLines, loading)
		agentLines = append(agentLines, loading)
	} else {
		for _, check := range homeChecks {
			homeLines = append(homeLines, trimGraphLine(healthCheckLine(check, check.Name == s.model.health.selected), columnWidth-4))
		}
		for _, check := range agentChecks {
			agentLines = append(agentLines, trimGraphLine(healthCheckLine(check, check.Name == s.model.health.selected), columnWidth-4))
		}
	}
	homePanel := panelStyle.Width(columnWidth).Render(strings.Join(homeLines, "\n"))
	agentPanel := panelStyle.Width(columnWidth).Render(strings.Join(agentLines, "\n"))
	lines := []string{lipgloss.JoinHorizontal(lipgloss.Top, homePanel, strings.Repeat(" ", gap), agentPanel)}

	warningLines := []string{titleStyle.Render("┤ warnings ├")}
	warnings := healthWarnings(s.model.health.result)
	if len(warnings) == 0 {
		warningLines = append(warningLines, "no warnings")
	} else {
		for _, check := range warnings {
			line := healthCheckLine(check, check.Name == s.model.health.selected) + ": " + check.Detail
			warningLines = append(warningLines, trimGraphLine(line, contentWidth-4))
		}
	}
	warningPanel := panelStyle.BorderForeground(lipgloss.Color("#ffb37a")).Width(contentWidth).Render(strings.Join(warningLines, "\n"))
	lines = append(lines, warningPanel)
	if height > 14 {
		if selected, ok := s.model.health.selectedCheck(); ok {
			lines = append(lines, lipgloss.NewStyle().Width(contentWidth).Render("DETAIL "+selected.Name+": "+selected.Detail))
		} else {
			lines = append(lines, "DETAIL no check selected")
		}
	}
	lines = append(lines, "capacity · later: capacity from qmeter. Not in this goal.")
	return strings.Join(lines, "\n")
}

func healthWarnings(result doctor.Result) []doctor.Check {
	warnings := make([]doctor.Check, 0)
	for _, check := range result.Checks {
		if check.Status == doctor.StatusWarn || check.Status == doctor.StatusFail {
			warnings = append(warnings, check)
		}
	}
	return warnings
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
