package tui

import (
	"context"
	"fmt"
	"image/color"
	"log/slog"
	"slices"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Harrison-Blair/hive/internal/doctor"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/tui/component"
	"github.com/charmbracelet/x/ansi"
)

const healthRestartOnly = "restart: coordinator and supervisor only"

// healthNote is the row under the agents table (health.md rule 4).
const healthNote = "restart: coordinator and supervisor only · hive asks to confirm only when the agent is live · other rows show a notice"

// healthAreas are the rows of the check grid, in order (rule 1).
var healthAreas = []string{"herdr", "home", "agents", "supervisor", "harness"}

// The columns of the agents table, each with its gap (rule 4).
const (
	healthNameWidth    = 17
	healthProfileWidth = 12
	healthHarnessWidth = 9
	healthStateWidth   = 14
	healthPaneWidth    = 10
	healthStartedWidth = 10
	healthHeader       = "AGENT             PROFILE      HARNESS   STATE          PANE       STARTED   RESTART"
)

// healthTileWidth is the most cells of a tile text.
const healthTileWidth = 30

// healthLabelWidth is the cells of a WHAT IT MEANS or NEXT STEP label and
// its gap.
const healthLabelWidth = 14

// healthScreen is the Health screen of golden-frames screens/health.md: the
// check grid of hive doctor, the details box of the selected tile, and the
// agents table. The grid and the table are its two boxes.
type healthScreen struct{ model *Model }

type healthState struct {
	result        doctor.Result
	checkedAt     time.Time
	agents        agentsData
	selected      string
	selectedAgent string
	agentOffset   int
	checking      bool
}

// loadHealth runs hive doctor in the TUI process. A panic in a check is the
// error of rule 7, so the TUI keeps running.
func loadHealth(ctx context.Context, home string, logger *slog.Logger) (result doctor.Result, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	return doctor.Run(ctx, home, logger), nil
}

func (healthScreen) DataRequest() dataRequest { return healthDataRequest }

func (healthScreen) KeyHints() string { return "" }

// View is not used: the frame draws the Health body with healthContent.
func (healthScreen) View(int, int) string { return "" }

func (s healthScreen) Update(msg tea.Msg) tea.Cmd {
	press, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil
	}
	m := s.model
	key := press.String()
	if m.agents.restartRunning {
		return nil
	}
	if m.agents.peekOpen && key == "esc" {
		pane := m.agents.peekPane
		m.agents.peekOpen, m.agents.peekText, m.agents.peekPane = false, "", ""
		return m.logAction("agent peek close", slog.String("pane_id", pane))
	}
	if m.regionFocus[tabHealth] == 0 {
		switch key {
		case "up":
			return s.moveTile(-1, 0)
		case "down":
			return s.moveTile(1, 0)
		case "left":
			return s.moveTile(0, -1)
		case "right":
			return s.moveTile(0, 1)
		}
		// health.md keys: r, a, and p act only while the table has the
		// focus.
		return nil
	}
	node, selected := m.health.selectedAgentNode()
	switch key {
	case "up":
		return s.moveAgent(-1)
	case "down":
		return s.moveAgent(1)
	case "g":
		return s.moveAgent(-len(m.health.agentList()))
	case "G":
		return s.moveAgent(len(m.health.agentList()))
	case "pgup":
		return s.moveAgent(-m.healthPage())
	case "pgdown":
		return s.moveAgent(m.healthPage())
	case "r":
		return m.requestRestart(node, selected)
	case "a":
		if selected && node.record.PaneID != "" {
			return batch(m.logAction("agent attach", slog.String("pane_id", node.record.PaneID)), m.attachCmd(node.record.PaneID))
		}
	case "p":
		if selected && node.record.PaneID != "" {
			if m.demo != nil {
				return m.demoBlocked("peek")
			}
			return batch(m.logAction("agent peek", slog.String("pane_id", node.record.PaneID)), m.startAgentsPeek(node.record.PaneID))
		}
	}
	return nil
}

