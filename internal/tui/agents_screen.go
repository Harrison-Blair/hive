package tui

import (
	"fmt"
	"image/color"
	"log/slog"
	"slices"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Harrison-Blair/hive/internal/lib/cellgrid"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/logs"
	"github.com/Harrison-Blair/hive/internal/tui/component"
	"github.com/charmbracelet/x/ansi"
)

// The Agents screen of decisions 10 and 14: one pod for each project, side by
// side, with the project agents in a tree or on a rail.
const (
	// podMinWidth is the narrowest pod: a card on the rail, the padding, and
	// the border.
	podMinWidth = component.CardWidth + 4
	// podMaxWide is the widest pod from 120 columns.
	podMaxWide = 64
	// treeSlot is the narrowest slot of one child in the tree form.
	treeSlot = component.CardWidth + 2
	// recentEvents is the number of RECENT lines of the details popup.
	recentEvents = 4
)

// The texts of the Agents screen.
const (
	agentsLegend   = "── spawn   ● working   ◐ idle   ○ stopped   one column per project"
	agentsNoneText = "no project agents"
	agentsScrolls  = "←→ at the edge scrolls"
)

// graphState is the state of the Agents screen: the selected agent, the
// first project column and the first pod row on the screen, the row of the
// selected card in the last view, and the hide flag of h.
type graphState struct {
	selected    string
	left        int
	top         int
	cardRow     int
	hideStopped bool
}

// graphPod is one project column: the manager and the agents that it spawned.
type graphPod struct {
	group    agentGroup
	manager  *agentNode
	children []agentNode
	levels   []int
	hidden   int
	// homeColumn marks the home pod: the general agents with no parent card.
	homeColumn bool
}

// nodes gives the agents of the pod in the order of the selection.
func (p graphPod) nodes() []agentNode {
	var out []agentNode
	if p.manager != nil {
		out = append(out, *p.manager)
	}
	return append(out, p.children...)
}

// tree reports that the children of the pod fit side by side in content
// width cells.
func (p graphPod) tree(content int) bool {
	return p.manager != nil && !p.nested() && (len(p.children) == 0 || content/len(p.children) >= treeSlot)
}

