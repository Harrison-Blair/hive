package tui

import (
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/Harrison-Blair/hive/internal/tui/component"
	"github.com/charmbracelet/x/ansi"
)

// frameRows is the rows of the frame: the header, the two tab bar rows, and
// the footer (decision 1).
const frameRows = 4

// The tabs of the tab bar, in key order (decision 3).
const (
	tabAgents = iota
	tabChat
	tabWork
	tabFriction
	tabHealth
	tabLogs
	tabCount
)

// frameScreen is one entry of the screen registry. The number of a screen is
// its index plus 1. Regions are the focus regions in tab order. A screen with
// listDetails has a list region and a details region: from 120 columns they
// sit side by side, and below 120 columns the details region is a full-width
// overlay (decisions 8, 9). Hints and compact give the key hints of each
// focus state from 120 columns and below 120 columns. A key is "view/focus"
// or "focus". Panels gives the list and the details of each view that wave 6
// replaced. The key is the view name.
type frameScreen struct {
	label       string
	views       []string
	regions     []string
	listDetails bool
	hints       map[string][]string
	compact     map[string][]string
	panels      map[string]screenPanels
}

// screenPanels draws one view of a list and details screen. The list gives
// the list region at width×height cells. The details give the title and the
// lines of the details region for a panel of width cells.
type screenPanels struct {
	list    func(m *Model, width, height int, focused bool) string
	details func(m *Model, width int) (string, []string)
	back    func(m *Model) bool
}

// screenRegistry holds the six screens of the tab bar.
var screenRegistry = [tabCount]frameScreen{
	tabAgents: {
		label:   "Agents",
		regions: []string{"graph"},
		hints: map[string][]string{
			"graph":   {"←↑↓→ move", "enter details", "a attach", "p peek", "t task", "h stopped", "? help"},
			"pods":    {"←→ select pod", "↑ coordinator", "enter open pod", "h stopped", "? help"},
			"pod":     {"←↑↓→ move", "enter details", "a attach", "p peek", "t task", "esc pods", "h stopped", "? help"},
			"details": {"a attach", "p peek", "t task", "↑↓ scroll", "esc close"},
		},
		compact: map[string][]string{
			"graph": {"←↑↓→ move", "enter details", "a attach", "p peek", "?"},
			"pods":  {"←→ select pod", "↑ coordinator", "enter open", "?"},
			"pod":   {"←↑↓→ move", "enter details", "esc pods", "?"},
		},
	},
	tabChat: {
		label:   "Chat",
		regions: []string{"feed", "input"},
		hints: map[string][]string{
			"input": {"enter send", "shift+enter newline", "ctrl+o events", "wheel or pgup/pgdn scroll"},
		},
		compact: map[string][]string{
			"input": {"enter send", "wheel or pgup/pgdn scroll"},
		},
	},
	tabWork: {
		label:       "Work",
		views:       []string{"Goals", "Queue", "Tasks"},
		regions:     []string{"list", "details"},
		listDetails: true,
		hints: map[string][]string{
			"Goals/list":    {"↑↓ move", "^] box", "[ ] view", "enter open", "o design", "t task", "? help"},
			"Queue/list":    {"↑↓ move", "^] box", "[ ] view", "enter open", "s stops only", "t task", "? help"},
			"Tasks/list":    {"↑↓ move", "^] box", "[ ] view", "enter open", "b brief", "v verified", "d deps", "? help"},
			"details":       {"↑↓ scroll", "^] box", "esc back to list", "? help"},
			"Tasks/details": {"↑↓ scroll", "^] box", "esc back to list", "b brief", "v verified", "d deps", "? help"},
		},
		compact: map[string][]string{
			"list":          {"↑↓ move", "enter details", "[ ] view", "?"},
			"details":       {"↑↓ scroll", "esc back to list"},
			"Tasks/details": {"↑↓ scroll", "esc back to list", "b brief", "v verified", "d deps"},
		},
		panels: map[string]screenPanels{
			"Goals": {list: (*Model).goalsList, details: (*Model).goalDetails},
			"Queue": {list: (*Model).queueList, details: (*Model).queueDetails},
			"Tasks": {list: (*Model).workTasksList, details: (*Model).workTasksDetails, back: func(m *Model) bool { return m.tasks.closeFile() }},
		},
	},
	tabFriction: {
		label:       "Friction",
		views:       []string{"Open", "Resolved"},
		regions:     []string{"list", "details"},
		listDetails: true,
		hints: map[string][]string{
			"list":    {"↑↓ move", "^] box", "[ ] view", "enter open", "g/G top/end", "? help"},
			"details": {"↑↓ scroll", "^] box", "esc back to list", "? help"},
		},
		compact: map[string][]string{
			"list":    {"↑↓ move", "enter details", "[ ] view", "?"},
			"details": {"↑↓ scroll", "esc back to list"},
		},
		panels: map[string]screenPanels{
			"Open":     {list: (*Model).frictionList, details: (*Model).frictionDetails},
			"Resolved": {list: (*Model).frictionList, details: (*Model).frictionDetails},
		},
	},
	tabHealth: {
		label:   "Health",
		regions: []string{"grid", "table"},
		hints: map[string][]string{
			"grid":  {"↑↓ ←→ tile", "^] agents", "R check again"},
			"table": {"↑↓ row", "g/G top/end", "^] checks", "r restart", "a attach", "p peek", "R check again"},
		},
	},
	tabLogs: {
		label:       "Logs",
		regions:     []string{"events", "json"},
		listDetails: true,
		hints: map[string][]string{
			"events": {"↑↓ move", "^] box", "enter open", "f follow", "< > day", "/ filter", "g/G top/end", "? help"},
			"json":   {"↑↓ scroll", "^] box", "f follow", "< > day", "/ filter", "g/G top/end", "? help"},
		},
		compact: map[string][]string{
			"events": {"↑↓ move", "enter details", "f follow", "/ filter", "?"},
			"json":   {"↑↓ scroll", "esc back to events", "f follow"},
		},
	},
}