// moveTile moves the grid selection by rows areas and cols tiles. Up and
// down keep the tile index as far as the row has tiles.
func (s healthScreen) moveTile(rows, cols int) tea.Cmd {
	m := s.model
	grid := m.health.grid()
	row, col := m.health.selectedTile(grid)
	if row < 0 {
		return nil
	}
	next := row
	for step := next + rows; rows != 0 && step >= 0 && step < len(grid); step += rows {
		if len(grid[step]) > 0 {
			next = step
			break
		}
	}
	col = min(max(col+cols, 0), len(grid[next])-1)
	name := grid[next][col].Name
	if name == m.health.selected {
		return nil
	}
	m.health.selected = name
	return m.logAction("health selection", slog.String("check", name))
}

// moveAgent moves the table selection by delta rows.
func (s healthScreen) moveAgent(delta int) tea.Cmd {
	m := s.model
	nodes := m.health.agentList()
	index := slices.IndexFunc(nodes, func(node agentNode) bool { return node.record.Key == m.health.selectedAgent })
	return m.selectHealthAgent(index + delta)
}

// selectHealthAgent selects the agent at index of the agents table.
func (m *Model) selectHealthAgent(index int) tea.Cmd {
	state := &m.health
	nodes := state.agentList()
	next := min(max(index, 0), len(nodes)-1)
	if next < 0 || nodes[next].record.Key == state.selectedAgent {
		return nil
	}
	state.selectedAgent = nodes[next].record.Key
	return m.logAction("agent selection", slog.String("agent_id", state.selectedAgent))
}

// requestRestart restarts the selected coordinator or supervisor with the
// behavior of task 3-tui-restart. It asks first only when the agent is live
// (health.md rule 6).
func (m *Model) requestRestart(node agentNode, ok bool) tea.Cmd {
	state := &m.agents
	kind := agentBuiltin(node.record)
	restartable := kind == records.AgentKindCoordinator || kind == records.AgentKindSupervisor
	if m.demo != nil && (!ok || restartable) {
		return m.demoBlocked("restart")
	}
	if !ok || !restartable {
		state.restartNotice = healthRestartOnly
		return m.logAction("agent restart request", slog.String("result", "unsupported target"))
	}
	if node.record.Status == records.AgentStarting {
		state.restartNotice = "cannot restart an agent while it is starting"
		return m.logAction("agent restart request", slog.String("agent_id", node.record.Key), slog.String("result", "agent is starting"))
	}
	state.restartNotice = ""
	state.restartTarget = node.record
	if kind == records.AgentKindCoordinator && node.record.Status == records.AgentLive {
		state.restartRunning = true
		state.restartNotice = "checking coordinator pane"
		return batch(
			m.logAction("agent restart request", slog.String("agent_id", node.record.Key), slog.String("profile", agentLabel(node.record)), slog.String("result", "checking pane")),
			m.startCoordinatorRestartCheck(node.record),
		)
	}
	if node.record.Status == records.AgentLive && (kind != records.AgentKindSupervisor || m.health.agents.supervisorLive) {
		state.restartConfirm = true
		state.restartNotice = "restart " + kind + "? y/n"
		return m.logAction("agent restart request", slog.String("agent_id", node.record.Key), slog.String("profile", agentLabel(node.record)), slog.String("result", "confirmation required"))
	}
	state.restartRunning = true
	state.restartNotice = "restarting"
	return batch(
		m.logAction("agent restart request", slog.String("agent_id", node.record.Key), slog.String("profile", agentLabel(node.record)), slog.String("result", "started immediately")),
		m.startAgentRestart(node.record),
	)
}

// accept takes a new doctor result and keeps the selected check. The first
// selection is the first tile that did not pass, else the first tile
// (rule 2).
func (s *healthState) accept(result doctor.Result) {
	s.result = result
	grid := s.grid()
	if row, _ := s.selectedTile(grid); row >= 0 {
		return
	}
	s.selected = ""
	for _, area := range grid {
		for _, check := range area {
			if s.selected == "" {
				s.selected = check.Name
			}
			if check.Status != doctor.StatusOK {
				s.selected = check.Name
				return
			}
		}
	}
}

// acceptAgents takes new agent data and keeps the selected agent.
func (s *healthState) acceptAgents(data agentsData) {
	s.agents = data
	nodes := s.agentList()
	if slices.ContainsFunc(nodes, func(node agentNode) bool { return node.record.Key == s.selectedAgent }) {
		return
	}
	s.selectedAgent = ""
	if len(nodes) != 0 {
		s.selectedAgent = nodes[0].record.Key
	}
}

