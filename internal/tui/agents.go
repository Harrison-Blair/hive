package tui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Harrison-Blair/hive/internal/agent"
	"github.com/Harrison-Blair/hive/internal/lib/inbox"
	"github.com/Harrison-Blair/hive/internal/lib/process"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
	"github.com/charmbracelet/x/ansi"
)

type agentsTab uint8

const (
	agentsTree agentsTab = iota
	agentsGraph
	agentsTable
)

type agentNode struct {
	record      records.Agent
	status      string
	age         string
	parentKey   string
	inferred    bool
	stopped     bool
	current     bool
	ghost       bool
	landedGhost bool
	landedLabel string
	taskStatus  string
	nextTask    string
	pid         string
	nudged      bool
}

type agentGroup struct {
	id    string
	title string
	keys  []string
}

type agentsData struct {
	nodes          []agentNode
	groups         []agentGroup
	supervisorLive bool
}

type agentsState struct {
	data           agentsData
	restartConfirm bool
	restartTarget  records.Agent
	restartRunning bool
	restartNotice  string
	tab            agentsTab
	selectedKey    string
	showHistory    bool
	scrollOffset   int
	scrollHeight   int
	graphOffset    int
	graphViewWidth int
	explorerOffset int
	peekPane       string
	peekText       string
	peekOpen       bool
}

func loadAgents(home string, now time.Time) (agentsData, error) {
	registry, err := store.Open(home)
	if err != nil {
		return agentsData{}, err
	}
	projects, err := listRecords[records.Project](registry, "projects")
	if err != nil {
		return agentsData{}, fmt.Errorf("projects: %w", err)
	}
	goals, err := listRecords[records.Goal](registry, "goals")
	if err != nil {
		return agentsData{}, fmt.Errorf("goals: %w", err)
	}
	tasks, err := listRecords[records.Task](registry, "tasks")
	if err != nil {
		return agentsData{}, fmt.Errorf("tasks: %w", err)
	}
	agentEnvelope := agent.List(home, agent.ListOptions{All: true})
	if agentEnvelope.Outcome != "ok" {
		return agentsData{}, fmt.Errorf("agents: %v", agentEnvelope.Error)
	}
	agents, ok := agentEnvelope.Result.([]records.Agent)
	if !ok {
		return agentsData{}, errors.New("agents: unexpected result")
	}
	fresh, supervisorClaim, err := supervisorIsLive(registry)
	if err != nil {
		return agentsData{}, err
	}

	activeGoalTasks := make(map[string]bool)
	for _, goal := range goals {
		if goal.Status != "active" {
			continue
		}
		for _, wave := range goal.Waves {
			for _, taskID := range wave.Tasks {
				activeGoalTasks[taskID] = true
			}
		}
	}
	activeTasks := make(map[string]bool, len(tasks))
	taskByID := make(map[string]records.Task, len(tasks))
	landedTaskLabels := make(map[string]string)
	nextTasks := make(map[string]string)
	for _, task := range tasks {
		taskByID[task.ID] = task
		landed := taskLanded(task)
		open := false
		if !landed {
			switch task.Status {
			case "created", "assigned", "completed":
				open = true
			case "verified":
				open = activeGoalTasks[task.ID]
			}
		}
		activeTasks[task.ID] = open
		if landed {
			landedTaskLabels[task.ID] = landedTaskLabel(task)
		}
		for _, dependency := range task.DependsOn {
			if nextTasks[dependency] == "" && task.Status != "cancelled" && !landed {
				nextTasks[dependency] = task.ID
			}
		}
	}
	currentAgents := currentAgentKeys(agents)
	nudgedRoles := make(map[string]bool)
	checkedInboxes := make(map[string]bool)
	for _, record := range agents {
		if !currentAgents[record.Key] || record.Name == "" || checkedInboxes[record.Name] || (record.Role != "coordinator" && record.Role != "manager") {
			continue
		}
		checkedInboxes[record.Name] = true
		inboxRecords, err := inbox.Unread(home, record.Name)
		if err != nil {
			return agentsData{}, fmt.Errorf("%s inbox: %w", record.Name, err)
		}
		for _, item := range inboxRecords {
			if item.NudgedAt != "" {
				nudgedRoles[record.Name] = true
				break
			}
		}
	}
	goalTitle := make(map[string]string)
	for _, goal := range goals {
		if goal.Status == "active" && goalTitle[goal.Project] == "" {
			goalTitle[goal.Project] = goal.Title
		}
	}
	data := agentsData{groups: []agentGroup{{id: "home", title: "home"}}, supervisorLive: fresh}
	groupByProject := make(map[string]int, len(projects))
	for _, project := range projects {
		groupByProject[project.ID] = len(data.groups)
		title := project.ID
		if goalTitle[project.ID] != "" {
			title += " · " + goalTitle[project.ID]
		}
		data.groups = append(data.groups, agentGroup{id: project.ID, title: title})
	}
	managerByProject := make(map[string]records.Agent)
	coordinatorKey := ""
	var coordinatorClaim records.Claim
	if err := registry.Read("claims", "coordinator", &coordinatorClaim); err == nil {
		coordinatorKey = coordinatorClaim.Agent
	} else if !errors.Is(err, os.ErrNotExist) {
		return agentsData{}, err
	}
	for _, record := range agents {
		if record.Role == "manager" {
			current, found := managerByProject[record.Project]
			if !found || (record.Status == records.AgentLive && current.Status != records.AgentLive) ||
				(record.Status == current.Status && record.UpdatedAt > current.UpdatedAt) {
				managerByProject[record.Project] = record
			}
		}
		if record.Role == "coordinator" && coordinatorKey == "" && record.Status == records.AgentLive {
			coordinatorKey = record.Key
		}
	}
	for _, record := range agents {
		groupIndex := 0
		if record.Role != "coordinator" && record.Role != "supervisor" {
			if index, ok := groupByProject[record.Project]; ok {
				groupIndex = index
			}
		}
		parentKey := record.Parent
		inferred := false
		if parentKey == "" {
			switch record.Role {
			case "manager":
				parentKey, inferred = coordinatorKey, true
			case "worker", "verifier":
				parentKey, inferred = managerByProject[record.Project].Key, true
			}
		}
		stopped := record.Status == records.AgentStopped || record.Status == records.AgentExited
		ghost := stopped && (record.Role == "worker" || record.Role == "verifier") && activeTasks[record.Task]
		landedLabel := landedTaskLabels[record.Task]
		landedGhost := stopped && record.Role == "worker" && landedLabel != ""
		status := agentStatus(record, fresh)
		age := activityAge(record.StatusChangedAt, now)
		pid := ""
		if record.Role == "supervisor" {
			if record.StatusChangedAt == "" {
				age = activityAge(record.StartedAt, now)
			}
			if record.ProcessPID > 0 {
				pid = fmt.Sprint(record.ProcessPID)
			} else if supervisorClaim.Agent == record.Key && supervisorClaim.PID > 0 {
				pid = fmt.Sprint(supervisorClaim.PID)
			}
		}
		taskStatus := ""
		if task, ok := taskByID[record.Task]; ok {
			taskStatus = task.Status
		}
		node := agentNode{
			record: record, status: status, age: age, parentKey: parentKey, inferred: inferred, stopped: stopped,
			current: currentAgents[record.Key], ghost: ghost, nudged: currentAgents[record.Key] && nudgedRoles[record.Name],
			landedGhost: landedGhost, landedLabel: landedLabel, taskStatus: taskStatus, nextTask: nextTasks[record.Task], pid: pid,
		}
		data.nodes = append(data.nodes, node)
		data.groups[groupIndex].keys = append(data.groups[groupIndex].keys, record.Key)
	}
	return data, nil
}

