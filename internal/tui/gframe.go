package tui

import (
	"fmt"
	"image/color"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/Harrison-Blair/hive/internal/lib/home"
	"github.com/Harrison-Blair/hive/internal/tui/component"
)

// chordTabs maps each chord action of golden-frames decision 16 to its
// screen, in status row order.
var chordTabs = []struct {
	action string
	tab    int
}{{"agents", tabAgents}, {"work", tabWork}, {"friction", tabFriction}, {"health", tabHealth}, {"logs", tabLogs}}

// The config of the TUI tables. A parameter named home hides the package in
// newModel and RunHome.
var (
	defaultTUI      = home.DefaultTUI
	readTUISettings = home.ReadTUISettings
)

// errorStyle is the error: prefix of the notice line (decision 14).
var errorStyle = lipgloss.NewStyle().Foreground(component.Error).Bold(true)

// statusRowsMin is the pane height from which the TUI shows each status row
// of the config (decision 8). Below it, the TUI shows only the first row.
const statusRowsMin = 30

// setTUIConfig takes the [tui.status] and [tui.keys] settings.
func (m *Model) setTUIConfig(cfg home.TUIConfig) {
	m.tuiConfig = cfg
	m.chordActions = make(map[string]string, len(cfg.Keys))
	for action, key := range cfg.Keys {
		m.chordActions[key] = action
	}
	m.invalidateView()
}

// updateChord handles a chord of decision 16 and the help key of decision
// 17. It reports false for each other key.
func (m *Model) updateChord(press tea.KeyPressMsg) (tea.Cmd, bool) {
	key := press.String()
	action, ok := m.chordActions[key]
	if !ok || action == "help" || m.helpOpen || m.agents.restartConfirm {
		return nil, false
	}
	for _, chord := range chordTabs {
		if chord.action != action {
			continue
		}
		logged := m.logAction("chord", slogString("key", key), slogString("chord", action))
		if m.activeTab() == chord.tab {
			return batch(logged, m.openChat()), true
		}
		m.popupOpen = false
		return batch(logged, m.openTab(chord.tab)), true
	}
	return nil, false
}

// openChat returns to Chat (decision 1).
func (m *Model) openChat() tea.Cmd {
	m.popupOpen = false
	return m.openTab(tabChat)
}

// chordText gives the status row form of a key: ctrl+g is ^g.
func chordText(key string) string {
	if rest, ok := strings.CutPrefix(key, "ctrl+"); ok {
		return "^" + rest
	}
	return key
}

// statusRows gives the status rows of decisions 8 to 11 at width cells.
func (m *Model) statusRows(width, height int) []string {
	lines := m.tuiConfig.StatusLines
	if height < statusRowsMin && len(lines) > 1 {
		lines = lines[:1]
	}
	rows := make([]string, 0, len(lines))
	for _, names := range lines {
		row := component.StatusRow{}
		for _, name := range names {
			if name == "clock" {
				row.Clock = m.now().Format("15:04")
				continue
			}
			if segment := m.statusSegment(name); segment != "" {
				row.Segments = append(row.Segments, segment)
			}
		}
		rows = append(rows, row.View(width))
	}
	return rows
}

// statusSegment renders one segment of decision 11. An empty segment is
// left out of the row.
func (m *Model) statusSegment(name string) string {
	switch name {
	case "state":
		return m.stateSegment()
	case "supervisor":
		switch {
		case m.header.supervisorOld:
			return component.GlyphText("▲ supervisor", component.Accent)
		case m.header.supervisor == "fresh":
			return component.GlyphText("● supervisor", component.Working)
		}
		return component.GlyphText("○ supervisor", component.Dim)
	case "asks":
		text := fmt.Sprintf("? %d asks", m.header.asks)
		if m.header.asks == 1 {
			text = "? 1 ask"
		}
		return countText(text, m.header.asks)
	case "needs_you":
		return countText(fmt.Sprintf("needs you %d", m.header.needsUser), m.header.needsUser)
	case "project":
		return component.DimText(m.selectedProjectID())
	case "screen":
		tab := m.activeTab()
		keys := screenRegistry[tab].keyHints(m.viewName(), m.hintKey(tab, m.focusState()), true)
		return component.StrongText(strings.ToLower(screenRegistry[tab].label)) + component.DimText(strings.Join(append([]string{""}, keys...), " · "))
	case "chords":
		parts := make([]string, 0, len(chordTabs))
		for _, chord := range chordTabs {
			if chord.tab == m.activeTab() {
				parts = append(parts, "esc chat")
				continue
			}
			parts = append(parts, chordText(m.tuiConfig.Keys[chord.action])+" "+chord.action)
		}
		return component.DimText(strings.Join(parts, "  "))
	case "hints":
		return component.DimText(chordText(m.tuiConfig.Keys["help"]) + " help")
	case "command":
		return m.statusCommandLine
	}
	return ""
}

// countText renders a count segment: accent bold when n is more than 0.
func countText(text string, n int) string {
	if n > 0 {
		return component.StrongText(text)
	}
	return component.DimText(text)
}

// stateSegment gives the coordinator state of decision 11. A live record
// with no agent status shows as starting.
func (m *Model) stateSegment() string {
	age := coordinatorStatusAge(m.header.coordinatorChangedAt, m.now())
	with := func(text string) string {
		if age == "" {
			return text
		}
		return text + " " + age
	}
	var glyph color.Color
	text := ""
	switch m.header.coordinator {
	case "working":
		glyph, text = component.Working, with("● working")
	case "idle":
		glyph, text = component.Dim, with("○ idle")
	case "blocked":
		glyph, text = component.Accent, with("▲ blocked")
	case "":
		glyph, text = component.Dim, "○ none"
	default:
		glyph, text = component.Dim, "● starting"
	}
	return component.GlyphText(text, glyph)
}

// selectedProjectID gives the project of the selected agent, or empty.
func (m *Model) selectedProjectID() string {
	if node, _, ok := m.selectedAgent(); ok {
		return node.record.Project
	}
	return ""
}

// noticeLine gives the notice line of decisions 14 and 15 at width cells, or
// empty when no notice exists.
func (m *Model) noticeLine(width int) string {
	notice := m.footerNotice()
	if notice == "" {
		return ""
	}
	if text, ok := strings.CutPrefix(notice, "error: "); ok {
		return fitLine(" "+errorStyle.Render("error:")+" "+text, width)
	}
	return fitLine(" "+component.AccentText(notice), width)
}

// fitLine cuts or pads one row to exactly width cells.
func fitLine(row string, width int) string {
	return fitBlock(row, width, 1)
}