// agentsPods gives one pod for each project, in the order of the records,
// and the home pod last when it has a card. The coordinator and the
// supervisor are in no pod (decision 14). A general agent is under the card
// of its parent when that card is on the graph, otherwise in home (decision
// 39 of general-agent).
func (m *Model) agentsPods() []graphPod {
	nodes := m.agents.nodeMap()
	hidden := func(node agentNode) bool { return m.graph.hideStopped && node.stopped }
	// kids gives the children of each card on the graph.
	kids := make(map[string][]agentNode)
	// podOf gives the pod index of each card on the graph.
	podOf := make(map[string]int)
	var pods []graphPod
	var roots [][]agentNode
	for _, group := range m.agents.data.groups {
		if group.id == "home" {
			continue
		}
		pod := graphPod{group: group}
		for _, key := range group.keys {
			node := nodes[key]
			if node.record.Role == "manager" && node.current && pod.manager == nil {
				pod.manager = &node
				podOf[key] = len(pods)
			}
		}
		var top []agentNode
		for _, key := range group.keys {
			node := nodes[key]
			if node.record.Role == "manager" || node.record.Role == generalRole || !node.current {
				continue
			}
			if hidden(node) {
				pod.hidden++
				continue
			}
			top = append(top, node)
			podOf[key] = len(pods)
		}
		if pod.manager != nil {
			kids[pod.manager.record.Key] = top
			top = nil
		}
		pods = append(pods, pod)
		roots = append(roots, top)
	}
	var pending []agentNode
	for _, node := range m.agents.data.nodes {
		if node.current && node.record.Role == generalRole {
			pending = append(pending, node)
		}
	}
	home := graphPod{group: agentGroup{id: "home"}, homeColumn: true}
	homeIndex := len(pods)
	// place puts each pending general agent under its parent card, until
	// no more agent moves. A hidden agent counts in the pod of its parent.
	place := func(hiddenIn func(index int) *graphPod) {
		for moved := true; moved; {
			moved = false
			rest := pending[:0:0]
			for _, node := range pending {
				index, ok := podOf[node.record.Parent]
				switch {
				case !ok:
					rest = append(rest, node)
					continue
				case hidden(node):
					hiddenIn(index).hidden++
				default:
					kids[node.record.Parent] = append(kids[node.record.Parent], node)
					podOf[node.record.Key] = index
				}
				moved = true
			}
			pending = rest
		}
	}
	place(func(index int) *graphPod { return &pods[index] })
	// The rest go to home. A home agent whose parent is also in home is
	// under it there.
	waiting := make(map[string]bool, len(pending))
	for _, node := range pending {
		waiting[node.record.Key] = true
	}
	var homeRoots []agentNode
	rest := pending[:0:0]
	for _, node := range pending {
		switch {
		case waiting[node.record.Parent]:
			rest = append(rest, node)
		case hidden(node):
			home.hidden++
		default:
			homeRoots = append(homeRoots, node)
			podOf[node.record.Key] = homeIndex
		}
	}
	pending = rest
	place(func(int) *graphPod { return &home })
	// Each agent that is left is in a loop of parents. It is a root.
	homeRoots = append(homeRoots, pending...)
	var out []graphPod
	for index, pod := range pods {
		if pod.manager != nil {
			pod.children, pod.levels = flattenCards(kids, kids[pod.manager.record.Key], 1)
		} else {
			pod.children, pod.levels = flattenCards(kids, roots[index], 0)
		}
		out = append(out, pod)
	}
	home.children, home.levels = flattenCards(kids, homeRoots, 0)
	if len(home.children) > 0 {
		out = append(out, home)
	}
	return out
}

// generalRole is the role of a general agent.
const generalRole = "agent"

// flattenCards gives the cards of a subtree in the order of the selection:
// each card, then its children in order of creation time. It also gives the
// level of each card.
func flattenCards(kids map[string][]agentNode, cards []agentNode, level int) ([]agentNode, []int) {
	cards = slices.Clone(cards)
	slices.SortStableFunc(cards, func(a, b agentNode) int { return compareAgentTimes(a.record.StartedAt, b.record.StartedAt) })
	var nodes []agentNode
	var levels []int
	for _, card := range cards {
		nodes = append(nodes, card)
		levels = append(levels, level)
		children, childLevels := flattenCards(kids, kids[card.record.Key], level+1)
		nodes = append(nodes, children...)
		levels = append(levels, childLevels...)
	}
	return nodes, levels
}

// projectPods gives the number of project pods: each pod except home.
func projectPods(pods []graphPod) int {
	if len(pods) > 0 && pods[len(pods)-1].homeColumn {
		return len(pods) - 1
	}
	return len(pods)
}

// nested reports that a card of the pod is the child of a child.
func (p graphPod) nested() bool {
	low := 1
	if p.manager == nil {
		low = 0
	}
	return slices.ContainsFunc(p.levels, func(level int) bool { return level > low })
}

// selectedAgent gives the selected agent of the graph. When the selection is
// not on the graph, the first agent is selected.
func (m *Model) selectedAgent() (agentNode, int, bool) {
	pods := m.agentsPods()
	for column, pod := range pods {
		for _, node := range pod.nodes() {
			if node.record.Key == m.graph.selected {
				return node, column, true
			}
		}
	}
	for column, pod := range pods {
		if nodes := pod.nodes(); len(nodes) > 0 {
			m.graph.selected = nodes[0].record.Key
			return nodes[0], column, true
		}
	}
	return agentNode{}, 0, false
}

