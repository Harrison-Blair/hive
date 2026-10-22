package tui

import (
	"context"
	"fmt"
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

// The views of the Health screen (decision 18).
const (
	healthHome = iota
	healthAgents
)

// Health column widths of the Health boards, each with its gap.
const (
	healthNameWidth      = 22
	healthCheckWidth     = 24
	healthRoleWide       = 12
	healthRoleCompact    = 14
	healthStateCompact   = 14
	healthUnresolvedID   = 18
	healthUnresolvedRole = 10
	healthUnresolvedSt   = 12
	healthPromptWidth    = 54
	healthPromptHeight   = 7
	healthPromptIndent   = 6
)

const healthRestartOnly = "restart: coordinator and supervisor only"

var healthAccentStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(colorAccent))

// healthScreen is tab 5 Health. Home lists the checks of hive doctor. Agents
// lists each agent and holds the restart move of decision 19.
type healthScreen struct{ model *Model }

type healthState struct {
	result        doctor.Result
	checkedAt     time.Time
	agents        agentsData
	selected      string
	selectedAgent string
	hideStopped   bool
	checkOffset   int
	agentOffset   int
}

func loadHealth(home string, logger *slog.Logger) doctor.Result {
	return doctor.Run(context.Background(), home, logger)
}

func (healthScreen) DataRequest() dataRequest { return healthDataRequest }

func (healthScreen) KeyHints() string { return "" }

// View is not used: the frame draws the Health regions from the registry.
func (healthScreen) View(int, int) string { return "" }

