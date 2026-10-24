package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/charmbracelet/x/ansi"
)

// AC2, golden-frames README decisions 8 to 13 and ask a27: the Health status
// row takes the form of the Health boards only while the first configured
// row is the default row. A custom row stays as configured.
func TestGfHealthCustomStatusLineStaysConfigured(t *testing.T) {
	m, _ := newFrameModel(t, 120, 40)
	m.selectScreen(tabHealth, 0)
	if got := ansi.Strip(m.statusRows(120, 40)[0]); !strings.Contains(got, "· health · ↑↓ ←→ tile") {
		t.Fatalf("default Health status row = %q; want the board form", got)
	}

	cfg := defaultTUI()
	cfg.StatusLines = [][]string{{"command", "clock"}}
	cfg.StatusLinesExplicit = true
	m.setTUIConfig(cfg)
	m.statusCommandLine = "custom-health-status"
	m.selectScreen(tabHealth, 0)
	got := ansi.Strip(m.statusRows(120, 40)[0])
	if !strings.Contains(got, "custom-health-status") {
		t.Fatalf("Health discarded the configured status row: %q", got)
	}
	if strings.Contains(got, "health · ↑↓") {
		t.Fatalf("Health replaced the configured status row: %q", got)
	}
}

// AC2, golden-frames decisions 6 and 20 and tui-v2 decisions 80 and 81: the
// wheel over the Health agents table moves its selection and keeps the
// focus, a click gives the table the focus and selects the row, and a click
// on the grid gives the grid the focus.
func TestGfHealthMouseOnTheTable(t *testing.T) {
	// The GfHealth table starts at screen row 18. Row 20 is the coordinator,
	// row 21 the supervisor, and row 22 manager-hive.
	t.Run("wheel", func(t *testing.T) {
		m := gfHealthModel(t, "GfHealth", nil)
		if m.health.selectedAgent != "supervisor-g4" || m.regionFocus[tabHealth] != 0 {
			t.Fatalf("start selection %q, focus %d; want supervisor-g4, grid", m.health.selectedAgent, m.regionFocus[tabHealth])
		}
		m, _ = raise(t, m, wheel(20, 22, tea.MouseWheelDown))
		if m.health.selectedAgent != "manager-hive-g2" {
			t.Fatalf("wheel down selected %q; want manager-hive-g2", m.health.selectedAgent)
		}
		if m.regionFocus[tabHealth] != 0 {
			t.Fatalf("wheel changed the focus to %d; want the grid", m.regionFocus[tabHealth])
		}
		m, _ = raise(t, m, wheel(20, 21, tea.MouseWheelUp))
		if m.health.selectedAgent != "supervisor-g4" {
			t.Fatalf("wheel up selected %q; want supervisor-g4", m.health.selectedAgent)
		}
	})
	t.Run("click row", func(t *testing.T) {
		m := gfHealthModel(t, "GfHealth", nil)
		m, _ = raise(t, m, click(20, 22))
		if m.health.selectedAgent != "manager-hive-g2" {
			t.Fatalf("click selected %q; want manager-hive-g2", m.health.selectedAgent)
		}
		if m.regionFocus[tabHealth] != 1 {
			t.Fatalf("click gave focus %d; want the table", m.regionFocus[tabHealth])
		}
	})
	t.Run("click below the last row selects the last row", func(t *testing.T) {
		m := gfHealthModel(t, "GfHealth", nil)
		m, _ = raise(t, m, click(20, 28))
		if m.health.selectedAgent != "plan-review-1" {
			t.Fatalf("click below the table selected %q; want plan-review-1", m.health.selectedAgent)
		}
	})
	t.Run("click grid", func(t *testing.T) {
		m := gfHealthModel(t, "GfHealth", nil)
		m.regionFocus[tabHealth] = 1
		m, _ = raise(t, m, click(20, 4))
		if m.regionFocus[tabHealth] != 0 {
			t.Fatalf("click on the grid gave focus %d; want the grid", m.regionFocus[tabHealth])
		}
	})
	t.Run("wheel over the grid does nothing", func(t *testing.T) {
		m := gfHealthModel(t, "GfHealth", nil)
		m, _ = raise(t, m, wheel(20, 4, tea.MouseWheelDown))
		if m.health.selectedAgent != "supervisor-g4" {
			t.Fatalf("wheel over the grid selected %q; want supervisor-g4", m.health.selectedAgent)
		}
	})
}

// AC2, golden-frames decision 6: the focused Health table takes pgup and
// pgdn, and each moves the selection by one page: the rows that the table
// box shows.
func TestGfHealthTablePageKeys(t *testing.T) {
	t.Run("board", func(t *testing.T) {
		m := gfHealthModel(t, "GfHealth", nil)
		m, _ = raise(t, m, ctrl(']'))
		if m.focusState() != "table" {
			t.Fatalf("ctrl+] gave focus %q; want table", m.focusState())
		}
		m, _ = raise(t, m, key(tea.KeyPgDown, ""))
		if m.health.selectedAgent != "plan-review-1" {
			t.Fatalf("pgdown selected %q; want plan-review-1", m.health.selectedAgent)
		}
		m, _ = raise(t, m, key(tea.KeyPgUp, ""))
		if m.health.selectedAgent != "coordinator-g2" {
			t.Fatalf("pgup selected %q; want coordinator-g2", m.health.selectedAgent)
		}
	})
	t.Run("many agents", func(t *testing.T) {
		m := gfHealthModel(t, "GfHealth", func(f *demoFixture) {
			f.Data.Agents = []records.Agent{{Key: "coordinator", Name: "coordinator", Role: "coordinator", Status: records.AgentLive, PaneID: "w1:p1"}}
			for index := range 60 {
				name := fmt.Sprintf("w%02d", index)
				f.Data.Agents = append(f.Data.Agents, records.Agent{Key: name, Name: name, Role: "worker", Status: records.AgentLive, PaneID: "w9:p1"})
			}
			f.UI.Selection = "w00"
		})
		m, _ = raise(t, m, ctrl(']'))
		visible := healthTableVisible(t, m)
		if visible < 2 {
			t.Fatalf("the table shows %d rows; want more than 1", visible)
		}
		m, _ = raise(t, m, key(tea.KeyPgDown, ""))
		if want := fmt.Sprintf("w%02d", visible); m.health.selectedAgent != want {
			t.Fatalf("pgdown selected %q; want %q, one page of %d rows", m.health.selectedAgent, want, visible)
		}
		m, _ = raise(t, m, key(tea.KeyPgUp, ""))
		if m.health.selectedAgent != "w00" {
			t.Fatalf("pgup selected %q; want w00", m.health.selectedAgent)
		}
	})
}

// healthTableVisible counts the agent rows that the agents table shows.
func healthTableVisible(t *testing.T, m *Model) int {
	t.Helper()
	rows := viewRows(m)
	start := -1
	for index, row := range rows {
		if strings.HasPrefix(row, "╭─ agents") {
			start = index
			break
		}
	}
	if start < 0 {
		t.Fatalf("no agents table:\n%s", strings.Join(rows, "\n"))
	}
	visible := 0
	for _, row := range rows[start+1:] {
		if strings.HasPrefix(row, "╰") {
			break
		}
		visible++
	}
	return visible - 1 // The header row is not an agent row.
}