// agentsColumns gives the number of pods on the screen and the pod width for
// a body of width cells.
func agentsColumns(pods, width int, wide bool) (int, int) {
	visible := max(1, min(pods, (width+1)/(podMinWidth+1)))
	podWidth := (width - (visible - 1)) / visible
	if wide {
		podWidth = min(podWidth, podMaxWide)
	}
	return visible, podWidth
}

// agentsGraphView renders the Agents screen at exactly width×height cells.
func (m *Model) agentsGraphView(width, height int, focused bool) string {
	pods := m.agentsPods()
	projects := projectPods(pods)
	title := fmt.Sprintf("project agents · %d projects", projects)
	if projects == 1 {
		title = "project agents · 1 project"
	}
	panel := component.Panel{Title: title, Focused: focused}
	bodyWidth, bodyHeight := max(width-4, 1), max(height-2, 0)
	if m.agents.peekOpen {
		panel.Title = "peek " + m.agents.peekPane + " · esc closes"
		panel.Lines = strings.Split(renderAgentPeek(&m.agents, bodyWidth, bodyHeight), "\n")
		return panel.View(width, height)
	}
	if projects == 0 {
		panel.Lines = []string{"", component.Notice{Text: agentsNoneText}.View(bodyWidth)}
		return panel.View(width, height)
	}
	wide := width >= component.LongFormWidth
	visible, podWidth := agentsColumns(len(pods), bodyWidth, wide)
	_, column, _ := m.selectedAgent()
	m.graph.left = min(max(m.graph.left, column-visible+1, 0), column, len(pods)-visible)
	scrolls := visible < len(pods)
	footer := 0
	if wide {
		footer++
	}
	if scrolls {
		footer++
	}
	canvas := newCellCanvas(bodyWidth, bodyHeight)
	group := visible*podWidth + visible - 1
	x := (bodyWidth - group) / 2
	room := bodyHeight - 1 - footer
	if footer > 0 {
		room--
	}
	podHeight := 0
	shown := pods[m.graph.left : m.graph.left+visible]
	for _, pod := range shown {
		podHeight = max(podHeight, m.podLines(pod, podWidth)+2)
	}
	// The pods are drawn in full on a tall canvas. When they do not fit,
	// the region shows room rows from top and keeps the selected card on
	// the screen.
	tall := newCellCanvas(bodyWidth, podHeight)
	m.graph.cardRow = -1
	for _, pod := range shown {
		m.drawPod(tall, pod, x, 0, podWidth, podHeight)
		x += podWidth + 1
	}
	room = max(room, 0)
	if row := m.graph.cardRow; row >= 0 {
		m.graph.top = min(max(m.graph.top, row+component.CardHeight-room), row)
	}
	m.graph.top = min(max(m.graph.top, 0), max(podHeight-room, 0))
	for row := range min(room, podHeight) {
		copy(canvas.cells[(row+1)*bodyWidth:(row+2)*bodyWidth], tall.cells[(row+m.graph.top)*bodyWidth:(row+m.graph.top+1)*bodyWidth])
	}
	panel.VMark = component.Thumb(bodyHeight, podHeight, room, m.graph.top)
	row := bodyHeight - footer
	if wide {
		canvas.text(0, row, agentsLegend, cellgrid.Dim, false)
		row++
	}
	if scrolls {
		end := min(m.graph.left+visible, projects)
		scroll := fmt.Sprintf("%s · projects %d–%d of %d", agentsScrolls, min(m.graph.left+1, end), end, projects)
		if m.graph.left+visible > projects {
			scroll += " · home"
		}
		canvas.text(0, row, scroll, cellgrid.Dim, false)
		var edges []string
		if m.graph.left > 0 {
			edges = append(edges, "← "+pods[m.graph.left-1].group.id)
		}
		if end := m.graph.left + visible; end < len(pods) {
			edges = append(edges, pods[end].group.id+" →")
		}
		right := strings.Join(edges, "   ")
		canvas.text(bodyWidth-ansi.StringWidth(right), row, right, cellgrid.Dim, false)
		panel.HMark = component.Thumb(width-2, len(pods), visible, m.graph.left)
	}
	panel.Lines = canvas.lines()
	return panel.View(width, height)
}