func (s healthScreen) Update(msg tea.Msg) tea.Cmd {
	press, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil
	}
	m := s.model
	state := &m.health
	key := press.String()
	agents := m.views[tabHealth] == healthAgents
	if agents && m.agents.restartRunning {
		return nil
	}
	if agents && m.agents.peekOpen && key == "esc" {
		pane := m.agents.peekPane
		m.agents.peekOpen, m.agents.peekText, m.agents.peekPane = false, "", ""
		return m.logAction("agent peek close", slog.String("pane_id", pane))
	}
	page := max(m.height-frameRows-4, 1)
	switch key {
	case "up", "k":
		return s.move(-1)
	case "down", "j":
		return s.move(1)
	case "pgup":
		return s.move(-page)
	case "pgdown":
		return s.move(page)
	case "g":
		return s.move(-1 << 20)
	case "G":
		return s.move(1 << 20)
	case "enter":
		m.regionFocus[tabHealth] = 1
		m.detailsOffset[tabHealth] = 0
		return m.logAction("focus region", slog.String("region", "details"))
	}
	if !agents {
		return nil
	}
	node, selected := state.selectedAgentNode()
	switch key {
	case "r":
		return m.requestRestart(node, selected)
	case "h":
		state.hideStopped = !state.hideStopped
		state.keepAgentSelection()
		return m.logAction("agent history visibility", slog.Bool("history", !state.hideStopped))
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

// move moves the selection of the active view by delta rows.
func (s healthScreen) move(delta int) tea.Cmd {
	m := s.model
	state := &m.health
	if m.views[tabHealth] == healthAgents {
		nodes := state.agentList()
		index := slices.IndexFunc(nodes, func(node agentNode) bool { return node.record.Key == state.selectedAgent })
		next := min(max(index+delta, 0), len(nodes)-1)
		if next < 0 || next == index {
			return nil
		}
		state.selectedAgent = nodes[next].record.Key
		m.detailsOffset[tabHealth] = 0
		return m.logAction("agent selection", slog.String("agent_id", state.selectedAgent))
	}
	checks := state.result.Checks
	index := slices.IndexFunc(checks, func(check doctor.Check) bool { return check.Name == state.selected })
	next := min(max(index+delta, 0), len(checks)-1)
	if next < 0 || next == index {
		return nil
	}
	state.selected = checks[next].Name
	m.detailsOffset[tabHealth] = 0
	return m.logAction("health selection", slog.String("check", state.selected))
}

// requestRestart restarts the selected coordinator or supervisor with the
// behavior of task 3-tui-restart. It asks first only when the agent is live
// (decision 19).
func (m *Model) requestRestart(node agentNode, ok bool) tea.Cmd {
	state := &m.agents
	restartable := node.record.Role == "coordinator" || node.record.Role == "supervisor"
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
	if node.record.Role == "coordinator" && node.record.Status == records.AgentLive {
		state.restartRunning = true
		state.restartNotice = "checking coordinator pane"
		return batch(
			m.logAction("agent restart request", slog.String("agent_id", node.record.Key), slog.String("role", node.record.Role), slog.String("result", "checking pane")),
			m.startCoordinatorRestartCheck(node.record),
		)
	}
	if node.record.Status == records.AgentLive && (node.record.Role != "supervisor" || m.health.agents.supervisorLive) {
		state.restartConfirm = true
		state.restartNotice = "restart " + node.record.Role + "? y/n"
		return m.logAction("agent restart request", slog.String("agent_id", node.record.Key), slog.String("role", node.record.Role), slog.String("result", "confirmation required"))
	}
	state.restartRunning = true
	state.restartNotice = "restarting"
	return batch(
		m.logAction("agent restart request", slog.String("agent_id", node.record.Key), slog.String("role", node.record.Role), slog.String("result", "started immediately")),
		m.startAgentRestart(node.record),
	)
}

// accept takes a new doctor result and keeps the selected check.
func (s *healthState) accept(result doctor.Result) {
	s.result = result
	if !slices.ContainsFunc(result.Checks, func(check doctor.Check) bool { return check.Name == s.selected }) {
		s.selected = ""
		if len(result.Checks) != 0 {
			s.selected = result.Checks[0].Name
		}
	}
}

// acceptAgents takes new agent data and keeps the selected agent.
func (s *healthState) acceptAgents(data agentsData) {
	s.agents = data
	s.keepAgentSelection()
}

func (s *healthState) keepAgentSelection() {
	nodes := s.agentList()
	if slices.ContainsFunc(nodes, func(node agentNode) bool { return node.record.Key == s.selectedAgent }) {
		return
	}
	s.selectedAgent = ""
	if len(nodes) != 0 {
		s.selectedAgent = nodes[0].record.Key
	}
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

// agentList gives the current agent of each identity, the coordinator and
// the supervisor first. Hide stopped removes each stopped agent.
func (s *healthState) agentList() []agentNode {
	nodes := make([]agentNode, 0, len(s.agents.nodes))
	for _, node := range s.agents.nodes {
		if node.current && (!s.hideStopped || !node.stopped) {
			nodes = append(nodes, node)
		}
	}
	slices.SortStableFunc(nodes, func(a, b agentNode) int { return healthRoleRank(a.record.Role) - healthRoleRank(b.record.Role) })
	return nodes
}

// stoppedCount gives the number of current stopped agents.
func (s *healthState) stoppedCount() int {
	count := 0
	for _, node := range s.agents.nodes {
		if node.current && node.stopped {
			count++
		}
	}
	return count
}

func healthRoleRank(role string) int {
	if index := slices.Index([]string{"coordinator", "supervisor", "manager", "worker", "verifier"}, role); index >= 0 {
		return index
	}
	return 5
}

// healthAgentName gives the name of an agent row. The coordinator and the
// supervisor show their key, which names the generation that restart
// replaces.
func healthAgentName(record records.Agent) string {
	if record.Role == "coordinator" || record.Role == "supervisor" || record.Name == "" {
		return record.Key
	}
	return record.Name
}

// healthAgentState gives the state word of an agent, its dot, the style of
// the dot, and the style of the word. The dot and the word both show the
// state, so colour is not the only signal.
func healthAgentState(record records.Agent) (string, string, lipgloss.Style, lipgloss.Style) {
	plain := lipgloss.NewStyle()
	switch record.Status {
	case records.AgentStopped, records.AgentExited:
		return "stopped", "○", dimStyle, dimStyle
	case records.AgentUnresolved, records.AgentStale:
		return record.Status, "●", healthAccentStyle, plain
	case records.AgentStarting:
		return "starting", "◐", healthAccentStyle, plain
	case records.AgentLive:
		switch record.AgentStatus {
		case "working":
			return "working", "●", workingStyle, plain
		case "idle", "done":
			return "idle", "◐", healthAccentStyle, plain
		case "blocked":
			return "blocked", "●", healthAccentStyle, plain
		}
		return "live", "●", workingStyle, plain
	}
	return record.Status, "·", dimStyle, plain
}

// healthPane gives the pane column of an agent row.
func healthPane(record records.Agent) string {
	switch {
	case record.Status == records.AgentUnresolved:
		return "gone"
	case record.Status == records.AgentStopped || record.Status == records.AgentExited || record.PaneID == "":
		return "—"
	}
	return record.PaneID
}

// healthCheckGlyph gives the glyph of a check status and its style.
func healthCheckGlyph(status string) (string, lipgloss.Style) {
	switch status {
	case doctor.StatusOK:
		return "✓", doneStyle.Bold(true)
	case doctor.StatusWarn:
		return "!", titleStyle
	case doctor.StatusFail:
		return "×", titleStyle
	}
	return "·", dimStyle.Bold(true)
}

// healthCheckName gives the short name of a check for its row.
func healthCheckName(name string) string {
	if name == "unpushed task branches" {
		return "unpushed branches"
	}
	return name
}

// healthCheckResult gives the short result of a check for its row.
func healthCheckResult(check doctor.Check) string {
	detail := check.Detail
	ok := check.Status == doctor.StatusOK
	switch {
	case check.Name == "herdr socket" && strings.Contains(detail, "protocol "):
		return detail[strings.Index(detail, "protocol "):]
	case check.Name == "git":
		return strings.TrimPrefix(detail, "git version ")
	case check.Name == "gh":
		version, _, _ := strings.Cut(strings.TrimPrefix(detail, "gh version "), " (")
		return version
	case check.Name == "supervisor" && strings.HasPrefix(detail, "supervisor "):
		return strings.Replace(strings.TrimPrefix(detail, "supervisor "), ": ", " · ", 1)
	case check.Name == "supervisor binary" && ok:
		return "current"
	case check.Name == "supervisor binary" && check.Status == doctor.StatusWarn:
		return "old"
	case strings.HasPrefix(check.Name, "harness ") && ok:
		return "found"
	}
	return detail
}

// fitHealthCell cuts text to width-1 cells with … and pads it to width
// cells, so one cell of gap stays.
func fitHealthCell(text string, width int) string {
	text = ansi.Truncate(text, max(width-1, 0), "…")
	return text + strings.Repeat(" ", max(width-lipgloss.Width(text), 0))
}

// fitHealthLine cuts text to width cells with … and pads it to width cells.
func fitHealthLine(text string, width int) string {
	text = ansi.Truncate(text, max(width, 0), "…")
	return text + strings.Repeat(" ", max(width-lipgloss.Width(text), 0))
}

// healthMarker gives the selection marker cell and its gap.
func healthMarker(selected bool) string {
	if selected {
		return titleStyle.Render("▶") + " "
	}
	return "  "
}

// healthPanel draws the list region of the active Health view.
func (m *Model) healthPanel(width, height int, focused bool) string {
	if m.views[tabHealth] == healthAgents {
		return m.healthAgentsPanel(width, height, focused)
	}
	state := &m.health
	line := max(width-4, 1)
	checks := state.result.Checks
	warn, fail := 0, 0
	rows := make([]string, 0, len(checks))
	selectedIndex := 0
	for index, check := range checks {
		switch check.Status {
		case doctor.StatusWarn:
			warn++
		case doctor.StatusFail:
			fail++
		}
		selected := check.Name == state.selected
		if selected {
			selectedIndex = index
		}
		glyph, glyphStyle := healthCheckGlyph(check.Status)
		name, result := lipgloss.NewStyle(), dimStyle
		if selected {
			name, result = titleStyle, healthAccentStyle
		}
		rest := max(line-4-healthNameWidth, 0)
		rows = append(rows, healthMarker(selected)+glyphStyle.Render(glyph)+" "+
			name.Render(fitHealthCell(healthCheckName(check.Name), healthNameWidth))+
			result.Render(fitHealthLine(healthCheckResult(check), rest)))
	}
	title := fmt.Sprintf("checks · %d", len(checks))
	if warn > 0 {
		title += fmt.Sprintf(" · %d warn", warn)
	}
	if fail > 0 {
		title += fmt.Sprintf(" · %d fail", fail)
	}
	header := "  " + dimStyle.Render(fitHealthCell("CHECK", healthCheckWidth)+"RESULT")
	if len(checks) == 0 {
		rows = append(rows, "  "+dimStyle.Render("loading checks…"))
	}
	return healthListPanel(title, focused, header, rows, nil, selectedIndex, &state.checkOffset, width, height)
}

// healthAgentsPanel draws the list of Health › Agents. Below 120 columns the
// list has the pane column and the notes of the CompactHealth board.
func (m *Model) healthAgentsPanel(width, height int, focused bool) string {
	state := &m.health
	wide := m.width >= component.LongFormWidth
	line := max(width-4, 1)
	nodes := state.agentList()
	stopped := state.stoppedCount()
	roleWidth, stateWidth := healthRoleCompact, healthStateCompact
	header := fitHealthCell("AGENT", healthNameWidth) + fitHealthCell("ROLE", healthRoleCompact) + fitHealthCell("STATE", healthStateCompact) + "PANE"
	if wide {
		roleWidth = healthRoleWide
		stateWidth = max(line-2-healthNameWidth-healthRoleWide, 3)
		header = fitHealthCell("AGENT", healthNameWidth) + fitHealthCell("ROLE", healthRoleWide) + "STATE"
	}
	rows := make([]string, 0, len(nodes))
	selectedIndex := 0
	for index, node := range nodes {
		selected := node.record.Key == state.selectedAgent
		if selected {
			selectedIndex = index
		}
		name := lipgloss.NewStyle()
		if selected {
			name = titleStyle
		}
		word, dot, dotStyle, wordStyle := healthAgentState(node.record)
		row := healthMarker(selected) + name.Render(fitHealthCell(healthAgentName(node.record), healthNameWidth)) +
			dimStyle.Render(fitHealthCell(node.record.Role, roleWidth)) +
			dotStyle.Render(dot) + " "
		if wide {
			row += wordStyle.Render(fitHealthLine(word, stateWidth-2))
		} else {
			row += wordStyle.Render(fitHealthCell(word, stateWidth-2)) + dimStyle.Render(healthPane(node.record))
		}
		rows = append(rows, row)
	}
	if len(nodes) == 0 {
		rows = append(rows, "  "+dimStyle.Render("no agents"))
	}
	title := fmt.Sprintf("agents · %d", len(nodes))
	var notes []string
	plural := "s"
	if stopped == 1 {
		plural = ""
	}
	switch {
	case !wide:
		notes = []string{"", "", dimStyle.Render("r restarts the coordinator or supervisor. It asks first only when live.")}
		if stopped > 0 && state.hideStopped {
			notes = append(notes, dimStyle.Render(fmt.Sprintf("h shows %d stopped agent%s.", stopped, plural)))
		} else if stopped > 0 {
			notes = append(notes, dimStyle.Render(fmt.Sprintf("h hides %d stopped agent%s.", stopped, plural)))
		}
	case stopped > 0 && state.hideStopped:
		title += fmt.Sprintf(" · h shows %d stopped", stopped)
	case stopped > 0:
		title += fmt.Sprintf(" · h hides %d stopped", stopped)
	}
	return healthListPanel(title, focused, "  "+dimStyle.Render(header), rows, notes, selectedIndex, &state.agentOffset, width, height)
}

// healthListPanel draws a list panel: the header, the rows from the offset
// that keeps the selected row visible, then the notes. The panel draws the
// scroll mark of decision 5 when the rows do not fit.
func healthListPanel(title string, focused bool, header string, rows, notes []string, selected int, offset *int, width, height int) string {
	visible := max(height-3-len(notes), 1)
	if visible < len(rows) {
		*offset = min(*offset, selected)
		*offset = max(*offset, selected-visible+1)
		*offset = min(max(*offset, 0), len(rows)-visible)
	} else {
		*offset = 0
	}
	shown := rows[*offset:min(*offset+visible, len(rows))]
	lines := append([]string{header}, shown...)
	lines = append(lines, notes...)
	mark := component.Thumb(visible, len(rows), visible, *offset)
	if mark.Size > 0 {
		mark.Start++
	}
	return component.Panel{Title: title, Focused: focused, Lines: lines, VMark: mark}.View(width, height)
}

// healthDetails gives the title and the lines of the details region of the
// active Health view.
func (m *Model) healthDetails(width int) (string, []string) {
	line := max(width-4, 1)
	if m.views[tabHealth] == healthAgents {
		return m.healthAgentDetails(line)
	}
	check, ok := m.health.selectedCheck()
	if !ok {
		return "details", []string{"", "no check selected"}
	}
	glyphStyle := doneStyle
	if check.Status != doctor.StatusOK {
		glyphStyle = titleStyle
	}
	result := check.Status + " · " + healthCheckResult(check)
	unresolved := []agentNode{}
	if check.Name == "unresolved agents" {
		for _, node := range m.health.agentList() {
			if node.record.Status == records.AgentUnresolved {
				unresolved = append(unresolved, node)
			}
		}
		plural := "s"
		if len(unresolved) == 1 {
			plural = ""
		}
		result = fmt.Sprintf("%s · %d agent%s", check.Status, len(unresolved), plural)
	}
	checked := "not checked"
	if !m.health.checkedAt.IsZero() {
		checked = m.health.checkedAt.Format("15:04:05")
	}
	lines := []string{
		"",
		component.DetailField("result", glyphStyle.Render(fitHealthLine(result, line-10))),
		component.DetailField("source", "hive doctor"),
		component.DetailField("checked", checked+" · R to check again"),
		"",
	}
	if check.Name == "unresolved agents" {
		lines = append(lines, component.DetailSection("AGENTS"),
			dimStyle.Render(fitHealthCell("ID", healthUnresolvedID)+fitHealthCell("ROLE", healthUnresolvedRole)+fitHealthCell("STATE", healthUnresolvedSt)+"PANE"))
		for _, node := range unresolved {
			lines = append(lines, fitHealthCell(healthAgentName(node.record), healthUnresolvedID)+fitHealthCell(node.record.Role, healthUnresolvedRole)+
				healthAccentStyle.Render(fitHealthCell(node.record.Status, healthUnresolvedSt))+dimStyle.Render(healthPane(node.record)))
		}
		lines = append(lines, "", component.DetailSection("WHAT IT MEANS"))
		lines = append(lines, strings.Split(ansi.Wordwrap("The record says the agent is live, but hive cannot find its pane. Hive does not change the record by itself.", line, ""), "\n")...)
	} else {
		lines = append(lines, component.DetailSection("DETAIL"))
		lines = append(lines, strings.Split(ansi.Wrap(check.Detail, line, ""), "\n")...)
	}
	if check.Status != doctor.StatusOK {
		lines = append(lines, "", component.DetailSection("NEXT STEP"), dimStyle.Render("Ask the coordinator in 2 Chat. Health only shows state."))
	}
	return check.Name, lines
}

// healthAgentDetails gives the details of the selected agent: its fields,
// its actions, and the restart prompt of decision 19 when it is open.
func (m *Model) healthAgentDetails(line int) (string, []string) {
	node, ok := m.health.selectedAgentNode()
	if !ok {
		return "details", []string{"", "no agent selected"}
	}
	record := node.record
	name := healthAgentName(record)
	word, _, dotStyle, _ := healthAgentState(record)
	restartable := record.Role == "coordinator" || record.Role == "supervisor"
	lines := []string{"", component.DetailField("role", record.Role), component.DetailField("state", dotStyle.Render(fitHealthLine(word, line-10)))}
	if node.pid != "" {
		lines = append(lines, component.DetailField("pid", node.pid))
	}
	if !restartable {
		if record.Project != "" {
			lines = append(lines, component.DetailField("project", record.Project))
		}
		if record.Task != "" {
			lines = append(lines, component.DetailField("task", record.Task))
		}
		if record.Harness != "" {
			lines = append(lines, component.DetailField("harness", record.Harness))
		}
	}
	pane := healthPane(record)
	if restartable && record.PaneID != "" && pane != "—" {
		pane += " · tab " + record.Name
	}
	lines = append(lines, component.DetailField("pane", pane))
	if record.Role == "supervisor" {
		for _, check := range m.health.result.Checks {
			if check.Name == "supervisor binary" {
				binary := healthCheckResult(check)
				if path, _, found := strings.Cut(strings.TrimPrefix(check.Detail, "supervisor "), " modified"); found {
					binary += " · " + path
				}
				lines = append(lines, component.DetailField("binary", binary))
			}
		}
	}
	if restartable && record.StartedByHive {
		lines = append(lines, component.DetailField("started", "by bare hive"))
	}
	lines = append(lines, "", component.DetailSection("ACTIONS"))
	switch record.Role {
	case "supervisor":
		lines = append(lines, titleStyle.Render("r")+" restart · runs hive supervise in a new tab")
	case "coordinator":
		lines = append(lines, titleStyle.Render("r")+" restart · starts a new coordinator in a new tab")
	}
	lines = append(lines, titleStyle.Render("a")+" attach to the pane", titleStyle.Render("p")+" peek at the pane", "")
	if m.agents.peekOpen && m.agents.peekPane == record.PaneID {
		lines = append(lines, component.DetailSection("PEEK · "+record.PaneID))
		lines = append(lines, strings.Split(strings.TrimRight(m.agents.peekText, "\n"), "\n")...)
		lines = append(lines, dimStyle.Render("esc closes the peek"), "")
	}
	if !restartable {
		return name, append(lines, dimStyle.Render("r restarts only the coordinator and the supervisor."))
	}
	if m.agents.restartConfirm && m.agents.restartTarget.Key == record.Key {
		lines = append(lines, "")
		indent := strings.Repeat(" ", min(healthPromptIndent, max(line-healthPromptWidth, 0)))
		for _, row := range strings.Split(restartPrompt(record.Key, min(healthPromptWidth, line)), "\n") {
			lines = append(lines, indent+row)
		}
		lines = append(lines, "")
	}
	return name, append(lines, "",
		dimStyle.Render("r asks to confirm only when the agent is live."),
		dimStyle.Render("A stale or stopped agent restarts at once."))
}

// restartPrompt draws the restart prompt box of the HealthAgents board.
func restartPrompt(key string, width int) string {
	return component.Panel{Title: "restart " + key + "?", Focused: true, Lines: []string{
		"",
		"It is live. Hive stops it and starts a new one.",
		dimStyle.Render("The old record stays on disk as stopped."),
		"",
		titleStyle.Render("y") + " restart   " + titleStyle.Render("esc") + " cancel",
	}}.View(width, healthPromptHeight)
}
