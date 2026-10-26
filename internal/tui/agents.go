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
	"github.com/Harrison-Blair/hive/internal/logs"
)

type agentNode struct {
	record   records.Agent
	status   string
	stopped  bool
	current  bool
	pid      string
	nudgedAt string
	result   string
	// The fields below are the stored data of the details popup
	// (agents.md rule 14): the task status, the worktree path and branch of
	// the task, the name of the parent, the unread records of the inbox of
	// the agent, and the last record that the agent sent.
	taskStatus string
	worktree   string
	branch     string
	parentName string
	unread     int
	lastSent   *records.Inbox
}

// agentProfile gives the profile label of a record. The profile is a free
// label, and the record keeps it as a fact.
func agentProfile(record records.Agent) string {
	return record.Profile
}

// agentLabel gives the label of a restart target: its built-in kind, or its
// profile.
func agentLabel(record records.Agent) string {
	if kind := agentBuiltin(record); kind != "" {
		return kind
	}
	return agentProfile(record)
}

// agentBuiltin gives the built-in kind of a record.
func agentBuiltin(record records.Agent) string {
	return record.Kind
}

// agentOwned reports that the record owns a project.
func agentOwned(record records.Agent) bool {
	return record.Owns != ""
}

// agentOwns reports that the record owns project.
func agentOwns(record records.Agent, project string) bool {
	return record.Owns == project
}

type agentGroup struct {
	id       string
	title    string
	keys     []string
	goal     string
	goalStop bool
	// wave is the wave progress of the pod title (agents.md rule 2).
	wave string
}

type agentsData struct {
	nodes          []agentNode
	groups         []agentGroup
	supervisorLive bool
	messages       []agentMessage
}

// agentMessage is one message between two agents (decision 66): an inbox
// record, or an agent prompt event of hive agent send.
type agentMessage struct {
	from, to string
	at       time.Time
}

type agentsState struct {
	data           agentsData
	restartConfirm bool
	restartTarget  records.Agent
	restartRunning bool
	restartNotice  string
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
	source := agentsSource{projects: projects, goals: goals, tasks: tasks, agents: agents, supervisorLive: fresh, supervisorClaim: supervisorClaim, unread: make(map[string][]records.Inbox)}
	for _, name := range inboxAgentNames(agents) {
		inboxRecords, err := inbox.Unread(home, name)
		if err != nil {
			return agentsData{}, fmt.Errorf("%s inbox: %w", name, err)
		}
		source.unread[name] = inboxRecords
	}
	events, _, err := logs.NewReaderWithClock(home, func() time.Time { return now }).Read(context.Background(), time.Time{})
	if err != nil {
		return agentsData{}, fmt.Errorf("logs: %w", err)
	}
	source.events = events
	for _, name := range messageInboxNames(agents) {
		items, err := inbox.All(home, name)
		if err != nil {
			return agentsData{}, fmt.Errorf("%s inbox: %w", name, err)
		}
		source.inbox = append(source.inbox, items...)
	}
	return buildAgents(source), nil
}

// agentsSource holds the records that the Agents screen derives its data from.
type agentsSource struct {
	projects        []records.Project
	goals           []records.Goal
	tasks           []records.Task
	agents          []records.Agent
	supervisorLive  bool
	supervisorClaim records.Claim
	unread          map[string][]records.Inbox
	events          []logs.Event
	// inbox holds each inbox record of a card, read or not: the lit lines
	// of decision 66.
	inbox []records.Inbox
}

// messageInboxNames gives the inbox key of each current agent with a card:
// its name, and "coordinator" for the coordinator.
func messageInboxNames(agents []records.Agent) []string {
	current := currentAgentKeys(agents)
	names := []string{}
	checked := make(map[string]bool)
	for _, record := range agents {
		name := record.Name
		if agentBuiltin(record) == records.AgentKindCoordinator {
			name = "coordinator"
		}
		if !current[record.Key] || name == "" || checked[name] || agentBuiltin(record) == records.AgentKindSupervisor {
			continue
		}
		checked[name] = true
		names = append(names, name)
	}
	return names
}

// agentMessages gives the messages of the inbox records and of the agent
// prompt events of hive agent send.
func agentMessages(items []records.Inbox, events []logs.Event) []agentMessage {
	var out []agentMessage
	for _, item := range items {
		if sent, err := time.Parse(time.RFC3339Nano, item.CreatedAt); err == nil {
			out = append(out, agentMessage{from: item.Agent, to: item.Inbox, at: sent})
		}
	}
	for _, event := range events {
		message, _ := event.Fields["msg"].(string)
		command, _ := event.Fields["command"].(string)
		from, _ := event.Fields["from"].(string)
		to, _ := event.Fields["agent_id"].(string)
		stamp, _ := event.Fields["time"].(string)
		sent, err := time.Parse(time.RFC3339Nano, stamp)
		if message != "agent prompt" || command != "agent send" || from == "" || to == "" || err != nil {
			continue
		}
		out = append(out, agentMessage{from: from, to: to, at: sent})
	}
	return out
}