// podLines gives the number of rows inside the border of a pod.
func (m *Model) podLines(pod graphPod, width int) int {
	content := width - 4
	rows := 1 // the goal line
	switch {
	case pod.homeColumn:
		rows += 5*len(pod.children) - 1
	case pod.manager == nil:
		rows += 2 + 5*len(pod.children)
	case pod.tree(content):
		rows += component.CardHeight
		if len(pod.children) > 0 {
			rows += 3 + component.CardHeight
		} else {
			rows += 2
		}
	default:
		rows += component.CardHeight + 5*len(pod.children)
	}
	if pod.note() != "" && len(pod.children) > 0 {
		rows += 2
	}
	return rows
}

// note gives the stopped note of a pod: the h key and the number of stopped
// agents.
func (p graphPod) note() string {
	if p.hidden > 0 {
		return fmt.Sprintf("h shows %d stopped", p.hidden)
	}
	stopped := 0
	for _, node := range p.children {
		if node.stopped {
			stopped++
		}
	}
	if stopped == 0 {
		return ""
	}
	return fmt.Sprintf("h hides %d stopped", stopped)
}

// drawPod draws one project column at x, y of width×height cells.
func (m *Model) drawPod(c *cellCanvas, pod graphPod, x, y, width, height int) {
	c.box(x, y, width, height, cellgrid.Border, false)
	c.text(x+2, y, " "+ansi.Truncate(pod.group.id, width-6, "…")+" ", cellgrid.Dim, true)
	left, content, top := x+2, width-4, y+1
	goal := cellgrid.Dim
	if pod.group.goalStop {
		goal = cellgrid.Accent
	}
	c.text(left, top, ansi.Truncate(pod.group.goal, content, "…"), goal, false)
	top++
	var next int
	switch {
	case pod.homeColumn:
		next = m.drawRail(c, pod, left, top-1)
	case pod.manager == nil:
		c.text(left, top+1, "no manager", cellgrid.Dim, false)
		next = m.drawRail(c, pod, left, top+2)
	case pod.tree(content):
		manager := *pod.manager
		mx := left + (content-component.CardWidth)/2
		m.drawCard(c, manager, mx, top)
		center := mx + component.CardWidth/2
		next = top + component.CardHeight
		if len(pod.children) == 0 {
			text := "no child agents"
			if note := pod.note(); note != "" {
				text = note
			}
			c.text(left, next+1, text, cellgrid.Dim, false)
			break
		}
		m.junction(c, center, next-1, manager, '┬', '┯')
		c.set(center, next, '│', cellgrid.Border)
		slot := content / len(pod.children)
		var centers []int
		for index := range pod.children {
			centers = append(centers, left+index*slot+(slot-component.CardWidth)/2+component.CardWidth/2)
		}
		first, last := centers[0], centers[len(centers)-1]
		for col := first; col <= last; col++ {
			c.set(col, next+1, '─', cellgrid.Border)
		}
		for _, col := range centers {
			c.set(col, next+1, '┬', cellgrid.Border)
			c.set(col, next+2, '│', cellgrid.Border)
		}
		if first != last {
			c.set(first, next+1, '╭', cellgrid.Border)
			c.set(last, next+1, '╮', cellgrid.Border)
		}
		mark := '┴'
		if slices.Contains(centers, center) {
			mark = '┼'
			if first == last {
				mark = '│'
			}
		}
		c.set(center, next+1, mark, cellgrid.Border)
		for index, child := range pod.children {
			cx := centers[index] - component.CardWidth/2
			m.drawCard(c, child, cx, next+3)
			m.junction(c, centers[index], next+3, child, '┴', '┷')
		}
		next += 3 + component.CardHeight
	default:
		m.drawCard(c, *pod.manager, left, top)
		next = m.drawRail(c, pod, left, top+component.CardHeight)
	}
	if note := pod.note(); note != "" && len(pod.children) > 0 {
		c.text(left, next+1, note, cellgrid.Dim, false)
	}
}

