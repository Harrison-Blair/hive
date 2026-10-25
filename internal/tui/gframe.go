package tui

import (
	"fmt"
	"image/color"
	"slices"
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

// escCloses reports that esc closes a part of the screen first (decision
// 17): a peek, the one-pod view of Agents, a details region, an open file,
// or the Logs field panel overlay. On Friction, esc returns to Chat from
// each box.
func (m *Model) escCloses() bool {
	switch tab := m.activeTab(); {
	case tab == tabLogs:
		// logs.md: esc returns to Chat from each box. Below 120 columns
		// it closes the field panel overlay first.
		return m.detailsFocused() && m.width < component.LongFormWidth
	case m.detailsFocused(), m.agents.peekOpen:
		return true
	case tab == tabAgents:
		return m.graph.podView
	case tab == tabWork:
		return m.tasks.popup != ""
	}
	return false
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

// screenStatusLine is the default first status row of each screen other
// than Chat (decision 34).
var screenStatusLine = []string{"state", "screen", "clock"}

// statusLines gives the status rows of the config. On a screen other than
// Chat, a config without tui.status.lines gives the first row of decision 34.
func (m *Model) statusLines() [][]string {
	lines := m.tuiConfig.StatusLines
	if m.activeTab() == tabChat || m.tuiConfig.StatusLinesExplicit {
		return lines
	}
	return append([][]string{screenStatusLine}, lines[1:]...)
}

// statusRows gives the status rows of decisions 8 to 11 and 34 at width
// cells.
func (m *Model) statusRows(width, height int) []string {
	lines := m.statusLines()
	if height < statusRowsMin && len(lines) > 1 {
		lines = lines[:1]
	}
	_, _, asking := m.focusedAsk()
	asking = asking && m.activeTab() == tabChat && !m.helpOpen
	rows := make([]string, 0, len(lines))
	for index, names := range lines {
		if index == 0 && asking && m.askFocused() {
			rows = append(rows, m.askStatusRow(width))
			continue
		}
		row := component.StatusRow{}
		for _, name := range names {
			switch {
			case name == "clock":
				row.Clock = m.now().Format("15:04")
				continue
			case asking && name == "chords":
				// Decision 33: while an ask is open, this segment is in
				// place of chords and hints.
				row.Segments = append(row.Segments, component.AccentText("^r answer the ask"))
				continue
			case asking && name == "hints":
				continue
			case name == "screen":
				// Decision 34: each key is its own segment, so that the
				// row drops keys from the right before the clock.
				row.Segments = append(row.Segments, m.screenSegments()...)
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

// askStatusRow gives the status row of decision 33 while the ask card has
// the focus: the state, the focus name, the keys of the focus, the asks
// segment when more than one ask is open, and the clock.
func (m *Model) askStatusRow(width int) string {
	focus := m.focusState()
	row := component.StatusRow{
		Segments: []string{m.stateSegment(), component.StrongText(focus)},
		Clock:    m.now().Format("15:04"),
	}
	for _, key := range m.askHints(focus) {
		row.Segments = append(row.Segments, component.DimText(key))
	}
	if len(m.openAsks()) > 1 {
		row.Tail = []string{m.statusSegment("asks")}
	}
	return row.View(width)
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

// screenSegments gives the screen segment of decisions 11 and 34: the screen
// name, or the name of the popup that has the focus, in accent bold, then
// each key of the focus in dim.
func (m *Model) screenSegments() []string {
	name := strings.ToLower(screenRegistry[m.activeTab()].label)
	if focus := m.focusState(); !slices.Contains(screenRegistry[m.activeTab()].regions, focus) {
		name = focus
	}
	segments := []string{component.StrongText(name)}
	for _, key := range m.focusHints(m.width) {
		segments = append(segments, component.DimText(key))
	}
	return segments
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
	text := ""
	switch m.header.coordinator {
	case "working":
		text = component.WorkingGlyph(m.markerFrame, "●") + " working"
		if age := coordinatorStatusAge(m.header.coordinatorChangedAt, m.now()); age != "" {
			text += " " + age
		}
	case "idle":
		text = "○ idle"
	case "blocked":
		text = "▲ blocked"
	case "":
		text = "○ none"
	default:
		text = "● starting"
	}
	return component.GlyphText(text, stateColor(m.header.coordinator))
}

// stateColor gives the glyph colour of a coordinator state of decision 11.
func stateColor(state string) color.Color {
	switch state {
	case "working":
		return component.Working
	case "blocked":
		return component.Accent
	}
	return component.Dim
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
	if m.agents.restartConfirm && notice == m.restartPrompt() {
		return fitLine(" "+component.StrongText(notice), width)
	}
	return fitLine(" "+component.AccentText(notice), width)
}

// fitLine cuts or pads one row to exactly width cells.
func fitLine(row string, width int) string {
	return fitBlock(row, width, 1)
}