// inboxAgentNames returns the current coordinator and manager names. Their
// unread inbox shows a nudge.
func inboxAgentNames(agents []records.Agent) []string {
	currentAgents := currentAgentKeys(agents)
	names := []string{}
	checked := make(map[string]bool)
	for _, record := range agents {
		if !currentAgents[record.Key] || record.Name == "" || checked[record.Name] || (agentBuiltin(record) != records.AgentKindCoordinator && !agentOwned(record)) {
			continue
		}
		checked[record.Name] = true
		names = append(names, record.Name)
	}
	return names
}

func buildAgents(source agentsSource) agentsData {
	projects, goals, tasks, agents := source.projects, source.goals, source.tasks, source.agents
	fresh, supervisorClaim := source.supervisorLive, source.supervisorClaim
	taskByID := make(map[string]records.Task, len(tasks))
	for _, task := range tasks {
		taskByID[task.ID] = task
	}
	currentAgents := currentAgentKeys(agents)
	names := make(map[string]string, len(agents))
	for _, record := range agents {
		names[record.Key] = record.Name
	}
	nudgedAt := make(map[string]string)
	for _, name := range inboxAgentNames(agents) {
		for _, item := range source.unread[name] {
			if item.NudgedAt != "" {
				nudgedAt[name] = max(nudgedAt[name], item.NudgedAt)
			}
		}
	}
	goalTitle := make(map[string]string)
	for _, goal := range goals {
		if goal.Status == "active" && goalTitle[goal.Project] == "" {
			goalTitle[goal.Project] = goal.Title
		}
	}
	data := agentsData{groups: []agentGroup{{id: "home", title: "home"}}, supervisorLive: fresh, messages: agentMessages(source.inbox, source.events)}
	groupByProject := make(map[string]int, len(projects))
	for _, project := range projects {
		groupByProject[project.ID] = len(data.groups)
		title := project.ID
		if goalTitle[project.ID] != "" {
			title += " · " + goalTitle[project.ID]
		}
		goal, stop := projectGoalLine(goals, project.ID)
		data.groups = append(data.groups, agentGroup{id: project.ID, title: title, goal: goal, goalStop: stop, wave: projectWave(goals, taskByID, project.ID)})
	}
	for _, record := range agents {
		groupIndex := 0
		if agentBuiltin(record) == "" {
			if index, ok := groupByProject[record.Project]; ok {
				groupIndex = index
			}
		}
		stopped := record.Status == records.AgentStopped || record.Status == records.AgentExited
		status := agentStatus(record, fresh)
		pid := ""
		if agentBuiltin(record) == records.AgentKindSupervisor {
			if record.ProcessPID > 0 {
				pid = fmt.Sprint(record.ProcessPID)
			} else if supervisorClaim.Agent == record.Key && supervisorClaim.PID > 0 {
				pid = fmt.Sprint(supervisorClaim.PID)
			}
		}
		result := ""
		if task, ok := taskByID[record.Task]; ok {
			if task.Verification != nil && task.Verification.Agent == record.Key {
				result = task.Verification.Result
			}
		}
		node := agentNode{
			record: record, status: status, stopped: stopped, current: currentAgents[record.Key],
			pid: pid, result: result, parentName: names[record.Parent],
		}
		if task, ok := taskByID[record.Task]; ok {
			node.taskStatus = task.Status
			if task.Worktree != nil {
				node.worktree, node.branch = task.Worktree.Path, task.Worktree.Branch
			}
		}
		inboxName := record.Name
		if agentBuiltin(record) == records.AgentKindCoordinator {
			inboxName = "coordinator"
		}
		for index, item := range source.inbox {
			if item.Inbox == inboxName && item.ReadAt == "" {
				node.unread++
			}
			if item.Agent == inboxName && (node.lastSent == nil || item.CreatedAt > node.lastSent.CreatedAt) {
				node.lastSent = &source.inbox[index]
			}
		}
		if node.current {
			node.nudgedAt = nudgedAt[record.Name]
		}
		data.nodes = append(data.nodes, node)
		data.groups[groupIndex].keys = append(data.groups[groupIndex].keys, record.Key)
	}
	return data
}