// The key hints of the focus states that each screen shares (decision 21).
var (
	helpHints    = []string{"↑↓ scroll", "? or esc close"}
	restartHints = []string{"y restart", "any other key cancels"}
	filterHints  = []string{"enter apply", "esc close"}
)

// keyHints gives the hints of focus in view. Below 120 columns it uses the
// compact hints when the screen has them.
func (s frameScreen) keyHints(view, focus string, wide bool) []string {
	tables := []map[string][]string{s.hints}
	if !wide {
		tables = []map[string][]string{s.compact, s.hints}
	}
	for _, table := range tables {
		for _, key := range []string{view + "/" + focus, focus} {
			if hints, ok := table[key]; ok {
				return hints
			}
		}
	}
	return nil
}

// activeTab gives the tab of the current screen.
func (m *Model) activeTab() int {
	switch m.screenIndex {
	case 0:
		return tabChat
	case 1:
		return tabAgents
	case 2, 3, 4:
		return tabWork
	case 5:
		return tabFriction
	case 6:
		return tabHealth
	}
	return tabLogs
}

// activeView gives the index of the view of the active tab.
func (m *Model) activeView() int {
	switch m.activeTab() {
	case tabWork:
		return m.screenIndex - 2
	case tabFriction:
		return int(m.friction.tab)
	}
	return 0
}

func (m *Model) viewName() string {
	if views := screenRegistry[m.activeTab()].views; len(views) > 0 {
		return views[m.activeView()]
	}
	return ""
}

// selectScreen shows tab in view. It keeps the current screen of each tab
// until wave 6 replaces it.
func (m *Model) selectScreen(tab, view int) {
	m.tab = tab
	switch tab {
	case tabAgents:
		m.screenIndex = 1
	case tabChat:
		m.screenIndex = 0
	case tabWork:
		m.screenIndex = 2 + view
	case tabFriction:
		m.screenIndex = 5
		if frictionTab(view) != m.friction.tab {
			m.friction.tab = frictionTab(view)
			m.friction.acceptSelection()
		}
	case tabHealth:
		m.screenIndex = 6
	default:
		m.screenIndex = 7
	}
	if len(screenRegistry[tab].views) > 0 {
		m.views[tab] = view
	}
}

