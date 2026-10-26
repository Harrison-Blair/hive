package tui

import (
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/tui/component"
)

// workNoGoal is the key of the pseudo-goal of work.md rule 1. A goal ID has
// the form g<n>-<slug>, so no goal has this key.
const workNoGoal = "no goal"

// The popups of the Work board.
const (
	workPopupTask = "task"
	workPopupDeps = "deps"
)

// tasksState is the state of the Work board. Goal is the key of the goal on
// the screen. Left is the first column on the screen, and tops gives the
// first card on the screen of each column by its key. A file of the
// selected task can take the place of the task popup text.
type tasksState struct {
	data          tasksData
	goal          string
	selectedID    string
	openTask      string
	left          int
	tops          map[string]int
	stopsOnly     bool
	showCancelled bool
	popup         string
	depsFromTask  bool
	popupOffset   int
	depsRow       int
	depsOffset    int
	fileOpen      bool
	fileLoading   bool
	fileTitle     string
	fileContent   string
}

// workGoal is one entry of the goal cycle. Goal is nil for the pseudo-goal.
type workGoal struct {
	key  string
	goal *records.Goal
}

// workColumn is one column of the board: a wave, the no goal column, or the
// no waves column. Cards are the cards that the filters show. Hidden counts
// the cancelled cards that hive hides.
type workColumn struct {
	key    string
	title  string
	foot   string
	done   bool
	cards  []taskItem
	hidden int
}

func (s *tasksState) accept(data tasksData) {
	s.data = data
	if s.openTask != "" {
		if _, ok := s.data.item(s.openTask); ok {
			// Keep the task that the Agents key opened across a stale
			// load that does not hold it.
			s.selectedID = s.openTask
			s.popup = workPopupTask
			s.popupOffset = 0
			s.closeFile()
			s.openTask = ""
		} else {
			s.selectedID = ""
			s.popup = ""
		}
	}
	s.normalize()
}

// cycle gives the goals in creation order, then the pseudo-goal when a task
// has no goal.
func (s *tasksState) cycle() []workGoal {
	goals := make([]workGoal, 0, len(s.data.goals)+1)
	for index := range s.data.goals {
		goals = append(goals, workGoal{key: s.data.goals[index].ID, goal: &s.data.goals[index]})
	}
	for _, item := range s.data.items {
		if item.goal == "" {
			goals = append(goals, workGoal{key: workNoGoal})
			break
		}
	}
	return goals
}

// goalIndex gives the index of the goal on the screen in the cycle, or -1.
func (s *tasksState) goalIndex() int {
	return slices.IndexFunc(s.cycle(), func(goal workGoal) bool { return goal.key == s.goal })
}

// itemGoal gives the goal key of a task.
func itemGoal(item taskItem) string {
	if item.goal == "" {
		return workNoGoal
	}
	return item.goal
}

// normalize keeps the selected task and shows its goal. Without a selected
// task, it keeps the goal on the screen, or shows the first active goal,
// and selects its first card.
func (s *tasksState) normalize() {
	cycle := s.cycle()
	if item, ok := s.data.item(s.selectedID); ok {
		if goal := itemGoal(item); goal != s.goal {
			s.goal, s.left, s.tops = goal, 0, nil
		}
		return
	}
	s.selectedID = ""
	if s.goalIndex() < 0 {
		s.goal, s.left, s.tops = "", 0, nil
		for _, goal := range cycle {
			if goal.goal != nil && goal.goal.Status == "active" {
				s.goal = goal.key
				break
			}
		}
		if s.goal == "" && len(cycle) > 0 {
			s.goal = cycle[0].key
		}
	}
	s.selectDefault()
}

// selectDefault selects the first card of the first wave that is not done,
// else the first card of the board.
func (s *tasksState) selectDefault() {
	s.selectedID = ""
	columns := s.columns()
	for _, column := range columns {
		if !column.done && len(column.cards) > 0 {
			s.selectedID = column.cards[0].record.ID
			return
		}
	}
	for _, column := range columns {
		if len(column.cards) > 0 {
			s.selectedID = column.cards[0].record.ID
			return
		}
	}
}

