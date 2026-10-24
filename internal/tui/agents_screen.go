package tui

import (
	"cmp"
	"fmt"
	"image/color"
	"log/slog"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

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
	// podGap is the space between two pods of the overview.
	podGap = 2
	// treeSlot is the narrowest slot of one child in the tree form.
	treeSlot = component.CardWidth + 2
	// recentEvents is the number of RECENT lines of the details popup.
	recentEvents = 4
)

// The texts of the Agents screen.
const (
	agentsLegend    = "── spawn   ● working   ◐ idle   ○ stopped   ━ ▶ message, lit 5s   one column per project"
	agentsLegendLit = "━ ▶"
	agentsNoneText  = "no project agents"
	agentsScrolls   = "←→ at the edge scrolls"
)

// graphState is the state of the Agents screen: the selected agent, the
// selected pod and the one-pod view (decision 62), the first project column
// and the first pod row on the screen, the first canvas column of the
// one-pod view on the screen, the cell of the selected card in the last
// view, and the hide flag of h.
type graphState struct {
	selected    string
	pod         string
	podView     bool
	left        int
	top         int
	shift       int
	cardRow     int
	cardCol     int
	hideStopped bool
	// The fields below hold the last view. onePod tells that it is the
	// one-pod view. focus is the key of the card with the heavy border. up
	// gives the parent of each card and each pod on the graph: the
	// coordinator is the root. lines gives the cells of the line from a
	// parent to a child, parent first. joints gives the joint of each
	// parent on a rail, and managerTop the top joint of the manager card.
	onePod     bool
	focus      string
	up         map[string]string
	lines      map[[2]string][]point
	joints     map[string]point
	managerTop point
	// roots gives the top cards of the tree of the one-pod view, left to
	// right: the children of the manager, or the cards with no parent.
	roots []treeRoot
	// litAt is the time of the lit timer that runs (decision 67).
	litAt time.Time
}

// point is one cell of a canvas.
type point struct{ x, y int }

// treeRoot is one top card of the tree of the one-pod view: its key, its
// name, and the canvas column and row of the card.
type treeRoot struct {
	key, name string
	x, y      int
}

// podKey gives the graph key of a pod.
func podKey(pod graphPod) string { return "pod:" + pod.group.id }

// graphPod is one project column: the manager and the agents that it spawned.
type graphPod struct {
	group    agentGroup
	manager  *agentNode
	children []agentNode
	levels   []int
	hidden   int
	// homeColumn marks the home pod: the general agents with no parent card.
	homeColumn bool
	// managerHidden marks a project pod whose manager card h hides. Its
	// cards hang from the pod top border as in the home pod (goal g9).
	managerHidden bool
}

// homeTree reports that the pod draws its cards as the home pod does: the
// home pod, or a pod with a hidden manager.
func (p graphPod) homeTree() bool { return p.homeColumn || p.managerHidden }

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

// agentsPods gives each pod with a visible card (goal g9).
func (m *Model) agentsPods() []graphPod {
	var out []graphPod
	for _, pod := range m.allPods() {
		if pod.manager != nil || len(pod.children) > 0 {
			out = append(out, pod)
		}
	}
	return out
}

// hiddenCards gives the number of cards that h hides, in all pods.
func (m *Model) hiddenCards() int {
	hidden := 0
	for _, pod := range m.allPods() {
		hidden += pod.hidden
	}
	return hidden
}

