package tui

import (
	"slices"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/charmbracelet/x/ansi"
)

// profileAgent gives a live agent with the D1 facts of a profile.
func profileAgent(key, profile, project, owns, parent, started string) records.Agent {
	task := ""
	if profile == records.ProfileBuilder || profile == records.ProfileVerifier {
		task = key
	}
	return records.Agent{
		Key: key, Name: key, Profile: profile, Project: project, Owns: owns, Parent: parent, Task: task,
		Harness: "pi", Status: records.AgentLive, AgentStatus: "idle",
		StatusChangedAt: "2026-10-21T14:19:00Z", StartedAt: started, UpdatedAt: started,
	}
}

// D11: the Agents screen groups an agent by its project and puts the owner
// first. An agent with no project is in the home group. The owner comes
// from owns, not from the wave-A role mirror.
func TestAgentsGroupsByProjectOwnerFirstAndHome(t *testing.T) {
	ownerAlpha := profileAgent("owner-alpha", records.ProfilePlanner, "alpha", "alpha", "", "2026-10-21T14:05:00Z")
	ownerAlpha.Role = "worker"
	source := agentsSource{
		projects: []records.Project{{ID: "alpha", Path: "/repo/alpha", Base: "dev"}, {ID: "beta", Path: "/repo/beta", Base: "dev"}},
		agents: []records.Agent{
			profileAgent("builder-alpha", records.ProfileBuilder, "alpha", "", "owner-alpha", "2026-10-21T14:01:00Z"),
			ownerAlpha,
			profileAgent("owner-beta", records.ProfilePlanner, "beta", "beta", "", "2026-10-21T14:02:00Z"),
			profileAgent("builder-home", records.ProfileBuilder, "", "", "", "2026-10-21T14:03:00Z"),
		},
	}
	m := agentsModel(t, source)
	pods := m.agentsPods()
	var got []string
	for _, pod := range pods {
		manager := ""
		if pod.manager != nil {
			manager = pod.manager.record.Key
		}
		var children []string
		for _, child := range pod.children {
			children = append(children, child.record.Key)
		}
		got = append(got, pod.group.id+":"+manager+":"+strings.Join(children, ","))
	}
	want := []string{"alpha:owner-alpha:builder-alpha", "beta:owner-beta:", "home::builder-home"}
	if !slices.Equal(got, want) {
		t.Fatalf("pods = %q; want %q", got, want)
	}
}

// D11: a general agent with a project is in the group of its project also
// when its parent card is not on the graph.
func TestAgentsGeneralAgentWithProjectStaysInItsProject(t *testing.T) {
	owner := profileAgent("owner-alpha", records.ProfilePlanner, "alpha", "alpha", "", "2026-10-21T14:00:00Z")
	owner.Role = "manager"
	review := profileAgent("review-alpha", records.ProfileGeneral, "alpha", "", "missing-gone", "2026-10-21T14:01:00Z")
	review.Role = "agent"
	source := agentsSource{
		projects: []records.Project{{ID: "alpha", Path: "/repo/alpha", Base: "dev"}},
		agents:   []records.Agent{owner, review},
	}
	m := agentsModel(t, source)
	if pod, _ := podOf(t, m, "review-alpha"); pod != "alpha" {
		t.Fatalf("review-alpha is in pod %q; want alpha", pod)
	}
}

// D11: an agent with no project is in the home group, also when its parent
// card is in a project pod.
func TestAgentsAgentWithoutProjectGoesHome(t *testing.T) {
	source := agentsSource{
		projects: []records.Project{{ID: "alpha", Path: "/repo/alpha", Base: "dev"}},
		agents: []records.Agent{
			profileAgent("owner-alpha", records.ProfilePlanner, "alpha", "alpha", "", "2026-10-21T14:00:00Z"),
			profileAgent("builder-alpha", records.ProfileBuilder, "alpha", "", "owner-alpha", "2026-10-21T14:01:00Z"),
			profileAgent("builder-home", records.ProfileBuilder, "", "", "builder-alpha", "2026-10-21T14:02:00Z"),
		},
	}
	m := agentsModel(t, source)
	if pod, _ := podOf(t, m, "builder-home"); pod != "home" {
		t.Fatalf("builder-home is in pod %q; want home", pod)
	}
}