// columns gives the columns of the goal on the screen.
func (s *tasksState) columns() []workColumn {
	if s.goal == workNoGoal {
		column := workColumn{key: workNoGoal, foot: "no done when"}
		var items []taskItem
		for _, item := range s.data.items {
			if item.goal == "" {
				items = append(items, item)
			}
		}
		s.fill(&column, items)
		column.title = fmt.Sprintf("no goal · %d", s.countOpen(items))
		return []workColumn{column}
	}
	index := s.goalIndex()
	if index < 0 {
		return nil
	}
	goal := s.cycle()[index].goal
	if len(goal.Waves) == 0 {
		return []workColumn{{key: "no waves", title: "no waves"}}
	}
	columns := make([]workColumn, 0, len(goal.Waves))
	for _, wave := range goal.Waves {
		var items []taskItem
		for _, id := range wave.Tasks {
			if item, ok := s.data.item(id); ok {
				items = append(items, item)
			}
		}
		column := workColumn{key: wave.ID, foot: "done when " + wave.DoneWhen}
		landed, total := s.landedOf(items)
		state := "planned"
		switch {
		case total > 0 && landed == total:
			state, column.done = "done", true
		case slices.ContainsFunc(items, func(item taskItem) bool {
			return !taskLanded(item.record) && (item.record.Status == "assigned" || item.record.Status == "verified")
		}):
			state = "running"
		}
		column.title = fmt.Sprintf("%s · %s · %d/%d", wave.ID, state, landed, total)
		s.fill(&column, items)
		columns = append(columns, column)
	}
	return columns
}

// fill puts the cards that the filters show in column: the cards that are
// not cancelled in order, then the cancelled cards when h shows them.
func (s *tasksState) fill(column *workColumn, items []taskItem) {
	var cancelled []taskItem
	for _, item := range items {
		switch {
		case item.record.Status == "cancelled":
			if s.showCancelled && !s.stopsOnly {
				cancelled = append(cancelled, item)
			} else if !s.stopsOnly {
				column.hidden++
			}
		case !s.stopsOnly || s.isStop(item):
			column.cards = append(column.cards, item)
		}
	}
	column.cards = append(column.cards, cancelled...)
}

// landedOf gives the landed tasks and the tasks that are not cancelled.
func (s *tasksState) landedOf(items []taskItem) (landed, total int) {
	for _, item := range items {
		if item.record.Status == "cancelled" {
			continue
		}
		total++
		if taskLanded(item.record) {
			landed++
		}
	}
	return landed, total
}

// countOpen gives the tasks that are not cancelled.
func (s *tasksState) countOpen(items []taskItem) int {
	_, total := s.landedOf(items)
	return total
}

// isStop reports a stop of work.md rule 2: a task with needs_user, or a
// task whose worker is unresolved.
func (s *tasksState) isStop(item taskItem) bool {
	task := item.record
	if taskLanded(task) || task.Status == "cancelled" {
		return false
	}
	if task.NeedsUser != "" {
		return true
	}
	agent, ok := s.data.agents[task.ActingAgent]
	return ok && task.ActingAgent != "" && agent.Status == records.AgentUnresolved
}

// position gives the column and the card index of the selected task, or
// -1 and -1.
func (s *tasksState) position(columns []workColumn) (int, int) {
	for c, column := range columns {
		for k, item := range column.cards {
			if item.record.ID == s.selectedID {
				return c, k
			}
		}
	}
	return -1, -1
}

// workColumnsFit gives the columns that fit in width cells: 28 cells each,
// 2 cells between, and a margin of 1 cell.
func workColumnsFit(width int) int {
	return max((width+1)/(component.WaveColumnWidth+2), 1)
}