func currentAgentKeys(agents []records.Agent) map[string]bool {
	replaced := make(map[string]bool)
	groups := make(map[string][]records.Agent)
	for _, record := range agents {
		if record.Replaces != "" {
			replaced[record.Replaces] = true
		}
		identity := agentIdentity(record)
		groups[identity] = append(groups[identity], record)
	}
	current := make(map[string]bool)
	for _, group := range groups {
		var newest records.Agent
		found := false
		for _, record := range group {
			if replaced[record.Key] {
				continue
			}
			if !found || newerAgent(record, newest) {
				newest = record
				found = true
			}
		}
		if found {
			current[newest.Key] = true
		}
	}
	return current
}

func agentIdentity(record records.Agent) string {
	if record.Parent == "" && record.Replaces == "" {
		return "name\x00" + record.Name
	}
	return strings.Join([]string{"lineage", record.Role, record.Project, record.Task}, "\x00")
}

func newerAgent(left, right records.Agent) bool {
	if comparison := compareAgentTimes(left.StatusChangedAt, right.StatusChangedAt); comparison != 0 {
		return comparison > 0
	}
	if comparison := compareAgentTimes(left.UpdatedAt, right.UpdatedAt); comparison != 0 {
		return comparison > 0
	}
	return left.Key > right.Key
}

func compareAgentTimes(left, right string) int {
	leftTime, leftErr := time.Parse(time.RFC3339Nano, left)
	rightTime, rightErr := time.Parse(time.RFC3339Nano, right)
	if leftErr == nil && rightErr != nil {
		return 1
	}
	if leftErr != nil && rightErr == nil {
		return -1
	}
	if leftErr != nil {
		return strings.Compare(left, right)
	}
	return leftTime.Compare(rightTime)
}

func landedTaskLabel(task records.Task) string {
	number, _, _ := strings.Cut(task.ID, "-")
	if number == "" {
		number = task.ID
	}
	commit := task.Delivery.Commit
	if len(commit) > 7 {
		commit = commit[:7]
	}
	return number + " landed · " + commit
}

func listRecords[T any](registry *store.Store, collection string) ([]T, error) {
	raw, err := registry.List(collection)
	if err != nil {
		return nil, err
	}
	result := make([]T, 0, len(raw))
	for _, data := range raw {
		var record T
		if err := json.Unmarshal(data, &record); err != nil {
			return nil, fmt.Errorf("record invalid: %s: %w", collection, err)
		}
		result = append(result, record)
	}
	return result, nil
}

func supervisorIsLive(registry *store.Store) (bool, records.Claim, error) {
	var claim records.Claim
	if err := registry.Read("claims", "supervisor", &claim); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, claim, nil
		}
		return false, claim, err
	}
	alive, err := process.Alive(claim.PID, claim.PIDStart)
	return err == nil && alive, claim, nil
}

func agentStatus(record records.Agent, supervisorLive bool) string {
	if !supervisorLive {
		return "unknown"
	}
	if record.Role == "supervisor" {
		if record.Status == records.AgentStopped || record.Status == records.AgentExited {
			return "stopped"
		}
		if record.Status != records.AgentLive {
			return "unknown"
		}
		if record.AgentStatus != "" {
			switch record.AgentStatus {
			case "working":
				return "working"
			case "idle", "done":
				return "idle"
			case "blocked":
				return "blocked"
			}
		}
		return "working"
	}
	if record.Status == records.AgentStopped || record.Status == records.AgentExited {
		return "stopped"
	}
	if record.Status == records.AgentStale {
		return "unknown"
	}
	switch record.AgentStatus {
	case "working":
		return "working"
	case "idle", "done":
		return "idle"
	case "blocked":
		return "blocked"
	default:
		return "unknown"
	}
}

func activityAge(timestamp string, now time.Time) string {
	changedAt, err := time.Parse(time.RFC3339Nano, timestamp)
	if err != nil || timestamp == "" {
		return "?"
	}
	age := now.Sub(changedAt)
	if age < 0 {
		age = 0
	}
	switch {
	case age < time.Minute:
		return "0m"
	case age < time.Hour:
		return fmt.Sprintf("%dm", int(age/time.Minute))
	case age < 24*time.Hour:
		return fmt.Sprintf("%dh", int(age/time.Hour))
	default:
		return fmt.Sprintf("%dd", int(age/(24*time.Hour)))
	}
}

func (s agentsState) visibleNodes() []agentNode {
	nodes := s.nodeMap()
	visible := make([]agentNode, 0, len(s.data.nodes))
	seen := make(map[string]bool, len(s.data.nodes))
	for _, group := range s.data.groups {
		for _, root := range s.groupRootsFrom(group, nodes) {
			appendVisibleNodes(&visible, seen, root, nodes, s)
		}
	}
	for _, node := range s.data.nodes {
		if !seen[node.record.Key] && s.nodeVisible(node) {
			visible = append(visible, node)
		}
	}
	return visible
}

func (s agentsState) nodeVisible(node agentNode) bool {
	return s.showHistory || (node.current && (!node.stopped || node.ghost))
}

func appendVisibleNodes(visible *[]agentNode, seen map[string]bool, node agentNode, nodes map[string]agentNode, state agentsState) {
	if seen[node.record.Key] || !state.nodeVisible(node) {
		return
	}
	seen[node.record.Key] = true
	*visible = append(*visible, node)
	for _, child := range state.children(node.record.Key) {
		appendVisibleNodes(visible, seen, child, nodes, state)
	}
}

func (s *agentsState) accept(data agentsData) {
	selected := s.selectedKey
	s.data = data
	for _, node := range s.visibleNodes() {
		if node.record.Key == selected {
			return
		}
	}
	visible := s.visibleNodes()
	s.selectedKey = ""
	if len(visible) != 0 {
		s.selectedKey = visible[0].record.Key
	}
	s.scrollOffset = 0
}

func (s agentsState) selected() (agentNode, bool) {
	for _, node := range s.visibleNodes() {
		if node.record.Key == s.selectedKey {
			return node, true
		}
	}
	return agentNode{}, false
}

func (s *agentsState) moveSelection(delta int) {
	visible := s.visibleNodes()
	if len(visible) == 0 {
		s.selectedKey = ""
		return
	}
	index := 0
	for i, node := range visible {
		if node.record.Key == s.selectedKey {
			index = i
			break
		}
	}
	index += delta
	if index < 0 {
		index = 0
	}
	if index >= len(visible) {
		index = len(visible) - 1
	}
	s.selectedKey = visible[index].record.Key
	s.keepSelectedVisible()
}

func (s *agentsState) keepSelectedVisible() {
	if s.scrollHeight < 1 {
		return
	}
	lines, positions := s.contentLines()
	selectedLine, ok := positions[s.selectedKey]
	if !ok {
		return
	}
	if selectedLine < s.scrollOffset {
		s.scrollOffset = selectedLine
	} else if selectedLine >= s.scrollOffset+s.scrollHeight {
		s.scrollOffset = selectedLine - s.scrollHeight + 1
	}
	maxOffset := len(lines) - s.scrollHeight
	if maxOffset < 0 {
		maxOffset = 0
	}
	if s.scrollOffset > maxOffset {
		s.scrollOffset = maxOffset
	}
}

