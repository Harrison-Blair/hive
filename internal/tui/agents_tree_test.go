package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/Harrison-Blair/hive/internal/lib/cellgrid"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/tui/component"
	"github.com/charmbracelet/x/ansi"
)

// treeBoards are the boards of the one-pod tree (decisions 74 to 78).
var treeBoards = []string{"TreeWide", "TreeNested", "CompactTreeWide", "CompactTreeNested"}

// Decision 74: the demo view of each tree board matches its golden grid
// cell for cell.
func TestTreeBoardsMatchTheirGrids(t *testing.T) {
	for _, board := range treeBoards {
		t.Run(board, func(t *testing.T) { checkBoardGolden(t, board) })
	}
}

// addAgent gives a fixture change that adds a live general agent of pi
// under parent, started at started, with agent status state.
func addAgent(key, parent, state, started string) func(*demoFixture) {
	return func(f *demoFixture) {
		f.Data.Agents = append(f.Data.Agents, records.Agent{
			Key: key, Name: key, Profile: records.ProfileGeneral, Harness: "pi", PaneID: "w90:p1", Status: "live",
			AgentStatus: state, Parent: parent, StatusChangedAt: "2026-10-21T14:21:00Z",
			Generation: 1, StartedAt: started, UpdatedAt: "2026-10-21T14:21:00Z",
		})
	}
}

// selectAgent gives a fixture change that selects key.
func selectAgent(key string) func(*demoFixture) {
	return func(f *demoFixture) { f.UI.Selection = key }
}

// both gives a fixture change that applies each change in order.
func both(changes ...func(*demoFixture)) func(*demoFixture) {
	return func(f *demoFixture) {
		for _, change := range changes {
			change(f)
		}
	}
}

// Decision 74: a nested child is one more tree level under its parent: the
// parent joint ┬, one │ row, and the child card with ┷ when it is selected.
// The rail is not used.
func TestTreeNestedChild(t *testing.T) {
	m := liveModel(t, "TreeNested", nil)
	requireCells(t, m,
		liveCell{45, 22, '┬', cellgrid.Working}, liveCell{45, 23, '│', cellgrid.Border},
		liveCell{45, 24, '┷', cellgrid.Accent}, liveCell{33, 24, '┏', cellgrid.Accent},
		liveCell{33, 19, '╭', cellgrid.Working}, liveCell{63, 19, '╭', cellgrid.Working},
		liveCell{28, 30, '╰', cellgrid.Border}, liveCell{91, 30, '╯', cellgrid.Border})
	requireView(t, m, "│ h shows 2 stopped")
	rejectView(t, m, "├┤", "╰┤")
}

// Decision 74: several children of one parent use a bus, as at level 1.
// The tree is 3 slots of 26 cells wide, so the pod grows to 82 cells.
func TestTreeNestedChildrenUseABus(t *testing.T) {
	m := liveModel(t, "TreeNested", both(addAgent("t5-research-2", "t5-worker-1", "idle", "2026-10-21T14:16:00Z"), selectAgent("t5-research-2")))
	requireCells(t, m,
		// The pod and the coordinator card above the manager card.
		liveCell{19, 10, '╭', cellgrid.Border}, liveCell{100, 10, '╮', cellgrid.Border},
		liveCell{48, 5, '╭', cellgrid.Working}, liveCell{60, 8, '┬', cellgrid.Working},
		liveCell{60, 10, '┼', cellgrid.Border}, liveCell{48, 12, '╭', cellgrid.Working},
		// The bus of level 1.
		liveCell{60, 15, '┬', cellgrid.Working}, liveCell{47, 17, '╭', cellgrid.Border},
		liveCell{60, 17, '┴', cellgrid.Border}, liveCell{86, 17, '╮', cellgrid.Border},
		liveCell{35, 19, '╭', cellgrid.Working}, liveCell{47, 19, '┴', cellgrid.Working},
		liveCell{74, 19, '╭', cellgrid.Working}, liveCell{86, 19, '┴', cellgrid.Working},
		// The bus of level 2 under t5-worker-1.
		liveCell{47, 22, '┬', cellgrid.Working}, liveCell{47, 23, '│', cellgrid.Border},
		liveCell{34, 24, '╭', cellgrid.Border}, liveCell{47, 24, '┴', cellgrid.Border},
		liveCell{60, 24, '╮', cellgrid.Border}, liveCell{34, 25, '│', cellgrid.Border},
		liveCell{60, 25, '│', cellgrid.Border}, liveCell{22, 26, '╭', cellgrid.Working},
		liveCell{34, 26, '┴', cellgrid.Working}, liveCell{48, 26, '┏', cellgrid.Accent},
		liveCell{60, 26, '┷', cellgrid.Accent})
	requireView(t, m, "t5-research-2")
}

