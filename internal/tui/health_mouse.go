package tui

import (
	"slices"

	tea "charm.land/bubbletea/v2"
)

// healthLayout gives the rows of the Health body below the check grid: the
// details box, the agents table, and the note under the table.
type healthLayout struct {
	details int
	table   int
	tableAt int
	note    bool
}

// healthLayout gives the layout at height content rows, detailLines body
// lines, and tableRows agent rows, as healthContent draws it.
func (m *Model) healthLayout(height, detailLines, tableRows int) healthLayout {
	rest := height - len(healthAreas) - 1
	tableNeed := tableRows + 3
	details := min(detailLines+2, max(3, rest-2-min(tableNeed, 5)), rest)
	room := rest - details - 1
	note := room-1 >= min(tableNeed, 4)
	if note {
		room--
	}
	table := min(tableNeed, room)
	if table < 3 {
		table = 0
	}
	tableAt := len(healthAreas) + 1
	if details >= 3 {
		tableAt += details + 1
	}
	return healthLayout{details: details, table: table, tableAt: tableAt, note: note}
}

// healthPage gives the rows of one page of the agents table: the rows that
// its box shows.
func (m *Model) healthPage() int {
	_, height := m.contentArea()
	_, lines := m.healthDetails(m.width)
	return max(m.healthLayout(height, len(lines), len(m.healthTableRows())).table-3, 1)
}

// healthTableRowIndex gives the agent index at content row row of the
// agents table, or -1 when the row is outside the table. A row above the
// first agent selects the first visible row, and a row below the last agent
// selects the last row that fits.
func (m *Model) healthTableRowIndex(row int) int {
	_, height := m.contentArea()
	_, lines := m.healthDetails(m.width)
	rows := m.healthTableRows()
	layout := m.healthLayout(height, len(lines), len(rows))
	if layout.table < 3 || row < layout.tableAt || row >= layout.tableAt+layout.table {
		return -1
	}
	nodes := m.health.agentList()
	selected := max(slices.IndexFunc(nodes, func(node agentNode) bool { return node.record.Key == m.health.selectedAgent }), 0)
	visible := max(layout.table-3, 1)
	offset := m.health.agentOffset
	if visible < len(rows) {
		offset = min(offset, selected)
		offset = max(offset, selected-visible+1)
		offset = min(max(offset, 0), len(rows)-visible)
	} else {
		offset = 0
	}
	index := offset + row - layout.tableAt - 2
	return min(max(index, offset), min(offset+visible, len(rows))-1)
}

// healthRegionAt gives the focus region of the Health screen at content row
// row: 0 for the grid, 1 for the agents table, or -1 outside both.
func (m *Model) healthRegionAt(row int) int {
	if row >= 0 && row < len(healthAreas) {
		return 0
	}
	if m.healthTableRowIndex(row) >= 0 {
		return 1
	}
	return -1
}

// healthWheel moves the agents table selection by step when the pointer is
// over it. The wheel does not change the focus (decisions 6 and 80).
func (m *Model) healthWheel(row, step int) tea.Cmd {
	if m.healthRegionAt(row) != 1 {
		return nil
	}
	nodes := m.health.agentList()
	index := slices.IndexFunc(nodes, func(node agentNode) bool { return node.record.Key == m.health.selectedAgent })
	return m.selectHealthAgent(index + step)
}

// healthClick gives the focus to the Health box at content row row, and a
// click on the agents table selects the row under the pointer
// (decisions 6 and 81).
func (m *Model) healthClick(row int) tea.Cmd {
	region := -1
	index := -1
	if row >= 0 && row < len(healthAreas) {
		region = 0
	} else if index = m.healthTableRowIndex(row); index >= 0 {
		region = 1
	}
	if region < 0 {
		return nil
	}
	var cmd tea.Cmd
	if m.regionFocus[tabHealth] != region {
		m.regionFocus[tabHealth] = region
		cmd = m.logAction("focus region", slogString("region", screenRegistry[tabHealth].regions[region]))
	}
	if region == 1 {
		cmd = batch(cmd, m.selectHealthAgent(index))
	}
	return cmd
}