type agentsScreen struct{ model *Model }

func (s agentsScreen) DataRequest() dataRequest { return agentsDataRequest }

func (s agentsScreen) KeyHints() string {
	return "←↑↓→ move · tab view · enter peek · a attach · h hide stopped · t task · r restart coordinator/supervisor · ask in Chat"
}

func (s agentsScreen) Update(msg tea.Msg) tea.Cmd {
	press, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil
	}
	state := &s.model.agents
	key := press.String()
	if state.restartConfirm {
		target := state.restartTarget
		state.restartConfirm = false
		if key == "y" {
			state.restartRunning = true
			state.restartNotice = "restarting"
			return batch(
				s.model.logAction("agent restart confirmation", slog.String("agent_id", target.Key), slog.String("result", "confirmed")),
				s.model.startAgentRestart(target),
			)
		}
		state.restartNotice = "restart cancelled"
		return s.model.logAction("agent restart confirmation", slog.String("agent_id", target.Key), slog.String("result", "cancelled"))
	}
	if state.restartRunning {
		return nil
	}
	if state.peekOpen {
		if key == "esc" {
			paneID := state.peekPane
			state.peekOpen = false
			state.peekText = ""
			state.peekPane = ""
			return s.model.logAction("agent peek close", slog.String("pane_id", paneID))
		}
		return nil
	}
	switch key {
	case "tab":
		state.tab = (state.tab + 1) % 3
		state.scrollOffset = 0
		state.graphOffset = 0
		state.keepSelectedVisible()
		return s.model.logAction("agent tab", slog.String("tab", []string{"tree", "graph", "table"}[state.tab]))
	case "r":
		node, ok := state.selected()
		if !ok || (node.record.Role != "coordinator" && node.record.Role != "supervisor") {
			state.restartNotice = "restart applies only to the coordinator and the supervisor"
			return s.model.logAction("agent restart request", slog.String("result", "unsupported target"))
		}
		if node.record.Status == records.AgentStarting {
			state.restartNotice = "cannot restart an agent while it is starting"
			return s.model.logAction("agent restart request", slog.String("agent_id", node.record.Key), slog.String("result", "agent is starting"))
		}
		state.restartNotice = ""
		state.restartTarget = node.record
		if node.record.Role == "coordinator" && node.record.Status == records.AgentLive {
			state.restartRunning = true
			state.restartNotice = "checking coordinator pane"
			return batch(
				s.model.logAction("agent restart request", slog.String("agent_id", node.record.Key), slog.String("role", node.record.Role), slog.String("result", "checking pane")),
				s.model.startCoordinatorRestartCheck(node.record),
			)
		}
		if node.record.Status == records.AgentLive && (node.record.Role != "supervisor" || state.data.supervisorLive) {
			state.restartConfirm = true
			state.restartNotice = "restart " + node.record.Role + "? y/n"
			return s.model.logAction("agent restart request", slog.String("agent_id", node.record.Key), slog.String("role", node.record.Role), slog.String("result", "confirmation required"))
		}
		state.restartRunning = true
		state.restartNotice = "restarting"
		return batch(
			s.model.logAction("agent restart request", slog.String("agent_id", node.record.Key), slog.String("role", node.record.Role), slog.String("result", "started immediately")),
			s.model.startAgentRestart(node.record),
		)
	case "h":
		state.showHistory = !state.showHistory
		if _, ok := state.selected(); !ok {
			state.accept(state.data)
		}
		return s.model.logAction("agent history visibility", slog.Bool("history", state.showHistory))
	case "up", "k", "left":
		previous := state.selectedKey
		if state.tab == agentsGraph && key == "left" {
			state.moveGraphSelection(-1, 0)
		} else if state.tab == agentsGraph && key == "up" {
			state.moveGraphSelection(0, -1)
		} else {
			state.moveSelection(-1)
		}
		if state.selectedKey != previous {
			return s.model.logAction("agent selection", slog.String("agent_id", state.selectedKey))
		}
	case "down", "j", "right":
		previous := state.selectedKey
		if state.tab == agentsGraph && key == "right" {
			state.moveGraphSelection(1, 0)
		} else if state.tab == agentsGraph && key == "down" {
			state.moveGraphSelection(0, 1)
		} else {
			state.moveSelection(1)
		}
		if state.selectedKey != previous {
			return s.model.logAction("agent selection", slog.String("agent_id", state.selectedKey))
		}
	case "enter":
		if selected, ok := state.selected(); ok && selected.record.PaneID != "" {
			paneID := selected.record.PaneID
			return batch(s.model.logAction("agent peek", slog.String("pane_id", paneID)), s.model.startAgentsPeek(paneID))
		}
	case "a":
		if selected, ok := state.selected(); ok && selected.record.PaneID != "" {
			paneID := selected.record.PaneID
			return batch(s.model.logAction("agent attach", slog.String("pane_id", paneID)), s.model.attachCmd(paneID))
		}
	case "t":
		if selected, ok := state.selected(); ok && selected.record.Task != "" {
			s.model.screenIndex = 4
			s.model.tasks.tab = tasksAll
			s.model.tasks.selectedID = selected.record.Task
			s.model.tasks.scrollOffset = 0
			s.model.screenLoading = false
			return batch(s.model.logAction("agent open task", slog.String("task_id", selected.record.Task)), s.model.startCurrentScreenLoad())
		}
	}
	return nil
}

type agentRestartMsg struct {
	agentKey string
	result   agent.RestartResult
	err      error
}

type coordinatorRestartCheckMsg struct {
	record records.Agent
	gone   bool
	err    error
}

func (m *Model) startCoordinatorRestartCheck(record records.Agent) tea.Cmd {
	home, logger := m.home, m.logger
	return func() tea.Msg {
		gone, err := coordinatorPaneGone(context.Background(), home, record, logger)
		return coordinatorRestartCheckMsg{record: record, gone: gone, err: err}
	}
}

func (m *Model) acceptCoordinatorRestartCheck(msg coordinatorRestartCheckMsg) tea.Cmd {
	m.invalidateView()
	state := &m.agents
	if !state.restartRunning || state.restartTarget.Key != msg.record.Key {
		return nil
	}
	if msg.gone {
		target := state.restartTarget
		target.Status = records.AgentStale
		state.restartTarget = target
		state.restartNotice = "restarting"
		return batch(
			m.logAction("agent restart pane check", slog.String("agent_id", target.Key), slog.String("outcome", "gone")),
			m.logAction("agent restart request", slog.String("agent_id", target.Key), slog.String("role", target.Role), slog.String("result", "started immediately")),
			m.startAgentRestart(target),
		)
	}
	state.restartRunning = false
	state.restartConfirm = true
	state.restartNotice = "restart coordinator? y/n"
	outcome := "present"
	attrs := []slog.Attr{slog.String("agent_id", msg.record.Key), slog.String("outcome", outcome)}
	if msg.err != nil {
		outcome = "failed"
		attrs[1] = slog.String("outcome", outcome)
		attrs = append(attrs, slog.String("error", boundedError(msg.err.Error())))
	}
	return batch(
		m.logAction("agent restart pane check", attrs...),
		m.logAction("agent restart request", slog.String("agent_id", msg.record.Key), slog.String("role", msg.record.Role), slog.String("result", "confirmation required")),
		m.logError("coordinator pane check", msg.err),
	)
}