// focusState gives the focus state of decision 21: a region of the screen,
// or "help", "restart", "filter", or the "details" popup of Agents.
func (m *Model) focusState() string {
	tab := m.activeTab()
	switch {
	case m.helpOpen:
		return "help"
	case m.agents.restartConfirm:
		return "restart"
	case tab == tabLogs && m.mode == editing:
		return "filter"
	case tab == tabAgents && m.popupOpen:
		return "details"
	case tab == tabChat:
		if m.chatInputActive {
			return "input"
		}
		if m.askFocused() {
			return m.askFocusState()
		}
		return "normal"
	}
	return screenRegistry[tab].regions[m.regionFocus[tab]]
}

// detailsFocused reports that the details region of a list and details
// screen has the focus. Below 120 columns it is the overlay.
func (m *Model) detailsFocused() bool {
	tab := m.activeTab()
	return screenRegistry[tab].listDetails && m.regionFocus[tab] == 1
}

// tooSmall gives the size to show and the size that hive needs, when the
// pane cannot hold the TUI (decisions 7, 36).
func (m *Model) tooSmall() (demoSize, demoSize, bool) {
	if m.demo != nil {
		return m.demoTooSmall()
	}
	if m.width < minimumWidth || m.height < minimumHeight {
		return demoSize{Width: m.width, Height: m.height}, demoSize{Width: minimumWidth, Height: minimumHeight}, true
	}
	return demoSize{}, demoSize{}, false
}

// frameView renders the screen at exactly the model size (golden-frames
// decisions 1 to 5): the title row and the rule on each screen except Chat,
// the content area, the notice line while a notice exists, and the status
// rows. A popup dims the full screen. Only the content area is cached.
// contentArea gives the first row and the rows of the content area of
// golden-frames decision 2: below the title row and the rule on each screen
// other than Chat, and above the notice line and the status rows.
func (m *Model) contentArea() (top, height int) {
	if m.activeTab() != tabChat {
		top = 2
	}
	height = m.height - top - len(m.statusRows(m.width, m.height))
	// On Chat the notice line is above the composer, inside the content
	// area (decision 14).
	if m.activeTab() != tabChat && m.noticeLine(m.width) != "" {
		height--
	}
	return top, max(1, height)
}

func (m *Model) frameView() string {
	width, height := m.width, m.height
	status := m.statusRows(width, height)
	notice := m.noticeLine(width)
	var rows []string
	if m.activeTab() != tabChat {
		title := component.TitleRow{Name: strings.ToLower(screenRegistry[m.activeTab()].label)}
		if m.activeTab() == tabHealth {
			title.Facts = m.healthFacts()
		}
		rows = append(rows, title.View(width))
	}
	_, contentHeight := m.contentArea()
	if !m.screenViewCacheValid || m.screenViewWidth != width || m.screenViewHeight != contentHeight || m.screenViewIndex != m.screenIndex {
		m.screenViewCache = m.contentView(width, contentHeight)
		m.screenViewCacheValid = true
		m.screenViewWidth = width
		m.screenViewHeight = contentHeight
		m.screenViewIndex = m.screenIndex
	}
	rows = append(rows, m.screenViewCache)
	if notice != "" && m.activeTab() != tabChat {
		rows = append(rows, notice)
	}
	screen := strings.Join(append(rows, status...), "\n")
	switch {
	case m.helpOpen:
		screen = component.Popup{Details: component.Details{Title: "keys · ? or esc closes", Lines: m.helpLines(), Offset: m.helpOffset}, Width: 84, Height: 29}.View(screen, width, height)
	case m.activeTab() == tabAgents && m.popupOpen:
		title, lines := m.agentDetailsPopup()
		m.popupOffset = min(m.popupOffset, max(len(lines)-agentPopupHeight+2, 0))
		screen = component.Popup{Details: component.Details{Title: title, Lines: lines, Offset: m.popupOffset}, Width: agentPopupWidth, Height: agentPopupHeight}.View(screen, width, height)
	}
	return screen
}

