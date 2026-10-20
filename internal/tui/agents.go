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
	"github.com/Harrison-Blair/hive/internal/lib/process"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

type agentsTab uint8

const (
	agentsTree agentsTab = iota
	agentsGraph
	agentsTable
)

type agentNode struct {
	record    records.Agent
	status    string
	age       string
	parentKey string
	inferred  bool
	stopped   bool
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
	data         agentsData
	tab          agentsTab
	selectedKey  string
	hideStopped  bool
	scrollOffset int
	scrollHeight int
	peekPane     string
	peekText     string
	peekOpen     bool
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
	agentEnvelope := agent.List(home, agent.ListOptions{All: true})
	if agentEnvelope.Outcome != "ok" {
		return agentsData{}, fmt.Errorf("agents: %v", agentEnvelope.Error)
	}
	agents, ok := agentEnvelope.Result.([]records.Agent)
	if !ok {
		return agentsData{}, errors.New("agents: unexpected result")
	}
	fresh, err := supervisorIsLive(registry)
	if err != nil {
		return agentsData{}, err
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
		node := agentNode{
			record: record, status: agentStatus(record, fresh), age: activityAge(record.StatusChangedAt, now),
			parentKey: parentKey, inferred: inferred, stopped: stopped,
		}
		data.nodes = append(data.nodes, node)
		data.groups[groupIndex].keys = append(data.groups[groupIndex].keys, record.Key)
	}
	return data, nil
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

func supervisorIsLive(registry *store.Store) (bool, error) {
	var claim records.Claim
	if err := registry.Read("claims", "supervisor", &claim); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	alive, err := process.Alive(claim.PID, claim.PIDStart)
	return err == nil && alive, nil
}

func agentStatus(record records.Agent, supervisorLive bool) string {
	if !supervisorLive {
		return "unknown"
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
		if !seen[node.record.Key] && (!s.hideStopped || !node.stopped) {
			visible = append(visible, node)
		}
	}
	return visible
}

func appendVisibleNodes(visible *[]agentNode, seen map[string]bool, node agentNode, nodes map[string]agentNode, state agentsState) {
	if seen[node.record.Key] || (state.hideStopped && node.stopped) {
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
	return "↑/↓ j/k move · tab tree/graph/table · h hide stopped · enter peek · a attach"
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
		state.keepSelectedVisible()
		return s.model.logAction("agent tab", slog.String("tab", []string{"tree", "graph", "table"}[state.tab]))
	case "h":
		state.hideStopped = !state.hideStopped
		if _, ok := state.selected(); !ok {
			state.accept(state.data)
		}
		return s.model.logAction("agent stopped visibility", slog.Bool("hide_stopped", state.hideStopped))
	case "up", "k", "left":
		previous := state.selectedKey
		state.moveSelection(-1)
		if state.selectedKey != previous {
			return s.model.logAction("agent selection", slog.String("agent_id", state.selectedKey))
		}
	case "down", "j", "right":
		previous := state.selectedKey
		state.moveSelection(1)
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
	end := state.scrollOffset + visibleHeight
	if end > len(lines) {
		end = len(lines)
	}
	if state.scrollOffset > len(lines) {
		state.scrollOffset = 0
	}
	body := []string{titleStyle.Render(tabLine)}
	body = append(body, lines[state.scrollOffset:end]...)
	body = append(body, detailLines...)
	for index, line := range body {
		body[index] = trimGraphLine(line, contentWidth)
	}
	return strings.Join(body, "\n")
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
	if state.hideStopped && node.stopped {
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

func (s agentsState) graphLines() ([]string, map[string]int) {
	lines := make([]string, 0, len(s.data.nodes)+len(s.data.groups)*3)
	positions := make(map[string]int, len(s.data.nodes))
	nodes := s.nodeMap()
	for _, group := range s.data.groups {
		visible := make([]agentNode, 0, len(group.keys))
		for _, key := range group.keys {
			if node, ok := nodes[key]; ok {
				visible = append(visible, node)
			}
		}
		if s.hideStopped {
			filtered := visible[:0]
			for _, node := range visible {
				if !node.stopped {
					filtered = append(filtered, node)
				}
			}
			visible = filtered
		}
		content := make([]string, 0, len(visible))
		groupPositions := make(map[string]int, len(visible))
		for _, root := range s.groupRootsFrom(group, nodes) {
			if s.hideStopped && root.stopped {
				continue
			}
			appendGraphNode(&content, &groupPositions, root, nodes, s, "", true)
		}
		if len(visible) == 0 {
			content = append(content, "  (no agents)")
		}
		boxWidth := max(24, lipgloss.Width(group.title)+3)
		for _, text := range content {
			boxWidth = max(boxWidth, lipgloss.Width(text))
		}
		lines = append(lines, "┌─ "+group.title+" "+strings.Repeat("─", boxWidth-lipgloss.Width(group.title)-3)+"┐")
		for index, text := range content {
			text = trimGraphLine(text, boxWidth)
			padding := max(0, boxWidth-lipgloss.Width(text))
			line := "│" + text + strings.Repeat(" ", padding) + "│"
			if nodeKey, ok := keyAtPosition(groupPositions, index); ok {
				positions[nodeKey] = len(lines)
				if nodeKey == s.selectedKey {
					line = "│" + selectedStyle.Render(text) + strings.Repeat(" ", padding) + "│"
				} else if node := nodes[nodeKey]; node.stopped {
					line = "│" + dimStyle.Render(text) + strings.Repeat(" ", padding) + "│"
				}
			}
			lines = append(lines, line)
		}
		lines = append(lines, "└"+strings.Repeat("─", boxWidth)+"┘")
	}
	for _, node := range s.data.nodes {
		if s.hideStopped && node.stopped {
			continue
		}
		if node.parentKey == "" {
			continue
		}
		parent, ok := nodes[node.parentKey]
		if !ok || groupOf(s.data.groups, parent.record.Key) == groupOf(s.data.groups, node.record.Key) {
			continue
		}
		kind := "──▶"
		if node.inferred {
			kind = "┄┄▶"
		}
		lines = append(lines, parent.record.Name+" "+kind+" "+node.record.Name)
	}
	return lines, positions
}

func appendGraphNode(lines *[]string, positions *map[string]int, node agentNode, nodes map[string]agentNode, state agentsState, prefix string, last bool) {
	branch := "├── "
	if last {
		branch = "└── "
	}
	edge := branch
	if node.inferred {
		edge = "┄┄─ "
	}
	if prefix == "" {
		if node.parentKey != "" {
			edge = "┄┄─ "
			if !node.inferred {
				edge = "─── "
			}
		} else {
			edge = "  "
		}
	} else {
		edge = prefix + edge
	}
	line := edge + state.graphNodeLabel(node)
	if node.record.Key == state.selectedKey {
		line = "▶ " + line
	}
	if node.parentKey != "" && groupOf(state.data.groups, node.record.Key) != groupOf(state.data.groups, node.parentKey) {
		if parent, ok := nodes[node.parentKey]; ok {
			line += " ← " + parent.record.Name
		}
	}
	if node.record.Replaces != "" {
		line += " ⋯ replaces " + displayAgent(nodes, node.record.Replaces)
	}
	(*positions)[node.record.Key] = len(*lines)
	*lines = append(*lines, line)
	children := state.children(node.record.Key)
	filtered := children[:0]
	for _, child := range children {
		if !state.hideStopped || !child.stopped {
			filtered = append(filtered, child)
		}
	}
	for index, child := range filtered {
		childPrefix := "│   "
		if last {
			childPrefix = "    "
		}
		appendGraphNode(lines, positions, child, nodes, state, childPrefix, index == len(filtered)-1)
	}
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
		if node.record.Key == s.selectedKey {
			line = selectedStyle.Render("▶ " + line)
		} else if node.stopped {
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
	return fmt.Sprintf("DETAIL %s · parent %s · replaces %s · model %s · task %s (%s) · pane %s · age %s",
		node.record.Name, s.parentLabel(node), emptyDash(displayAgent(s.nodeMap(), node.record.Replaces)),
		emptyDash(node.record.Model), emptyDash(node.record.Task), node.status,
		emptyDash(node.record.PaneID), node.age)
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
	label := statusMark(node.status) + " " + node.record.Name + " · " + node.status + " · " + node.age
	if node.record.Replaces != "" {
		label += " ⋯ replaces " + displayAgent(s.nodeMap(), node.record.Replaces)
	}
	if node.record.Key == s.selectedKey {
		return selectedStyle.Render("▶ " + label)
	}
	if node.stopped {
		return dimStyle.Render(label)
	}
	return label
}

func (s agentsState) graphNodeLabel(node agentNode) string {
	label := statusMark(node.status) + " " + node.record.Name + " · " + emptyDash(node.record.Harness) + " · "
	if node.record.Task != "" {
		label += node.record.Task
	} else {
		label += node.age
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
		if node, ok := nodes[key]; ok && (!s.hideStopped || !node.stopped) {
			keys[key] = struct{}{}
		}
	}
	roots := make([]agentNode, 0)
	for _, key := range group.keys {
		node, ok := nodes[key]
		if !ok || (s.hideStopped && node.stopped) {
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

func keyAtPosition(positions map[string]int, position int) (string, bool) {
	for key, line := range positions {
		if line == position {
			return key, true
		}
	}
	return "", false
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
	case "parent", "replaces", "model", "task", "pane", "age":
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
