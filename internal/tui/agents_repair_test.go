package tui

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
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
			loaded, err := loadDemoFixture(gfDesignDir + "/GfAgents.fixture.json")
			if err != nil {
				t.Fatal(err)
			}
			loaded.Data.Supervisor.Live = true
			loaded.UI.HideStopped = false
			for index := range loaded.Data.Agents {
				agent := &loaded.Data.Agents[index]
				if agent.Key == "t6-worker-1" {
					agent.Status = tc.status
					agent.AgentStatus = tc.agentStatus
				}
			}
			m := newDemoModel(t.TempDir(), &demoState{fixture: loaded})
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