// homeLabel gives the home of the welcome block: the fixture home in demo
// mode, else the home with ~ for the user home folder.
func (m *Model) homeLabel() string {
	if m.demo != nil {
		return m.demo.fixture.Home
	}
	if user, err := os.UserHomeDir(); err == nil && user != "" && strings.HasPrefix(m.home, user) {
		return "~" + strings.TrimPrefix(m.home, user)
	}
	return m.home
}

// hintKey gives the key of the hints of focus. On the Agents graph it is
// pods for the overview and pod for the one-pod view of decision 62. With
// one project the one-pod view has no esc pods, so it keeps the graph hints.
func (m *Model) hintKey(tab int, focus string) string {
	if tab != tabAgents || focus != "graph" {
		return focus
	}
	pods := m.agentsPods()
	switch {
	case len(pods) == 0 || len(pods) == 1 && projectPods(pods) == 1:
		return focus
	case m.inPodView(pods):
		return "pod"
	}
	return "pods"
}

// focusHints gives the keys of the focus for the screen segment of
// golden-frames decision 11.
func (m *Model) focusHints(width int) []string {
	tab := m.activeTab()
	focus := m.focusState()
	var hints []string
	switch focus {
	case "help":
		hints = helpHints
	case "restart":
		hints = restartHints
	case "filter":
		hints = filterHints
	case "ask", "note", "other", "review":
		hints = m.askHints(focus)
	default:
		hints = screenRegistry[tab].keyHints(m.viewName(), m.hintKey(tab, focus), width >= component.LongFormWidth)
	}
	return hints
}

func (m *Model) footerNotice() string {
	switch {
	case m.quitNotice:
		return "press ctrl+c again to quit"
	case m.demo != nil && m.demo.notice != "":
		return m.demo.notice
	case m.agents.restartConfirm:
		return m.restartPrompt()
	case m.screenIndex == 6 && m.agents.restartNotice != "":
		return m.agents.restartNotice
	case m.startupNotice != "":
		return m.startupNotice
	case m.screenIndex == 2 && m.goalsQueue.designNotice != "":
		return m.goalsQueue.designNotice
	}
	if err := m.footerError(); err != "" {
		return "error: " + err
	}
	return ""
}

// restartPrompt gives the restart prompt of health.md rule 6.
func (m *Model) restartPrompt() string {
	target := m.agents.restartTarget.Role
	if target == "" {
		target = m.agents.restartTarget.Key
	}
	return "restart the " + target + "? y restarts · any other key cancels"
}

func (m *Model) footerError() string {
	errs := []string{m.startupErr, m.attachErr, m.headerErr, m.screenErr, m.logErr}
	if m.activeTab() == tabChat {
		errs = append(errs, m.chatErr)
	}
	for _, text := range errs {
		if text != "" {
			return text
		}
	}
	return ""
}

// contentView renders the content area of the active screen at exactly
// width×height cells.
func (m *Model) contentView(width, height int) string {
	tab := m.activeTab()
	entry := screenRegistry[tab]
	var rows []string
	bodyHeight := height
	overlay := entry.listDetails && m.detailsFocused() && width < component.LongFormWidth
	if len(entry.views) > 0 {
		subTabs := component.SubTabs{Views: entry.views, Active: m.activeView()}
		if overlay {
			subTabs.Hint = "list › details"
		}
		rows = append(rows, subTabs.View(width))
		bodyHeight--
	}
	focus := m.focusState()
	var body string
	switch {
	case tab == tabChat:
		body = fitBlock(m.currentScreen().View(width, bodyHeight), width, bodyHeight)
	case tab == tabAgents:
		body = m.agentsGraphView(width, bodyHeight, focus == "graph")
	case tab == tabLogs:
		body = m.logsContent(width, bodyHeight)
	case tab == tabHealth:
		body = m.healthContent(width, bodyHeight)
	case entry.listDetails && width >= component.LongFormWidth:
		listWidth := listRegionWidth(tab, width)
		list := strings.Split(m.legacyPanel(listWidth, bodyHeight, focus == entry.regions[0]), "\n")
		details := strings.Split(m.detailsPanel(width-listWidth, bodyHeight, focus == entry.regions[1], ""), "\n")
		for index := range list {
			list[index] += details[index]
		}
		body = strings.Join(list, "\n")
	case entry.listDetails && m.detailsFocused():
		body = m.detailsPanel(width, bodyHeight, focus == entry.regions[1], " · esc back")
	default:
		body = m.legacyPanel(width, bodyHeight, focus == entry.regions[0])
	}
	return strings.Join(append(rows, body), "\n")
}