// drawRail draws the children of a pod in the rail form from row next and
// gives the first row below them. A card of level 0 is on content column
// left. A card of level L is on a rail at left+2(L-1), with 5 rows for each
// card: a │ row, then the card with ├┤ or ╰┤ on its second row.
func (m *Model) drawRail(c *cellCanvas, pod graphPod, left, next int) int {
	// continues reports that a later card of the pod has level, before a
	// card of a lower level.
	continues := func(index, level int) bool {
		for _, later := range pod.levels[index+1:] {
			if later <= level {
				return later == level
			}
		}
		return false
	}
	for index, child := range pod.children {
		level := pod.levels[index]
		if level == 0 {
			m.drawCard(c, child, left, next+1)
			next += 5
			continue
		}
		rail := left + 2*(level-1)
		if index == 0 || pod.levels[index-1] < level {
			parent, light, heavy := pod.manager, '├', '┡'
			if index > 0 {
				parent = &pod.children[index-1]
			}
			if level > 1 {
				light, heavy = '┬', '┯'
			}
			m.junction(c, rail, next-1, *parent, light, heavy)
		}
		for ancestor := 1; ancestor < level; ancestor++ {
			if continues(index, ancestor) {
				for row := next; row < next+5; row++ {
					c.set(left+2*(ancestor-1), row, '│', cellgrid.Border)
				}
			}
		}
		c.set(rail, next, '│', cellgrid.Border)
		c.set(rail, next+1, '│', cellgrid.Border)
		m.drawCard(c, child, rail+1, next+1)
		if continues(index, level) {
			c.set(rail, next+2, '├', cellgrid.Border)
			c.set(rail, next+3, '│', cellgrid.Border)
			c.set(rail, next+4, '│', cellgrid.Border)
		} else {
			c.set(rail, next+2, '╰', cellgrid.Border)
		}
		m.junction(c, rail+1, next+2, child, '┤', '┨')
		next += 5
	}
	return next
}

// junction sets a connector cell on the border of the card of node: light in
// border, or in working for a working card (decision 60), or for the selected
// card, which has a heavy border, the mixed-weight rune heavy in accent.
func (m *Model) junction(c *cellCanvas, x, y int, node agentNode, light, heavy rune) {
	switch {
	case node.record.Key == m.graph.selected:
		c.set(x, y, heavy, cellgrid.Accent)
	case node.status == "working":
		c.set(x, y, light, cellgrid.Working)
	default:
		c.set(x, y, light, cellgrid.Border)
	}
}

// drawCard draws the card of one agent with component.Card.
func (m *Model) drawCard(c *cellCanvas, node agentNode, x, y int) {
	card := component.Card{Name: node.record.Name, Sub: agentCardSub(node), State: node.status, Selected: node.record.Key == m.graph.selected}
	if card.Selected {
		m.graph.cardRow = y
	}
	c.blit(x, y, card.View())
}

// agentCardSub gives the sub line of a card: the harness, the state, and the
// age of the state or the result of a stopped agent.
func agentCardSub(node agentNode) string {
	parts := []string{node.record.Harness, node.status}
	switch {
	case node.stopped && node.result != "":
		parts = append(parts, node.result)
	case node.stopped && node.record.AgentStatus != "":
		parts = append(parts, node.record.AgentStatus)
	case !node.stopped && node.age != "?":
		parts = append(parts, node.age)
	}
	return strings.Join(slices.DeleteFunc(parts, func(text string) bool { return text == "" }), " · ")
}