// boardLeft gives the first column on the screen. The board moves only as
// far as necessary to show the selected column (rule 6).
func (s *tasksState) boardLeft(columns []workColumn, fit int) int {
	left := min(max(s.left, 0), max(len(columns)-fit, 0))
	if c, _ := s.position(columns); c >= 0 {
		left = min(max(left, c-fit+1), c)
	}
	return left
}

// columnTop gives the first card on the screen of a column with slots
// cards. The column moves only as far as necessary to show the selected
// card (rule 7).
func (s *tasksState) columnTop(column workColumn, slots int) int {
	top := min(max(s.tops[column.key], 0), max(len(column.cards)-slots, 0))
	for k, item := range column.cards {
		if item.record.ID == s.selectedID {
			top = min(max(top, k-slots+1), k)
		}
	}
	return top
}

// workScrollRow reports that the board needs the scroll row of rule 6.
func workScrollRow(columns []workColumn, width int) bool {
	return len(columns) > workColumnsFit(width)
}

// workColumnHeight gives the rows of a column in a content area of height
// rows: a blank row, the columns, an optional scroll row, a blank row, and
// the legend.
func workColumnHeight(height int, scroll bool) int {
	if scroll {
		return max(height-4, 3)
	}
	return max(height-3, 3)
}

// workKeep stores the scroll of the board and of the column of the
// selection for the size of the content area.
func (m *Model) workKeep() {
	s := &m.tasks
	columns := s.columns()
	s.left = s.boardLeft(columns, workColumnsFit(m.width))
	if c, _ := s.position(columns); c >= 0 {
		if s.tops == nil {
			s.tops = map[string]int{}
		}
		s.tops[columns[c].key] = s.columnTop(columns[c], m.workSlots(columns, columns[c]))
	}
	m.invalidateView()
}

// workSlots gives the cards that fit in column of the board. The hint of
// the hidden cancelled cards takes one row.
func (m *Model) workSlots(columns []workColumn, column workColumn) int {
	_, height := m.contentArea()
	return workColumnSlots(column, workColumnHeight(height, workScrollRow(columns, m.width)))
}

// workColumnSlots gives the cards that fit in column at rows rows.
func workColumnSlots(column workColumn, rows int) int {
	if column.hidden > 0 {
		rows--
	}
	return component.WaveColumnSlots(rows)
}

type workScreen struct{ model *Model }

func (workScreen) DataRequest() dataRequest { return tasksDataRequest }

func (workScreen) KeyHints() string { return "" }

func (s workScreen) View(width, height int) string { return s.model.workBoardView(width, height) }

func (s workScreen) Update(msg tea.Msg) tea.Cmd {
	press, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil
	}
	m := s.model
	switch m.tasks.popup {
	case workPopupTask:
		return m.workTaskPopupKey(press.String())
	case workPopupDeps:
		return m.workDepsPopupKey(press.String())
	}
	return m.workBoardKey(press.String())
}

// workBoardKey handles a key of the board (work.md key table).
func (m *Model) workBoardKey(key string) tea.Cmd {
	s := &m.tasks
	// A user action gives the selection to the board again.
	s.openTask = ""
	switch key {
	case "left", "right", "up", "down", "k", "j", "g", "G":
		before := s.selectedID
		m.workMove(key)
		if s.selectedID == before {
			return nil
		}
		return m.logAction("work selection", slog.String("task_id", s.selectedID))
	case "[", "]":
		cycle := s.cycle()
		if len(cycle) < 2 {
			return nil
		}
		step := 1
		if key == "[" {
			step = len(cycle) - 1
		}
		s.goal = cycle[(max(s.goalIndex(), 0)+step)%len(cycle)].key
		s.left, s.tops = 0, nil
		s.selectDefault()
		m.workKeep()
		return m.logAction("work goal", slog.String("goal", s.goal))
	case "enter":
		if _, ok := s.data.item(s.selectedID); ok && m.workSelectedVisible() {
			s.popup, s.popupOffset = workPopupTask, 0
			m.invalidateView()
			return m.logAction("work task popup open", slog.String("task_id", s.selectedID))
		}
	case "d":
		if _, ok := s.data.item(s.selectedID); ok && m.workSelectedVisible() {
			s.popup, s.depsFromTask, s.depsRow, s.depsOffset = workPopupDeps, false, 0, 0
			m.invalidateView()
			return m.logAction("work deps popup open", slog.String("task_id", s.selectedID))
		}
	case "s":
		s.stopsOnly = !s.stopsOnly
		if s.stopsOnly && !m.workSelectedVisible() {
			for _, column := range s.columns() {
				if len(column.cards) > 0 {
					s.selectedID = column.cards[0].record.ID
					break
				}
			}
		}
		m.workKeep()
		return m.logAction("work stops only", slog.Bool("stops_only", s.stopsOnly))
	case "h":
		s.showCancelled = !s.showCancelled
		if !m.workSelectedVisible() {
			s.selectDefault()
		}
		m.workKeep()
		return m.logAction("work cancelled", slog.Bool("show_cancelled", s.showCancelled))
	}
	return nil
}