func (m *Model) startAgentRestart(record records.Agent) tea.Cmd {
	home, logger := m.home, m.logger
	return func() tea.Msg {
		result, err := agent.Restart(context.Background(), home, record, logger)
		return agentRestartMsg{agentKey: record.Key, result: result, err: err}
	}
}

func (s agentsScreen) View(width, height int) string {
	state := &s.model.agents
	contentWidth := width - 4
	if contentWidth < 1 {
		contentWidth = 1
	}
	if state.peekOpen {
		return renderAgentPeek(state, contentWidth, height-3)
	}
	if state.tab == agentsGraph {
		state.graphViewWidth = contentWidth
	}
	tabLine := agentsTabLine(state.tab)
	var lines []string
	var graphPositions map[string]graphPosition
	if state.tab == agentsGraph {
		lines, graphPositions = state.graphLayout()
	} else {
		lines, _ = state.contentLines()
	}
	selected, hasSelected := state.selected()
	var detailLines []string
	if hasSelected {
		detailLines = agentDetailStrip(state.detailLine(selected), contentWidth)
	}
	visibleHeight := height - 4 - len(detailLines)
	if visibleHeight < 1 {
		visibleHeight = 1
	}
	state.scrollHeight = visibleHeight
	state.keepSelectedVisible()
	var graphStyles [][]string
	if state.tab == agentsGraph {
		state.keepGraphSelectedVisible(contentWidth)
		graphStyles = state.graphColorStyles(lines, graphPositions)
	}
	end := state.scrollOffset + visibleHeight
	if end > len(lines) {
		end = len(lines)
	}
	if state.scrollOffset > len(lines) {
		state.scrollOffset = 0
	}
	if state.tab == agentsGraph {
		legend := dimStyle.Render("edges: ─ delegates · ┄ nudges · ⋯ replaced")
		gap := contentWidth - lipgloss.Width(tabLine) - lipgloss.Width(legend)
		if gap > 0 {
			tabLine += strings.Repeat(" ", gap) + legend
		} else {
			tabLine += "\n" + lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Right).Render(legend)
		}
	}
	body := []string{tabLine}
	for index, line := range lines[state.scrollOffset:end] {
		if state.tab == agentsGraph {
			line = colorGraphLine(line, graphStyles[state.scrollOffset+index])
			line = ansi.Cut(line, state.graphOffset, state.graphOffset+contentWidth)
		} else {
			line = trimGraphLine(line, contentWidth)
		}
		body = append(body, line)
	}
	body = append(body, detailLines...)
	return strings.Join(body, "\n")
}

func agentsTabLine(selected agentsTab) string {
	labels := [...]string{"TREE", "GRAPH", "TABLE"}
	parts := make([]string, 0, len(labels))
	for index, label := range labels {
		if agentsTab(index) == selected {
			parts = append(parts, selectedStyle.Render("▶ [ "+label+" ]"))
		} else {
			parts = append(parts, dimStyle.Render(label))
		}
	}
	return strings.Join(parts, "  ")
}

func (s agentsState) explorerLines() ([]string, map[string]int) {
	lines := make([]string, 0, len(s.data.nodes)+len(s.data.groups))
	positions := make(map[string]int, len(s.data.nodes))
	nodes := s.nodeMap()
	for _, group := range s.data.groups {
		title := group.id
		if group.id == "home" {
			title = "home"
			lines = append(lines, "▾ "+title)
		} else {
			working := 0
			for _, key := range group.keys {
				if node, ok := nodes[key]; ok && node.status == "working" && !node.stopped {
					working++
				}
			}
			lines = append(lines, "▾ "+title+" "+dimStyle.Render(fmt.Sprintf("· %d working", working)))
		}
		roots := s.groupRootsFrom(group, nodes)
		if group.id == "home" {
			roots = roots[:0]
			for _, key := range group.keys {
				node := nodes[key]
				if s.nodeVisible(node) && (node.record.Role == "coordinator" || node.record.Role == "supervisor") {
					roots = append(roots, node)
				}
			}
		}
		for _, root := range roots {
			appendExplorerNode(&lines, positions, root, nodes, s, "  ", group.id != "home")
		}
	}
	return lines, positions
}

func appendExplorerNode(lines *[]string, positions map[string]int, node agentNode, nodes map[string]agentNode, state agentsState, prefix string, expand bool) {
	positions[node.record.Key] = len(*lines)
	mark := statusMark(node.status)
	if state.showHistory && node.landedGhost {
		mark = "✓"
	}
	label := mark + " " + node.record.Name
	if node.record.Key == state.selectedKey {
		label = selectedStyle.Render("▶ " + label)
	} else if node.stopped || !node.current {
		label = dimStyle.Render(label)
	} else {
		label = agentStatusStyle(node).Render(mark) + " " + node.record.Name
	}
	*lines = append(*lines, prefix+label)
	if !expand {
		return
	}
	for _, child := range state.visibleChildren(node.record.Key) {
		appendExplorerNode(lines, positions, child, nodes, state, prefix+"  ", true)
	}
}

func (s agentsState) treeLines() ([]string, map[string]int) {
	lines := make([]string, 0, len(s.data.nodes)+len(s.data.groups))
	positions := make(map[string]int, len(s.data.nodes))
	nodes := s.nodeMap()
	for _, group := range s.data.groups {
		lines = append(lines, "▾ "+group.title)
		roots := s.groupRootsFrom(group, nodes)
		for i, root := range roots {
			last := i == len(roots)-1
			appendTreeNode(&lines, &positions, root, nodes, s, "  ", last)
		}
	}
	return lines, positions
}

func appendTreeNode(lines *[]string, positions *map[string]int, node agentNode, nodes map[string]agentNode, state agentsState, prefix string, last bool) {
	if !state.nodeVisible(node) {
		return
	}
	(*positions)[node.record.Key] = len(*lines)
	branch := "├─ "
	if last {
		branch = "└─ "
	}
	if node.inferred {
		branch = "┄┄ "
	}
	*lines = append(*lines, prefix+branch+state.treeLabel(node))
	children := state.visibleChildren(node.record.Key)
	for index, child := range children {
		childPrefix := prefix + "│  "
		if last {
			childPrefix = prefix + "   "
		}
		appendTreeNode(lines, positions, child, nodes, state, childPrefix, index == len(children)-1)
	}
}

type graphPosition struct {
	row   int
	col   int
	width int
}