// listRegionWidth gives the width of the list region of a list and details
// screen from 120 columns. The Logs event list has 8/15 of the width, and
// each other list has 13/30.
func listRegionWidth(tab, width int) int {
	if tab == tabLogs {
		return width * 8 / 15
	}
	return width * 13 / 30
}

// legacyPanel draws the current screen of the tab in one panel. The panel
// has the accent border when its region has the focus.
func (m *Model) legacyPanel(width, height int, focused bool) string {
	if panels, ok := screenRegistry[m.activeTab()].panels[m.viewName()]; ok {
		return panels.list(m, width, height, focused)
	}
	lines := strings.Split(m.currentScreen().View(width, height), "\n")
	return component.Panel{Title: strings.ToLower(m.legacyName()), Focused: focused, Lines: lines}.View(width, height)
}

// detailsPanel draws the details of the selected item of the current screen.
func (m *Model) detailsPanel(width, height int, focused bool, suffix string) string {
	tab := m.activeTab()
	title, lines := "details", []string(nil)
	if panels, ok := screenRegistry[tab].panels[m.viewName()]; ok {
		title, lines = panels.details(m, width)
	}
	body := max(height-2, 0)
	m.detailsOffset[tab] = min(max(m.detailsOffset[tab], 0), max(len(lines)-body, 0))
	return component.Details{Title: title + suffix, Focused: focused, Lines: lines, Offset: m.detailsOffset[tab]}.View(width, height)
}

func (m *Model) legacyName() string {
	return [...]string{"Chat", "Agents", "Goals", "Queue", "Tasks", "Friction", "Health", "Logs"}[m.screenIndex]
}

// helpLines gives the key map of the Help overlay. It keeps the layout of
// tui-v2 decision 21 and lists the keys of golden-frames decision 16 with
// the keys of the config (Help rule 3).
func (m *Model) helpLines() []string {
	key := func(text string, width int) string { return titleStyle.Render(padCell(text, width)) }
	pair := func(leftKey, left, rightKey, right string) string {
		line := key(leftKey, 11) + padCell(left, 30)
		if rightKey != "" {
			line += key(rightKey, 6) + right
		}
		return line
	}
	keys := m.tuiConfig.Keys
	return []string{
		"",
		titleStyle.Render("SCREENS · same on every screen") + strings.Repeat(" ", 11) + titleStyle.Render("ACTIONS · one meaning each"),
		"",
		pair(keys["agents"], "Agents", "a", "attach to pane"),
		pair(keys["work"], "Work", "p", "peek pane"),
		pair(keys["friction"], "Friction", "r", "restart · Health › Agents"),
		pair(keys["health"], "Health", "h", "show or hide stopped agents"),
		pair(keys["logs"], "Logs", "t", "open task"),
		pair("esc", "close, then back to Chat", "b", "task brief"),
		pair(keys["help"], "this help", "v", "VERIFIED.md"),
		pair("ctrl+c ×2", "quit", "d", "dependency graph"),
		pair("ctrl+]", "other box", "o", "goal design"),
		pair("[ ]", "previous / next view", "s", "stops only"),
		pair("↑↓ j k", "move", "f", "follow logs"),
		pair("←→", "move on the graph", "< >", "previous / next day"),
		pair("pgup pgdn", "page", "R", "refresh now"),
		pair("g G", "top / end", "", ""),
		pair("enter", "open / details", "", ""),
		pair("/", "filter bar", "", ""),
		"",
		titleStyle.Render("FILTER FIELDS"),
		"level:  comp:  proj:  agent:  task:  req:  text",
		dimStyle.Render("example  / level:warn agent:manager-hive restart"),
		"",
		titleStyle.Render("CHAT"),
		dimStyle.Render("the composer has the focus · enter sends · shift+enter adds a line"),
		dimStyle.Render("ctrl+o opens or folds all chat events"),
		dimStyle.Render("ctrl+r moves the focus between input and ask card · esc to input"),
		dimStyle.Render("ask card: ↑↓ move · 1–9 pick · space mark · n note · x dismiss · J/K card"),
		dimStyle.Render("ask review: enter on the last question · ↑↓ question · enter send or change"),
	}
}

