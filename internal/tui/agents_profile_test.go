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