func (s agentsState) graphLayout() ([]string, map[string]graphPosition) {
	type pod struct {
		lines     []string
		positions map[string]int
		width     int
		cardWidth int
		column    int
	}
	pods := make([]pod, 0, len(s.data.groups))
	nodes := s.nodeMap()
	rowCount := 0
	width := 0
	columnWidth := 0
	for _, group := range s.data.groups {
		_, _, podWidth, _ := s.graphPod(group, nodes)
		columnWidth = max(columnWidth, podWidth)
	}
	for _, group := range s.data.groups {
		podLines, nodeRows, podWidth, cardWidth := s.graphPodAtWidth(group, nodes, columnWidth)
		pods = append(pods, pod{lines: podLines, positions: nodeRows, width: podWidth, cardWidth: cardWidth, column: width})
		rowCount = max(rowCount, len(podLines))
		width += podWidth + 3
	}
	if len(pods) > 0 {
		width -= 3
	}
	lines := make([]string, rowCount)
	positions := make(map[string]graphPosition, len(s.data.nodes))
	for index, item := range pods {
		for row := range lines {
			part := ""
			if row < len(item.lines) {
				part = item.lines[row]
			}
			lines[row] += part + strings.Repeat(" ", item.width-lipgloss.Width(part))
		}
		if index+1 < len(pods) {
			for row := range lines {
				lines[row] += "   "
			}
		}
		cardOffset := (item.width - item.cardWidth) / 2
		for key, row := range item.positions {
			positions[key] = graphPosition{row: row, col: item.column + cardOffset, width: item.cardWidth}
		}
	}

	anchors := make(map[int]bool)
	coordinatorCol := -1
	managerCols := make([]int, 0)
	for _, node := range s.data.nodes {
		if !s.nodeVisible(node) || (node.record.Role != "coordinator" && node.record.Role != "manager") {
			continue
		}
		position, ok := positions[node.record.Key]
		if !ok {
			continue
		}
		center := position.col + position.width/2
		anchors[center] = true
		if node.record.Role == "coordinator" {
			coordinatorCol = center
		} else {
			managerCols = append(managerCols, center)
		}
	}
	busRows := make([]string, 0, 2)
	if coordinatorCol >= 0 && len(managerCols) > 0 {
		minimum, maximum := coordinatorCol, coordinatorCol
		for _, managerCol := range managerCols {
			minimum = min(minimum, managerCol)
			maximum = max(maximum, managerCol)
		}
		busRows = append(busRows, graphBusLine(width, minimum, maximum, anchors), graphStemLine(width, anchors))
	}
	var supervisor agentNode
	hasSupervisor := false
	for _, node := range s.data.nodes {
		if node.record.Role == "supervisor" && node.current && s.nodeVisible(node) {
			supervisor, hasSupervisor = node, true
			break
		}
	}
	type nudgeRoute struct{ source, target, label int }
	crossNudges := make([]nudgeRoute, 0)
	if hasSupervisor {
		for _, target := range s.data.nodes {
			if !target.nudged || !s.nodeVisible(target) || groupOf(s.data.groups, supervisor.record.Key) == groupOf(s.data.groups, target.record.Key) {
				continue
			}
			sourceGroup := groupOf(s.data.groups, supervisor.record.Key)
			targetGroup := groupOf(s.data.groups, target.record.Key)
			if sourceGroup < 0 || targetGroup < 0 {
				continue
			}
			sourcePosition := positions[supervisor.record.Key]
			targetPosition := positions[target.record.Key]
			sourceColumn := pods[sourceGroup].column + pods[sourceGroup].width - 2
			targetColumn := pods[targetGroup].column + 1
			if sourceGroup < targetGroup {
				lines[sourcePosition.row] = setGraphCell(lines[sourcePosition.row], sourcePosition.col+sourcePosition.width-1, '├')
				lines[sourcePosition.row] = setGraphCell(lines[sourcePosition.row], sourceColumn, '┄')
				lines[targetPosition.row] = setGraphCell(lines[targetPosition.row], targetPosition.col, '┤')
				lines[targetPosition.row] = setGraphCell(lines[targetPosition.row], targetColumn, '┄')
			} else {
				lines[sourcePosition.row] = setGraphCell(lines[sourcePosition.row], sourcePosition.col, '┤')
				lines[sourcePosition.row] = setGraphCell(lines[sourcePosition.row], sourceColumn, '┄')
				lines[targetPosition.row] = setGraphCell(lines[targetPosition.row], targetPosition.col+targetPosition.width-1, '├')
				lines[targetPosition.row] = setGraphCell(lines[targetPosition.row], targetColumn, '┄')
			}
			for row := sourcePosition.row + 1; row < rowCount; row++ {
				mark := '│'
				if row == len(pods[sourceGroup].lines)-1 {
					mark = '┼'
				}
				lines[row] = setGraphCell(lines[row], sourceColumn, mark)
			}
			for row := targetPosition.row + 1; row < rowCount; row++ {
				mark := '│'
				if row == len(pods[targetGroup].lines)-1 {
					mark = '┼'
				}
				lines[row] = setGraphCell(lines[row], targetColumn, mark)
			}
			sourceCenter := sourcePosition.col + sourcePosition.width/2
			targetCenter := targetPosition.col + targetPosition.width/2
			crossNudges = append(crossNudges, nudgeRoute{source: sourceColumn, target: targetColumn, label: (sourceCenter + targetCenter) / 2})
		}
	}
	for column := range anchors {
		lines[0] = setGraphRune(lines[0], column, '┬')
	}
	graph := append(busRows, lines...)
	for key, position := range positions {
		position.row += len(busRows)
		positions[key] = position
	}
	for _, route := range crossNudges {
		branches := map[int]bool{route.source: true, route.target: true}
		graph = append(graph, graphStemLine(width, branches), graphNudgeLine(width, route.source, route.target), graphNudgeLabelLine(width, route.label))
	}
	return graph, positions
}

func (s agentsState) graphColorStyles(lines []string, positions map[string]graphPosition) [][]string {
	styles := make([][]string, len(lines))
	for row, line := range lines {
		styles[row] = make([]string, ansi.StringWidth(line))
	}
	setCell := func(row, column int, color string) {
		if row >= 0 && row < len(styles) && column >= 0 && column < len(styles[row]) {
			styles[row][column] = color
		}
	}
	setRange := func(row, start, end int, color string) {
		for column := start; column < end; column++ {
			setCell(row, column, color)
		}
	}

	for row, line := range lines {
		column := 0
		for _, mark := range ansi.Strip(line) {
			if strings.ContainsRune("─│┌┐└┘┬┴├┤┼┄⋯╭╮╰╯", mark) {
				setCell(row, column, colorWorking)
			}
			column += ansi.StringWidth(string(mark))
		}
	}

	busRows := 0
	if len(lines) > 0 && strings.Contains(ansi.Strip(lines[0]), "┌") {
		busRows = 2
	}
	columnWidth := 0
	nodes := s.nodeMap()
	for _, group := range s.data.groups {
		_, _, width, _ := s.graphPod(group, nodes)
		columnWidth = max(columnWidth, width)
	}
	column := 0
	for _, group := range s.data.groups {
		pod, _, width, _ := s.graphPodAtWidth(group, nodes, columnWidth)
		for row, line := range pod {
			graphRow := busRows + row
			if row == 0 || row == len(pod)-1 {
				cell := 0
				for _, mark := range ansi.Strip(line) {
					if mark != ' ' {
						setCell(graphRow, column+cell, colorInfo)
					}
					cell += ansi.StringWidth(string(mark))
				}
			} else {
				setCell(graphRow, column, colorInfo)
				setCell(graphRow, column+width-1, colorInfo)
			}
		}
		column += width + 3
	}

	for key, position := range positions {
		node := nodes[key]
		live := node.current && node.record.Status == records.AgentLive && !node.stopped
		borderColor := ""
		if key == s.selectedKey {
			borderColor = colorAccent
		} else if live {
			borderColor = colorWorking
		}
		for column := position.col; column < position.col+position.width; column++ {
			setCell(position.row-1, column, borderColor)
			setCell(position.row+2, column, borderColor)
		}
		for row := position.row; row <= position.row+1; row++ {
			setCell(row, position.col, borderColor)
			setCell(row, position.col+position.width-1, borderColor)
		}
		if !live {
			continue
		}
		if node.status == "working" && key != s.selectedKey {
			setCell(position.row, position.col+2, colorWorking)
		}
		if node.record.Role == "supervisor" {
			continue
		}
		switch graphHarness(node.record.Harness) {
		case "cc":
			setRange(position.row+1, position.col+2, position.col+4, colorAccent)
		case "pi":
			setRange(position.row+1, position.col+2, position.col+4, colorInfo)
		}
	}
	return styles
}