// treeScrollLine gives the text of the scroll line of the view of m, without
// the padding. It is the last body row above the status row. From 120
// columns the legend and one blank row are below it (golden-frames
// agents.md rule 7).
func treeScrollLine(m *Model) string {
	lines := strings.Split(ansi.Strip(m.View().Content), "\n")
	row := len(lines) - 2
	if m.width >= component.LongFormWidth {
		row = len(lines) - 4
	}
	return strings.TrimSpace(lines[row])
}

// scrollLine gives the scroll line of a body of width cells: left, and
// right at the right edge.
func scrollLine(width int, left, right string) string {
	return left + strings.Repeat(" ", width-ansi.StringWidth(left)-ansi.StringWidth(right)) + right
}

// requireSelected fails when the selection is not want.
func requireSelected(t *testing.T, m *Model, want string) {
	t.Helper()
	if m.graph.selected != want {
		t.Fatalf("selected = %q; want %q", m.graph.selected, want)
	}
}

// Decisions 75 and 77: at 80 columns the view scrolls sideways to keep the
// selected card at the centre when possible. The scroll line names the
// children on the screen and the next hidden child. Golden-frames agents.md
// rules 1 and 8: the body has no scroll thumb, and the scroll line has the
// full width less one cell of margin on each side.
func TestTreeScrollsSidewaysAt80(t *testing.T) {
	m := liveModel(t, "CompactTreeWide", nil)
	steps := []struct {
		press    rune
		selected string
		cardX    int
		line     string
	}{
		{0, "t6-worker-1", 28, scrollLine(78, "←→ at the edge scrolls · children 2–4 of 5", "t7-review-1 →")},
		{tea.KeyLeft, "t5-verifier-1", 28, scrollLine(78, "←→ at the edge scrolls · children 1–3 of 5", "t6-verifier-1 →")},
		{tea.KeyLeft, "t5-worker-1", 4, scrollLine(78, "←→ at the edge scrolls · children 1–3 of 5", "t6-verifier-1 →")},
		{tea.KeyLeft, "t5-worker-1", 4, scrollLine(78, "←→ at the edge scrolls · children 1–3 of 5", "t6-verifier-1 →")},
		{tea.KeyRight, "t5-verifier-1", 28, scrollLine(78, "←→ at the edge scrolls · children 1–3 of 5", "t6-verifier-1 →")},
		{tea.KeyRight, "t6-worker-1", 28, scrollLine(78, "←→ at the edge scrolls · children 2–4 of 5", "t7-review-1 →")},
		{tea.KeyRight, "t6-verifier-1", 28, scrollLine(78, "←→ at the edge scrolls · children 3–5 of 5", "← t5-verifier-1")},
		{tea.KeyRight, "t7-review-1", 53, scrollLine(78, "←→ at the edge scrolls · children 3–5 of 5", "← t5-verifier-1")},
		{tea.KeyRight, "t7-review-1", 53, scrollLine(78, "←→ at the edge scrolls · children 3–5 of 5", "← t5-verifier-1")},
	}
	for _, step := range steps {
		if step.press != 0 {
			m, _ = raise(t, m, key(step.press, ""))
		}
		requireSelected(t, m, step.selected)
		requireCells(t, m, liveCell{step.cardX, 17, '┏', cellgrid.Accent}, liveCell{step.cardX + 12, 17, '┷', cellgrid.Accent})
		if got := treeScrollLine(m); got != step.line {
			t.Fatalf("after %s: scroll line = %q; want %q", step.selected, got, step.line)
		}
	}
}

