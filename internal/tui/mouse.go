package tui

import (
	"log/slog"
	"sort"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/Harrison-Blair/hive/internal/tui/component"
	"github.com/charmbracelet/x/ansi"
)

// wheelLines is the lines that one wheel step scrolls a text region.
const wheelLines = 3

// updateMouse handles one click or one wheel step. The wheel scrolls the
// region under the pointer. A click on a region gives it the focus, and a
// click on a list row selects the row. Golden-frames decision 2 removes the
// tab bar, so no click opens a screen. A
// mouse action that changes the screen is logged.
func (m *Model) updateMouse(msg tea.MouseMsg) tea.Cmd {
	if _, _, small := m.tooSmall(); small || m.agents.restartConfirm {
		return nil
	}
	mouse := msg.Mouse()
	before := m.frameView()
	var action string
	var cmd tea.Cmd
	switch msg.(type) {
	case tea.MouseWheelMsg:
		step := 0
		switch mouse.Button {
		case tea.MouseWheelUp:
			step = -1
		case tea.MouseWheelDown:
			step = 1
		}
		if step == 0 {
			return nil
		}
		action, cmd = "mouse wheel", m.wheel(mouse.X, mouse.Y, step)
	case tea.MouseClickMsg:
		if mouse.Button != tea.MouseLeft {
			return nil
		}
		action, cmd = "mouse click", m.click(mouse.X, mouse.Y)
	}
	m.invalidateView()
	if m.frameView() == before {
		return cmd
	}
	return batch(m.logAction(action, slog.Int("x", mouse.X), slog.Int("y", mouse.Y), slogString("focus", m.focusState())), cmd)
}

// contentRow gives the row of the content area at screen row y.
func (m *Model) contentRow(y int) (int, bool) {
	top, height := m.contentArea()
	row := y - top
	return row, row >= 0 && row < height
}

// wheel scrolls the region at x, y by step.
func (m *Model) wheel(x, y, step int) tea.Cmd {
	row, ok := m.contentRow(y)
	if !ok {
		return nil
	}
	tab := m.activeTab()
	switch {
	case m.helpOpen:
		m.helpOffset = min(max(m.helpOffset+step, 0), len(m.helpLines()))
		return nil
	case tab == tabAgents && m.popupOpen:
		m.popupOffset = max(m.popupOffset+step, 0)
		return nil
	case tab == tabChat:
		// chat.md rule 7: the wheel scrolls the transcript. The card has no
		// wheel, so the wheel over the composer scrolls the transcript too.
		m.scrollChat(step * wheelLines)
		return nil
	case tab == tabHealth:
		return m.healthWheel(row, step)
	case !screenRegistry[tab].listDetails:
		// Decision 77: the wheel does not move the Agents graph.
		return nil
	}
	region := m.regionAt(tab, x)
	press := tea.KeyPressMsg{Code: tea.KeyDown}
	if step < 0 {
		press = tea.KeyPressMsg{Code: tea.KeyUp}
	}
	repeat := 1
	if region == 1 {
		repeat = wheelLines
	}
	// The scroll keys go to the region, also while the filter is open.
	saved, savedMode := m.regionFocus[tab], m.mode
	m.regionFocus[tab], m.mode = region, viewing
	var cmds []tea.Cmd
	for range repeat {
		cmds = append(cmds, m.updateFrameKey(press))
	}
	m.regionFocus[tab], m.mode = saved, savedMode
	return batch(cmds...)
}

// click handles a left click at x, y.
func (m *Model) click(x, y int) tea.Cmd {
	row, ok := m.contentRow(y)
	tab := m.activeTab()
	if !ok || m.helpOpen || tab == tabAgents || m.focusState() == "filter" {
		return nil
	}
	if tab == tabChat {
		top := m.chatFeed.composerTop
		switch {
		case row < m.chatFeed.height:
			return m.clickTranscript(row)
		case row > top && row <= top+len(m.askRows):
			return m.clickAsk(row - top - 1)
		case row >= top:
			return m.startChatInput()
		}
		return nil
	}
	if tab == tabHealth {
		return m.healthClick(row)
	}
	entry := screenRegistry[tab]
	if !entry.listDetails || (m.detailsFocused() && m.width < component.LongFormWidth) {
		return nil
	}
	region := m.regionAt(tab, x)
	var cmd tea.Cmd
	if m.regionFocus[tab] != region {
		m.regionFocus[tab] = region
		cmd = m.logAction("focus region", slogString("region", entry.regions[region]))
	}
	if region == 0 {
		return batch(cmd, m.selectListRow(row))
	}
	return cmd
}

// regionAt gives the region of a list and details screen at cell x: 0 for
// the list and 1 for the details. Below 120 columns the details are an
// overlay, so the region is the one with the focus.
func (m *Model) regionAt(tab, x int) int {
	if m.width < component.LongFormWidth {
		return m.regionFocus[tab]
	}
	if x < listRegionWidth(tab, m.width) {
		return 0
	}
	return 1
}

// selectListRow moves the list selection to the row of the content area.
// The selected row of each list has the marker ▶ at cell 2. The selection
// moves one row at a time until the marker is on the row, or until it stops.
// A row outside the panel of the list changes nothing.
func (m *Model) selectListRow(row int) tea.Cmd {
	marker := func() []string {
		m.invalidateView()
		_, height := m.contentArea()
		return strings.Split(ansi.Strip(m.contentView(m.width, height)), "\n")
	}
	rows := marker()
	selected := markerRow(rows)
	if selected < 0 || !samePanel(rows, selected, row) {
		return nil
	}
	var cmds []tea.Cmd
	for selected != row {
		press, toward := tea.KeyPressMsg{Code: tea.KeyDown}, 1
		if row < selected {
			press, toward = tea.KeyPressMsg{Code: tea.KeyUp}, -1
		}
		cmds = append(cmds, m.updateFrameKey(press))
		next := markerRow(marker())
		if next == selected || (next-row)*toward > 0 {
			break
		}
		selected = next
	}
	return batch(cmds...)
}

// markerRow gives the row with the list marker ▶ at cell 2, or -1.
func markerRow(rows []string) int {
	for index, line := range rows {
		if ansi.Strip(ansi.Cut(line, 2, 3)) == "▶" {
			return index
		}
	}
	return -1
}

// samePanel reports that row is a body row of the panel that holds the row
// at from. Cell 0 of a body row is a side border.
func samePanel(rows []string, from, row int) bool {
	if row < 0 || row >= len(rows) {
		return false
	}
	low, high := min(from, row), max(from, row)
	for index := low; index <= high; index++ {
		if first := ansi.Cut(rows[index], 0, 1); first != "│" && first != "┃" {
			return false
		}
	}
	return true
}

// clickTranscript opens or folds the event at row of the transcript
// (chat.md rule 3).
func (m *Model) clickTranscript(row int) tea.Cmd {
	feed := &m.chatFeed
	line := row
	if len(feed.lines) > feed.height {
		line = feed.top + row - 1
	}
	if line < 0 || line >= len(feed.lines) || len(feed.starts) == 0 {
		return nil
	}
	index := sort.SearchInts(feed.starts, line+1) - 1
	if index < 0 || index >= len(feed.msgs) || feed.msgs[index].Kind != component.ChatEvent {
		return nil
	}
	if m.chatEventOpen == nil {
		m.chatEventOpen = map[uint64]bool{}
	}
	open := !feed.msgs[index].Open
	m.chatEventOpen[feed.seqs[index]] = open
	m.rebuildChat()
	m.invalidateView()
	return m.logAction("chat event toggle", slog.Bool("open", open))
}