// agentDetailsPopup gives the title and the lines of the details popup of
// the selected agent.
func (m *Model) agentDetailsPopup() (string, []string) {
	node, column, ok := m.selectedAgent()
	if !ok {
		return "details", []string{"no agent selected"}
	}
	record := node.record
	status := record.Status
	if record.Status == records.AgentLive {
		status += " · " + node.status
	}
	field := func(label, value string) string {
		if value == "" {
			return ""
		}
		return component.DetailField(label, value)
	}
	nudged := ""
	if node.nudgedAt != "" {
		nudged = activityAge(node.nudgedAt, m.now()) + " ago by supervisor"
	}
	lines := []string{"",
		field("role", record.Role),
		field("project", record.Project),
		component.DetailField("status", stripCardDot(node.status)+" "+status),
		field("nudged", nudged),
		field("harness", record.Harness),
		field("model", record.Model),
		field("pane", record.PaneID),
		field("tab", record.Name),
		field("goal", m.agentsPods()[column].group.goal),
		field("parent", record.Parent),
		field("replaces", record.Replaces),
	}
	lines = slices.DeleteFunc(lines[1:], func(line string) bool { return line == "" })
	lines = append([]string{""}, lines...)
	key := lipgloss.NewStyle().Foreground(component.Accent).Bold(true)
	dim := lipgloss.NewStyle().Foreground(component.Dim)
	task := key.Render("t") + "  open task · " + record.Task
	if record.Task == "" {
		task = dim.Bold(true).Render("t") + dim.Render("  open task · none")
	}
	lines = append(lines, "", component.DetailSection("ACTIONS"),
		key.Render("a")+"  attach to pane",
		key.Render("p")+"  peek pane",
		task,
	)
	if len(node.recent) > 0 {
		lines = append(lines, "", component.DetailSection("RECENT"))
		for index, text := range node.recent {
			if index > 0 {
				text = dim.Render(text)
			}
			lines = append(lines, text)
		}
	}
	return record.Name + " · esc closes", lines
}

// stripCardDot gives the dot of a state in the style of a card.
func stripCardDot(state string) string {
	line := strings.Split(component.Card{State: state}.View(), "\n")[1]
	// The dot is the third cell of the first body row: the border, then
	// one cell of padding.
	return ansi.Cut(line, 2, 3)
}

// updateAgentsKey handles a key of the Agents graph (decisions 14, 22, 23).
func (m *Model) updateAgentsKey(key string) tea.Cmd {
	node, _, ok := m.selectedAgent()
	switch key {
	case "left", "right", "up", "down", "j", "k":
		if m.moveAgentSelection(key) {
			return m.logAction("agent selection", slog.String("agent_id", m.graph.selected))
		}
	case "enter":
		if ok {
			m.popupOpen = true
			m.popupOffset = 0
			return m.logAction("agent details open", slog.String("agent_id", node.record.Key))
		}
	case "h":
		m.graph.hideStopped = !m.graph.hideStopped
		return m.logAction("agent hide stopped", slog.Bool("hide_stopped", m.graph.hideStopped))
	case "a", "p", "t":
		if ok {
			return m.agentAction(key, node)
		}
	}
	return nil
}

// agentAction runs the action key a, p, or t on node.
func (m *Model) agentAction(key string, node agentNode) tea.Cmd {
	record := node.record
	switch {
	case key == "a" && record.PaneID != "":
		return batch(m.logAction("agent attach", slog.String("pane_id", record.PaneID)), m.attachCmd(record.PaneID))
	case key == "p" && record.PaneID != "":
		m.popupOpen = false
		return batch(m.logAction("agent peek", slog.String("pane_id", record.PaneID)), m.startAgentsPeek(record.PaneID))
	case key == "t" && record.Task != "":
		m.popupOpen = false
		m.selectScreen(tabWork, 2)
		m.tasks.selectedID = record.Task
		m.tasks.scrollOffset = 0
		m.tasks.fileOpen = false
		m.tasks.fileLoading = false
		m.screenLoading = false
		return batch(m.logAction("agent open task", slog.String("task_id", record.Task)), m.startCurrentScreenLoad())
	}
	return nil
}