func colorGraphLine(line string, styles []string) string {
	width := min(ansi.StringWidth(line), len(styles))
	var rendered strings.Builder
	for start := 0; start < width; {
		color := styles[start]
		end := start + 1
		for end < width && styles[end] == color {
			end++
		}
		part := ansi.Cut(line, start, end)
		if color != "" {
			part = lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(ansi.Strip(part))
		}
		rendered.WriteString(part)
		start = end
	}
	if width < ansi.StringWidth(line) {
		rendered.WriteString(ansi.Cut(line, width, ansi.StringWidth(line)))
	}
	return rendered.String()
}

func (s agentsState) graphPod(group agentGroup, nodes map[string]agentNode) ([]string, map[string]int, int, int) {
	return s.graphPodAtWidth(group, nodes, 0)
}

func (s agentsState) graphPodAtWidth(group agentGroup, nodes map[string]agentNode, columnWidth int) ([]string, map[string]int, int, int) {
	nodeList := make([]agentNode, 0, len(group.keys))
	var appendNodes func(agentNode)
	appendNodes = func(node agentNode) {
		nodeList = append(nodeList, node)
		for _, child := range s.visibleChildren(node.record.Key) {
			appendNodes(child)
		}
	}
	for _, root := range s.groupRootsFrom(group, nodes) {
		appendNodes(root)
	}
	cardWidth := 18
	for _, node := range nodeList {
		first := "▶ " + statusMark(node.status) + " " + node.record.Name
		second := s.graphNodeSubLabel(node)
		cardWidth = max(cardWidth, lipgloss.Width(first)+4, lipgloss.Width(second)+4)
	}
	if cardWidth%2 != 0 {
		cardWidth++
	}
	title := "≡ " + group.title
	hasDelegationStem := len(nodeList) != 0 && (nodeList[0].record.Role == "coordinator" || nodeList[0].record.Role == "manager")
	width := max(24, lipgloss.Width(title)+6, cardWidth+4, 2*(lipgloss.Width("≡ "+group.id)+6))
	if hasDelegationStem {
		width = max(width, 2*(lipgloss.Width(title)+5))
	}
	if s.graphViewWidth > 0 && width > s.graphViewWidth {
		width = max(s.graphViewWidth, cardWidth+4)
	}
	if columnWidth > 0 {
		width = max(width, columnWidth)
	} else if (width-cardWidth)%2 != 0 {
		width++
	}
	lines := []string{
		graphPodTitleBorderLine(width, title, hasDelegationStem),
		"│" + strings.Repeat(" ", width-2) + "│",
		"│" + strings.Repeat(" ", width-2) + "│",
	}
	if hasDelegationStem {
		lines[1] = graphPodLine(width, width/2, '│')
		if nodeList[0].inferred {
			lines[2] = graphPodLabelLineWithStem(width, "┄ inferred")
		} else {
			lines[2] = graphPodLine(width, width/2, '│')
		}
	}
	positions := make(map[string]int, len(nodeList))
	previous := agentNode{}
	for index, node := range nodeList {
		if index > 0 {
			inner := strings.Repeat(" ", width-2)
			middle := width / 2
			switch {
			case agentsNudgePair(previous, node):
				lines = append(lines, graphPodLabelLine(width, "┄┄ nudges ┄┄"))
				inner = ""
			case node.inferred && previous.record.Key == node.parentKey:
				lines = append(lines, graphPodLabelLine(width, "┄ inferred"))
				inner = ""
			case node.parentKey == previous.record.Key || node.parentKey != "" && node.parentKey == previous.parentKey:
				inner = inner[:middle-1] + "│" + inner[middle:]
			case node.record.Replaces == previous.record.Key || previous.record.Replaces == node.record.Key:
				inner = inner[:middle-1] + "⋯" + inner[middle:]
			}
			if inner != "" {
				lines = append(lines, "│"+inner+"│")
			}
		}
		first := graphNodeMark(node) + " " + node.record.Name
		second := s.graphNodeSubLabel(node)
		if node.record.Key == s.selectedKey {
			first = selectedStyle.Render("▶ " + first)
		} else if node.stopped || !node.current {
			first = dimStyle.Render(first)
			second = dimStyle.Render(second)
		}
		border := "─"
		if node.ghost {
			border = "┄"
		}
		pad := func(text string) string {
			return " " + text + strings.Repeat(" ", max(0, cardWidth-4-lipgloss.Width(text))) + " "
		}
		card := []string{
			"┌" + strings.Repeat(border, cardWidth-2) + "┐",
			"│" + pad(first) + "│",
			"│" + pad(second) + "│",
			"└" + strings.Repeat(border, cardWidth-2) + "┘",
		}
		middle := cardWidth / 2
		if node.parentKey != "" || node.record.Role == "coordinator" || node.record.Role == "manager" {
			card[0] = setGraphRune(card[0], middle, '┬')
		}
		if len(s.visibleChildren(node.record.Key)) != 0 || index+1 < len(nodeList) && nodeList[index+1].parentKey == node.parentKey && node.parentKey != "" {
			card[3] = setGraphRune(card[3], middle, '┴')
		}
		positions[node.record.Key] = len(lines) + 1
		cardOffset := (width - cardWidth) / 2
		leftPadding := cardOffset - 1
		rightPadding := width - cardOffset - cardWidth - 1
		for _, cardLine := range card {
			lines = append(lines, "│"+strings.Repeat(" ", leftPadding)+cardLine+strings.Repeat(" ", rightPadding)+"│")
		}
		previous = node
	}
	if len(nodeList) == 0 {
		lines = append(lines, "│"+strings.Repeat(" ", width-2)+"│")
	}
	lines = append(lines, "╰"+strings.Repeat("─", width-2)+"╯")
	return lines, positions, width, cardWidth
}

func (s agentsState) graphNodeSubLabel(node agentNode) string {
	harness := graphHarness(node.record.Harness)
	activity := node.record.Task
	switch node.record.Role {
	case "supervisor":
		return "pid " + emptyQuestion(node.pid) + " · " + emptyQuestion(node.age)
	case "coordinator":
		activity = "you · now"
	case "manager":
		if activity == "" {
			activity = node.status + " · " + node.age
		}
	default:
		if activity == "" {
			activity = node.age
		}
	}
	label := emptyQuestion(harness) + " " + emptyQuestion(activity)
	if s.showHistory && node.landedGhost {
		label += " · " + node.landedLabel
	}
	return label
}

func graphHarness(harness string) string {
	if harness == "claude" {
		return "cc"
	}
	return harness
}

func agentsNudgePair(left, right agentNode) bool {
	return left.record.Role == "supervisor" && right.nudged || right.record.Role == "supervisor" && left.nudged
}