// padCell pads text with spaces to width cells.
func padCell(text string, width int) string {
	return text + strings.Repeat(" ", max(width-ansi.StringWidth(text), 0))
}

// fitBlock cuts or pads text to exactly width×height cells.
func fitBlock(text string, width, height int) string {
	lines := strings.Split(text, "\n")
	out := make([]string, height)
	for index := range out {
		line := ""
		if index < len(lines) {
			line = ansi.Truncate(lines[index], width, "")
		}
		out[index] = line + strings.Repeat(" ", max(width-ansi.StringWidth(line), 0))
	}
	return strings.Join(out, "\n")
}

// updateFrameKey routes one key press by the focus state (decisions 21 to
// 24). It gives the key to the current screen when the frame does not take
// it.
func (m *Model) updateFrameKey(press tea.KeyPressMsg) tea.Cmd {
	key := press.String()
	tab := m.activeTab()
	entry := screenRegistry[tab]
	switch m.focusState() {
	case "help":
		switch key {
		case m.tuiConfig.Keys["help"], "esc":
			m.helpOpen = false
			return m.logAction("help close")
		case "up":
			m.helpOffset = max(m.helpOffset-1, 0)
		case "down":
			m.helpOffset = min(m.helpOffset+1, len(m.helpLines()))
		}
		return nil
	case "restart":
		return m.screens[1].Update(press)
	case "filter":
		return m.currentScreen().Update(press)
	case "input":
		switch key {
		case "tab", "esc", "ctrl+]":
			return nil
		case "ctrl+r":
			return m.focusAskCard()
		case m.tuiConfig.Keys["help"]:
			if m.chatInput.Value() == "" {
				return m.openHelp()
			}
		case "pgup", "pgdown":
			return m.scrollChatFeed(press)
		}
		return m.currentScreen().Update(press)
	case "ask", "note", "other", "review":
		if cmd, handled := m.updateAskKey(press); handled {
			return cmd
		}
	case "details":
		if tab == tabAgents {
			switch key {
			case "esc":
				m.popupOpen = false
				return m.logAction("agent details close")
			case "up", "k", "pgup":
				m.popupOffset = max(m.popupOffset-1, 0)
			case "down", "j", "pgdown":
				m.popupOffset++
			case "a", "p", "t":
				if node, _, ok := m.selectedAgent(); ok {
					return m.agentAction(key, node)
				}
			}
			return nil
		}
	}
	if entry.listDetails && m.detailsFocused() && m.width < component.LongFormWidth {
		// Decision 21: the overlay takes its scroll keys and its action
		// keys, and esc closes it. It drops each other key.
		switch key {
		case "esc", "up", "k", "down", "j", "pgup", "pgdown":
		case "<", ">":
			if tab != tabLogs {
				return nil
			}
		default:
			if len(key) != 1 || !strings.Contains(actionKeys, key) {
				return nil
			}
		}
	}
	switch key {
	case m.tuiConfig.Keys["help"]:
		return m.openHelp()
	case "R":
		if tab == tabHealth {
			// health.md rule 7: the screen shows checking until the result.
			m.health.checking = true
		}
		return batch(m.logAction("refresh"), m.startHeaderLoad(), m.startCurrentScreenLoad())
	case "tab":
		// Golden-frames decision 19: there is no tab region move.
		return nil
	case "ctrl+]":
		return m.nextRegion()
	case "[", "]":
		if len(entry.views) == 0 {
			return nil
		}
		step := 1
		if key == "[" {
			step = len(entry.views) - 1
		}
		view := (m.activeView() + step) % len(entry.views)
		m.selectScreen(tab, view)
		return batch(m.logAction("view select", slogString("view", entry.views[view])), m.startCurrentScreenLoad())
	}
	if key == "esc" && !m.escCloses() {
		// Golden-frames decision 17: with no popup open, esc returns to Chat.
		return m.openChat()
	}
	if tab == tabAgents {
		if m.agents.peekOpen {
			return m.currentScreen().Update(press)
		}
		// Decision 19: restart is in Health › Agents, so r does nothing.
		return m.updateAgentsKey(key)
	}
	if tab == tabLogs && (key == "<" || key == ">") {
		day := "["
		if key == ">" {
			day = "]"
		}
		return m.currentScreen().Update(tea.KeyPressMsg{Code: rune(day[0]), Text: day})
	}
	if entry.listDetails && m.detailsFocused() {
		switch key {
		case "esc":
			if panels, ok := entry.panels[m.viewName()]; ok && panels.back != nil && panels.back(m) {
				return m.logAction("details back")
			}
			m.regionFocus[tab] = 0
			return m.logAction("focus region", slogString("region", entry.regions[0]))
		case "up", "k":
			m.detailsOffset[tab]--
		case "down", "j":
			m.detailsOffset[tab]++
		case "pgup":
			m.detailsOffset[tab] -= max(m.height-frameRows-3, 1)
		case "pgdown":
			m.detailsOffset[tab] += max(m.height-frameRows-3, 1)
		default:
			return m.currentScreen().Update(press)
		}
		return nil
	}
	if entry.listDetails && key == "enter" && m.width < component.LongFormWidth {
		m.regionFocus[tab] = 1
		m.detailsOffset[tab] = 0
		return m.logAction("details open", slogString("region", entry.regions[1]))
	}
	return m.currentScreen().Update(press)
}