// allPods gives one pod for each project, in the order of the records, and
// the home pod last. The coordinator and the supervisor are in no pod
// (decision 14). A general agent is under the card of its parent when that
// card is on the graph, otherwise in home (decision 39 of general-agent).
func (m *Model) allPods() []graphPod {
	nodes := m.agents.nodeMap()
	hidden := func(node agentNode) bool { return m.graph.hideStopped && node.hideable() }
	// kids gives the children of each card on the graph.
	kids := make(map[string][]agentNode)
	// podOf gives the pod index of each card on the graph, and of each
	// hidden manager.
	podOf := make(map[string]int)
	// hiddenManager marks each hidden manager. Its children are top cards of
	// its pod (goal g9).
	hiddenManager := make(map[string]bool)
	var pods []graphPod
	var roots [][]agentNode
	for _, group := range m.agents.data.groups {
		if group.id == "home" {
			continue
		}
		pod := graphPod{group: group}
		for _, key := range group.keys {
			node := nodes[key]
			if node.record.Role == "manager" && node.current && pod.manager == nil && !pod.managerHidden {
				if hidden(node) {
					pod.hidden++
					pod.managerHidden = true
					hiddenManager[key] = true
					podOf[key] = len(pods)
					continue
				}
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
				case hiddenManager[node.record.Parent]:
					roots[index] = append(roots[index], node)
					podOf[node.record.Key] = index
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
	return append(out, home)
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

// coordinatorNode gives the current coordinator, which has the card at the
// top of the screen (decision 65).
func (m *Model) coordinatorNode() (agentNode, bool) {
	for _, node := range m.agents.data.nodes {
		if node.current && node.record.Role == "coordinator" {
			return node, true
		}
	}
	return agentNode{}, false
}

// coordinatorCard gives the coordinator when the screen draws its card:
// with no registered project the screen is the notice alone (decision 10).
func (m *Model) coordinatorCard() (agentNode, bool) {
	node, ok := m.coordinatorNode()
	return node, ok && len(m.agents.data.groups) > 1
}

// inPodView reports that the screen shows the one-pod view: after enter on
// a pod, or at once with one project and no home pod (decision 64).
func (m *Model) inPodView(pods []graphPod) bool {
	return len(pods) > 0 && (m.graph.podView || len(pods) == 1 && projectPods(pods) == 1)
}

// podIndex gives the selected pod: the pod of m.graph.pod, else the pod
// with the id or the card of the selection, else the first pod.
func (m *Model) podIndex(pods []graphPod) int {
	found := func() int {
		for index, pod := range pods {
			if pod.group.id == m.graph.pod {
				return index
			}
		}
		for index, pod := range pods {
			if pod.group.id == m.graph.selected {
				return index
			}
		}
		for index, pod := range pods {
			for _, node := range pod.nodes() {
				if node.record.Key == m.graph.selected {
					return index
				}
			}
		}
		return -1
	}()
	if found < 0 {
		// The remembered pod is not on the graph, for example while h hides
		// it. It is selected again when it comes back.
		return 0
	}
	if found < len(pods) {
		m.graph.pod = pods[found].group.id
	}
	return found
}

// coordinatorSelected reports that the coordinator card is selected in the
// overview.
func (m *Model) coordinatorSelected(pods []graphPod) bool {
	node, ok := m.coordinatorCard()
	return ok && !m.inPodView(pods) && m.graph.selected == node.record.Key
}

// selectedAgent gives the selected card and its pod. In the one-pod view,
// the first card of the pod is selected when the selection is not in it. In
// the overview, only the coordinator card can be selected; it is in no pod.
func (m *Model) selectedAgent() (agentNode, int, bool) {
	pods := m.agentsPods()
	if m.coordinatorSelected(pods) {
		node, _ := m.coordinatorNode()
		return node, -1, true
	}
	column := m.podIndex(pods)
	if !m.inPodView(pods) {
		return agentNode{}, column, false
	}
	nodes := pods[column].nodes()
	for _, node := range nodes {
		if node.record.Key == m.graph.selected {
			return node, column, true
		}
	}
	if len(nodes) > 0 {
		m.graph.selected = nodes[0].record.Key
		return nodes[0], column, true
	}
	return agentNode{}, column, false
}

// agentsColumns gives the number of pods on the screen and the pod width for
// a body of width cells.
func agentsColumns(pods, width int, wide bool) (int, int) {
	visible := max(1, min(pods, (width+podGap)/(podMinWidth+podGap)))
	podWidth := (width - podGap*(visible-1)) / visible
	if wide {
		podWidth = min(podWidth, podMaxWide)
	}
	return visible, podWidth
}

// agentsGraphView renders the Agents screen at exactly width×height cells:
// the overview of all pods, or the one-pod view (decisions 62 to 65), with
// the coordinator card at the top and the lit lines of decisions 66 to 68.
func (m *Model) agentsGraphView(width, height int, focused bool) string {
	pods := m.agentsPods()
	projects := projectPods(pods)
	if m.agents.peekOpen {
		panel := component.Panel{Title: "peek " + m.agents.peekPane + " · esc closes", Focused: focused}
		panel.Lines = strings.Split(renderAgentPeek(&m.agents, max(width-4, 1), max(height-2, 0)), "\n")
		return panel.View(width, height)
	}
	// The body has no box: it uses the full width with one cell of margin
	// on each side (agents.md rule 1).
	bodyWidth, bodyHeight := max(width-2, 1), height
	coordinator, hasCoordinator := m.coordinatorNode()
	// With no registered project, or with nothing on the graph, the screen
	// is the notice. When h hides each project pod, the notice is a body
	// line below the coordinator card and the home pod (goal g9).
	if registered := len(m.agents.data.groups) - 1; projects == 0 && (registered == 0 || len(pods) == 0 && !hasCoordinator) {
		return fitBlock(strings.Join([]string{"", " " + component.Notice{Text: agentsNoneText}.View(bodyWidth)}, "\n"), width, height)
	}
	wide := width >= component.LongFormWidth
	if wide {
		// One blank row stays below the legend, above the status rows.
		bodyHeight = max(bodyHeight-1, 0)
	}
	selected := m.podIndex(pods)
	m.graph.onePod = len(pods) > 0 && m.inPodView(pods)
	m.graph.focus = ""
	switch {
	case m.coordinatorSelected(pods):
		m.graph.focus = coordinator.record.Key
	case m.graph.onePod:
		node, _, _ := m.selectedAgent()
		m.graph.focus = node.record.Key
	}
	var shown []graphPod
	visible, podWidth := agentsColumns(len(pods), bodyWidth, wide)
	if m.graph.onePod {
		shown = pods[selected : selected+1]
		visible, podWidth = agentsColumns(1, bodyWidth, wide)
		// The pod grows to the width of its tree (decision 75).
		podWidth = max(podWidth, layoutTree(shown[0], podWidth-4).width+4)
	} else {
		visible = min(visible, len(pods))
		m.graph.left = max(min(max(m.graph.left, selected-visible+1, 0), selected, len(pods)-visible), 0)
		shown = pods[m.graph.left : m.graph.left+visible]
	}
	// The canvas columns 0 and inner-1 are the padding cells of the body. A
	// tree that is wider than the region makes the canvas wider, and the
	// view scrolls sideways (decision 75).
	inner := width
	canvasWidth := inner
	if m.graph.onePod {
		canvasWidth = max(inner, podWidth+1)
	}
	sideways := canvasWidth > inner
	scrolls := !m.graph.onePod && visible < len(pods)
	footer := 0
	if wide {
		footer++
	}
	if projects == 0 {
		footer++
	}
	if scrolls || sideways {
		footer++
	}
	// From 120 columns one blank row is above the coordinator card and one
	// is above the footer lines. Below 120 columns there is none.
	gap := 0
	if wide {
		gap = 1
	}
	room := bodyHeight - footer - gap
	if footer > 0 {
		room -= gap
	}
	head := 0
	if hasCoordinator {
		head = component.CardHeight + 1
	}
	podHeight := 0
	for _, pod := range shown {
		podHeight = max(podHeight, m.podLines(pod, podWidth)+2)
	}
	// The coordinator card and the pods are drawn in full on a tall canvas.
	// When they do not fit, the region shows room rows from top and keeps
	// the selected card on the screen.
	tall := newCellCanvas(canvasWidth, head+podHeight)
	m.graph.cardRow = -1
	m.graph.roots = nil
	m.graph.up = make(map[string]string)
	m.graph.lines = make(map[[2]string][]point)
	m.graph.joints = make(map[string]point)
	group := visible*podWidth + podGap*(visible-1)
	first := 1 + max(bodyWidth-group, 0)/2
	x := first
	var drops []int
	for index, pod := range shown {
		lit := !m.graph.onePod && m.graph.focus == "" && m.graph.left+index == selected
		m.drawPod(tall, pod, x, head, podWidth, podHeight, lit)
		// The drop is at the pod center, or 1 cell after the title when
		// the title is wider than the left half of the pod.
		drops = append(drops, max(x+(podWidth-1)/2, x+2+ansi.StringWidth(podTitle(pod, podWidth))+1))
		x += podWidth + podGap
	}
	if hasCoordinator {
		// In the one-pod view the coordinator card is above the manager
		// card, or at the center of the pod, and moves with the tree.
		cx := (canvasWidth - component.CardWidth) / 2
		if m.graph.onePod {
			cx = first + (podWidth-component.CardWidth)/2
			if shown[0].manager != nil {
				cx = m.graph.managerTop.x - component.CardWidth/2
			}
		}
		m.drawCoordinator(tall, coordinator, cx, shown, drops, head)
	}
	m.lightLines(tall)
	room = max(room, 0)
	if row := m.graph.cardRow; row >= 0 {
		m.graph.top = min(max(m.graph.top, row+component.CardHeight-room), row)
	}
	m.graph.top = min(max(m.graph.top, 0), max(tall.height-room, 0))
	// Sideways, the selected card stays on the screen: from 120 columns
	// with the smallest move, below 120 columns at the center when possible
	// (decision 75).
	if col := m.graph.cardCol; m.graph.onePod && m.graph.cardRow >= 0 {
		m.graph.shift = min(max(m.graph.shift, col+component.CardWidth-inner), col)
		if !wide {
			m.graph.shift = col + component.CardWidth/2 - inner/2
		}
	}
	m.graph.shift = min(max(m.graph.shift, 0), canvasWidth-inner)
	canvas := newCellCanvas(inner, bodyHeight)
	for row := range min(room, tall.height) {
		from := (row+m.graph.top)*canvasWidth + m.graph.shift
		copy(canvas.cells[(row+gap)*inner:(row+gap+1)*inner], tall.cells[from:from+inner])
	}
	row := bodyHeight - footer
	if projects == 0 {
		canvas.text(1, row, agentsNoneText, cellgrid.Dim, false)
		row++
	}
	if wide {
		// The legend is the last row of the body (agents.md rule 7).
		legend := bodyHeight - 1
		canvas.text(1, legend, agentsLegend+agentsLegendWave, cellgrid.Dim, false)
		canvas.text(1+ansi.StringWidth(agentsLegend[:strings.Index(agentsLegend, agentsLegendLit)]), legend, agentsLegendLit, cellgrid.Info, false)
	}
	if scrolls {
		end := min(m.graph.left+visible, projects)
		scroll := fmt.Sprintf("%s · projects %d–%d of %d", agentsScrolls, min(m.graph.left+1, end), end, projects)
		if m.graph.left+visible > projects {
			scroll += " · home"
		}
		canvas.text(1, row, scroll, cellgrid.Dim, false)
		var edges []string
		if m.graph.left > 0 {
			edges = append(edges, "← "+pods[m.graph.left-1].group.id)
		}
		if end := m.graph.left + visible; end < len(pods) {
			edges = append(edges, pods[end].group.id+" →")
		}
		right := strings.Join(edges, "   ")
		canvas.text(inner-1-ansi.StringWidth(right), row, right, cellgrid.Dim, false)
	}
	if sideways {
		left, right := m.treeScrollText(inner)
		canvas.text(1, row, left, cellgrid.Dim, false)
		canvas.text(inner-1-ansi.StringWidth(right), row, right, cellgrid.Dim, false)
	}
	return fitBlock(strings.Join(canvas.lines(), "\n"), width, height)
}

// treeScrollText gives the scroll line of the one-pod view on a screen of
// inner columns (decision 75): the top cards that are fully on the screen,
// and the next hidden top card on the right, or else on the left.
func (m *Model) treeScrollText(inner int) (string, string) {
	roots := m.graph.roots
	first, last := -1, -1
	right := ""
	for index, root := range roots {
		switch {
		case root.x+component.CardWidth > m.graph.shift+inner:
			if right == "" {
				right = root.name + " →"
			}
		case root.x >= m.graph.shift:
			if first < 0 {
				first = index
			}
			last = index
		}
	}
	if right == "" {
		for index := len(roots) - 1; index >= 0; index-- {
			if roots[index].x < m.graph.shift {
				right = "← " + roots[index].name
				break
			}
		}
	}
	if first < 0 {
		return fmt.Sprintf("%s · %d children", agentsScrolls, len(roots)), right
	}
	return fmt.Sprintf("%s · children %d–%d of %d", agentsScrolls, first+1, last+1, len(roots)), right
}

// drawCoordinator draws the coordinator card at column cx and its line
// (decision 65). The joint is at the card x+12. In the overview, one bus row
// below the card goes to a drop at the center of each pod on the page; a
// drop 1 column or less from the joint moves to the joint. In the one-pod
// view, the line goes down through the pod top border to the top joint of
// the manager card, or to the pod top border when the pod has no manager.
// In home, the line goes to each top card, with a bus above the cards when
// there are several (rule A of decision 78).
func (m *Model) drawCoordinator(c *cellCanvas, node agentNode, cx int, shown []graphPod, drops []int, podTop int) {
	m.drawCard(c, node, cx, 0)
	if len(shown) == 0 {
		return
	}
	joint := cx + component.CardWidth/2
	bus := component.CardHeight
	c.addArms(joint, bus-1, component.Arms{Down: component.Light})
	root := node.record.Key
	if m.graph.onePod {
		pod := shown[0]
		target := point{joint, podTop}
		if pod.manager != nil {
			target = m.graph.managerTop
		}
		if pod.homeTree() && len(m.graph.roots) > 0 {
			for _, top := range m.graph.roots {
				center := top.x + component.CardWidth/2
				path := []point{{joint, bus - 1}, {joint, bus}, {center, bus}, {center, top.y}}
				if len(m.graph.roots) > 1 {
					path = []point{{joint, bus - 1}, {joint, top.y - 2}, {center, top.y - 2}, {center, top.y}}
				}
				m.graph.up[top.key] = root
				m.graph.lines[[2]string{root, top.key}] = m.connect(c, path)
			}
			return
		}
		cells := m.connect(c, []point{{joint, bus - 1}, {joint, bus}, {target.x, bus}, target})
		if pod.manager != nil {
			m.graph.up[pod.manager.record.Key] = root
			m.graph.lines[[2]string{root, pod.manager.record.Key}] = cells
			return
		}
		m.graph.up[podKey(pod)] = root
		m.graph.lines[[2]string{root, podKey(pod)}] = cells
		return
	}
	for index, drop := range drops {
		if drop >= joint-1 && drop <= joint+1 {
			drops[index] = joint
		}
	}
	low, high := slices.Min(append(drops, joint)), slices.Max(append(drops, joint))
	for x := low; x <= high; x++ {
		arms := component.Arms{}
		if x > low {
			arms.Left = component.Light
		}
		if x < high {
			arms.Right = component.Light
		}
		if x == joint {
			arms.Up = component.Light
		}
		if slices.Contains(drops, x) {
			arms.Down = component.Light
		}
		c.addArms(x, bus, arms)
	}
	for index, pod := range shown {
		drop := drops[index]
		c.addArms(drop, podTop, component.Arms{Up: component.Light})
		path := []point{{joint, bus - 1}, {joint, bus}}
		for x := joint; x != drop; {
			if x < drop {
				x++
			} else {
				x--
			}
			path = append(path, point{x, bus})
		}
		m.graph.up[podKey(pod)] = root
		m.graph.lines[[2]string{root, podKey(pod)}] = append(path, point{drop, podTop})
	}
}

// connect draws a line through the corners of path and gives each of its
// cells, first corner first. Each corner and each end of the line adds the
// arms of the line to the cell that is there.
func (m *Model) connect(c *cellCanvas, path []point) []point {
	cells := pathCells(path)
	for index, cell := range cells {
		var arms component.Arms
		for _, other := range []int{index - 1, index + 1} {
			if other >= 0 && other < len(cells) {
				arms = withArm(arms, cell, cells[other], component.Light)
			}
		}
		c.addArms(cell.x, cell.y, arms)
	}
	return cells
}

// pathCells gives each cell of a line through the corners of path, first
// corner first.
func pathCells(path []point) []point {
	cells := []point{path[0]}
	for _, corner := range path[1:] {
		for last := cells[len(cells)-1]; last != corner; last = cells[len(cells)-1] {
			next := last
			switch {
			case last.x < corner.x:
				next.x++
			case last.x > corner.x:
				next.x--
			case last.y < corner.y:
				next.y++
			default:
				next.y--
			}
			cells = append(cells, next)
		}
	}
	return cells
}

// withArm gives arms with the arm of from toward to set to weight.
func withArm(arms component.Arms, from, to point, weight component.Weight) component.Arms {
	switch {
	case to.y < from.y:
		arms.Up = weight
	case to.y > from.y:
		arms.Down = weight
	case to.x < from.x:
		arms.Left = weight
	case to.x > from.x:
		arms.Right = weight
	}
	return arms
}

// podTitle gives the title of a pod on its top border. The title leaves 1
// border cell and the drop before the corner.
func podTitle(pod graphPod, width int) string {
	title := pod.group.id
	if pod.group.wave != "" {
		title += " · " + pod.group.wave
	}
	return " " + ansi.Truncate(title, width-7, "…") + " "
}

// podLines gives the number of rows inside the border of a pod.
func (m *Model) podLines(pod graphPod, width int) int {
	content := width - 4
	rows := 1 // the goal line
	tree := layoutTree(pod, content)
	switch {
	case m.graph.onePod && pod.homeTree():
		rows += tree.height
	case m.graph.onePod && pod.manager == nil:
		rows += 2
		if len(pod.children) > 0 {
			rows += 1 + tree.height
		}
	case pod.homeTree():
		rows += 5*len(pod.children) - 1
	case pod.manager == nil:
		rows += 2 + 5*len(pod.children)
	case m.graph.onePod || pod.tree(content):
		rows += tree.height
		if len(pod.children) == 0 {
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

// hideable reports that h hides the card of node: a stopped agent, or a
// stale agent. A stale card is current only with no live successor.
func (n agentNode) hideable() bool {
	return n.stopped || n.record.Status == records.AgentStale
}

// note gives the stopped note of a pod: the h key and the number of stopped
// agents.
func (p graphPod) note() string {
	if p.hidden > 0 {
		return fmt.Sprintf("h shows %d stopped", p.hidden)
	}
	stopped := 0
	for _, node := range p.nodes() {
		if node.hideable() {
			stopped++
		}
	}
	if stopped == 0 {
		return ""
	}
	return fmt.Sprintf("h hides %d stopped", stopped)
}

// drawPod draws one project column at x, y of width×height cells. The
// selected pod of the overview has a heavy accent border and an accent title
// (decision 62).
func (m *Model) drawPod(c *cellCanvas, pod graphPod, x, y, width, height int, selected bool) {
	token, title := cellgrid.Border, cellgrid.Dim
	if selected {
		token, title = cellgrid.Accent, cellgrid.Accent
	}
	c.box(x, y, width, height, token, selected)
	c.text(x+2, y, podTitle(pod, width), title, true)
	left, content, top := x+2, width-4, y+1
	goal := cellgrid.Dim
	if pod.group.goalStop {
		goal = cellgrid.Accent
	}
	// In the one-pod view the coordinator line goes down through the goal
	// row at the center of the manager card.
	goalRoom := content
	switch {
	case m.graph.onePod && pod.manager != nil:
		goalRoom = min(content, (layoutTree(pod, content).width-component.CardWidth)/2+component.CardWidth/2-1)
	case m.graph.onePod && pod.managerHidden:
		// The line goes down from the coordinator joint at the pod center.
		goalRoom = min(content, (width-component.CardWidth)/2+component.CardWidth/2-3)
	}
	c.text(left, top, ansi.Truncate(pod.group.goal, goalRoom, "…"), goal, false)
	top++
	if pod.manager != nil {
		m.graph.up[pod.manager.record.Key] = podKey(pod)
	}
	var next int
	switch {
	case m.graph.onePod && pod.homeTree():
		next = m.drawTree(c, pod, left, top, content)
	case m.graph.onePod && pod.manager == nil:
		c.text(left, top+1, "no manager", cellgrid.Dim, false)
		next = m.drawTree(c, pod, left, top+3, content)
	case pod.homeTree():
		next = m.drawRail(c, pod, left, top-1)
	case pod.manager == nil:
		c.text(left, top+1, "no manager", cellgrid.Dim, false)
		next = m.drawRail(c, pod, left, top+2)
	case m.graph.onePod || pod.tree(content):
		next = m.drawTree(c, pod, left, top, content)
		if len(pod.children) == 0 {
			text := "no child agents"
			if note := pod.note(); note != "" {
				text = note
			}
			c.text(left, next+1, text, cellgrid.Dim, false)
		}
	default:
		m.drawCard(c, *pod.manager, left, top)
		m.graph.managerTop = point{left + component.CardWidth/2, top}
		next = m.drawRail(c, pod, left, top+component.CardHeight)
	}
	if note := pod.note(); note != "" && len(pod.children) > 0 {
		c.text(left, next+1, note, cellgrid.Dim, false)
	}
}

// treeLayout is the tree form of a pod (decision 74), relative to the first
// content cell of the tree: the width and the height of the tree, and for
// each child of the pod its card column and row and the index of its parent
// card, or -1 for a top card.
type treeLayout struct {
	width, height int
	x, y, parent  []int
}

// layoutTree gives the tree form of a pod in content cells. Each card is at
// the center of its span. A span has one slot for each card at the bottom of
// the subtree. A slot is content divided by the slots of the tree, and at
// least treeSlot cells. A tree that does not fit is wider than content. A
// level is one │ row below its parents, or 3 rows with a bus when a parent
// has more than one child. The children of the manager always use the bus.
func layoutTree(pod graphPod, content int) treeLayout {
	n := len(pod.children)
	t := treeLayout{x: make([]int, n), y: make([]int, n), parent: make([]int, n)}
	kids := make([][]int, n)
	var roots []int
	for index, level := range pod.levels {
		t.parent[index] = -1
		for before := index - 1; before >= 0; before-- {
			if pod.levels[before] == level-1 {
				t.parent[index] = before
				break
			}
		}
		if parent := t.parent[index]; parent >= 0 {
			kids[parent] = append(kids[parent], index)
		} else {
			roots = append(roots, index)
		}
	}
	units := make([]int, n)
	for index := n - 1; index >= 0; index-- {
		for _, kid := range kids[index] {
			units[index] += units[kid]
		}
		units[index] = max(units[index], 1)
	}
	total := 0
	for _, root := range roots {
		total += units[root]
	}
	slot := treeSlot
	if total > 0 {
		slot = max(treeSlot, content/total)
	}
	t.width = max(content, total*slot)
	var place func(index, start int)
	place = func(index, start int) {
		t.x[index] = start + (units[index]*slot-component.CardWidth)/2
		for _, kid := range kids[index] {
			place(kid, start)
			start += units[kid] * slot
		}
	}
	start := 0
	for _, root := range roots {
		place(root, start)
		start += units[root] * slot
	}
	low, row := 0, 0
	if pod.manager != nil {
		low, row = 1, component.CardHeight
	}
	rows := make(map[int]int)
	for level := low; level <= slices.Max(append(slices.Clone(pod.levels), low-1)); level++ {
		gap := 1
		switch {
		case level == low && pod.homeTree() && len(roots) > 1:
			// The bus of the coordinator line (rule A of decision 78).
			gap = 2
		case level == low && pod.manager == nil:
			gap = 0
		case level == low:
			gap = 3
		default:
			for index, above := range pod.levels {
				if above == level-1 && len(kids[index]) > 1 {
					gap = 3
				}
			}
		}
		rows[level] = row + gap
		row = rows[level] + component.CardHeight
	}
	for index, level := range pod.levels {
		t.y[index] = rows[level]
	}
	t.height = row
	return t
}

// drawTree draws the manager and the children of a pod in the tree form
// from column left and row top, and gives the first row below them. Each
// parent has the joint ┬ on its bottom border and a │ row below it. A level
// with a bus has the bus ╭─┴─╮ and one more │ row. Each child has the joint
// ┴ on its top border.
func (m *Model) drawTree(c *cellCanvas, pod graphPod, left, top, content int) int {
	tree := layoutTree(pod, content)
	if pod.manager != nil {
		mx := left + (tree.width-component.CardWidth)/2
		m.drawCard(c, *pod.manager, mx, top)
		m.graph.managerTop = point{mx + component.CardWidth/2, top}
	}
	kids := make(map[int][]int)
	for index, child := range pod.children {
		m.drawCard(c, child, left+tree.x[index], top+tree.y[index])
		kids[tree.parent[index]] = append(kids[tree.parent[index]], index)
		if tree.parent[index] < 0 && m.graph.onePod {
			m.graph.roots = append(m.graph.roots, treeRoot{key: child.record.Key, name: child.record.Name, x: left + tree.x[index], y: top + tree.y[index]})
		}
	}
	for parent := -1; parent < len(pod.children); parent++ {
		if len(kids[parent]) == 0 {
			continue
		}
		var node agentNode
		var center, bottom int
		switch {
		case parent >= 0:
			node = pod.children[parent]
			center, bottom = left+tree.x[parent]+component.CardWidth/2, top+tree.y[parent]+component.CardHeight-1
		case pod.manager != nil:
			node = *pod.manager
			center, bottom = m.graph.managerTop.x, top+component.CardHeight-1
		default:
			for _, root := range kids[parent] {
				m.graph.up[pod.children[root].record.Key] = podKey(pod)
			}
			continue
		}
		kidTop := top + tree.y[kids[parent][0]]
		bus := bottom + 1
		m.junction(c, center, bottom, node, '┬', '┯')
		c.set(center, bottom+1, '│', cellgrid.Border)
		var centers []int
		for _, kid := range kids[parent] {
			centers = append(centers, left+tree.x[kid]+component.CardWidth/2)
		}
		if kidTop-bottom > 2 {
			bus++
			first, last := centers[0], centers[len(centers)-1]
			for col := first; col <= last; col++ {
				c.set(col, bus, '─', cellgrid.Border)
			}
			for _, col := range centers {
				c.set(col, bus, '┬', cellgrid.Border)
				c.set(col, bus+1, '│', cellgrid.Border)
			}
			if first != last {
				c.set(first, bus, '╭', cellgrid.Border)
				c.set(last, bus, '╮', cellgrid.Border)
			}
			mark := '┴'
			if slices.Contains(centers, center) {
				mark = '┼'
				if first == last {
					mark = '│'
				}
			}
			c.set(center, bus, mark, cellgrid.Border)
		}
		for index, kid := range kids[parent] {
			child := pod.children[kid]
			m.junction(c, centers[index], kidTop, child, '┴', '┷')
			m.graph.up[child.record.Key] = node.record.Key
			m.graph.lines[[2]string{node.record.Key, child.record.Key}] = pathCells([]point{{center, bottom}, {center, bus}, {centers[index], bus}, {centers[index], kidTop}})
		}
	}
	return top + tree.height
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
	// parent gives the card above a card of the rail: the nearest card
	// before it with a level of one less, or the manager.
	parent := func(index int) string {
		for before := index - 1; before >= 0; before-- {
			if pod.levels[before] == pod.levels[index]-1 {
				return pod.children[before].record.Key
			}
		}
		if pod.manager != nil {
			return pod.manager.record.Key
		}
		return podKey(pod)
	}
	for index, child := range pod.children {
		level := pod.levels[index]
		if level == 0 {
			m.graph.up[child.record.Key] = podKey(pod)
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
			m.graph.joints[parent.record.Key] = point{rail, next - 1}
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
		from := parent(index)
		joint := m.graph.joints[from]
		var line []point
		for row := joint.y; row <= next+2; row++ {
			line = append(line, point{rail, row})
		}
		m.graph.up[child.record.Key] = from
		m.graph.lines[[2]string{from, child.record.Key}] = append(line, point{rail + 1, next + 2})
		next += 5
	}
	return next
}

// junction sets a connector cell on the border of the card of node: light in
// border, or in working for a working card (decision 60), or for the selected
// card, which has a heavy border, the mixed-weight rune heavy in accent.
func (m *Model) junction(c *cellCanvas, x, y int, node agentNode, light, heavy rune) {
	switch {
	case node.record.Key == m.graph.focus:
		c.set(x, y, heavy, cellgrid.Accent)
	case node.status == "working":
		c.set(x, y, light, cellgrid.Working)
	default:
		c.set(x, y, light, cellgrid.Border)
	}
}

// drawCard draws the card of one agent with component.Card. Only the
// focus card has the heavy border.
func (m *Model) drawCard(c *cellCanvas, node agentNode, x, y int) {
	card := component.Card{Name: node.record.Name, Sub: agentCardSub(node), State: node.status, Selected: node.record.Key == m.graph.focus}
	if card.Selected {
		m.graph.cardRow, m.graph.cardCol = y, x
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
	// The coordinator card is in no pod and has no goal line.
	goal := ""
	if column >= 0 {
		goal = m.agentsPods()[column].group.goal
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
		field("goal", goal),
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

// updateAgentsKey handles a key of the Agents graph (decisions 14, 22, 23,
// and 62 to 64).
func (m *Model) updateAgentsKey(key string) tea.Cmd {
	pods := m.agentsPods()
	node, _, ok := m.selectedAgent()
	if !m.inPodView(pods) {
		return m.updatePodsKey(key, pods, node, ok)
	}
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
	case "esc":
		if m.graph.podView {
			m.graph.podView = false
			m.graph.selected = m.graph.pod
			return m.logAction("agent pod close", slog.String("pod", m.graph.pod))
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

// updatePodsKey handles a key of the overview of decision 62: ← and →
// select a pod, ↑ selects the coordinator card, ↓ selects the pod again,
// and enter opens the pod or the details of the coordinator.
func (m *Model) updatePodsKey(key string, pods []graphPod, node agentNode, ok bool) tea.Cmd {
	column := m.podIndex(pods)
	coordinator, hasCoordinator := m.coordinatorCard()
	selected := m.coordinatorSelected(pods)
	switch key {
	case "left", "right":
		next := column + 1
		if key == "left" {
			next = column - 1
		}
		if selected || next < 0 || next >= len(pods) {
			return nil
		}
		m.graph.pod = pods[next].group.id
		m.graph.selected = m.graph.pod
		return m.logAction("agent pod select", slog.String("pod", m.graph.pod))
	case "up", "k":
		if hasCoordinator && !selected {
			m.graph.selected = coordinator.record.Key
			return m.logAction("agent selection", slog.String("agent_id", m.graph.selected))
		}
	case "down", "j":
		if selected {
			m.graph.selected = m.graph.pod
			return m.logAction("agent pod select", slog.String("pod", m.graph.pod))
		}
	case "enter":
		if selected {
			m.popupOpen = true
			m.popupOffset = 0
			return m.logAction("agent details open", slog.String("agent_id", node.record.Key))
		}
		if len(pods) == 0 {
			return nil
		}
		m.graph.podView = true
		// The first card of the pod is selected when the selection is not
		// in it.
		m.selectedAgent()
		return m.logAction("agent pod open", slog.String("pod", m.graph.pod), slog.String("agent_id", m.graph.selected))
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

// moveAgentSelection moves the selection in the one-pod view by one arrow
// key in the tree (decision 74). ↑ selects the parent, ↓ the first child,
// and ← and → the card before or after on the same level. The selection
// stays in the pod.
func (m *Model) moveAgentSelection(key string) bool {
	node, column, ok := m.selectedAgent()
	if !ok || column < 0 {
		return false
	}
	pod := m.agentsPods()[column]
	tree := layoutTree(pod, 0)
	index := slices.IndexFunc(pod.children, func(n agentNode) bool { return n.record.Key == node.record.Key })
	target := -2
	switch key {
	case "up", "k":
		if index >= 0 {
			target = tree.parent[index]
		}
		if target == -1 && pod.manager == nil {
			target = -2
		}
	case "down", "j":
		target = slices.Index(tree.parent, index)
		if target < 0 {
			target = -2
		}
	case "left", "right":
		if index < 0 {
			break
		}
		var level []int
		for other := range pod.children {
			if pod.levels[other] == pod.levels[index] {
				level = append(level, other)
			}
		}
		at := slices.Index(level, index)
		if key == "left" && at > 0 {
			target = level[at-1]
		}
		if key == "right" && at < len(level)-1 {
			target = level[at+1]
		}
	}
	switch {
	case target >= 0:
		m.graph.selected = pod.children[target].record.Key
	case target == -1 && pod.manager != nil:
		m.graph.selected = pod.manager.record.Key
	default:
		return false
	}
	return true
}

// projectGoalLine gives the goal line of a project pod: the active goal and
// its open wave, or the last goal and its status. The flag tells that the
// open wave needs the user.
func projectGoalLine(goals []records.Goal, project string) (string, bool) {
	goal := projectGoal(goals, project)
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

// projectGoal gives the goal of a project pod: the active goal, or the last
// goal. It gives nil when the project has no goal.
func projectGoal(goals []records.Goal, project string) *records.Goal {
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
	return goal
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

// litWindow is the time that a line stays lit after a message (decision 67).
const litWindow = 5 * time.Second

// recentMessages gives the messages of the last 5 seconds.
func (m *Model) recentMessages() []agentMessage {
	now := m.now()
	var out []agentMessage
	for _, message := range m.agents.data.messages {
		if !message.at.After(now) && now.Before(message.at.Add(litWindow)) {
			out = append(out, message)
		}
	}
	return out
}

// litMessages gives the messages that light a line now, with the card keys
// of the last view as sender and receiver: for each line, the newest
// message of the last 5 seconds, oldest first. A record key and an agent
// name of one card name one line.
func (m *Model) litMessages() []agentMessage {
	newest := make(map[[2]string]agentMessage)
	for _, message := range m.recentMessages() {
		from, okFrom := m.graphKey(message.from)
		to, okTo := m.graphKey(message.to)
		if !okFrom || !okTo || from == to {
			continue
		}
		message.from, message.to = from, to
		pair := [2]string{from, to}
		if to < from {
			pair = [2]string{to, from}
		}
		if old, ok := newest[pair]; !ok || message.at.After(old.at) {
			newest[pair] = message
		}
	}
	out := make([]agentMessage, 0, len(newest))
	for _, message := range newest {
		out = append(out, message)
	}
	slices.SortFunc(out, func(a, b agentMessage) int {
		return cmp.Or(a.at.Compare(b.at), strings.Compare(a.from, b.from), strings.Compare(a.to, b.to))
	})
	return out
}

// agentsLitMsg tells that a lit line goes dark: the Agents screen draws
// again.
type agentsLitMsg struct{}

// litTick starts a timer to the time at which the next lit line goes dark.
// The demo has a fixed clock and no timer.
func (m *Model) litTick() tea.Cmd {
	next, ok := m.nextLitExpiry()
	if m.demo != nil || !ok || next.Equal(m.graph.litAt) {
		return nil
	}
	m.graph.litAt = next
	return tea.Tick(next.Sub(m.now()), func(time.Time) tea.Msg { return agentsLitMsg{} })
}

// nextLitExpiry gives the time at which the next lit line goes dark.
func (m *Model) nextLitExpiry() (time.Time, bool) {
	var next time.Time
	for _, message := range m.recentMessages() {
		if end := message.at.Add(litWindow); next.IsZero() || end.Before(next) {
			next = end
		}
	}
	return next, !next.IsZero()
}

// graphKey gives the key of the card of an inbox name or an agent name on
// the last view: "coordinator" is the coordinator card.
func (m *Model) graphKey(name string) (string, bool) {
	for _, node := range m.agents.data.nodes {
		if !node.current || node.record.Role == "supervisor" {
			continue
		}
		if node.record.Key == name || node.record.Name == name || name == "coordinator" && node.record.Role == "coordinator" {
			_, onGraph := m.graph.up[node.record.Key]
			if onGraph || node.record.Role == "coordinator" {
				return node.record.Key, true
			}
		}
	}
	return "", false
}

// graphPath gives the cells of the line from the card of from to the card
// of to on the last view, through their nearest common parent.
func (m *Model) graphPath(from, to string) []point {
	chain := func(key string) []string {
		out := []string{key}
		for {
			parent, ok := m.graph.up[key]
			if !ok {
				return out
			}
			out = append(out, parent)
			key = parent
		}
	}
	up, down := chain(from), chain(to)
	common := ""
	for _, key := range up {
		if slices.Contains(down, key) {
			common = key
			break
		}
	}
	if common == "" {
		return nil
	}
	// cells gives the cells from the common parent down to key.
	cells := func(keys []string) []point {
		var out []point
		at := slices.Index(keys, common)
		for index := at; index > 0; index-- {
			for _, cell := range m.graph.lines[[2]string{keys[index], keys[index-1]}] {
				if len(out) == 0 || out[len(out)-1] != cell {
					out = append(out, cell)
				}
			}
		}
		return out
	}
	sender, receiver := cells(up), cells(down)
	shared := 0
	for shared < len(sender) && shared < len(receiver) && sender[shared] == receiver[shared] {
		shared++
	}
	// The two lines share their first cells: the path turns at the last
	// shared cell.
	pivot := max(shared-1, 0)
	var path []point
	for index := len(sender) - 1; index >= pivot; index-- {
		path = append(path, sender[index])
	}
	if shared == 0 {
		return append(path, receiver...)
	}
	return append(path, receiver[pivot+1:]...)
}

// lightLines draws each lit line of decision 66 on c: heavy arms in info on
// each cell of the path, and one arrow on a straight cell near the
// receiver that points to the receiver.
func (m *Model) lightLines(c *cellCanvas) {
	for _, message := range m.litMessages() {
		path := m.graphPath(message.from, message.to)
		if len(path) < 2 {
			continue
		}
		for index, cell := range path {
			arms := c.arms(cell.x, cell.y)
			for _, other := range []int{index - 1, index + 1} {
				if other >= 0 && other < len(path) {
					arms = withArm(arms, cell, path[other], component.Heavy)
				}
			}
			if r, ok := component.BoxRune(arms); ok {
				c.set(cell.x, cell.y, r, cellgrid.Info)
			}
		}
		// The arrow is on the first straight cell from the receiver. On a
		// horizontal line it is one cell further, so that it is not next
		// to a corner.
		for index := len(path) - 2; index > 0; index-- {
			cell := path[index]
			arms := c.arms(cell.x, cell.y)
			if !straight(arms) {
				continue
			}
			if arms.Left != component.None && index > 1 && straight(c.arms(path[index-1].x, path[index-1].y)) {
				index--
				cell = path[index]
			}
			c.set(cell.x, cell.y, arrow(cell, path[index+1]), cellgrid.Info)
			break
		}
	}
}

// straight reports that arms are one straight line.
func straight(arms component.Arms) bool {
	vertical := arms.Up != component.None && arms.Down != component.None && arms.Left == component.None && arms.Right == component.None
	horizontal := arms.Left != component.None && arms.Right != component.None && arms.Up == component.None && arms.Down == component.None
	return vertical || horizontal
}

// arrow gives the arrow that points from cell to next.
func arrow(cell, next point) rune {
	switch {
	case next.y < cell.y:
		return '▲'
	case next.y > cell.y:
		return '▼'
	case next.x < cell.x:
		return '◀'
	}
	return '▶'
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

// box draws a rounded border of width×height cells from x, y, or a heavy
// border with square corners.
func (c *cellCanvas) box(x, y, width, height int, token cellgrid.Token, heavy bool) {
	glyphs := []string{"─", "│", "╭", "╮", "╰", "╯"}
	if heavy {
		glyphs = []string{"━", "┃", "┏", "┓", "┗", "┛"}
	}
	for col := x + 1; col < x+width-1; col++ {
		c.put(col, y, canvasCell{text: glyphs[0], token: token})
		c.put(col, y+height-1, canvasCell{text: glyphs[0], token: token})
	}
	for row := y + 1; row < y+height-1; row++ {
		c.put(x, row, canvasCell{text: glyphs[1], token: token})
		c.put(x+width-1, row, canvasCell{text: glyphs[1], token: token})
	}
	c.put(x, y, canvasCell{text: glyphs[2], token: token})
	c.put(x+width-1, y, canvasCell{text: glyphs[3], token: token})
	c.put(x, y+height-1, canvasCell{text: glyphs[4], token: token})
	c.put(x+width-1, y+height-1, canvasCell{text: glyphs[5], token: token})
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

// arms gives the box arms of the cell at x, y.
func (c *cellCanvas) arms(x, y int) component.Arms {
	if x < 0 || x >= c.width || y < 0 || y >= c.height {
		return component.Arms{}
	}
	r, _ := utf8.DecodeRuneInString(c.cells[y*c.width+x].text)
	arms, _ := component.BoxArms(r)
	return arms
}

// addArms adds arms to the box glyph at x, y. A blank cell gets a glyph in
// border. A glyph keeps its token.
func (c *cellCanvas) addArms(x, y int, add component.Arms) {
	if x < 0 || x >= c.width || y < 0 || y >= c.height {
		return
	}
	cell := c.cells[y*c.width+x]
	arms := c.arms(x, y)
	if arms == (component.Arms{}) {
		cell.token = cellgrid.Border
	}
	arms.Up, arms.Down = max(arms.Up, add.Up), max(arms.Down, add.Down)
	arms.Left, arms.Right = max(arms.Left, add.Left), max(arms.Right, add.Right)
	if r, ok := component.BoxRune(arms); ok {
		cell.text = string(r)
		c.cells[y*c.width+x] = cell
	}
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
