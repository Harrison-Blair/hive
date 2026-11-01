package graph

import (
	"sort"
	"strings"

	"github.com/Harrison-Blair/hive/internal/agent/list"
)

// HomeTitle is the title of the pod of the agents outside each project pod.
const HomeTitle = "home"

const coordinatorSuffix = "-coordinator"

// Pod is one column of the graph: the agents of a project, or of home.
type Pod struct {
	Title string
	Roots []*Node
}

// Node is one card of a pod and the cards of its children in the same pod.
type Node struct {
	Entry    list.Entry
	Children []*Node
}

// Order returns the nodes of the pod in pre-order, the order of the cards on
// the screen.
func (p Pod) Order() []*Node {
	var nodes []*Node
	var walk func([]*Node)
	walk = func(level []*Node) {
		for _, n := range level {
			nodes = append(nodes, n)
			walk(n.Children)
		}
	}
	walk(p.Roots)
	return nodes
}

// Pods groups the entries into pods. An agent named <project>-coordinator
// starts the pod of that project, and each descendant of it through the
// parent field joins that pod. Each other agent is in the home pod. The
// project pods come in title order, then home. Inside a pod, an agent is a
// child of its parent when the parent is in the same pod and the parent chain
// has no cycle. Otherwise it is a root.
func Pods(entries []list.Entry) []Pod {
	sorted := append([]list.Entry(nil), entries...)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].Name < sorted[j].Name })
	byName := make(map[string]list.Entry, len(sorted))
	for _, e := range sorted {
		byName[e.Name] = e
	}
	podOf := func(name string) string {
		seen := map[string]bool{}
		for name != "" && !seen[name] {
			seen[name] = true
			if project, ok := strings.CutSuffix(name, coordinatorSuffix); ok && project != "" {
				return project
			}
			name = byName[name].Parent
		}
		return HomeTitle
	}
	// cyclic reports whether the parent chain of name comes back to name.
	cyclic := func(name string) bool {
		seen := map[string]bool{}
		for at := byName[name].Parent; at != "" && !seen[at]; at = byName[at].Parent {
			if at == name {
				return true
			}
			seen[at] = true
		}
		return false
	}
	nodes := make(map[string]*Node, len(sorted))
	for _, e := range sorted {
		nodes[e.Name] = &Node{Entry: e}
	}
	pods := map[string]*Pod{}
	for _, e := range sorted {
		title := podOf(e.Name)
		pod := pods[title]
		if pod == nil {
			pod = &Pod{Title: title}
			pods[title] = pod
		}
		parent, ok := nodes[e.Parent]
		if ok && podOf(e.Parent) == title && !cyclic(e.Name) {
			parent.Children = append(parent.Children, nodes[e.Name])
			continue
		}
		pod.Roots = append(pod.Roots, nodes[e.Name])
	}
	titles := make([]string, 0, len(pods))
	for title := range pods {
		if title != HomeTitle {
			titles = append(titles, title)
		}
	}
	sort.Strings(titles)
	if pods[HomeTitle] != nil {
		titles = append(titles, HomeTitle)
	}
	out := make([]Pod, 0, len(titles))
	for _, title := range titles {
		out = append(out, *pods[title])
	}
	return out
}