// openTab shows tab in its last view.
func (m *Model) openTab(tab int) tea.Cmd {
	view := m.views[tab]
	if tab == m.activeTab() {
		view = m.activeView()
	}
	m.selectScreen(tab, view)
	return batch(m.logAction("screen select", slogString("screen", screenRegistry[tab].label)), m.startCurrentScreenLoad())
}

// openHelp opens the Help overlay.
func (m *Model) openHelp() tea.Cmd {
	m.helpOpen = true
	m.helpOffset = 0
	return m.logAction("help open")
}

// nextRegion moves the focus to the other box of the screen (golden-frames
// decision 6). Chat has no second box, so ctrl+] does nothing there.
func (m *Model) nextRegion() tea.Cmd {
	tab := m.activeTab()
	if tab == tabChat {
		return nil
	}
	regions := screenRegistry[tab].regions
	m.regionFocus[tab] = (m.regionFocus[tab] + 1) % len(regions)
	return m.logAction("focus region", slogString("region", regions[m.regionFocus[tab]]))
}

// actionKeys are the action keys of decision 23 that a details overlay
// gives to its screen.
const actionKeys = "aprhtbvdosf"

// startChatInput gives the input line the focus (chat.md rule 17). The
// text stays.
func (m *Model) startChatInput() tea.Cmd {
	if m.chatCallerBlocked || m.chatInputActive {
		return nil
	}
	m.chatInputActive = true
	m.invalidateView()
	return batch(m.chatInput.Focus(), m.logAction("chat focus", slogString("focus", "input")))
}

// focusAskCard gives the ask card the focus (chat.md rule 17). The text of
// the input line stays.
func (m *Model) focusAskCard() tea.Cmd {
	if _, _, ok := m.focusedAsk(); !ok || !m.chatInputActive {
		return nil
	}
	m.chatInputActive = false
	m.chatInput.Blur()
	m.invalidateView()
	return m.logAction("chat focus", slogString("focus", "ask"))
}