// moveAgentSelection moves the selection by one arrow key. In a tree, ← and
// → move between the children, and ↑ and ↓ move between the manager and the
// children. At the edge of a pod, ← and → move to the next pod.
func (m *Model) moveAgentSelection(key string) bool {
	node, column, ok := m.selectedAgent()
	if !ok {
		return false
	}
	pods := m.agentsPods()
	nodes := pods[column].nodes()
	index := slices.IndexFunc(nodes, func(n agentNode) bool { return n.record.Key == node.record.Key })
	tree := pods[column].tree(m.podContentWidth(len(pods)))
	target := index
	switch key {
	case "up", "k":
		target = index - 1
		if tree && index > 0 {
			target = 0
		}
	case "down", "j":
		target = index + 1
	case "left", "right":
		step := 1
		if key == "left" {
			step = -1
		}
		if tree && index > 0 && index+step > 0 && index+step < len(nodes) {
			target = index + step
			break
		}
		next := column + step
		for next >= 0 && next < len(pods) && len(pods[next].nodes()) == 0 {
			next += step
		}
		if next < 0 || next >= len(pods) {
			return false
		}
		others := pods[next].nodes()
		m.graph.selected = others[min(index, len(others)-1)].record.Key
		return true
	}
	if target < 0 || target >= len(nodes) || target == index {
		return false
	}
	m.graph.selected = nodes[target].record.Key
	return true
}

// podContentWidth gives the content width of a pod at the current size.
func (m *Model) podContentWidth(pods int) int {
	_, podWidth := agentsColumns(pods, max(m.width-4, 1), m.width >= component.LongFormWidth)
	return podWidth - 4
}

// projectGoalLine gives the goal line of a project pod: the active goal and
// its open wave, or the last goal and its status. The flag tells that the
// open wave needs the user.
func projectGoalLine(goals []records.Goal, project string) (string, bool) {
	var goal *records.Goal
	for index := range goals {
		candidate := &goals[index]
		if candidate.Project != project {
			continue
		}
		if goal == nil || (candidate.Status == "active") != (goal.Status == "active") && candidate.Status == "active" ||
			(candidate.Status == "active") == (goal.Status == "active") && candidate.UpdatedAt > goal.UpdatedAt {
			goal = candidate
		}
	}
	if goal == nil {
		return "", false
	}
	if goal.Status == "active" {
		for _, wave := range goal.Waves {
			if wave.Status == "done" {
				continue
			}
			if wave.NeedsUser != "" {
				return "needs you · " + wave.NeedsUser, true
			}
			return goal.ID + " · wave " + wave.ID, false
		}
	}
	return goal.ID + " · " + goal.Status, false
}

// recentAgentEvents gives the last log events of each agent, newest first,
// as "HH:MM message" in zone.
func recentAgentEvents(events []logs.Event, zone *time.Location) map[string][]string {
	type entry struct {
		at   time.Time
		text string
	}
	byAgent := make(map[string][]entry)
	for _, event := range events {
		agentID, _ := event.Fields["agent_id"].(string)
		stamp, _ := event.Fields["time"].(string)
		message, _ := event.Fields["msg"].(string)
		at, err := time.Parse(time.RFC3339Nano, stamp)
		if agentID == "" || err != nil {
			continue
		}
		byAgent[agentID] = append(byAgent[agentID], entry{at: at, text: at.In(zone).Format("15:04") + " " + message})
	}
	out := make(map[string][]string, len(byAgent))
	for agentID, entries := range byAgent {
		slices.SortStableFunc(entries, func(a, b entry) int { return b.at.Compare(a.at) })
		for _, item := range entries[:min(len(entries), recentEvents)] {
			out[agentID] = append(out[agentID], item.text)
		}
	}
	return out
}

// cellCanvas is a grid of cells that the Agents screen draws on. Each cell
// has a character, a palette token, and a bold flag.
type cellCanvas struct {
	width, height int
	cells         []canvasCell
}

