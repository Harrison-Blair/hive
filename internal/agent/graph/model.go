package graph

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/Harrison-Blair/hive/internal/agent/list"
	"github.com/charmbracelet/x/ansi"
)

const (
	refreshInterval = time.Second
	fetchTimeout    = 5 * time.Second
	// chromeRows are the rows around the pod area: the title, the status,
	// the legend, and the keys.
	chromeRows   = 4
	defaultWidth = 80
)

// Options holds the dependencies of the model.
type Options struct {
	// Fetch returns the current entries, as hive agent list does.
	Fetch func(context.Context) ([]list.Entry, error)
	// Now returns the time of the clock and of the elapsed times.
	Now func() time.Time
}

type fetchedMsg struct {
	entries []list.Entry
	err     error
}

type clockMsg time.Time

type frameMsg struct{ gen int }

// Model is the Bubble Tea model of the graph.
type Model struct {
	options Options

	entries   []list.Entry
	pods      []Pod
	loaded    bool
	fetchedAt time.Time
	err       string
	fetching  bool

	now           time.Time
	width, height int
	selected      string
	details       bool
	scroll        int

	spinning bool
	frame    int
	frameGen int
}

// New returns a model that fetches at start and then each second.
func New(options Options) Model {
	return Model{options: options, now: options.Now(), fetching: true}
}

// Init starts the first fetch and the clock.
func (m Model) Init() tea.Cmd {
	return tea.Batch(m.fetch(), clockTick())
}

func clockTick() tea.Cmd {
	return tea.Every(refreshInterval, func(t time.Time) tea.Msg { return clockMsg(t) })
}

func frameTick(gen int) tea.Cmd {
	return tea.Tick(frameInterval, func(time.Time) tea.Msg { return frameMsg{gen: gen} })
}

func (m Model) fetch() tea.Cmd {
	fetch := m.options.Fetch
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), fetchTimeout)
		defer cancel()
		entries, err := fetch(ctx)
		return fetchedMsg{entries: entries, err: err}
	}
}

// startFetch returns a fetch unless one runs.
func (m *Model) startFetch() tea.Cmd {
	if m.fetching {
		return nil
	}
	m.fetching = true
	return m.fetch()
}

// Update handles one message.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case fetchedMsg:
		m.fetching = false
		if msg.err != nil {
			m.err = msg.err.Error()
			break
		}
		m.err, m.loaded, m.fetchedAt = "", true, m.now
		m.entries, m.pods = msg.entries, Pods(msg.entries)
		if _, ok := m.find(m.selected); !ok {
			m.selected = ""
			if order := m.order(); len(order) > 0 {
				m.selected = order[0].Entry.Name
			}
		}
	case clockMsg:
		m.now = time.Time(msg)
		cmds = append(cmds, clockTick(), m.startFetch())
	case frameMsg:
		if msg.gen != m.frameGen || !m.spinning {
			return m, nil
		}
		m.frame++
		if m.wantSpinner() {
			return m, frameTick(m.frameGen)
		}
		m.spinning, m.frame = false, 0
		return m, nil
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "r":
			cmds = append(cmds, m.startFetch())
		case "enter":
			m.details = !m.details && m.selected != ""
		case "esc":
			m.details = false
		case "down", "j":
			m.step(1)
		case "up", "k":
			m.step(-1)
		case "right", "l":
			m.jump(1)
		case "left", "h":
			m.jump(-1)
		}
	}
	m.follow()
	cmds = append(cmds, m.syncSpinner())
	return m, tea.Batch(cmds...)
}

func (m Model) order() []*Node {
	var nodes []*Node
	for _, p := range m.pods {
		nodes = append(nodes, p.Order()...)
	}
	return nodes
}

// find returns the pod index and the index in that pod of name.
func (m Model) find(name string) ([2]int, bool) {
	for i, p := range m.pods {
		for j, n := range p.Order() {
			if n.Entry.Name == name {
				return [2]int{i, j}, true
			}
		}
	}
	return [2]int{}, false
}

func (m *Model) step(delta int) {
	order := m.order()
	for i, n := range order {
		if n.Entry.Name == m.selected {
			m.selected = order[max(0, min(len(order)-1, i+delta))].Entry.Name
			return
		}
	}
}

// jump selects the card at the same index in the pod delta pods away, or
// the last card of that pod when it has fewer cards.
func (m *Model) jump(delta int) {
	at, ok := m.find(m.selected)
	if !ok {
		return
	}
	nodes := m.pods[max(0, min(len(m.pods)-1, at[0]+delta))].Order()
	m.selected = nodes[min(at[1], len(nodes)-1)].Entry.Name
}

func (m Model) screenWidth() int {
	if m.width <= 0 {
		return defaultWidth
	}
	return m.width
}

// area is the number of rows of the pod area, or 0 when the height is not
// known and the view shows each row.
func (m Model) area() int {
	if m.height <= 0 {
		return 0
	}
	return max(1, m.height-chromeRows)
}

func (m Model) board() board {
	return drawBoard(m.pods, m.screenWidth(), m.now, m.selected, m.frame)
}