// D11: each details row shows the profile label. No row shows a role name.
// A record written before the facts migration keeps its profile label,
// derived from the facts.
func TestAgentsDetailsShowProfileLabels(t *testing.T) {
	legacy := profileAgent("legacy-worker", "", "alpha", "", "owner-alpha", "2026-10-21T14:04:00Z")
	legacy.Role = "worker"
	source := agentsSource{
		projects: []records.Project{{ID: "alpha", Path: "/repo/alpha", Base: "dev"}},
		agents: []records.Agent{
			profileAgent("owner-alpha", records.ProfilePlanner, "alpha", "alpha", "", "2026-10-21T14:00:00Z"),
			profileAgent("builder-alpha", records.ProfileBuilder, "alpha", "", "owner-alpha", "2026-10-21T14:01:00Z"),
			profileAgent("verifier-alpha", records.ProfileVerifier, "alpha", "", "owner-alpha", "2026-10-21T14:02:00Z"),
			profileAgent("general-alpha", records.ProfileGeneral, "alpha", "", "owner-alpha", "2026-10-21T14:03:00Z"),
			legacy,
		},
	}
	m := agentsModel(t, source)
	for _, tc := range []struct {
		key  string
		want string
	}{
		{"owner-alpha", "profile   planner"},
		{"builder-alpha", "profile   builder"},
		{"verifier-alpha", "profile   verifier"},
		{"general-alpha", "profile   general"},
		{"legacy-worker", "profile   builder"},
	} {
		m.graph.selected = tc.key
		_, lines, _ := m.agentDetailsPopup()
		plain := make([]string, 0, len(lines))
		for _, line := range lines {
			plain = append(plain, ansi.Strip(line))
		}
		if !slices.Contains(plain, tc.want) {
			t.Fatalf("%s details lines = %q; want %q", tc.key, plain, tc.want)
		}
		for _, line := range plain {
			if strings.HasPrefix(line, "role ") || strings.HasPrefix(line, "role      ") {
				t.Fatalf("%s details show a role row: %q", tc.key, line)
			}
		}
	}
}

// d11Agent gives a live record with D1 facts and the wave-A role mirror that
// hive writes from those facts.
func d11Agent(key, profile, project, owns, parent, started string) records.Agent {
	record := records.Agent{
		Key: key, Name: key, Profile: profile, Project: project, Owns: owns, Parent: parent,
		Harness: "pi", Status: records.AgentLive, AgentStatus: "idle",
		StatusChangedAt: "2026-10-21T14:19:00Z", StartedAt: started, UpdatedAt: started,
	}
	record.SyncRole()
	return record
}

// F1 of D11: only owns gives ownership. A profile label is free, so a
// planner with empty owns is a project member, and the actual owner is first
// even when its profile is not planner.
func TestAgentsOwnerComesOnlyFromOwns(t *testing.T) {
	planner := d11Agent("a-planner", records.ProfilePlanner, "alpha", "", "", "2026-10-21T14:01:00Z")
	owner := d11Agent("z-owner", "review-specialist", "alpha", "alpha", "", "2026-10-21T14:02:00Z")
	m := agentsModel(t, agentsSource{
		projects: []records.Project{{ID: "alpha", Path: "/repo/alpha", Base: "dev"}},
		agents:   []records.Agent{planner, owner},
	})
	pods := m.agentsPods()
	if len(pods) != 1 || pods[0].manager == nil || pods[0].manager.record.Key != "z-owner" {
		t.Fatalf("pods = %#v; want alpha headed by z-owner from owns", pods)
	}
	var children []string
	for _, node := range pods[0].children {
		children = append(children, node.record.Key)
	}
	if !slices.Equal(children, []string{"a-planner"}) {
		t.Fatalf("alpha children = %q; want a-planner as a member", children)
	}
}