func graphPodLine(width, middle int, mark rune) string {
	line := "│" + strings.Repeat(" ", width-2) + "│"
	return setGraphRune(line, middle, mark)
}

func graphPodTitleBorderLine(width int, title string, hasStem bool) string {
	cells := []rune("╭" + strings.Repeat("─", width-2) + "╮")
	label := " " + title + " "
	limit := width - 4
	if hasStem {
		limit = width/2 - 3
		cells[width/2] = '┬'
	}
	if lipgloss.Width(label) > limit {
		label = ansi.Truncate(" "+title, limit-1, "…") + " "
	}
	copy(cells[2:], []rune(label))
	return string(cells)
}

func graphTextAroundStem(cells []rune, text string, center, start, end int) {
	leftEnd := max(start, center-1)
	rightStart := min(end, center+2)
	textRunes := []rune(text)
	leftCount := min(len(textRunes), leftEnd-start)
	rightTextStart := leftCount
	if leftCount < len(textRunes) {
		for split := leftCount; split > 1; split-- {
			if textRunes[split-1] == ' ' {
				leftCount = split - 1
				rightTextStart = split
				break
			}
			if textRunes[split] == ' ' {
				leftCount = split
				rightTextStart = split + 1
				break
			}
		}
	}
	copy(cells[start:leftEnd], textRunes[:leftCount])
	copy(cells[rightStart:end], textRunes[rightTextStart:])
	cells[center-1] = ' '
	cells[center+1] = ' '
}

func graphPodLabelLine(width int, label string) string {
	cells := []rune(strings.Repeat(" ", width-2))
	start := max(0, (len(cells)-len([]rune(label)))/2)
	copy(cells[start:], []rune(label))
	return "│" + string(cells) + "│"
}

func graphPodLabelLineWithStem(width int, label string) string {
	cells := []rune("│" + strings.Repeat(" ", width-2) + "│")
	center := width / 2
	graphTextAroundStem(cells, label, center, 2, width-1)
	cells[center] = '│'
	return string(cells)
}

func graphBusLine(width, minimum, maximum int, anchors map[int]bool) string {
	cells := []rune(strings.Repeat(" ", width))
	for column := minimum; column <= maximum && column < len(cells); column++ {
		if column >= 0 {
			cells[column] = '─'
		}
	}
	for column := range anchors {
		if column >= 0 && column < len(cells) {
			cells[column] = '┬'
		}
	}
	if minimum >= 0 && minimum < len(cells) {
		cells[minimum] = '┌'
	}
	if maximum >= 0 && maximum < len(cells) {
		cells[maximum] = '┐'
	}
	return string(cells)
}

func graphStemLine(width int, anchors map[int]bool) string {
	cells := []rune(strings.Repeat(" ", width))
	for column := range anchors {
		if column >= 0 && column < len(cells) {
			cells[column] = '│'
		}
	}
	return string(cells)
}

func graphNudgeLabelLine(width, center int) string {
	label := "nudges"
	start := max(0, min(width-len(label), center-len(label)/2))
	return strings.Repeat(" ", start) + label
}

func graphNudgeLine(width, left, right int) string {
	if left > right {
		left, right = right, left
	}
	cells := []rune(strings.Repeat(" ", width))
	for column := left; column <= right && column < len(cells); column++ {
		if column >= 0 {
			cells[column] = '┄'
		}
	}
	return string(cells)
}

func setGraphCell(line string, column int, value rune) string {
	width := lipgloss.Width(line)
	if column < 0 || column >= width {
		return line
	}
	return ansi.Cut(line, 0, column) + string(value) + ansi.Cut(line, column+1, width)
}

func setGraphRune(line string, column int, value rune) string {
	cells := []rune(line)
	if column >= 0 && column < len(cells) {
		cells[column] = value
	}
	return string(cells)
}

func graphNodeMark(node agentNode) string {
	if node.landedGhost {
		return "✓"
	}
	return statusMark(node.status)
}

func (s agentsState) graphLines() ([]string, map[string]int) {
	lines, graphPositions := s.graphLayout()
	positions := make(map[string]int, len(graphPositions))
	for key, position := range graphPositions {
		positions[key] = position.row
	}
	return lines, positions
}

func (s *agentsState) keepGraphSelectedVisible(width int) {
	if width < 1 {
		return
	}
	lines, positions := s.graphLayout()
	selected, ok := positions[s.selectedKey]
	if ok {
		if selected.col < s.graphOffset {
			s.graphOffset = selected.col
		} else if selected.col+selected.width > s.graphOffset+width {
			s.graphOffset = selected.col + selected.width - width
		}
	}
	maxWidth := 0
	for _, line := range lines {
		maxWidth = max(maxWidth, lipgloss.Width(line))
	}
	maxOffset := max(0, maxWidth-width)
	if s.graphOffset > maxOffset {
		s.graphOffset = maxOffset
	}
}

func (s *agentsState) moveGraphSelection(dx, dy int) {
	_, positions := s.graphLayout()
	current, ok := positions[s.selectedKey]
	if !ok {
		return
	}
	currentGroup := groupOf(s.data.groups, s.selectedKey)
	bestKey := ""
	bestScore := int(^uint(0) >> 1)
	for key, candidate := range positions {
		if key == s.selectedKey {
			continue
		}
		candidateGroup := groupOf(s.data.groups, key)
		if dx != 0 {
			if (dx < 0 && candidate.col >= current.col) || (dx > 0 && candidate.col <= current.col) {
				continue
			}
			score := absInt(candidate.col-current.col) + 2*absInt(candidate.row-current.row)
			if score < bestScore {
				bestKey, bestScore = key, score
			}
			continue
		}
		if candidateGroup != currentGroup || (dy < 0 && candidate.row >= current.row) || (dy > 0 && candidate.row <= current.row) {
			continue
		}
		score := absInt(candidate.row - current.row)
		if score < bestScore {
			bestKey, bestScore = key, score
		}
	}
	if bestKey != "" {
		s.selectedKey = bestKey
		s.keepSelectedVisible()
	}
}

func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

func (s agentsState) tableLines() ([]string, map[string]int) {
	lines := []string{"name · role · project · task · harness · model · status · age · parent"}
	positions := make(map[string]int, len(s.data.nodes))
	for _, node := range s.visibleNodes() {
		positions[node.record.Key] = len(lines)
		parent := s.parentLabel(node)
		line := fmt.Sprintf("%s · %s · %s · %s · %s · %s · %s · %s · %s",
			node.record.Name, emptyDash(node.record.Role), emptyDash(node.record.Project), emptyDash(node.record.Task),
			emptyDash(node.record.Harness), emptyDash(node.record.Model), node.status, node.age, parent)
		if s.showHistory && node.landedGhost {
			line = "✓ " + line + " · " + node.landedLabel
		}
		if node.record.Key == s.selectedKey {
			line = selectedStyle.Render("▶ " + line)
		} else if node.stopped || !node.current {
			line = dimStyle.Render(line)
		}
		lines = append(lines, line)
	}
	return lines, positions
}

func (s agentsState) contentLines() ([]string, map[string]int) {
	switch s.tab {
	case agentsGraph:
		return s.graphLines()
	case agentsTable:
		return s.tableLines()
	default:
		return s.treeLines()
	}
}

