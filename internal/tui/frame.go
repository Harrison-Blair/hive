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
			"feed":  {"i or type to write", "j/k scroll", "g/G top/end", "ctrl+o open events", "tab region", "? help"},
			"input": {"enter send", "shift+enter newline", "ctrl+o open events", "esc leave input", "pgup/pgdn scroll feed", "tab region"},
		},
		compact: map[string][]string{
			"feed":  {"i type", "pgup/pgdn feed", "ctrl+o events", "tab region", "?"},
			"input": {"enter send", "ctrl+o events", "esc leave input", "pgup/pgdn feed"},
		},
	},
	tabWork: {
		label:       "Work",
		views:       []string{"Goals", "Queue", "Tasks"},
		regions:     []string{"list", "details"},
		listDetails: true,
		hints: map[string][]string{
			"Goals/list":    {"↑↓ move", "tab region", "[ ] view", "enter open", "o design", "t task", "? help"},
			"Queue/list":    {"↑↓ move", "tab region", "[ ] view", "enter open", "s stops only", "t task", "? help"},
			"Tasks/list":    {"↑↓ move", "tab region", "[ ] view", "enter open", "b brief", "v verified", "d deps", "? help"},
			"details":       {"↑↓ scroll", "tab region", "esc back to list", "? help"},
			"Tasks/details": {"↑↓ scroll", "tab region", "esc back to list", "b brief", "v verified", "d deps", "? help"},
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
			"list":    {"↑↓ move", "tab region", "[ ] view", "enter open", "g/G top/end", "? help"},
			"details": {"↑↓ scroll", "tab region", "esc back to list", "? help"},
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
		label:       "Health",
		views:       []string{"Home", "Agents"},
		regions:     []string{"list", "details"},
		listDetails: true,
		hints: map[string][]string{
			"Home/list":   {"↑↓ move", "tab region", "[ ] view", "enter open", "R refresh", "? help"},
			"Agents/list": {"↑↓ move", "tab region", "[ ] view", "enter open", "r restart", "R refresh", "? help"},
			"details":     {"↑↓ scroll", "tab region", "esc back to list", "? help"},
		},
		compact: map[string][]string{
			"Home/list":   {"↑↓ move", "enter details", "[ ] view", "R refresh", "?"},
			"Agents/list": {"↑↓ move", "enter details", "[ ] view", "r restart", "R refresh", "?"},
			"details":     {"↑↓ scroll", "esc back to list"},
		},
		panels: map[string]screenPanels{
			"Home":   {list: (*Model).healthPanel, details: (*Model).healthDetails},
			"Agents": {list: (*Model).healthPanel, details: (*Model).healthDetails},
		},
	},
	tabLogs: {
		label:       "Logs",
		regions:     []string{"events", "json"},
		listDetails: true,
		hints: map[string][]string{
			"events": {"↑↓ move", "tab region", "enter open", "f follow", "< > day", "/ filter", "g/G top/end", "? help"},
			"json":   {"↑↓ scroll", "tab region", "f follow", "< > day", "/ filter", "g/G top/end", "? help"},
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
	case tabHealth:
		return m.views[tabHealth]
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
		return "feed"
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

// frameView renders the header, the tab bar, the content area, and the
// footer at exactly the model size. Only the content area is cached.
func (m *Model) frameView() string {
	width, height := m.width, m.height
	contentHeight := max(1, height-frameRows)
	if !m.screenViewCacheValid || m.screenViewWidth != width || m.screenViewHeight != contentHeight || m.screenViewIndex != m.screenIndex {
		m.screenViewCache = m.contentView(width, contentHeight)
		m.screenViewCacheValid = true
		m.screenViewWidth = width
		m.screenViewHeight = contentHeight
		m.screenViewIndex = m.screenIndex
	}
	top := m.headerView().View(width) + "\n" + component.TabBar{Active: m.activeTab()}.View(width)
	if m.helpOpen || (m.activeTab() == tabAgents && m.popupOpen) {
		top = component.DimScreen(top, width, 3)
	}
	return top + "\n" + m.screenViewCache + "\n" + m.footerView(width)
}

// headerView gives the header of decision 2 from the header state.
func (m *Model) headerView() component.Header {
	home := m.home
	if m.demo != nil {
		home = m.demo.fixture.Home
	} else if user, err := os.UserHomeDir(); err == nil && user != "" && strings.HasPrefix(home, user) {
		home = "~" + strings.TrimPrefix(home, user)
	}
	supervisor := "stopped"
	switch {
	case m.header.supervisorOld:
		supervisor = "old"
	case m.header.supervisor == "fresh":
		supervisor = "live"
	}
	coordinator := m.header.coordinator
	if coordinator == "" {
		coordinator = "stopped"
	}
	return component.Header{
		Home:        home,
		Coordinator: coordinator,
		Supervisor:  supervisor,
		NeedsYou:    m.header.needsUser,
		Clock:       m.now().Format("15:04"),
	}
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

// footerView gives the footer of decision 4: the focus state and its key
// hints. A notice takes the place of the hints.
func (m *Model) footerView(width int) string {
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
	default:
		hints = screenRegistry[tab].keyHints(m.viewName(), m.hintKey(tab, focus), width >= component.LongFormWidth)
		if tab == tabChat && width >= component.LongFormWidth && (focus == "feed" || focus == "input") {
			toggleHint := "ctrl+o open events"
			if m.chatEventsOpen {
				toggleHint = "ctrl+o fold events"
			}
			for i, hint := range hints {
				if strings.HasPrefix(hint, "ctrl+o ") {
					hints[i] = toggleHint
				}
			}
		}
	}
	if notice := m.footerNotice(); notice != "" {
		hints = []string{notice}
	}
	return component.KeyHints{Region: focus, Hints: hints}.View(width)
}

func (m *Model) footerNotice() string {
	switch {
	case m.quitNotice:
		return "press ctrl+c again to quit"
	case m.demo != nil && m.demo.notice != "":
		return m.demo.notice
	case m.agents.restartConfirm:
		return ""
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
	case entry.listDetails && width >= component.LongFormWidth:
		listWidth := width * 13 / 30
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
	content := strings.Join(append(rows, body), "\n")
	switch {
	case m.helpOpen:
		content = component.Popup{Details: component.Details{Title: "keys · ? or esc closes", Lines: helpLines(), Offset: m.helpOffset}, Width: 84, Height: 29}.View(content, width, height)
	case tab == tabAgents && m.popupOpen:
		title, lines := m.agentDetailsPopup()
		m.popupOffset = min(m.popupOffset, max(len(lines)-agentPopupHeight+2, 0))
		content = component.Popup{Details: component.Details{Title: title, Lines: lines, Offset: m.popupOffset}, Width: agentPopupWidth, Height: agentPopupHeight}.View(content, width, height)
	}
	return content
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

// helpLines gives the key map of the Help board (decision 21).
func helpLines() []string {
	key := func(text string, width int) string { return titleStyle.Render(padCell(text, width)) }
	pair := func(leftKey, left, rightKey, right string) string {
		line := key(leftKey, 11) + padCell(left, 30)
		if rightKey != "" {
			line += key(rightKey, 6) + right
		}
		return line
	}
	return []string{
		"",
		titleStyle.Render("GLOBAL · same on every screen") + strings.Repeat(" ", 12) + titleStyle.Render("ACTIONS · one meaning each"),
		"",
		pair("1–6", "switch screen", "a", "attach to pane"),
		pair("tab", "next region", "p", "peek pane"),
		pair("[ ]", "previous / next view", "r", "restart · Health › Agents"),
		pair("↑↓ j k", "move", "h", "show or hide stopped agents"),
		pair("←→", "move on the graph", "t", "open task"),
		pair("pgup pgdn", "page", "b", "task brief"),
		pair("g G", "top / end", "v", "VERIFIED.md"),
		pair("enter", "open / details", "d", "dependency graph"),
		pair("esc", "back / close", "o", "goal design"),
		pair("/", "filter bar", "s", "stops only"),
		pair("?", "this help", "f", "follow logs"),
		pair("R", "refresh now", "< >", "previous / next day"),
		pair("ctrl+c ×2", "quit", "", ""),
		"",
		"",
		titleStyle.Render("FILTER FIELDS"),
		"level:  comp:  proj:  agent:  task:  req:  text",
		dimStyle.Render("example  / level:warn agent:manager-hive restart"),
		"",
		titleStyle.Render("CHAT"),
		dimStyle.Render("i or enter types in chat · esc leaves the input · 1–6 work again"),
		dimStyle.Render("ctrl+o opens or folds all chat events"),
		dimStyle.Render("r restart: coordinator or supervisor, asks first only when live"),
		"",
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
		case "?", "esc":
			m.helpOpen = false
			return m.logAction("help close")
		case "up":
			m.helpOffset = max(m.helpOffset-1, 0)
		case "down":
			m.helpOffset = min(m.helpOffset+1, len(helpLines()))
		}
		return nil
	case "restart":
		return m.screens[1].Update(press)
	case "filter":
		return m.currentScreen().Update(press)
	case "input":
		switch key {
		case "tab":
			m.chatInputActive = false
			m.chatInput.Blur()
			return m.logAction("focus region", slogString("region", "feed"))
		case "pgup", "pgdown":
			return m.scrollChatFeed(press)
		}
		return m.currentScreen().Update(press)
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
	if index, ok := screenKey(key); ok {
		view := m.views[index]
		if index == tab {
			view = m.activeView()
		}
		m.selectScreen(index, view)
		return batch(m.logAction("screen select", slogString("screen", screenRegistry[index].label)), m.startCurrentScreenLoad())
	}
	switch key {
	case "?":
		m.helpOpen = true
		m.helpOffset = 0
		return m.logAction("help open")
	case "R":
		return batch(m.logAction("refresh"), m.startHeaderLoad(), m.startCurrentScreenLoad())
	case "tab":
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
	if tab == tabChat {
		return m.updateChatFeedKey(press)
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

// screenKey gives the tab of a key 1 to 6.
func screenKey(key string) (int, bool) {
	if len(key) == 1 && key[0] >= '1' && key[0] < '1'+tabCount {
		return int(key[0] - '1'), true
	}
	return 0, false
}

// nextRegion moves the focus to the next region of the screen (decision 6).
func (m *Model) nextRegion() tea.Cmd {
	tab := m.activeTab()
	if tab == tabChat {
		if m.chatCallerBlocked {
			return nil
		}
		m.chatInputActive = true
		return batch(m.chatInput.Focus(), m.logAction("focus region", slogString("region", "input")))
	}
	regions := screenRegistry[tab].regions
	m.regionFocus[tab] = (m.regionFocus[tab] + 1) % len(regions)
	return m.logAction("focus region", slogString("region", regions[m.regionFocus[tab]]))
}

// actionKeys are the action keys of decision 23 that a details overlay
// gives to its screen.
const actionKeys = "aprhtbvdosf"

// chatFeedGlobals are the keys that stay global when the chat feed has the
// focus (decision 24).
const chatFeedGlobals = "?/jkgGRf"

// updateChatFeedKey handles a key in the chat feed. A printable key gives the
// focus to the input and types the key. i and enter give the focus to the
// input.
func (m *Model) updateChatFeedKey(press tea.KeyPressMsg) tea.Cmd {
	key := press.String()
	switch {
	case key == "pgup" || key == "pgdown":
		return m.scrollChatFeed(press)
	case key == "i" || key == "enter":
		return m.nextRegion()
	case press.Text != "" && len([]rune(key)) == 1 && !strings.Contains(chatFeedGlobals, key):
		if m.chatCallerBlocked {
			return nil
		}
		cmd := m.nextRegion()
		updated, typed := m.chatInput.Update(press)
		m.chatInput = updated
		return batch(cmd, typed)
	}
	return m.currentScreen().Update(press)
}
