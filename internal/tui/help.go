package tui

import (
	"log/slog"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/Harrison-Blair/hive/internal/tui/component"
)

// The Help box of golden-frames screens/help.md, at 120×40.
const (
	helpBoxTitle  = "help · esc or ? closes · ↑↓ scroll"
	helpBoxWidth  = 104
	helpLeftWidth = 48
	helpKeyWidth  = 11
)

// helpKeyStyle is a key of the Help body: fg bold (help.md rule 2).
var helpKeyStyle = lipgloss.NewStyle().Bold(true)

// helpSection is one section of a Help column: a name and its key rows.
type helpSection struct {
	name string
	keys [][2]string
}

// helpColumns gives the sections of the left and the right column of
// help.md rule 2. SCREENS shows the keys of the config (rule 3).
func (m *Model) helpColumns() (left, right []helpSection) {
	keys := m.tuiConfig.Keys
	left = []helpSection{
		{"SCREENS", [][2]string{
			{keys["agents"], "agents"}, {keys["work"], "work"}, {keys["friction"], "friction"},
			{keys["health"], "health"}, {keys["logs"], "logs"},
			{"esc", "back to chat · the same chord too"}, {keys["help"], "this help, from an empty composer"},
			{"ctrl+c ×2", "quit"},
		}},
		{"CHAT", [][2]string{
			{"enter", "send"}, {"shift+enter", "newline"}, {"ctrl+o", "open or fold all events"},
			{"pgup pgdn", "scroll · the wheel too"}, {"click", "caret · click on an event opens it"},
			{"tab", "reserved for autocomplete"},
		}},
		{"ASK CARD", [][2]string{
			{"ctrl+r", "card focus, or back to the input"}, {"↑↓ · 1..9", "move · pick"},
			{"space", "mark, when the ask allows many"}, {"enter", "review, then send"}, {"n", "note"},
			{"x", "dismiss"}, {"J K", "next or previous card"}, {"esc", "back to the input"},
		}},
	}
	right = []helpSection{
		{"AGENTS", [][2]string{
			{"←→ ↑↓", "pod · card"}, {"enter", "details popup"}, {"a · p · t", "attach · peek · task"},
			{"h", "show or hide stopped"},
		}},
		{"WORK", [][2]string{
			{"←→ ↑↓", "wave · card"}, {"[ ]", "goal"}, {"enter", "task popup"},
			{"b · v · d", "brief · verified · deps"}, {"s", "stops only"},
		}},
		{"FRICTION", [][2]string{
			{"↑↓", "move"}, {"enter", "open the resolved fold"}, {"/", "filter"}, {"ctrl+]", "details box"},
		}},
		{"HEALTH", [][2]string{
			{"ctrl+]", "grid or agents"}, {"enter", "details"}, {"r", "restart coordinator or supervisor"},
			{"R", "check again"},
		}},
		{"LOGS", [][2]string{
			{"/", "filter · level: comp: proj: agent: req:"}, {"f", "follow"}, {"← →", "day"},
			{"j", "raw JSON"}, {"ctrl+]", "field panel"},
		}},
	}
	return left, right
}

// helpColumnRows gives the rows of one column: each section name, its key
// rows, and one blank row between two sections.
func helpColumnRows(sections []helpSection) []string {
	var rows []string
	for index, section := range sections {
		if index > 0 {
			rows = append(rows, "")
		}
		rows = append(rows, titleStyle.Render(section.name))
		for _, pair := range section.keys {
			rows = append(rows, helpKeyStyle.Render(padCell(pair[0], helpKeyWidth))+" "+dimStyle.Render(pair[1]))
		}
	}
	return rows
}

// helpLines gives the body of the Help box: one blank row, the two columns,
// one blank row, and STATUS ROW (help.md rule 2).
func (m *Model) helpLines() []string {
	left, right := m.helpColumns()
	leftRows, rightRows := helpColumnRows(left), helpColumnRows(right)
	lines := []string{""}
	for index := range max(len(leftRows), len(rightRows)) {
		row := ""
		if index < len(leftRows) {
			row = leftRows[index]
		}
		if index < len(rightRows) {
			row = padCell(row, helpLeftWidth) + rightRows[index]
		}
		lines = append(lines, row)
	}
	return append(lines, "",
		titleStyle.Render("STATUS ROW"),
		dimStyle.Render("segments  state · supervisor · asks · needs_you · project · chords · hints · clock · command"),
		dimStyle.Render(`config    [tui.status] lines = [[…]] · command = "…" · 1 to 3 rows · cut from the right when short`),
	)
}

// helpBoxSize gives the size of the Help box at width×height: 104 columns
// and the rows of the body, cut to the screen. While a notice exists, the
// box at the center stays clear of the notice line above the status rows
// (rule 4).
func (m *Model) helpBoxSize(lines, width, height int) (int, int) {
	if m.noticeLine(width) != "" {
		height -= 2 * (len(m.statusRows(width, height)) + 1)
	}
	return min(helpBoxWidth, width), min(lines+2, height)
}

// helpMaxOffset gives the last scroll offset of the Help body (rule 4).
func (m *Model) helpMaxOffset() int {
	lines := len(m.helpLines())
	_, h := m.helpBoxSize(lines, m.width, m.height)
	return max(lines-max(h-2, 0), 0)
}

// helpView draws the Help box at the center of screen, a text of
// width×height cells, and dims the screen (rule 1). A body taller than the
// box scrolls with the ┃ thumb on the right border. The notice line shows
// on the row above the status rows (rule 4).
func (m *Model) helpView(screen string, width, height int) string {
	lines := m.helpLines()
	w, h := m.helpBoxSize(len(lines), width, height)
	body := max(h-2, 0)
	m.helpOffset = min(max(m.helpOffset, 0), max(len(lines)-body, 0))
	box := component.Panel{
		Title:   helpBoxTitle,
		Focused: true,
		Lines:   lines[m.helpOffset:],
		VMark:   component.Thumb(body, len(lines), body, m.helpOffset),
	}.View(w, h)
	screen = component.Overlay(screen, box, width, height)
	notice := m.noticeLine(width)
	if notice == "" {
		return screen
	}
	rows := strings.Split(screen, "\n")
	if row := height - len(m.statusRows(width, height)) - 1; row >= 0 && row < len(rows) {
		rows[row] = notice
	}
	return strings.Join(rows, "\n")
}

// updateHelpKey handles a key while Help is open: ? or esc closes it, ↑ and
// ↓ scroll it, and each other key does nothing.
func (m *Model) updateHelpKey(key string) tea.Cmd {
	switch key {
	case m.tuiConfig.Keys["help"], "esc":
		m.helpOpen = false
		return m.logAction("help close")
	case "up", "down":
		step := 1
		if key == "up" {
			step = -1
		}
		return m.scrollHelp(step)
	}
	return nil
}

// scrollHelp moves the Help body by step rows and logs the scroll.
func (m *Model) scrollHelp(step int) tea.Cmd {
	m.helpOffset = min(max(m.helpOffset+step, 0), m.helpMaxOffset())
	return m.logAction("help scroll", slog.Int("offset", m.helpOffset))
}
