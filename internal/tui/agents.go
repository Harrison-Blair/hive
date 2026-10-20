package tui

import (
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
	tab            agentsTab
	selectedKey    string
	showHistory    bool
	scrollOffset   int
	scrollHeight   int
	graphOffset    int
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
	history := "off"
	if s.model.agents.showHistory {
		history = "on"
	}
	return "↑/↓←/→ move · tab tree/graph/table · h history: " + history + " · enter peek · a attach · t task · ask in Chat"
}

func (s agentsScreen) Update(msg tea.Msg) tea.Cmd {
	press, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil
	}
	state := &s.model.agents
	key := press.String()
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

func (s agentsScreen) View(width, height int) string {
	state := &s.model.agents
	contentWidth := width - 4
	if contentWidth < 1 {
		contentWidth = 1
	}
	if state.peekOpen {
		return renderAgentPeek(state, contentWidth, height-3)
	}
	tabLine := []string{"[ TREE ]  GRAPH  TABLE", "TREE  [ GRAPH ]  TABLE", "TREE  GRAPH  [ TABLE ]"}[state.tab]
	lines, _ := state.contentLines()
	selected, hasSelected := state.selected()
	var detailLines []string
	if hasSelected {
		detailLines = wrapAgentLine(state.detailLine(selected), contentWidth)
	}
	visibleHeight := height - 4 - len(detailLines)
	if visibleHeight < 1 {
		visibleHeight = 1
	}
	state.scrollHeight = visibleHeight
	state.keepSelectedVisible()
	if state.tab == agentsGraph {
		state.keepGraphSelectedVisible(contentWidth)
	}
	end := state.scrollOffset + visibleHeight
	if end > len(lines) {
		end = len(lines)
	}
	if state.scrollOffset > len(lines) {
		state.scrollOffset = 0
	}
	legend := ""
	if state.tab == agentsGraph {
		legend = "  edges: ─ delegates · ┄ nudges · ⋯ replaced"
	}
	body := []string{trimGraphLine(titleStyle.Render(tabLine+legend), contentWidth)}
	for _, line := range lines[state.scrollOffset:end] {
		if state.tab == agentsGraph {
			line = ansi.Cut(line, state.graphOffset, state.graphOffset+contentWidth)
		} else {
			line = trimGraphLine(line, contentWidth)
		}
		body = append(body, line)
	}
	body = append(body, detailLines...)
	return strings.Join(body, "\n")
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
			lines = append(lines, fmt.Sprintf("▾ %s · %d working", title, working))
		}
		roots := s.groupRootsFrom(group, nodes)
		for index, root := range roots {
			appendExplorerNode(&lines, positions, root, nodes, s, "  ", index == len(roots)-1)
		}
	}
	return lines, positions
}

func appendExplorerNode(lines *[]string, positions map[string]int, node agentNode, nodes map[string]agentNode, state agentsState, prefix string, last bool) {
	positions[node.record.Key] = len(*lines)
	branch := "├─ "
	if last {
		branch = "└─ "
	}
	if node.inferred {
		branch = "┄┄ "
	}
	label := statusMark(node.status) + " " + node.record.Name
	if state.showHistory && node.landedGhost {
		label = "✓ " + node.record.Name
	}
	if node.record.Key == state.selectedKey {
		label = selectedStyle.Render("▶ " + label)
	} else if node.stopped || !node.current {
		label = dimStyle.Render(label)
	}
	*lines = append(*lines, prefix+branch+label)
	children := state.children(node.record.Key)
	visibleChildren := make([]agentNode, 0, len(children))
	for _, child := range children {
		if state.nodeVisible(child) {
			visibleChildren = append(visibleChildren, child)
		}
	}
	for index, child := range visibleChildren {
		childPrefix := prefix + "│  "
		if last {
			childPrefix = prefix + "   "
		}
		appendExplorerNode(lines, positions, child, nodes, state, childPrefix, index == len(visibleChildren)-1)
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
	children := state.children(node.record.Key)
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
	}
	pods := make([]pod, 0, len(s.data.groups))
	nodes := s.nodeMap()
	for _, group := range s.data.groups {
		podLines, nodeRows, width, cardWidth := s.graphPod(group, nodes)
		pods = append(pods, pod{lines: podLines, positions: nodeRows, width: width, cardWidth: cardWidth})
	}
	rowCount := 0
	for _, item := range pods {
		rowCount = max(rowCount, len(item.lines))
	}
	lines := make([]string, rowCount)
	positions := make(map[string]graphPosition, len(s.data.nodes))
	column := 0
	for index, item := range pods {
		for row := range lines {
			part := ""
			if row < len(item.lines) {
				part = item.lines[row]
			}
			lines[row] += part + strings.Repeat(" ", item.width-lipgloss.Width(part))
		}
		for key, row := range item.positions {
			positions[key] = graphPosition{row: row, col: column + 2, width: item.cardWidth}
		}
		column += item.width
		if index+1 < len(pods) {
			for row := range lines {
				lines[row] += "   "
			}
			column += 3
		}
	}
	for _, node := range s.data.nodes {
		if !s.nodeVisible(node) {
			continue
		}
		if node.parentKey != "" {
			parent, ok := nodes[node.parentKey]
			if ok && s.nodeVisible(parent) && groupOf(s.data.groups, parent.record.Key) != groupOf(s.data.groups, node.record.Key) {
				label := ""
				if node.inferred {
					label = " · inferred"
				}
				lines = append(lines, parent.record.Name+" ──▶ "+node.record.Name+label)
			}
		}
		if node.record.Replaces != "" {
			lines = append(lines, node.record.Name+" ⋯ replaces "+displayAgent(nodes, node.record.Replaces))
		}
	}
	supervisor := ""
	for _, node := range s.data.nodes {
		if node.record.Role == "supervisor" && node.current && s.nodeVisible(node) {
			supervisor = node.record.Name
			break
		}
	}
	if supervisor != "" {
		for _, node := range s.data.nodes {
			if node.nudged && s.nodeVisible(node) {
				lines = append(lines, supervisor+" ┄┄▶ "+node.record.Name+" · nudge")
			}
		}
	}
	return lines, positions
}

