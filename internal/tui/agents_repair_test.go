package tui

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/tui/component"
	"github.com/charmbracelet/x/ansi"
)

// agents.md rule 12: ↓ on the coordinator selects the last card of the pod
// that had the selection, not its first card.
func TestGfAgentsCoordinatorDownSelectsLastCard(t *testing.T) {
	m := newGfDemoModel(t, "GfAgents")
	for range 3 {
		m, _ = raise(t, m, key(tea.KeyUp, ""))
	}
	if node, _, _ := m.selectedAgent(); node.record.Key != "coordinator-g8" {
		t.Fatalf("up traversal selected %q; want the coordinator", node.record.Key)
	}
	m, _ = raise(t, m, key(tea.KeyDown, ""))
	if node, _, _ := m.selectedAgent(); node.record.Key != "t6-worker-1" {
		t.Fatalf("down from the coordinator selected %q; want the last card t6-worker-1 of hive", node.record.Key)
	}
}

// The Agents key table gives h and ? to any focus, including the details
// popup.
func TestGfAgentsDetailsAnyKeys(t *testing.T) {
	t.Run("help", func(t *testing.T) {
		m := newGfDemoModel(t, "GfAgentsDetails")
		m, _ = raise(t, m, key('?', "?"))
		if !m.helpOpen {
			t.Fatal("? in the details popup did not open Help")
		}
	})
	t.Run("stopped", func(t *testing.T) {
		m := newGfDemoModel(t, "GfAgentsDetails")
		m, _ = raise(t, m, key('h', "h"))
		if m.graph.hideStopped {
			t.Fatal("h in the details popup did not show the stopped agents")
		}
	})
}

// AC2: each handled details scroll key logs a user action.
func TestGfAgentsDetailsScrollLogs(t *testing.T) {
	m := newGfDemoModel(t, "GfAgentsDetails")
	m.invalidateView()
	m.View()
	var logs bytes.Buffer
	m.logger = slog.New(slog.NewJSONHandler(&logs, nil))
	m, cmd := raise(t, m, key(tea.KeyDown, ""))
	m = deliverAgentCommands(t, m, cmd)
	if m.popupOffset != 1 {
		t.Fatalf("details down moved the offset to %d; want 1", m.popupOffset)
	}
	if !strings.Contains(logs.String(), `"msg":"tui action"`) {
		t.Fatalf("details scroll moved but logged no tui action: %s", logs.String())
	}
}

// gfAgentsStateFixture gives the small synthetic fixture of one Agents state
// case. It has one project, one live supervisor, and two worker cards. The
// second worker is selected, so a broken selection does not give it by
// default. The test reads no design file, so it runs in a checkout without
// the local spec.
func gfAgentsStateFixture(status, agentStatus string) *demoFixture {
	const (
		nowText     = "2026-10-23T14:22:00Z"
		stampText   = "2026-10-23T14:21:00Z"
		earlierText = "2026-10-23T14:19:00Z"
	)
	return &demoFixture{
		Board:    "GfAgents.dc.html",
		Terminal: demoSize{Width: 120, Height: 40},
		Clock:    demoClock{Time: nowText, Zone: "UTC"},
		Home:     "~/source/hive",
		Data: demoData{
			Supervisor: demoSupervisor{Live: true, Agent: "supervisor-g4", PID: 1620686},
			Projects: []records.Project{
				{ID: "hive", Path: "~/source/hive", Base: "dev", RegisteredAt: stampText},
			},
			Agents: []records.Agent{
				{
					Key: "supervisor-g4", Name: "supervisor", Kind: records.AgentKindSupervisor, Harness: "hive",
					Status: records.AgentLive, AgentStatus: "working", StartedAt: stampText, UpdatedAt: stampText,
				},
				{
					Key: "t5-worker-1", Name: "t5-worker-1", Profile: records.ProfileBuilder, Project: "hive",
					Task: "5-tui-v2-spec", Harness: "pi", Status: records.AgentLive, AgentStatus: "working",
					Parent: "manager-hive-g2", StatusChangedAt: earlierText, Generation: 1,
					StartedAt: earlierText, UpdatedAt: earlierText, Spawned: true,
				},
				{
					Key: "t6-worker-1", Name: "t6-worker-1", Profile: records.ProfileBuilder, Project: "hive",
					Task: "6-tui-v2-boards", Harness: "pi", Status: status, AgentStatus: agentStatus,
					Parent: "manager-hive-g2", StatusChangedAt: stampText, Generation: 1,
					StartedAt: stampText, UpdatedAt: stampText, Spawned: true,
				},
			},
		},
		UI: demoUIState{
			Screen: "agents", Focus: "graph", Selection: "t6-worker-1",
			HideStopped: false, Follow: true,
		},
		now: time.Date(2026, 10, 23, 14, 22, 0, 0, time.UTC),
	}
}

// agents.md state table: starting, blocked, unresolved, and stale keep their
// stored word. A blocked card uses the triangle of tui-v2 decision 55, and
// an idle agent keeps the Agents ◐ override. A live supervisor is required
// for each state.
func TestGfAgentsCardStates(t *testing.T) {
	for _, tc := range []struct {
		name        string
		status      string
		agentStatus string
		word        string
		glyph       string
	}{
		{"starting lifecycle", "starting", "", "starting", "○"},
		{"starting harness", "live", "starting", "starting", "○"},
		{"blocked harness", "live", "blocked", "blocked", "▲"},
		{"unresolved lifecycle", "unresolved", "", "unresolved", "○"},
		{"stale lifecycle", "stale", "working", "stale", "○"},
		{"idle override", "live", "idle", "idle", "◐"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := newDemoModel(t.TempDir(), &demoState{fixture: gfAgentsStateFixture(tc.status, tc.agentStatus)})
			t.Cleanup(m.Stop)
			if !m.agents.data.supervisorLive {
				t.Fatal("precondition: the supervisor is not fresh and live")
			}
			node, _, ok := m.selectedAgent()
			if !ok || node.record.Key != "t6-worker-1" {
				t.Fatalf("selected agent = %q (%t); want t6-worker-1", node.record.Key, ok)
			}
			card := ansi.Strip(component.Card{Name: "probe", State: node.status, Sub: agentCardSub(node)}.View())
			if node.status != tc.word {
				t.Errorf("derived state = %q; want %q", node.status, tc.word)
			}
			if !strings.Contains(card, "│ "+tc.glyph+" probe") {
				t.Errorf("card lacks the glyph %q:\n%s", tc.glyph, card)
			}
		})
	}
}