// workSelectedVisible reports that the filters show the selected card.
func (m *Model) workSelectedVisible() bool {
	c, _ := m.tasks.position(m.tasks.columns())
	return c >= 0
}

// workMove moves the selection by one key of rule 8 and keeps it on the
// screen.
func (m *Model) workMove(key string) {
	s := &m.tasks
	columns := s.columns()
	c, k := s.position(columns)
	if c < 0 {
		s.selectDefault()
		m.workKeep()
		return
	}
	cards := columns[c].cards
	switch key {
	case "up", "k":
		k = max(k-1, 0)
	case "down", "j":
		k = min(k+1, len(cards)-1)
	case "g":
		k = 0
	case "G":
		k = len(cards) - 1
	case "left", "right":
		step := 1
		if key == "left" {
			step = -1
		}
		row := k - s.columnTop(columns[c], m.workSlots(columns, columns[c]))
		for next := c + step; next >= 0 && next < len(columns); next += step {
			if len(columns[next].cards) == 0 {
				continue
			}
			top := s.columnTop(columns[next], m.workSlots(columns, columns[next]))
			k, cards = min(top+row, len(columns[next].cards)-1), columns[next].cards
			break
		}
	}
	s.selectedID = cards[k].record.ID
	m.workKeep()
}

// workTaskPopupKey handles a key of the task popup (rule 9).
func (m *Model) workTaskPopupKey(key string) tea.Cmd {
	s := &m.tasks
	item, ok := s.data.item(s.selectedID)
	if !ok {
		s.popup = ""
		return nil
	}
	switch key {
	case "esc":
		if s.closeFile() {
			return m.logAction("task file close", slog.String("file", s.fileTitle))
		}
		s.popup = ""
		m.invalidateView()
		return m.logAction("work task popup close", slog.String("task_id", s.selectedID))
	case "up", "k":
		s.popupOffset = max(s.popupOffset-1, 0)
	case "down", "j":
		s.popupOffset = min(s.popupOffset+1, m.workPopupMaxOffset())
	case "pgup":
		s.popupOffset = max(s.popupOffset-workPopupPage, 0)
	case "pgdown":
		s.popupOffset = min(s.popupOffset+workPopupPage, m.workPopupMaxOffset())
	case "b", "v":
		if item.projectPath == "" || item.record.Feature == "" {
			return nil
		}
		name, rel := "brief.md", filepath.Join(taskDir(item.record), "brief.md")
		if key == "v" {
			name, rel = "VERIFIED.md", filepath.Join(taskDir(item.record), "evidence", "VERIFIED.md")
		}
		read := m.startTaskFileRead(filepath.Join(item.projectPath, rel), name)
		return batch(m.logAction("task file open", slog.String("task_id", item.record.ID), slog.String("file", name)), read)
	case "d":
		s.closeFile()
		s.popup, s.depsFromTask, s.depsRow, s.depsOffset = workPopupDeps, true, 0, 0
		return m.logAction("work deps popup open", slog.String("task_id", s.selectedID))
	default:
		return nil
	}
	m.invalidateView()
	return m.logAction("work task popup scroll", slog.String("task_id", s.selectedID), slog.Int("offset", s.popupOffset))
}