// grid gives the checks of each area of healthAreas, in doctor order.
func (s *healthState) grid() [][]doctor.Check {
	grid := make([][]doctor.Check, len(healthAreas))
	for _, check := range s.result.Checks {
		index := slices.Index(healthAreas, healthArea(check.Name))
		grid[index] = append(grid[index], check)
	}
	return grid
}

// selectedTile gives the row and the column of the selected check, or -1.
func (s *healthState) selectedTile(grid [][]doctor.Check) (int, int) {
	for row, area := range grid {
		for col, check := range area {
			if check.Name == s.selected {
				return row, col
			}
		}
	}
	return -1, -1
}

func (s *healthState) selectedCheck() (doctor.Check, bool) {
	for _, check := range s.result.Checks {
		if check.Name == s.selected {
			return check, true
		}
	}
	return doctor.Check{}, false
}

func (s *healthState) selectedAgentNode() (agentNode, bool) {
	for _, node := range s.agentList() {
		if node.record.Key == s.selectedAgent {
			return node, true
		}
	}
	return agentNode{}, false
}

// agentList gives the current record of each agent, the coordinator and the
// supervisor first (rule 4).
func (s *healthState) agentList() []agentNode {
	nodes := make([]agentNode, 0, len(s.agents.nodes))
	for _, node := range s.agents.nodes {
		if node.current {
			nodes = append(nodes, node)
		}
	}
	slices.SortStableFunc(nodes, func(a, b agentNode) int { return healthAgentRank(a.record) - healthAgentRank(b.record) })
	return nodes
}

// healthAgentRank gives the row order: the coordinator, the supervisor, an
// owner, then the first profiles. It reads only facts.
func healthAgentRank(record records.Agent) int {
	switch record.Kind {
	case records.AgentKindCoordinator:
		return 0
	case records.AgentKindSupervisor:
		return 1
	}
	if record.Owns != "" {
		return 2
	}
	switch record.Profile {
	case records.ProfilePlanner:
		return 3
	case records.ProfileBuilder:
		return 4
	case records.ProfileVerifier:
		return 5
	default:
		return 6
	}
}

// healthArea gives the grid row of a check by its name (rule 1).
func healthArea(name string) string {
	switch {
	case strings.HasPrefix(name, "herdr "):
		return "herdr"
	case strings.HasPrefix(name, "supervisor"):
		return "supervisor"
	case strings.HasPrefix(name, "harness "):
		return "harness"
	}
	switch name {
	case "starting agents", "unresolved agents", "ownerless inbox records", "stale worktrees", "unpushed task branches":
		return "agents"
	}
	return "home"
}

// healthTile gives the label and the short result of the tile of a check.
// The tile text is the two, joined by a space.
func healthTile(check doctor.Check) (string, string) {
	detail, ok := check.Detail, check.Status == doctor.StatusOK
	first, _, _ := strings.Cut(detail, ":")
	count := func() string {
		if detail == "none" {
			return detail
		}
		return fmt.Sprint(len(strings.Split(detail, ", ")))
	}
	switch name := check.Name; {
	case name == "herdr socket":
		if index := strings.Index(detail, "protocol "); index >= 0 {
			return "socket ·", detail[index:]
		}
		return "socket ·", detail
	case name == "herdr pane":
		return "pane", detail
	case name == "home":
		return "", detail
	case name == "state store":
		return "store ·", detail
	case name == "git":
		return "git", strings.TrimPrefix(detail, "git version ")
	case name == "home version" && strings.HasPrefix(detail, "development build"):
		return "version", "dev"
	case name == "home version":
		return "version", detail
	case name == "gh":
		version, _, _ := strings.Cut(strings.TrimPrefix(detail, "gh version "), " (")
		return "gh", version
	case name == "starting agents":
		return "starting", detail
	case name == "unresolved agents":
		return "unresolved", detail
	case name == "ownerless inbox records":
		return "ownerless inbox", first
	case name == "stale worktrees":
		return "stale worktrees", count()
	case name == "unpushed task branches":
		return "unpushed", count()
	case name == "supervisor":
		return "", strings.Replace(strings.TrimPrefix(detail, "supervisor "), ": ", " · ", 1)
	case name == "supervisor binary" && ok:
		return "binary", "current"
	case name == "supervisor binary" && (strings.Contains(detail, "older") || strings.Contains(detail, "deleted")):
		return "binary", "old"
	case strings.HasPrefix(name, "harness ") && ok:
		return strings.TrimPrefix(name, "harness "), ""
	case strings.HasPrefix(name, "harness "):
		return strings.TrimPrefix(name, "harness "), detail
	}
	return check.Name, first
}