// follow scrolls the pod area so that the selected card shows.
func (m *Model) follow() {
	area := m.area()
	if area == 0 {
		m.scroll = 0
		return
	}
	b := m.board()
	if card, ok := b.cards[m.selected]; ok {
		if card.top < m.scroll {
			m.scroll = card.top
		}
		if bottom := card.top + cardHeight; bottom > m.scroll+area {
			m.scroll = bottom - area
		}
	}
	m.scroll = max(0, min(m.scroll, len(b.lines)-area))
}

// wantSpinner reports whether a working card shows in the pod area.
func (m Model) wantSpinner() bool {
	if m.details {
		return false
	}
	area := m.area()
	for _, card := range m.board().cards {
		if card.status != "working" {
			continue
		}
		if area == 0 || card.top+cardHeight > m.scroll && card.top < m.scroll+area {
			return true
		}
	}
	return false
}

func (m *Model) syncSpinner() tea.Cmd {
	want := m.wantSpinner()
	switch {
	case want && !m.spinning:
		m.spinning = true
		m.frameGen++
		return frameTick(m.frameGen)
	case !want && m.spinning:
		m.spinning, m.frame = false, 0
	}
	return nil
}

// View draws the screen.
func (m Model) View() tea.View {
	width := m.screenWidth()
	agents := len(m.entries)
	title := fmt.Sprintf(" hive agents · %d %s · %d %s", len(m.pods), plural(len(m.pods), "pod"), agents, plural(agents, "agent"))
	lines := []string{split(title, m.now.Format("15:04:05")+" ", width), m.status()}
	var body []string
	switch {
	case m.details:
		body = m.detailLines(width)
	case !m.loaded:
	case len(m.pods) == 0:
		body = []string{" no agents"}
	default:
		body = m.board().lines
		if area := m.area(); area > 0 {
			body = body[min(m.scroll, len(body)):]
			body = body[:min(area, len(body))]
		}
	}
	if area := m.area(); area > 0 {
		body = body[:min(area, len(body))]
		for len(body) < area {
			body = append(body, "")
		}
	}
	lines = append(lines, body...)
	lines = append(lines, legend(), m.keys())
	for i, line := range lines {
		lines[i] = ansi.Truncate(line, width, "")
	}
	view := tea.NewView(strings.Join(lines, "\n"))
	view.AltScreen = true
	return view
}

func (m Model) status() string {
	switch {
	case m.err != "" && m.loaded:
		return color(styleError, " ! refresh failed: "+m.err+" · showing "+m.fetchedAt.Format("15:04:05"))
	case m.err != "":
		return color(styleError, " ! "+m.err)
	case !m.loaded:
		return color(styleDim, " loading…")
	}
	return color(styleDim, " updated "+m.fetchedAt.Format("15:04:05"))
}

func (m Model) keys() string {
	if m.details {
		return color(styleDim, " [details] esc close · ↑↓ move · r refresh · q quit")
	}
	return color(styleDim, " [graph] ←↑↓→ move · enter details · r refresh · q quit")
}

func legend() string {
	parts := make([]string, 0, len(statuses))
	for _, s := range statuses {
		parts = append(parts, color(s.style, s.glyph)+" "+s.status)
	}
	return " " + color(styleBorder, "──") + " " + strings.Join(parts, "   ")
}

func (m Model) detailLines(width int) []string {
	at, ok := m.find(m.selected)
	if !ok {
		return nil
	}
	e := m.pods[at[0]].Order()[at[1]].Entry
	mark, markStyle := glyph(e.Status, 0)
	started := ""
	if !e.StartedAt.IsZero() {
		started = e.StartedAt.Local().Format("2006-01-02 15:04:05")
	}
	rows := [][2]string{
		{"pod", m.pods[at[0]].Title},
		{"status", color(markStyle, mark) + " " + e.Status},
		{"harness", e.Harness},
		{"model", e.Model},
		{"pane", e.Pane},
		{"parent", e.Parent},
		{"cwd", e.Cwd},
		{"worktree", e.Worktree},
		{"session", e.HarnessSessionID},
		{"started", started},
		{"elapsed", since(e, m.now)},
	}
	w := min(width-2, 72)
	inner := w - 4
	head := "┏━ " + e.Name + " · esc close "
	head += strings.Repeat("━", max(0, w-1-ansi.StringWidth(head))) + "┓"
	lines := []string{" " + color(styleSelected, head)}
	for _, row := range rows {
		value := row[1]
		if value == "" {
			value = color(styleDim, "-")
		}
		text := ansi.Truncate(fmt.Sprintf("%-9s %s", row[0], value), inner, "…")
		pad := max(0, inner-ansi.StringWidth(text))
		lines = append(lines, " "+color(styleSelected, "┃")+" "+text+strings.Repeat(" ", pad)+" "+color(styleSelected, "┃"))
	}
	lines = append(lines, " "+color(styleSelected, "┗"+strings.Repeat("━", max(0, w-2))+"┛"))
	return lines
}

// split puts left and right at the two ends of width, and cuts left when
// both do not fit.
func split(left, right string, width int) string {
	room := width - ansi.StringWidth(right)
	if ansi.StringWidth(left) >= room {
		left = ansi.Truncate(left, max(0, room-1), "…")
	}
	return left + strings.Repeat(" ", max(0, room-ansi.StringWidth(left))) + right
}

func color(style, text string) string {
	return "\x1b[" + style + "m" + text + "\x1b[0m"
}

func plural(n int, word string) string {
	if n == 1 {
		return word
	}
	return word + "s"
}
