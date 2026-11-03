package graph

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/Harrison-Blair/hive/internal/agent/list"
	"github.com/Harrison-Blair/hive/internal/lib/style"
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
	// Checkout reads the current checkout for an agent's start directory.
	Checkout func(context.Context, string) string
	// Now returns the time of the clock and of the elapsed times.
	Now func() time.Time
}

type fetchedMsg struct {
	entries   []list.Entry
	checkouts map[string]string
	err       error
	// at is the time when the agent list returned.
	at time.Time
}

type clockMsg time.Time

type frameMsg struct{ gen int }

// stateKey names one agent record. A record replaced under the same name has
// a new start time and so a new key.
type stateKey struct {
	name    string
	started int64
}

// stateSince is the last status that the graph saw for an agent, and the
// time of the refresh that first saw it. changed is false until the graph
// sees a change, and the time in state is then at least the time since the
// start.
type stateSince struct {
	status  string
	since   time.Time
	changed bool
}

// Model is the Bubble Tea model of the graph.
type Model struct {
	options Options

	entries   []list.Entry
	checkouts map[string]string
	states    map[stateKey]stateSince
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
	hidden        bool // hides the outline of the selected card; the selection stays
	xscroll       int
	// list is true in list mode and false in tree mode.
	list bool

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
		at := m.options.Now()
		checkouts := make(map[string]string, len(entries))
		if err == nil && m.options.Checkout != nil {
			for _, entry := range entries {
				checkouts[entry.Name] = m.options.Checkout(ctx, entry.Cwd)
			}
		}
		return fetchedMsg{entries: entries, checkouts: checkouts, err: err, at: at}
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
		m.err, m.loaded, m.fetchedAt = "", true, msg.at
		m.entries, m.pods, m.checkouts = msg.entries, Pods(msg.entries), msg.checkouts
		m.states = m.track(msg.entries, msg.at)
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
		if m.hidden && !m.details {
			switch msg.String() {
			case "enter", "down", "j", "up", "k", "right", "l", "left", "h":
				m.hidden = false
				m.follow()
				return m, m.syncSpinner()
			}
		}
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "r":
			cmds = append(cmds, m.startFetch())
		case "enter":
			m.details = !m.details && m.selected != ""
		case "esc":
			m.hidden = m.hidden || !m.details
			m.details = false
		case "g":
			m.list = !m.list
		case "down", "j":
			m.vertical(1)
		case "up", "k":
			m.vertical(-1)
		case "right", "l":
			m.horizontal(1)
		case "left", "h":
			m.horizontal(-1)
		}
	}
	m.follow()
	cmds = append(cmds, m.syncSpinner())
	return m, tea.Batch(cmds...)
}

// track returns the state of each agent in entries after a refresh that
// returned at at. It drops the agents that left the list.
func (m Model) track(entries []list.Entry, at time.Time) map[stateKey]stateSince {
	states := make(map[stateKey]stateSince, len(entries))
	for _, e := range entries {
		k := stateKey{e.Name, e.StartedAt.UnixNano()}
		old, ok := m.states[k]
		switch {
		case !ok:
			states[k] = stateSince{status: e.Status}
		case old.status != e.Status:
			states[k] = stateSince{status: e.Status, since: at, changed: true}
		default:
			states[k] = old
		}
	}
	return states
}

// inState formats the time of e in its current state. Before the graph sees
// a change, it is ≥ and the time since the start.
func (m Model) inState(e list.Entry) string {
	if s := m.states[stateKey{e.Name, e.StartedAt.UnixNano()}]; s.changed {
		return elapsed(m.now.Sub(s.since))
	}
	if age := since(e, m.now); age != "" {
		return "≥" + age
	}
	return ""
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
	ages := make(map[string]string, len(m.entries))
	for _, e := range m.entries {
		ages[e.Name] = m.inState(e)
	}
	selected := m.selected
	if m.hidden {
		selected = ""
	}
	return drawBoard(m.pods, m.screenWidth(), ages, selected, m.frame, m.checkouts, !m.list)
}