// healthTileText gives the text of the tile of a check.
func healthTileText(check doctor.Check) string {
	label, short := healthTile(check)
	text := strings.TrimSpace(label + " " + short)
	return ansi.Truncate(text, healthTileWidth, "…")
}

// healthGridShowsSelection reports that the grid shows the ▶ and the accent
// of its selected tile. The boards show it only while the grid has the
// focus. Ask a28 decides this rule.
func healthGridShowsSelection(focused bool) bool { return focused }

// healthTitle gives the title row of the Health screen (health.md rule 1).
func (m *Model) healthTitle() component.TitleRow {
	return component.TitleRow{Name: "health", Facts: m.healthFacts()}
}

// healthFacts gives the facts of the title row: the counts, the time of the
// last result, and the R key or the checking mark (rule 7).
func (m *Model) healthFacts() string {
	state := &m.health
	checks := state.result.Checks
	if len(checks) == 0 && state.checkedAt.IsZero() {
		return "· checking…"
	}
	passed := 0
	for _, check := range checks {
		if check.Status == doctor.StatusOK {
			passed++
		}
	}
	facts := "· " + plural(len(checks), "check") + fmt.Sprintf(" · %d ok", passed)
	if warn := len(checks) - passed; warn > 0 {
		facts += fmt.Sprintf(" · %d warn", warn)
	}
	if !state.checkedAt.IsZero() {
		facts += " · checked " + state.checkedAt.Format("15:04:05")
	}
	if state.checking {
		return facts + " · checking…"
	}
	return facts + " · R checks again"
}

// healthContent draws the Health body at exactly width×height cells: the
// check grid, the details box, the agents table, and the note under it. The
// details box and the table take the rows that they need while the rows
// last. The table scrolls when its rows do not fit.
func (m *Model) healthContent(width, height int) string {
	state := &m.health
	gridFocused := m.regionFocus[tabHealth] == 0
	shown := healthGridShowsSelection(gridFocused)
	grid := state.grid()
	rows := make([]string, 0, height)
	for index, area := range healthAreas {
		row := component.CheckRow{Area: area}
		for _, check := range grid[index] {
			row.Tiles = append(row.Tiles, component.CheckTile{
				Text:     healthTileText(check),
				Passed:   check.Status == doctor.StatusOK,
				Selected: shown && check.Name == state.selected,
				Dim:      state.checking,
			})
		}
		rows = append(rows, row.View(width))
	}
	rows = append(rows, "")
	title, lines := m.healthDetails(width)
	tableRows := m.healthTableRows()
	layout := m.healthLayout(height, len(lines), len(tableRows))
	if layout.details >= 3 {
		rows = append(rows, component.Panel{Title: title, Lines: lines}.View(width, layout.details), "")
	}
	if layout.table >= 3 {
		rows = append(rows, m.healthTable(tableRows, width, layout.table, !gridFocused))
		if layout.note {
			rows = append(rows, component.Notice{Text: healthNote}.View(width))
		}
	}
	return fitBlock(strings.Join(rows, "\n"), width, height)
}