// workPopupPage is the rows of one pgup or pgdn in the task popup.
const workPopupPage = 10

// workDepsPopupKey handles a key of the dependency popup (rule 10 and the
// board GfWorkDeps).
func (m *Model) workDepsPopupKey(key string) tea.Cmd {
	s := &m.tasks
	rows := s.depsRows()
	switch key {
	case "esc":
		s.popup = ""
		if s.depsFromTask {
			s.popup = workPopupTask
		}
		s.depsOffset = 0
		m.invalidateView()
		return m.logAction("work deps popup close", slog.String("task_id", s.selectedID))
	case "up", "k":
		s.depsRow = max(s.depsRow-1, 0)
	case "down", "j":
		s.depsRow = min(s.depsRow+1, max(len(rows)-1, 0))
	case "enter":
		if s.depsRow >= len(rows) {
			return nil
		}
		s.selectedID = rows[s.depsRow].record.ID
		s.normalize()
		s.popup, s.popupOffset, s.depsFromTask, s.depsOffset = workPopupTask, 0, false, 0
		m.workKeep()
		return m.logAction("work task popup open", slog.String("task_id", s.selectedID))
	default:
		return nil
	}
	lines, selected := m.workDepsLines()
	s.depsOffset = s.depsVisibleOffset(selected, len(lines), m.workPopupBody())
	m.invalidateView()
	selectedID := s.selectedID
	if s.depsRow < len(rows) {
		selectedID = rows[s.depsRow].record.ID
	}
	return m.logAction("work deps selection", slog.String("task_id", selectedID))
}

// depsRows gives the rows of the dependency popup: the dependencies of the
// selected task in record order, then the tasks that depend on it in ID
// order.
func (s *tasksState) depsRows() []taskItem {
	item, ok := s.data.item(s.selectedID)
	if !ok {
		return nil
	}
	var rows []taskItem
	for _, id := range item.record.DependsOn {
		if dependency, ok := s.data.item(id); ok {
			rows = append(rows, dependency)
		}
	}
	return append(rows, s.dependents(item.record.ID)...)
}

// dependents gives the tasks that depend on id.
func (s *tasksState) dependents(id string) []taskItem {
	var rows []taskItem
	for _, candidate := range s.data.items {
		if slices.Contains(candidate.record.DependsOn, id) {
			rows = append(rows, candidate)
		}
	}
	return rows
}

// waitsOn gives the dependencies of a task that did not land.
func (s *tasksState) waitsOn(task records.Task) []string {
	var ids []string
	for _, id := range task.DependsOn {
		if dependency, ok := s.data.item(id); !ok || !taskLanded(dependency.record) {
			ids = append(ids, id)
		}
	}
	return ids
}

// verificationItemPattern finds the VN heading of a brief and the VN word of
// a verification reason.
var (
	verificationHeading = regexp.MustCompile(`(?m)^#+\s+V(\d+)\b`)
	verificationWord    = regexp.MustCompile(`\bV(\d+)\b`)
)