// Decision 75: from 120 columns the view scrolls sideways only as far as
// it must to keep the selected card on the screen. Golden-frames agents.md
// rule 8: the body has no scroll thumb.
func TestTreeScrollsSidewaysAt120(t *testing.T) {
	m := liveModel(t, "TreeWide", nil)
	steps := []struct {
		press    rune
		selected string
		cardX    int
		line     string
	}{
		{0, "t5-worker-1", 4, "children 1–4 of 5"},
		{tea.KeyRight, "t5-verifier-1", 30, "children 1–4 of 5"},
		{tea.KeyRight, "t6-worker-1", 56, "children 1–4 of 5"},
		{tea.KeyRight, "t6-verifier-1", 82, "children 1–4 of 5"},
		{tea.KeyRight, "t7-review-1", 96, "children 2–5 of 5"},
		{tea.KeyLeft, "t6-verifier-1", 70, "children 2–5 of 5"},
		{tea.KeyLeft, "t6-worker-1", 44, "children 2–5 of 5"},
		{tea.KeyLeft, "t5-verifier-1", 18, "children 2–5 of 5"},
		{tea.KeyLeft, "t5-worker-1", 0, "children 1–4 of 5"},
	}
	for _, step := range steps {
		if step.press != 0 {
			m, _ = raise(t, m, key(step.press, ""))
		}
		requireSelected(t, m, step.selected)
		requireCells(t, m, liveCell{step.cardX, 19, '┏', cellgrid.Accent})
		if got := treeScrollLine(m); !strings.Contains(got, step.line) {
			t.Fatalf("after %s: scroll line = %q; want %q", step.selected, got, step.line)
		}
	}
	// Golden-frames agents.md rule 7: the legend is the last body row, below
	// the scroll line.
	lines := strings.Split(ansi.Strip(m.View().Content), "\n")
	if !strings.HasPrefix(lines[36], " ←→ at the edge scrolls") || !strings.HasPrefix(lines[37], " ── spawn") {
		t.Fatalf("rows 36 and 37 = %q, %q; want the scroll line, then the legend", lines[36], lines[37])
	}
}

// Decision 76: when the tree is higher than the region, the view scrolls
// vertically and keeps the selected card on the screen. Golden-frames
// agents.md rule 8: the body has no scroll thumb.
func TestTreeScrollsVertically(t *testing.T) {
	m := liveModel(t, "CompactTreeNested", both(addAgent("t5-research-2", "t5-research-1", "idle", "2026-10-21T14:16:00Z"), selectAgent("manager-hive-g2")))
	steps := []struct {
		press    rune
		selected string
		cardX    int
		cardY    int
	}{
		{0, "manager-hive-g2", 28, 10},
		{tea.KeyDown, "t5-worker-1", 9, 17},
		{tea.KeyDown, "t5-research-1", 9, 20},
		{tea.KeyDown, "t5-research-2", 9, 20},
		{tea.KeyDown, "t5-research-2", 9, 20},
		{tea.KeyUp, "t5-research-1", 9, 15},
		{tea.KeyUp, "t5-worker-1", 9, 10},
		{tea.KeyUp, "manager-hive-g2", 28, 3},
		{tea.KeyUp, "manager-hive-g2", 28, 3},
	}
	for _, step := range steps {
		if step.press != 0 {
			m, _ = raise(t, m, key(step.press, ""))
		}
		requireSelected(t, m, step.selected)
		requireCells(t, m, liveCell{step.cardX, step.cardY, '┏', cellgrid.Accent})
	}
}