// healthDetails gives the title and the body of the details box of the
// selected tile (rule 3), or the peek of the selected agent.
func (m *Model) healthDetails(width int) (string, []string) {
	state := &m.health
	line := max(width-4, 1)
	if m.agents.peekOpen {
		lines := strings.Split(strings.TrimRight(m.agents.peekText, "\n"), "\n")
		return "peek · " + m.agents.peekPane + " · esc closes", lines
	}
	check, ok := state.selectedCheck()
	switch {
	case !ok && state.checkedAt.IsZero():
		return "checks", []string{dimStyle.Render("loading checks…")}
	case !ok:
		return "checks", []string{dimStyle.Render("no checks")}
	}
	passed := check.Status == doctor.StatusOK
	body := lipgloss.NewStyle()
	source := "source hive doctor · checked " + state.checkedAt.Format("15:04:05")
	if state.checking {
		// health.md rule 7: the body is dim until the new result.
		body = dimStyle
		source = "source hive doctor · last result " + state.checkedAt.Format("15:04:05") + " · new result pending"
	}
	title := "✓ " + check.Name
	if !passed {
		_, short := healthTile(check)
		title = "! " + check.Name
		if short != "" {
			title += " · " + short
		}
	}
	var lines []string
	if check.Name == "unresolved agents" && !passed {
		lines = append(lines, dimStyle.Bold(true).Render("ID               PROFILE   STATE         PANE"))
		for _, node := range state.agentList() {
			if record := node.record; record.Status == records.AgentUnresolved {
				lines = append(lines, body.Render(fitHealthCell(healthAgentName(record), 17)+fitHealthCell(agentLabel(record), 10)+fitHealthCell(record.Status, 14)+healthPane(record)))
			}
		}
	} else {
		for _, row := range strings.Split(ansi.Wrap(check.Detail, line, ""), "\n") {
			lines = append(lines, body.Render(row))
		}
	}
	if !passed {
		lines = append(lines, "")
		lines = append(lines, healthLabelled("WHAT IT MEANS", healthMeaning(check), line, body)...)
		lines = append(lines, healthLabelled("NEXT STEP", "Ask the coordinator in Chat. Health only shows state.", line, body)...)
		lines = append(lines, "")
	}
	return title, append(lines, dimStyle.Render(source))
}

// healthMeaning gives the WHAT IT MEANS text of a check that did not pass.
func healthMeaning(check doctor.Check) string {
	if check.Name == "unresolved agents" {
		return "The record says the agent is live, but hive cannot find its pane. Hive does not change the record."
	}
	return "hive doctor gives " + check.Status + " for " + check.Name + ": " + check.Detail
}

// healthLabelled gives a label in accent bold and its text, wrapped to line
// cells. Each next row starts below the text.
func healthLabelled(label, text string, line int, body lipgloss.Style) []string {
	rows := strings.Split(ansi.Wordwrap(text, max(line-healthLabelWidth, 1), ""), "\n")
	out := make([]string, len(rows))
	for index, row := range rows {
		head := strings.Repeat(" ", healthLabelWidth)
		if index == 0 {
			head = titleStyle.Render(label) + strings.Repeat(" ", healthLabelWidth-len(label))
		}
		out[index] = head + body.Render(row)
	}
	return out
}

// healthTableRows gives the rows of the agents table (rule 4).
func (m *Model) healthTableRows() []string {
	state := &m.health
	nodes := state.agentList()
	rows := make([]string, 0, len(nodes))
	for _, node := range nodes {
		record := node.record
		selected := record.Key == state.selectedAgent
		marker, name, restart := "  ", lipgloss.NewStyle(), dimStyle
		if selected {
			marker, name, restart = titleStyle.Render("▶")+" ", titleStyle, healthAccentStyle
		}
		word, dot, dotColor := healthAgentState(record, m.now(), m.markerFrame)
		started := ""
		if at, err := time.Parse(time.RFC3339Nano, record.StartedAt); err == nil {
			started = at.In(m.now().Location()).Format("15:04")
		}
		row := marker + lipgloss.NewStyle().Foreground(dotColor).Render(dot) + " " +
			name.Render(fitHealthCell(healthAgentName(record), healthNameWidth)) + " " +
			dimStyle.Render(fitHealthCell(agentLabel(record), healthProfileWidth)) + " " +
			dimStyle.Render(fitHealthCell(record.Harness, healthHarnessWidth)) + " " +
			fitHealthCell(word, healthStateWidth) + " " +
			dimStyle.Render(fitHealthCell(healthPane(record), healthPaneWidth)) + " " +
			dimStyle.Render(started) + strings.Repeat(" ", max(healthStartedWidth-lipgloss.Width(started), 0))
		if kind := agentBuiltin(record); kind == records.AgentKindCoordinator || kind == records.AgentKindSupervisor {
			row += restart.Render("r restart")
		}
		rows = append(rows, row)
	}
	if len(rows) == 0 {
		rows = append(rows, "    "+dimStyle.Render("no agents"))
	}
	return rows
}