// follow scrolls the pod area down and sideways so that the selected card
// shows. While the outline is hidden, it only keeps both scrolls in the pod
// area.
func (m *Model) follow() {
	b := m.board()
	card, ok := b.cards[m.selected]
	ok = ok && !m.hidden
	width := m.screenWidth()
	if ok {
		if card.left < m.xscroll {
			m.xscroll = card.left
		}
		if right := card.left + card.width; right > m.xscroll+width {
			m.xscroll = right - width
		}
	}
	m.xscroll = max(0, min(m.xscroll, b.width-width))
	area := m.area()
	if area == 0 {
		m.scroll = 0
		return
	}
	if ok {
		if card.top < m.scroll {
			m.scroll = card.top
		}
		if bottom := card.top + cardHeight; bottom > m.scroll+area {
			m.scroll = bottom - area
		}
	}
	m.scroll = max(0, min(m.scroll, len(b.lines)-area))
}

// wantSpinner reports whether a working card shows in the pod area, or the
// details show a working agent.
func (m Model) wantSpinner() bool {
	if m.details {
		e, ok := m.selectedEntry()
		return ok && e.Status == "working"
	}
	area, width := m.area(), m.screenWidth()
	for _, card := range m.board().cards {
		if card.status != "working" || card.left+card.width <= m.xscroll || card.left >= m.xscroll+width {
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
		for i, line := range body {
			body[i] = ansi.Cut(line, m.xscroll, m.xscroll+width)
		}
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
	lines = append(lines, legend(m.frame), m.keys())
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
		return style.Color(style.Error, " ! refresh failed: "+m.err+" · showing "+m.fetchedAt.Format("15:04:05"))
	case m.err != "":
		return style.Color(style.Error, " ! "+m.err)
	case !m.loaded:
		return style.Color(style.Dim, " loading…")
	}
	return style.Color(style.Dim, " updated "+m.fetchedAt.Format("15:04:05"))
}

func (m Model) keys() string {
	if m.details {
		return style.Color(style.Dim, " [details] esc close · ↑↓ move · r refresh · q quit")
	}
	if m.list {
		return style.Color(style.Dim, " [list] ←↑↓→ move · enter details · esc hide · g tree · r refresh · q quit")
	}
	return style.Color(style.Dim, " [tree] ←↑↓→ move · enter details · esc hide · g list · r refresh · q quit")
}

// legend shows the glyph of each status. The working glyph follows frame.
func legend(frame int) string {
	parts := make([]string, 0, len(statuses))
	for _, s := range statuses {
		mark, markStyle := glyph(s.status, frame)
		parts = append(parts, style.Color(markStyle, mark)+" "+s.status)
	}
	return " " + style.Color(style.Border, "──") + " " + strings.Join(parts, "   ")
}

func (m Model) selectedEntry() (list.Entry, bool) {
	at, ok := m.find(m.selected)
	if !ok {
		return list.Entry{}, false
	}
	return m.pods[at[0]].Order()[at[1]].Entry, true
}

func (m Model) detailLines(width int) []string {
	at, ok := m.find(m.selected)
	if !ok {
		return nil
	}
	e := m.pods[at[0]].Order()[at[1]].Entry
	mark, markStyle := glyph(e.Status, m.frame)
	status := style.Color(markStyle, mark) + " " + e.Status
	if age := m.inState(e); age != "" {
		status += " · " + age
	}
	started := ""
	if !e.StartedAt.IsZero() {
		started = e.StartedAt.Local().Format("2006-01-02 15:04:05")
	}
	rows := [][2]string{
		{"pod", m.pods[at[0]].Title},
		{"status", status},
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
	lines := []string{" " + style.Color(style.Selected, head)}
	for _, row := range rows {
		value := row[1]
		if value == "" {
			value = style.Color(style.Dim, "-")
		}
		text := ansi.Truncate(fmt.Sprintf("%-9s %s", row[0], value), inner, "…")
		pad := max(0, inner-ansi.StringWidth(text))
		lines = append(lines, " "+style.Color(style.Selected, "┃")+" "+text+strings.Repeat(" ", pad)+" "+style.Color(style.Selected, "┃"))
	}
	lines = append(lines, " "+style.Color(style.Selected, "┗"+strings.Repeat("━", max(0, w-2))+"┛"))
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

func plural(n int, word string) string {
	if n == 1 {
		return word
	}
	return word + "s"
}