// cardLine gives line 2 of the card of a task and reports that it is in
// accent (work.md rule 4). VN of a pass is the highest VN heading of the
// brief. VN of a fail is the first VN word of the reason.
func (s *tasksState) cardLine(item taskItem, now time.Time) (string, bool) {
	task := item.record
	verification := latestTaskVerification(task)
	agent, hasAgent := s.data.agents[task.ActingAgent]
	switch {
	case task.Status == "cancelled":
		return "cancelled · " + task.CancelReason, false
	case taskLanded(task):
		return "landed · " + shortSHA(task.Delivery.Commit), false
	case task.NeedsUser != "":
		return "needs you · " + task.NeedsUser, false
	case s.isStop(item):
		return "stop · unresolved", false
	case task.Status == "verified" && verification != nil && verification.Result == "pass":
		highest := 0
		for _, match := range verificationHeading.FindAllStringSubmatch(item.brief, -1) {
			n, _ := strconv.Atoi(match[1])
			highest = max(highest, n)
		}
		if highest == 0 {
			return "verified · pass", false
		}
		return fmt.Sprintf("verified · V%d pass", highest), false
	case verification != nil && verification.Result == "fail":
		if match := verificationWord.FindString(verification.Reason); match != "" {
			return "verify fail · " + match, true
		}
		return "verify fail", true
	case task.ActingAgent != "" && hasAgent:
		if age := coordinatorStatusAge(agent.StatusChangedAt, now); age != "" {
			return fitLabelDuration(task.ActingAgent, age, component.CardWidth-4), false
		}
		return task.ActingAgent, false
	}
	if ids := s.waitsOn(task); len(ids) > 0 {
		numbers := make([]string, len(ids))
		for index, id := range ids {
			numbers[index], _, _ = strings.Cut(id, "-")
		}
		return "waits on " + strings.Join(numbers, ", "), false
	}
	return task.Status, false
}

// shortSHA gives the first 6 characters of a commit, as the cards show it.
func shortSHA(commit string) string {
	if len(commit) > 6 {
		return commit[:6]
	}
	return commit
}

// taskWorkingMarker reports that a task card shows the working marker. The
// card cases cancel, land, stop, and needs-you before this marker, so the
// timer must use the same order.
func (s *tasksState) taskWorkingMarker(item taskItem) bool {
	task := item.record
	return task.Status == "verified" && !taskLanded(task) && task.NeedsUser == "" && !s.isStop(item)
}

// card gives the card of a task (rule 4).
func (s *tasksState) card(item taskItem, now time.Time, frame int) component.TaskCard {
	line, accent := s.cardLine(item, now)
	card := component.TaskCard{ID: item.record.ID, Line: line, Dot: "○", DotColor: component.Dim, Selected: item.record.ID == s.selectedID}
	if accent {
		card.LineColor = component.Accent
	}
	switch task := item.record; {
	case task.Status == "cancelled":
		card.Dot = "×"
	case taskLanded(task):
		card.Dot, card.DotColor = "●", component.Done
	case s.isStop(item):
		card.Dot, card.DotColor = "!", component.Accent
	case s.taskWorkingMarker(item):
		card.Dot, card.DotColor = component.WorkingGlyph(frame, "◐"), component.Working
	}
	if agent, ok := s.data.agents[item.record.ActingAgent]; ok && item.record.ActingAgent != "" && agent.AgentStatus == "working" {
		card.Working = true
	}
	return card
}

// closeFile closes an open file and reports that one was open.
func (s *tasksState) closeFile() bool {
	if !s.fileOpen {
		return false
	}
	s.fileOpen = false
	s.fileLoading = false
	s.fileContent = ""
	return true
}

type taskFileMsg struct {
	title   string
	content string
	err     error
}

// startTaskFileRead opens a file of the selected task in the task popup.
func (m *Model) startTaskFileRead(path, title string) tea.Cmd {
	s := &m.tasks
	s.fileOpen = true
	s.fileLoading = true
	s.fileTitle = title
	s.fileContent = ""
	s.popupOffset = 0
	m.invalidateView()
	return m.startLoad(func() tea.Msg {
		content, err := m.readScreenFile(path)
		return taskFileMsg{title: title, content: content, err: err}
	})
}

func (m *Model) acceptTaskFile(msg taskFileMsg) {
	m.tasks.fileLoading = false
	if errors.Is(msg.err, errDemoNoFile) {
		m.tasks.fileContent = demoNoFile
		m.demo.notice = demoNoFile
		return
	}
	if msg.err != nil {
		m.tasks.fileContent = "error: " + boundedError(msg.err.Error())
		return
	}
	m.tasks.fileTitle = msg.title
	m.tasks.fileContent = msg.content
}