// healthTable draws the agents table box at width×height cells. The rows
// start at the offset that keeps the selected row visible, and the ┃ thumb
// shows when they do not fit (rule 5).
func (m *Model) healthTable(rows []string, width, height int, focused bool) string {
	state := &m.health
	nodes := state.agentList()
	selected := max(slices.IndexFunc(nodes, func(node agentNode) bool { return node.record.Key == state.selectedAgent }), 0)
	visible := max(height-3, 1)
	offset := &state.agentOffset
	if visible < len(rows) {
		*offset = min(*offset, selected)
		*offset = max(*offset, selected-visible+1)
		*offset = min(max(*offset, 0), len(rows)-visible)
	} else {
		*offset = 0
	}
	lines := append([]string{"    " + dimStyle.Bold(true).Render(healthHeader)}, rows[*offset:min(*offset+visible, len(rows))]...)
	mark := component.Thumb(visible, len(rows), visible, *offset)
	if mark.Size > 0 {
		mark.Start++
	}
	return component.Panel{Title: m.healthTableTitle(nodes), Focused: focused, Lines: lines, VMark: mark}.View(width, height)
}

// healthTableTitle gives the title of the agents table: the count of the
// agents and the count of each state that has one.
func (m *Model) healthTableTitle(nodes []agentNode) string {
	counts := map[string]int{}
	for _, node := range nodes {
		record := node.record
		switch {
		case record.Status == records.AgentLive && record.AgentStatus == "idle":
			counts["idle"]++
		case record.Status == records.AgentLive:
			counts["live"]++
		case record.Status == records.AgentExited:
			counts["stopped"]++
		default:
			counts[record.Status]++
		}
	}
	title := fmt.Sprintf("agents · %d", len(nodes))
	for _, word := range []string{"live", "idle", records.AgentStarting, records.AgentStopped, records.AgentUnresolved, records.AgentStale} {
		if counts[word] > 0 {
			title += fmt.Sprintf(" · %d %s", counts[word], word)
		}
	}
	return title
}

var healthAccentStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(colorAccent))

// healthAgentName gives the name of an agent row.
func healthAgentName(record records.Agent) string {
	if record.Name == "" {
		return record.Key
	}
	return record.Name
}

// healthAgentState gives the state word of an agent, its dot, and the
// colour of the dot, as tui-v2 decision 55 gives them. A live state has the
// whole minutes from status_changed_at. The word also shows the state, so
// colour is not the only signal.
func healthAgentState(record records.Agent, now time.Time, frame int) (string, string, color.Color) {
	with := func(word string) string {
		if age := coordinatorStatusAge(record.StatusChangedAt, now); age != "" {
			return word + " " + age
		}
		return word
	}
	switch record.Status {
	case records.AgentLive:
		switch record.AgentStatus {
		case "working":
			return with("working"), component.WorkingGlyph(frame, "●"), component.Working
		case "blocked":
			return with("blocked"), "▲", component.Accent
		case "":
			return with("live"), "●", component.Working
		}
		return with(record.AgentStatus), "○", component.Dim
	case records.AgentStarting:
		return "starting", "●", component.Dim
	case records.AgentStopped, records.AgentExited:
		return "stopped", "○", component.Dim
	}
	return record.Status, "○", component.Dim
}

// healthPane gives the pane column of an agent row: the pane of a live
// agent, else gone.
func healthPane(record records.Agent) string {
	switch {
	case record.Status != records.AgentLive && record.Status != records.AgentStarting:
		return "gone"
	case record.PaneID == "":
		return "—"
	}
	return record.PaneID
}

// fitHealthCell cuts text to width cells with … and pads it to width cells.
func fitHealthCell(text string, width int) string {
	text = ansi.Truncate(text, max(width, 0), "…")
	return text + strings.Repeat(" ", max(width-lipgloss.Width(text), 0))
}