func (s agentsState) graphPod(group agentGroup, nodes map[string]agentNode) ([]string, map[string]int, int, int) {
	nodeList := make([]agentNode, 0, len(group.keys))
	var appendNodes func(agentNode)
	appendNodes = func(node agentNode) {
		nodeList = append(nodeList, node)
		for _, child := range s.children(node.record.Key) {
			if s.nodeVisible(child) {
				appendNodes(child)
			}
		}
	}
	for _, root := range s.groupRootsFrom(group, nodes) {
		appendNodes(root)
	}
	cardWidth := 18
	for _, node := range nodeList {
		first := "▶ " + statusMark(node.status) + " " + node.record.Name
		second := s.graphNodeSubLabel(node)
		cardWidth = max(cardWidth, lipgloss.Width(first)+2, lipgloss.Width(second)+2)
	}
	width := max(24, lipgloss.Width(group.title)+5, cardWidth+4)
	cardWidth = width - 4
	lines := []string{"┌─ " + group.title + " " + strings.Repeat("─", width-lipgloss.Width(group.title)-5) + "┐"}
	positions := make(map[string]int, len(nodeList))
	previous := agentNode{}
	for index, node := range nodeList {
		if index > 0 {
			inner := strings.Repeat(" ", width-2)
			if previous.record.Key == node.parentKey {
				middle := width / 2
				if node.inferred {
					label := "┄ inferred"
					inner = inner[:middle-1] + label + inner[min(len(inner), middle-1+lipgloss.Width(label)):]
				} else {
					inner = inner[:middle-1] + "│" + inner[middle:]
				}
			} else if previous.parentKey == node.parentKey && node.parentKey != "" {
				middle := width / 2
				inner = inner[:middle-1] + "├" + inner[middle:]
			}
			lines = append(lines, "│"+inner+"│")
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
			return text + strings.Repeat(" ", max(0, cardWidth-2-lipgloss.Width(text)))
		}
		card := []string{
			"┌" + strings.Repeat(border, cardWidth-2) + "┐",
			"│" + pad(first) + "│",
			"│" + pad(second) + "│",
			"└" + strings.Repeat(border, cardWidth-2) + "┘",
		}
		for _, cardLine := range card {
			lines = append(lines, "│ "+cardLine+" │")
		}
		positions[node.record.Key] = len(lines) - 3
		previous = node
	}
	if len(nodeList) == 0 {
		lines = append(lines, "│"+strings.Repeat(" ", width-2)+"│")
	}
	lines = append(lines, "└"+strings.Repeat("─", width-2)+"┘")
	return lines, positions, width, cardWidth
}

func (s agentsState) graphNodeSubLabel(node agentNode) string {
	if node.record.Role == "supervisor" {
		return "pid " + emptyQuestion(node.pid) + " · " + emptyQuestion(node.age)
	}
	activity := node.record.Task
	if activity == "" {
		activity = node.age
	}
	label := emptyQuestion(node.record.Harness) + " · " + emptyQuestion(activity)
	if s.showHistory && node.landedGhost {
		label += " · " + node.landedLabel
	}
	return label
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
	return fmt.Sprintf("DETAIL %s · parent %s · model %s · task %s · %s · pane %s · next %s",
		node.record.Name, parent, emptyQuestion(node.record.Model), emptyQuestion(node.record.Task),
		emptyQuestion(node.taskStatus), emptyQuestion(node.record.PaneID), emptyQuestion(node.nextTask))
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
	return label
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
		if line != "" && lipgloss.Width(line+" "+word) > width {
			lines = append(lines, line)
			line = ""
		}
		if line == "" {
			line = word
		} else {
			line += " " + word
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