// Decision 77: the view scrolls only to keep the selection on the screen.
// pgup and pgdn do not move it. CompactTreeWide and CompactTreeNested start
// scrolled. Golden-frames agents.md rule 8: the body has no thumb, so the
// check reads the scroll offsets of the model.
func TestTreePageKeysDoNotScroll(t *testing.T) {
	for _, tc := range []struct {
		board      string
		shift, top bool
	}{
		{"CompactTreeWide", true, false},
		{"CompactTreeNested", false, true},
		{"TreeWide", false, false},
	} {
		t.Run(tc.board, func(t *testing.T) {
			m := liveModel(t, tc.board, nil)
			m.View()
			if got := m.graph.shift > 0; got != tc.shift {
				t.Fatalf("scrolled sideways = %v; want %v", got, tc.shift)
			}
			if got := m.graph.top > 0; got != tc.top {
				t.Fatalf("scrolled down = %v; want %v", got, tc.top)
			}
			want := m.View().Content
			for _, press := range []rune{tea.KeyPgDown, tea.KeyPgDown, tea.KeyPgUp, tea.KeyPgUp, tea.KeyPgUp} {
				m, _ = raise(t, m, key(press, ""))
				if got := m.View().Content; got != want {
					t.Fatalf("view changed after a page key:\n%s\nwant:\n%s", ansi.Strip(got), ansi.Strip(want))
				}
			}
		})
	}
}

// Rule A of decision 78: in the one-pod view of home, the coordinator line
// goes down through the pod top border (┼) and the empty goal row to the
// top card, as it goes to the manager card of a project pod.
func TestHomeTreeCoordinatorLineGoesToTheTopCard(t *testing.T) {
	m := liveModel(t, "AgentsCoordinatorChild", nil)
	requireCells(t, m,
		liveCell{60, 8, '┬', cellgrid.Working}, liveCell{60, 9, '│', cellgrid.Border},
		liveCell{60, 10, '┼', cellgrid.Border}, liveCell{60, 11, '│', cellgrid.Border},
		liveCell{48, 12, '╭', cellgrid.Working}, liveCell{60, 12, '┴', cellgrid.Working})
	if got := m.graph.up["plan-review-1"]; got != "coordinator-g2" {
		t.Fatalf("the line above plan-review-1 comes from %q; want coordinator-g2", got)
	}
}

// Rule A of decision 78: with several top cards in home, the coordinator
// line uses a bus below the pod top border, as at a tree level: the │ goal
// row, the bus ╭─┴─╮, one more │ row, and the joint ┴ on each top card.
func TestHomeTreeCoordinatorLineUsesABus(t *testing.T) {
	m := liveModel(t, "AgentsCoordinatorChild", addAgent("plan-review-3", "coordinator-g2", "working", "2026-10-21T14:05:00Z"))
	grid := renderGrid(t, m, m.width, m.height)
	_, first := findText(t, grid, "plan-review-1")
	_, second := findText(t, grid, "plan-review-3")
	// The first cell of the name is 4 cells right of the card.
	a, b := first-4+12, second-4+12
	joint := (a + b) / 2
	requireCells(t, m,
		liveCell{joint, 8, '┬', cellgrid.Working}, liveCell{joint, 10, '┼', cellgrid.Border},
		liveCell{joint, 11, '│', cellgrid.Border},
		liveCell{a, 12, '╭', cellgrid.Border}, liveCell{joint, 12, '┴', cellgrid.Border},
		liveCell{b, 12, '╮', cellgrid.Border}, liveCell{a, 13, '│', cellgrid.Border},
		liveCell{b, 13, '│', cellgrid.Border}, liveCell{a, 14, '┴', cellgrid.Working},
		liveCell{b, 14, '┴', cellgrid.Working})
	for _, key := range []string{"plan-review-1", "plan-review-3"} {
		if got := m.graph.up[key]; got != "coordinator-g2" {
			t.Fatalf("the line above %s comes from %q; want coordinator-g2", key, got)
		}
	}
}