type canvasCell struct {
	text  string
	token cellgrid.Token
	bold  bool
}

func newCellCanvas(width, height int) *cellCanvas {
	c := &cellCanvas{width: width, height: height, cells: make([]canvasCell, width*height)}
	for index := range c.cells {
		c.cells[index].text = " "
	}
	return c
}

func (c *cellCanvas) set(x, y int, r rune, token cellgrid.Token) {
	c.put(x, y, canvasCell{text: string(r), token: token})
}

func (c *cellCanvas) put(x, y int, cell canvasCell) {
	if x >= 0 && x < c.width && y >= 0 && y < c.height {
		c.cells[y*c.width+x] = cell
	}
}

// text writes s from x, y, one cell for each character.
func (c *cellCanvas) text(x, y int, s string, token cellgrid.Token, bold bool) {
	for _, r := range s {
		c.put(x, y, canvasCell{text: string(r), token: token, bold: bold})
		x++
	}
}

// box draws a rounded border of width×height cells from x, y.
func (c *cellCanvas) box(x, y, width, height int, token cellgrid.Token, bold bool) {
	for col := x + 1; col < x+width-1; col++ {
		c.put(col, y, canvasCell{text: "─", token: token, bold: bold})
		c.put(col, y+height-1, canvasCell{text: "─", token: token, bold: bold})
	}
	for row := y + 1; row < y+height-1; row++ {
		c.put(x, row, canvasCell{text: "│", token: token, bold: bold})
		c.put(x+width-1, row, canvasCell{text: "│", token: token, bold: bold})
	}
	c.put(x, y, canvasCell{text: "╭", token: token, bold: bold})
	c.put(x+width-1, y, canvasCell{text: "╮", token: token, bold: bold})
	c.put(x, y+height-1, canvasCell{text: "╰", token: token, bold: bold})
	c.put(x+width-1, y+height-1, canvasCell{text: "╯", token: token, bold: bold})
}

// blit copies the cells of a rendered component to x, y.
func (c *cellCanvas) blit(x, y int, view string) {
	for row, line := range strings.Split(view, "\n") {
		grid, err := cellgrid.Convert(line, ansi.StringWidth(line), 1)
		if err != nil {
			c.text(x, y+row, ansi.Strip(line), cellgrid.FG, false)
			continue
		}
		for col, cell := range grid.Cells {
			c.put(x+col, y+row, canvasCell{text: string(cell.Rune), token: cell.Token, bold: cell.Bold})
		}
	}
}

// lines gives each row of the canvas as styled text.
func (c *cellCanvas) lines() []string {
	out := make([]string, c.height)
	for row := range c.height {
		var b strings.Builder
		cells := c.cells[row*c.width : (row+1)*c.width]
		for start := 0; start < len(cells); {
			end := start + 1
			for end < len(cells) && cells[end].token == cells[start].token && cells[end].bold == cells[start].bold {
				end++
			}
			var run strings.Builder
			for _, cell := range cells[start:end] {
				run.WriteString(cell.text)
			}
			b.WriteString(tokenStyle(cells[start].token, cells[start].bold).Render(run.String()))
			start = end
		}
		out[row] = b.String()
	}
	return out
}

// tokenStyle gives the style of a palette token.
func tokenStyle(token cellgrid.Token, bold bool) lipgloss.Style {
	style := lipgloss.NewStyle().Bold(bold)
	colors := map[cellgrid.Token]color.Color{
		cellgrid.Border: component.Border, cellgrid.Dim: component.Dim, cellgrid.Accent: component.Accent,
		cellgrid.Working: component.Working, cellgrid.Done: component.Done, cellgrid.Info: component.Info,
		cellgrid.User: component.User, cellgrid.Code: component.Code,
	}
	if value, ok := colors[token]; ok {
		style = style.Foreground(value)
	}
	return style
}

// The size of the details popup of the Agents screen.
const (
	agentPopupWidth  = 60
	agentPopupHeight = 27
)
