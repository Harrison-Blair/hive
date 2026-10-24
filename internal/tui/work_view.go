package tui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/Harrison-Blair/hive/internal/tui/component"
)

// workLegend is the legend of the Work layout row 38, with the × cancelled
// entry of the boards.
var workLegend = []struct{ glyph, text, color string }{
	{"●", " landed  ", colorDone},
	{"◐", " verified  ", colorWorking},
	{"○", " assigned or created  ", colorDim},
	{"!", " needs you  ", colorAccent},
	{"×", " cancelled  ", colorDim},
	{"─▶", " depends on", colorDim},
}

// workScrollTrack is the first column of the track of the scroll row of
// rule 6.
const workScrollTrack = 44

// titleRow gives the title row and the row below it of the active screen.
func (m *Model) titleRow() component.TitleRow {
	if m.activeTab() == tabWork {
		return m.workTitleRow()
	}
	return component.TitleRow{Name: strings.ToLower(screenRegistry[m.activeTab()].label)}
}

// workTitleRow gives the title row and the wave line of the Work board
// (work.md layout, rules 1, 2, and 11). With no goal and no task, row 1 is
// the rule.
func (m *Model) workTitleRow() component.TitleRow {
	s := &m.tasks
	cycle := s.cycle()
	index := s.goalIndex()
	if index < 0 {
		name := "work"
		if s.data.project != "" {
			name += " · " + s.data.project
		}
		return component.TitleRow{Name: name, Facts: "· no goals"}
	}
	row := component.TitleRow{Right: fmt.Sprintf("[ ] goal %d of %d", index+1, len(cycle))}
	if s.stopsOnly {
		row.Note = "· stops only"
	}
	var line string
	var onScreen []taskItem
	goal := cycle[index].goal
	if goal == nil {
		row.Name, row.Facts = "work · "+workNoGoal, "· tasks outside every goal"
		stops := 0
		for _, item := range s.data.items {
			if item.goal == "" && item.record.Status != "cancelled" {
				onScreen = append(onScreen, item)
				if s.isStop(item) {
					stops++
				}
			}
		}
		line = plural(len(onScreen), "task") + " · " + plural(stops, "stop") + " · no grant"
	} else {
		row.Name, row.Facts = "work · "+goal.ID, "· "+goal.Title
		landed, total, needs := 0, 0, 0
		current := ""
		for _, column := range s.columns() {
			if !column.done && current == "" {
				current = column.key
			}
		}
		for _, wave := range goal.Waves {
			for _, id := range wave.Tasks {
				item, ok := s.data.item(id)
				if !ok || item.record.Status == "cancelled" {
					continue
				}
				onScreen = append(onScreen, item)
				total++
				if taskLanded(item.record) {
					landed++
				}
				if item.record.NeedsUser != "" {
					needs++
				}
			}
		}
		var parts []string
		if len(goal.Waves) > 0 {
			if current == "" {
				current = goal.Waves[len(goal.Waves)-1].ID
			}
			parts = append(parts, fmt.Sprintf("wave %s of %d", current, len(goal.Waves)))
		}
		grant := "no grant"
		if goal.Grant != "" {
			grant = "grant " + goal.Grant
		}
		parts = append(parts, fmt.Sprintf("%d/%d landed", landed, total), fmt.Sprintf("%d needs you", needs), grant)
		line = strings.Join(parts, " · ")
	}
	row.Line = dimStyle.Render(line)
	if elsewhere := s.stopsElsewhere(onScreen); elsewhere > 0 {
		row.Line += "  " + accentStyle.Render(fmt.Sprintf("· %d stops elsewhere", elsewhere))
	}
	return row
}

// stopsElsewhere counts the stops of rule 2 that are not on the screen.
func (s *tasksState) stopsElsewhere(onScreen []taskItem) int {
	shown := make(map[string]bool, len(onScreen))
	for _, item := range onScreen {
		shown[item.record.ID] = true
	}
	count := 0
	for _, item := range s.data.items {
		if !shown[item.record.ID] && s.isStop(item) {
			count++
		}
	}
	return count
}

// workHints gives the keys of the screen segment of the Work board, the
// task popup, and the dependency popup.
func (m *Model) workHints(focus string) []string {
	switch focus {
	case workPopupTask:
		return []string{"b brief", "v verified", "d deps", "↑↓ scroll", "esc close"}
	case workPopupDeps:
		return []string{"↑↓ move", "enter task", "esc close"}
	}
	stops, cancelled := "s stops", "h cancelled"
	if m.tasks.stopsOnly {
		stops = "s all"
	}
	if m.tasks.showCancelled {
		cancelled = "h hide cancelled"
	}
	return []string{"←→ ↑↓ card", "[ ] goal", "enter task", "d deps", stops, cancelled}
}