func (s agentsState) detailLine(node agentNode) string {
	parent := s.parentLabel(node)
	if parent == "none" {
		parent = "?"
	}
	return fmt.Sprintf("%s  parent %s  model %s  task %s  %s  pane %s  next %s",
		node.record.Name, parent, emptyQuestion(node.record.Model), emptyQuestion(node.record.Task),
		emptyQuestion(node.taskStatus), emptyQuestion(node.record.PaneID), emptyQuestion(node.nextTask))
}

func agentDetailStrip(text string, width int) []string {
	if width < 5 {
		return wrapAgentLine(text, max(1, width))
	}
	innerWidth := width - 4
	wrapped := wrapAgentLine(text, innerWidth)
	lines := make([]string, 0, len(wrapped)+2)
	border := lipgloss.NewStyle().Foreground(lipgloss.Color(colorAccent))
	lines = append(lines, border.Render("╭"+strings.Repeat("─", width-2)+"╮"))
	for index, line := range wrapped {
		line = trimGraphLine(line, innerWidth)
		line = styleAgentDetailLine(line, index == 0)
		lines = append(lines, "│ "+line+strings.Repeat(" ", max(0, innerWidth-lipgloss.Width(line)))+" │")
	}
	lines = append(lines, border.Render("╰"+strings.Repeat("─", width-2)+"╯"))
	return lines
}

func styleAgentDetailLine(line string, first bool) string {
	words := strings.Fields(line)
	var styled strings.Builder
	index := 0
	if first && len(words) > 0 {
		styled.WriteString(selectedStyle.Render(words[0]))
		index = 1
	}
	for index < len(words) {
		if styled.Len() > 0 {
			styled.WriteString("  ")
		}
		if isAgentDetailField(words[index]) && index+1 < len(words) {
			styled.WriteString(dimStyle.Render(words[index]))
			styled.WriteByte(' ')
			styled.WriteString(words[index+1])
			index += 2
		} else {
			styled.WriteString(words[index])
			index++
		}
	}
	return styled.String()
}

func (s agentsState) parentLabel(node agentNode) string {
	if node.inferred {
		return "inferred"
	}
	if node.parentKey == "" {
		return "none"
	}
	return displayAgent(s.nodeMap(), node.parentKey)
}

func (s agentsState) treeLabel(node agentNode) string {
	mark := statusMark(node.status)
	if s.showHistory && node.landedGhost {
		mark = "✓"
	}
	label := mark + " " + node.record.Name + " · " + node.status + " · " + node.age
	if s.showHistory && node.landedGhost {
		label += " · " + node.landedLabel
	}
	if node.record.Replaces != "" {
		label += " ⋯ replaces " + displayAgent(s.nodeMap(), node.record.Replaces)
	}
	if node.record.Key == s.selectedKey {
		return selectedStyle.Render("▶ " + label)
	}
	if node.stopped || !node.current {
		return dimStyle.Render(label)
	}
	return agentStatusStyle(node).Render(mark) + strings.TrimPrefix(label, mark)
}

func (s agentsState) nodeMap() map[string]agentNode {
	result := make(map[string]agentNode, len(s.data.nodes))
	for _, node := range s.data.nodes {
		result[node.record.Key] = node
	}
	return result
}

func (s agentsState) groupRootsFrom(group agentGroup, nodes map[string]agentNode) []agentNode {
	keys := make(map[string]struct{}, len(group.keys))
	for _, key := range group.keys {
		if node, ok := nodes[key]; ok && s.nodeVisible(node) {
			keys[key] = struct{}{}
		}
	}
	roots := make([]agentNode, 0)
	for _, key := range group.keys {
		node, ok := nodes[key]
		if !ok || !s.nodeVisible(node) {
			continue
		}
		if _, hasParent := keys[node.parentKey]; !hasParent {
			roots = append(roots, node)
		}
	}
	return roots
}

func (s agentsState) children(parent string) []agentNode {
	children := make([]agentNode, 0)
	for _, node := range s.data.nodes {
		if node.parentKey == parent && node.record.Key != parent {
			if groupOf(s.data.groups, node.record.Key) == groupOf(s.data.groups, parent) {
				children = append(children, node)
			}
		}
	}
	return children
}

func (s agentsState) visibleChildren(parent string) []agentNode {
	children := make([]agentNode, 0)
	for _, child := range s.children(parent) {
		if s.nodeVisible(child) {
			children = append(children, child)
		}
	}
	return children
}

func groupOf(groups []agentGroup, key string) int {
	for index, group := range groups {
		for _, item := range group.keys {
			if item == key {
				return index
			}
		}
	}
	return -1
}

func displayAgent(nodes map[string]agentNode, key string) string {
	if key == "" {
		return "—"
	}
	if node, ok := nodes[key]; ok && node.record.Name != "" {
		return node.record.Name
	}
	return key
}

func agentStatusStyle(node agentNode) lipgloss.Style {
	if node.stopped || !node.current {
		return dimStyle
	}
	switch node.status {
	case "working":
		return workingStyle
	case "blocked":
		return infoStyle
	default:
		return dimStyle
	}
}

func statusMark(status string) string {
	switch status {
	case "working":
		return "●"
	case "idle":
		return "○"
	case "blocked":
		return "!"
	case "stopped":
		return "×"
	default:
		return "?"
	}
}

func emptyDash(text string) string {
	if text == "" {
		return "—"
	}
	return text
}

func emptyQuestion(text string) string {
	if text == "" {
		return "?"
	}
	return text
}

func wrapAgentLine(text string, width int) []string {
	words := strings.Fields(text)
	lines := make([]string, 0, 1)
	line := ""
	for index := 0; index < len(words); index++ {
		word := words[index]
		if isAgentDetailField(word) && index+1 < len(words) {
			word += " " + words[index+1]
			index++
		}
		if line != "" && lipgloss.Width(line+"  "+word) > width {
			lines = append(lines, line)
			line = ""
		}
		if line == "" {
			line = word
		} else {
			line += "  " + word
		}
	}
	if line != "" {
		lines = append(lines, line)
	}
	return lines
}

func isAgentDetailField(word string) bool {
	switch word {
	case "parent", "replaces", "model", "task", "pane", "age", "next":
		return true
	default:
		return false
	}
}

func trimGraphLine(text string, width int) string {
	if width < 1 {
		return ""
	}
	if lipgloss.Width(text) <= width {
		return text
	}
	var result strings.Builder
	used := 0
	for _, r := range text {
		runeWidth := lipgloss.Width(string(r))
		if used+runeWidth > width-1 {
			break
		}
		result.WriteRune(r)
		used += runeWidth
	}
	result.WriteRune('…')
	return result.String()
}

func renderAgentPeek(state *agentsState, width, height int) string {
	lines := strings.Split(strings.TrimRight(state.peekText, "\n"), "\n")
	if len(lines) > 20 {
		lines = lines[len(lines)-20:]
	}
	if len(lines) == 0 {
		lines = []string{"(empty pane)"}
	}
	content := []string{titleStyle.Render("PEEK · " + state.peekPane)}
	content = append(content, lines...)
	maxLines := height
	if maxLines < 1 {
		maxLines = 1
	}
	if len(content) > maxLines {
		content = content[len(content)-maxLines:]
	}
	for index, line := range content {
		content[index] = trimGraphLine(line, width)
	}
	return strings.Join(content, "\n")
}