// F1 of D11: with no owns fact, a planner does not become an owner.
func TestAgentsNoOwnerIsInventedFromProfile(t *testing.T) {
	planner := d11Agent("a-planner", records.ProfilePlanner, "alpha", "", "", "2026-10-21T14:01:00Z")
	m := agentsModel(t, agentsSource{
		projects: []records.Project{{ID: "alpha", Path: "/repo/alpha", Base: "dev"}},
		agents:   []records.Agent{planner},
	})
	pods := m.agentsPods()
	if len(pods) != 1 || pods[0].manager != nil {
		t.Fatalf("pods = %#v; want no invented alpha owner", pods)
	}
}

// F2 of D11: a hidden parent does not change the project fact of its child.
// The hidden count stays 1, and h keeps the selection on the child in alpha.
func TestAgentsHiddenParentKeepsTheChildInItsProject(t *testing.T) {
	owner := d11Agent("owner", records.ProfilePlanner, "alpha", "alpha", "", "2026-10-21T14:00:00Z")
	parent := d11Agent("parent", records.ProfileBuilder, "alpha", "", "owner", "2026-10-21T14:01:00Z")
	parent.Task = "task-parent"
	parent.Status = records.AgentStopped
	parent.SyncRole()
	child := d11Agent("child", records.ProfileGeneral, "alpha", "", "parent", "2026-10-21T14:02:00Z")
	m := agentsModel(t, agentsSource{
		projects: []records.Project{{ID: "alpha", Path: "/repo/alpha", Base: "dev"}},
		agents:   []records.Agent{owner, parent, child},
	})
	m.graph.hideStopped = true
	if pod, _ := podOf(t, m, "child"); pod != "alpha" {
		t.Fatalf("hidden-parent child is in pod %q; want alpha", pod)
	}
	if got := m.hiddenCards(); got != 1 {
		t.Fatalf("hidden cards = %d; want 1", got)
	}
	m.graph.focus, m.graph.selected = "child", "child"
	m, _ = raise(t, m, key('h', "h"))
	if m.graph.focus != "child" || m.graph.selected != "child" {
		t.Fatalf("selection = %q/%q after h; want child/child", m.graph.focus, m.graph.selected)
	}
	if pod, _ := podOf(t, m, "child"); pod != "alpha" {
		t.Fatalf("shown-parent child is in pod %q; want alpha", pod)
	}
}

// F3 of D11: a shared free profile label does not join distinct agents. Two
// custom-profile siblings with different names and no replaces relation stay
// distinct and current in their project.
func TestAgentsCustomProfileSiblingsStayDistinct(t *testing.T) {
	owner := d11Agent("owner", records.ProfilePlanner, "alpha", "alpha", "", "2026-10-21T14:00:00Z")
	first := d11Agent("review-a", "review-specialist", "alpha", "", "owner", "2026-10-21T14:01:00Z")
	second := d11Agent("review-b", "review-specialist", "alpha", "", "owner", "2026-10-21T14:02:00Z")
	agents := []records.Agent{owner, first, second}
	current := currentAgentKeys(agents)
	for _, key := range []string{"review-a", "review-b"} {
		if !current[key] {
			t.Errorf("current[%s] = false; want each distinct sibling current", key)
		}
	}
	m := agentsModel(t, agentsSource{
		projects: []records.Project{{ID: "alpha", Path: "/repo/alpha", Base: "dev"}},
		agents:   agents,
	})
	var cards []string
	for _, pod := range m.agentsPods() {
		if pod.group.id != "alpha" {
			continue
		}
		for _, node := range pod.children {
			cards = append(cards, node.record.Key)
		}
	}
	if !slices.Equal(cards, []string{"review-a", "review-b"}) {
		t.Fatalf("alpha children = %q; want both custom-profile siblings", cards)
	}
}