// workBoardView renders the content area of the Work board at exactly
// width×height cells: a blank row, the columns, the scroll row of rule 6
// when the columns do not fit, a blank row, and the legend.
func (m *Model) workBoardView(width, height int) string {
	s := &m.tasks
	columns := s.columns()
	scroll := workScrollRow(columns, width)
	columnHeight := workColumnHeight(height, scroll)
	rows := make([]string, 0, height)
	rows = append(rows, "")
	if s.goalIndex() < 0 {
		body := make([]string, columnHeight)
		notice := "no goals · hive goal create starts one"
		body[(columnHeight-1)/2] = strings.Repeat(" ", max(1+(width-ansi.StringWidth(notice))/2, 0)) + dimStyle.Render(notice)
		rows = append(rows, body...)
	} else {
		rows = append(rows, m.workColumnsView(columns, width, columnHeight)...)
	}
	if scroll {
		rows = append(rows, m.workScrollView(columns, width))
	}
	for len(rows) < height-1 {
		rows = append(rows, "")
	}
	var legend strings.Builder
	legend.WriteString(" ")
	for _, entry := range workLegend {
		legend.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(entry.color)).Render(entry.glyph) + dimStyle.Render(entry.text))
	}
	rows = append(rows[:height-1], legend.String())
	return fitBlock(strings.Join(rows, "\n"), width, height)
}

// workColumnsView renders the columns on the screen and the rails between
// them at columnHeight rows.
func (m *Model) workColumnsView(columns []workColumn, width, columnHeight int) []string {
	s := &m.tasks
	fit := workColumnsFit(width)
	left := s.boardLeft(columns, fit)
	shown := columns[left:min(left+fit, len(columns))]
	selected, _ := s.position(columns)
	tops := make([]int, len(shown))
	slots := make([]int, len(shown))
	for index, column := range shown {
		slots[index] = workColumnSlots(column, columnHeight)
		tops[index] = s.columnTop(column, slots[index])
	}
	views := make([]component.WaveColumn, len(shown))
	for index, column := range shown {
		view := component.WaveColumn{Title: column.title, Selected: left+index == selected, Foot: column.foot, Done: column.done, Height: columnHeight}
		if column.hidden > 0 {
			view.Hint = fmt.Sprintf("h shows %d cancelled", column.hidden)
		}
		cards := column.cards[tops[index]:]
		for _, item := range cards[:min(len(cards), slots[index])] {
			view.Cards = append(view.Cards, s.card(item, m.now()))
		}
		view.Thumb = component.Thumb(columnHeight-2, len(column.cards), slots[index], tops[index])
		views[index] = view
	}
	rails := make([]component.Rail, max(len(shown)-1, 0))
	for index := range rails {
		rail := component.Rail{Height: columnHeight}
		from, to := shown[index], shown[index+1]
		for k2 := tops[index+1]; k2 < min(len(to.cards), tops[index+1]+slots[index+1]); k2++ {
			for _, id := range to.cards[k2].record.DependsOn {
				for k1 := tops[index]; k1 < min(len(from.cards), tops[index]+slots[index]); k1++ {
					if from.cards[k1].record.ID != id {
						continue
					}
					rowFrom, rowTo := 3+(k1-tops[index])*5, 3+(k2-tops[index+1])*5
					rail.Links = append(rail.Links, component.RailLink{From: rowFrom, To: rowTo})
					views[index].RailsOut = append(views[index].RailsOut, rowFrom)
					views[index+1].RailsIn = append(views[index+1].RailsIn, rowTo)
				}
			}
		}
		rails[index] = rail
	}
	parts := make([][]string, 0, 2*len(views))
	for index, view := range views {
		parts = append(parts, strings.Split(view.View(), "\n"))
		if index < len(rails) {
			parts = append(parts, strings.Split(rails[index].View(), "\n"))
		}
	}
	rows := make([]string, columnHeight)
	for row := range rows {
		var line strings.Builder
		line.WriteString(" ")
		for _, part := range parts {
			line.WriteString(part[row])
		}
		rows[row] = line.String()
	}
	return rows
}

// workScrollView gives the scroll row of rule 6: the waves on the screen on
// the left, and the track with the ━ thumb on the right.
func (m *Model) workScrollView(columns []workColumn, width int) string {
	s := &m.tasks
	fit := workColumnsFit(width)
	left := s.boardLeft(columns, fit)
	last := min(left+fit, len(columns)) - 1
	text := fmt.Sprintf("←→ at the edge scrolls · waves %s–%s of %d", columns[left].key, columns[last].key, len(columns))
	track := width - 1 - workScrollTrack
	if track < 1 {
		return " " + dimStyle.Render(text)
	}
	thumb := component.Thumb(track, len(columns), fit, left)
	bar := workBorderStyle.Render(strings.Repeat("─", thumb.Start)) +
		accentStyle.Render(strings.Repeat("━", thumb.Size)) +
		workBorderStyle.Render(strings.Repeat("─", track-thumb.Start-thumb.Size))
	return padCell(" "+dimStyle.Render(text), workScrollTrack) + bar
}

// workBorderStyle is the style of the border token.
var workBorderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(colorBorder))