// currentAgentKeys gives the current record of each identity: the live
// record, else the newest record that no other record replaces.
func currentAgentKeys(agents []records.Agent) map[string]bool {
	replaced := make(map[string]bool)
	// root joins the identities of one agent: an old record with no parent
	// and no replaces has only its name, and a newer record of the same
	// name also has its lineage.
	root := make(map[string]string)
	var find func(identity string) string
	find = func(identity string) string {
		parent, ok := root[identity]
		if !ok || parent == identity {
			return identity
		}
		root[identity] = find(parent)
		return root[identity]
	}
	for _, record := range agents {
		if record.Replaces != "" {
			replaced[record.Replaces] = true
		}
		identities := agentIdentities(record)
		first := find(identities[0])
		for _, identity := range identities[1:] {
			root[find(identity)] = first
		}
	}
	groups := make(map[string][]records.Agent)
	for _, record := range agents {
		identity := find(agentIdentities(record)[0])
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
			if !found || currentBefore(record, newest) {
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

// agentIdentities gives each identity of a record. Records that share one
// identity are generations of one agent.
func agentIdentities(record records.Agent) []string {
	// Each general agent is its own identity: many can share a parent.
	if agentProfile(record) == records.ProfileGeneral {
		return []string{"key\x00" + record.Key}
	}
	identities := []string{"name\x00" + record.Name}
	return identities
}

// currentBefore reports that left is a better current record than right: a
// live record, else the newer record. Of two records with one name, the
// newer record has the higher generation.
func currentBefore(left, right records.Agent) bool {
	if leftLive, rightLive := left.Status == records.AgentLive, right.Status == records.AgentLive; leftLive != rightLive {
		return leftLive
	}
	if left.Name == right.Name && left.Generation != right.Generation {
		return left.Generation > right.Generation
	}
	return newerAgent(left, right)
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
	if agentBuiltin(record) == records.AgentKindSupervisor {
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
	switch record.Status {
	case records.AgentStopped, records.AgentExited:
		return "stopped"
	case records.AgentStarting:
		return "starting"
	case records.AgentUnresolved:
		return "unresolved"
	case records.AgentStale:
		return "stale"
	}
	switch record.AgentStatus {
	case "working":
		return "working"
	case "idle", "done":
		return "idle"
	case "blocked":
		return "blocked"
	case records.AgentStarting:
		return "starting"
	case records.AgentUnresolved:
		return "unresolved"
	case records.AgentStale:
		return "stale"
	default:
		return "unknown"
	}
}

// agentStatusAge gives the HH:MM:SS age of the state of an agent, or "?" when
// its status time and fallback time are missing or invalid. The supervisor
// falls back to its started time, as its row always did.
func agentStatusAge(node agentNode, now time.Time) string {
	timestamp := node.record.StatusChangedAt
	if agentBuiltin(node.record) == records.AgentKindSupervisor && timestamp == "" {
		timestamp = node.record.StartedAt
	}
	if age, ok := statusDuration(timestamp, now); ok {
		return age
	}
	return "?"
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

func (s *agentsState) accept(data agentsData) {
	s.data = data
}

type agentsScreen struct{ model *Model }

func (s agentsScreen) DataRequest() dataRequest { return agentsDataRequest }

func (s agentsScreen) KeyHints() string {
	return "←↑↓→ move · tab view · enter peek · a attach · h stopped · t task · r restart coordinator/supervisor · ask in Chat"
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
		if key == "y" && s.model.demo != nil {
			return s.model.demoBlocked("restart")
		}
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
	home, logger := m.home, m.gatedLogger
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
			m.logAction("agent restart request", slog.String("agent_id", target.Key), slog.String("profile", agentLabel(target)), slog.String("result", "started immediately")),
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
		m.logAction("agent restart request", slog.String("agent_id", msg.record.Key), slog.String("profile", agentLabel(msg.record)), slog.String("result", "confirmation required")),
		m.logError("coordinator pane check", msg.err),
	)
}

func (m *Model) startAgentRestart(record records.Agent) tea.Cmd {
	home, logger := m.home, m.gatedLogger
	return func() tea.Msg {
		result, err := agent.Restart(context.Background(), home, record, logger)
		return agentRestartMsg{agentKey: record.Key, result: result, err: err}
	}
}

// View is not used: the frame draws the Agents graph (agentsGraphView).
func (agentsScreen) View(int, int) string { return "" }

func (s agentsState) nodeMap() map[string]agentNode {
	result := make(map[string]agentNode, len(s.data.nodes))
	for _, node := range s.data.nodes {
		result[node.record.Key] = node
	}
	return result
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
