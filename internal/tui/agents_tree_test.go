package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/Harrison-Blair/hive/internal/lib/cellgrid"
	"github.com/Harrison-Blair/hive/internal/lib/records"
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
			Key: key, Name: key, Role: "agent", Harness: "pi", PaneID: "w90:p1", Status: "live",
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
	requireView(t, m, "│ h shows 1 stopped")
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

// treeScrollLine gives the text of the scroll line of the view of m: the
// row above the bottom border, without the borders and the padding.
func treeScrollLine(m *Model) string {
	lines := strings.Split(ansi.Strip(m.View().Content), "\n")
	return strings.TrimSpace(strings.Trim(lines[len(lines)-3], "│┃"))
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
// children on the screen and the next hidden child. The bottom border has
// the ━ thumb of a track of 78 cells: 45 cells for 78 of 135 columns.
func TestTreeScrollsSidewaysAt80(t *testing.T) {
	m := liveModel(t, "CompactTreeWide", nil)
	steps := []struct {
		press     rune
		selected  string
		cardX     int
		line      string
		thumbFrom int
	}{
		{0, "t6-worker-1", 28, scrollLine(76, "←→ at the edge scrolls · children 2–4 of 5", "t7-review-1 →"), 17},
		{tea.KeyLeft, "t5-verifier-1", 28, scrollLine(76, "←→ at the edge scrolls · children 1–3 of 5", "t6-verifier-1 →"), 2},
		{tea.KeyLeft, "t5-worker-1", 5, scrollLine(76, "←→ at the edge scrolls · children 1–2 of 5", "t6-worker-1 →"), 1},
		{tea.KeyLeft, "t5-worker-1", 5, scrollLine(76, "←→ at the edge scrolls · children 1–2 of 5", "t6-worker-1 →"), 1},
		{tea.KeyRight, "t5-verifier-1", 28, scrollLine(76, "←→ at the edge scrolls · children 1–3 of 5", "t6-verifier-1 →"), 2},
		{tea.KeyRight, "t6-worker-1", 28, scrollLine(76, "←→ at the edge scrolls · children 2–4 of 5", "t7-review-1 →"), 17},
		{tea.KeyRight, "t6-verifier-1", 28, scrollLine(76, "←→ at the edge scrolls · children 3–5 of 5", "← t5-verifier-1"), 32},
		{tea.KeyRight, "t7-review-1", 52, scrollLine(76, "←→ at the edge scrolls · children 4–5 of 5", "← t6-worker-1"), 34},
		{tea.KeyRight, "t7-review-1", 52, scrollLine(76, "←→ at the edge scrolls · children 4–5 of 5", "← t6-worker-1"), 34},
	}
	for _, step := range steps {
		if step.press != 0 {
			m, _ = raise(t, m, key(step.press, ""))
		}
		requireSelected(t, m, step.selected)
		requireCells(t, m, liveCell{step.cardX, 17, '┏', cellgrid.Accent}, liveCell{step.cardX + 12, 17, '┷', cellgrid.Accent},
			liveCell{step.thumbFrom, 22, '━', cellgrid.Accent}, liveCell{step.thumbFrom + 44, 22, '━', cellgrid.Accent})
		if step.thumbFrom > 1 {
			requireCells(t, m, liveCell{step.thumbFrom - 1, 22, '─', cellgrid.Accent})
		}
		if step.thumbFrom+45 < 79 {
			requireCells(t, m, liveCell{step.thumbFrom + 45, 22, '─', cellgrid.Accent})
		}
		if got := treeScrollLine(m); got != step.line {
			t.Fatalf("after %s: scroll line = %q; want %q", step.selected, got, step.line)
		}
	}
}

// Decision 75: from 120 columns the view scrolls sideways only as far as
// it must to keep the selected card on the screen. The thumb is 103 of 118
// cells.
func TestTreeScrollsSidewaysAt120(t *testing.T) {
	m := liveModel(t, "TreeWide", nil)
	steps := []struct {
		press     rune
		selected  string
		cardX     int
		line      string
		thumbFrom int
	}{
		{0, "t5-worker-1", 5, "children 1–4 of 5", 1},
		{tea.KeyRight, "t5-verifier-1", 31, "children 1–4 of 5", 1},
		{tea.KeyRight, "t6-worker-1", 57, "children 1–4 of 5", 1},
		{tea.KeyRight, "t6-verifier-1", 83, "children 1–4 of 5", 1},
		{tea.KeyRight, "t7-review-1", 95, "children 2–5 of 5", 13},
		{tea.KeyLeft, "t6-verifier-1", 69, "children 2–5 of 5", 13},
		{tea.KeyLeft, "t6-worker-1", 43, "children 2–5 of 5", 13},
		{tea.KeyLeft, "t5-verifier-1", 17, "children 2–5 of 5", 13},
		{tea.KeyLeft, "t5-worker-1", 1, "children 1–4 of 5", 4},
	}
	for _, step := range steps {
		if step.press != 0 {
			m, _ = raise(t, m, key(step.press, ""))
		}
		requireSelected(t, m, step.selected)
		requireCells(t, m, liveCell{step.cardX, 19, '┏', cellgrid.Accent},
			liveCell{step.thumbFrom, 38, '━', cellgrid.Accent}, liveCell{step.thumbFrom + 102, 38, '━', cellgrid.Accent},
			liveCell{step.thumbFrom + 103, 38, '─', cellgrid.Accent})
		if got := treeScrollLine(m); !strings.Contains(got, step.line) {
			t.Fatalf("after %s: scroll line = %q; want %q", step.selected, got, step.line)
		}
	}
	// The legend is above the scroll line (decision 33).
	requireView(t, m, "one column per project"+strings.Repeat(" ", 29)+"│\n│ ←→ at the edge scrolls")
}

// Decision 76: when the tree is higher than the region, the view scrolls
// vertically and keeps the selected card on the screen. The right border
// has the ┃ thumb: 10 of 18 rows for 18 of 31 rows.
func TestTreeScrollsVertically(t *testing.T) {
	m := liveModel(t, "CompactTreeNested", both(addAgent("t5-research-2", "t5-research-1", "idle", "2026-10-21T14:16:00Z"), selectAgent("manager-hive-g2")))
	steps := []struct {
		press    rune
		selected string
		cardX    int
		cardY    int
		from     int
	}{
		{0, "manager-hive-g2", 28, 11, 4},
		{tea.KeyDown, "t5-worker-1", 10, 18, 4},
		{tea.KeyDown, "t5-research-1", 10, 18, 7},
		{tea.KeyDown, "t5-research-2", 10, 18, 10},
		{tea.KeyDown, "t5-research-2", 10, 18, 10},
		{tea.KeyUp, "t5-research-1", 10, 13, 10},
		{tea.KeyUp, "t5-worker-1", 10, 8, 10},
		{tea.KeyUp, "manager-hive-g2", 28, 4, 8},
		{tea.KeyUp, "manager-hive-g2", 28, 4, 8},
	}
	for _, step := range steps {
		if step.press != 0 {
			m, _ = raise(t, m, key(step.press, ""))
		}
		requireSelected(t, m, step.selected)
		requireCells(t, m, liveCell{step.cardX, step.cardY, '┏', cellgrid.Accent},
			liveCell{79, step.from, '┃', cellgrid.Accent}, liveCell{79, step.from + 9, '┃', cellgrid.Accent},
			liveCell{79, step.from + 10, '│', cellgrid.Accent})
		if step.from > 4 {
			requireCells(t, m, liveCell{79, step.from - 1, '│', cellgrid.Accent})
		}
	}
}

// Decision 77: the view scrolls only to keep the selection on the screen.
// pgup and pgdn do not move it.
func TestTreePageKeysDoNotScroll(t *testing.T) {
	for _, board := range []string{"CompactTreeWide", "CompactTreeNested", "TreeWide"} {
		t.Run(board, func(t *testing.T) {
			m := liveModel(t, board, nil)
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
