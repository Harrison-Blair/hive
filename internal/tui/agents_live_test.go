package tui

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/Harrison-Blair/hive/internal/lib/cellgrid"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/charmbracelet/x/ansi"
)

// liveBoards are the boards of the live Agents screen (decisions 62 to 72).
var liveBoards = []string{"LiveOverview", "LivePod", "CompactLiveOverview", "CompactLivePod"}

// AC10: the demo view of each live board matches its golden grid cell for
// cell.
func TestLiveBoardsMatchTheirGrids(t *testing.T) {
	for _, board := range liveBoards {
		t.Run(board, func(t *testing.T) { checkBoardGolden(t, board) })
	}
}

// liveCell is one expected cell of a rendered view: the column, the row,
// the glyph, and the token.
type liveCell struct {
	x, y  int
	glyph rune
	token cellgrid.Token
}

// requireCells fails when a cell of the view of m is not as want.
func requireCells(t *testing.T, m *Model, want ...liveCell) {
	t.Helper()
	grid := renderGrid(t, m, m.width, m.height)
	for _, cell := range want {
		got := grid.At(cell.y, cell.x)
		if got.Rune != cell.glyph || got.Token != cell.token {
			t.Errorf("cell %d,%d = %q %s; want %q %s", cell.x, cell.y, string(got.Rune), got.Token, string(cell.glyph), cell.token)
		}
	}
	if t.Failed() {
		t.Logf("view:\n%s", m.View().Content)
	}
}

// liveModel gives the demo model of board after change edits its fixture.
func liveModel(t *testing.T, board string, change func(*demoFixture)) *Model {
	t.Helper()
	return healthDemoModel(t, board, change)
}

// at sets the fixture clock to the clock plus offset.
func at(offset time.Duration) func(*demoFixture) {
	return func(f *demoFixture) { f.now = f.now.Add(offset) }
}

// promptEvent gives the agent prompt log event of hive agent send.
func promptEvent(t *testing.T, at time.Time, from, agentID string) json.RawMessage {
	t.Helper()
	event := map[string]any{"time": at.UTC().Format(time.RFC3339Nano), "level": "DEBUG", "msg": "agent prompt", "component": "agent", "command": "agent send", "agent_id": agentID}
	if from != "" {
		event["from"] = from
	}
	data, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// The LivePod cells of the line from manager-hive to t6-worker-1, lit and
// not lit.
var (
	podLineLit = []liveCell{
		{60, 15, '┰', cellgrid.Info}, {60, 16, '┃', cellgrid.Info}, {60, 17, '┺', cellgrid.Info},
		{70, 17, '━', cellgrid.Info}, {75, 17, '┓', cellgrid.Info}, {75, 18, '▼', cellgrid.Info}, {75, 19, '┸', cellgrid.Info},
		{46, 17, '─', cellgrid.Border},
	}
	podLineDark = []liveCell{
		{60, 15, '┬', cellgrid.Working}, {60, 16, '│', cellgrid.Border}, {60, 17, '┴', cellgrid.Border},
		{70, 17, '─', cellgrid.Border}, {75, 17, '╮', cellgrid.Border}, {75, 18, '│', cellgrid.Border}, {75, 19, '┴', cellgrid.Working},
	}
)

// Decision 67: a line stays lit for 5 seconds after the time of the message
// (14:21:59 in LivePod).
func TestLiveLineIsLitForFiveSeconds(t *testing.T) {
	for _, tc := range []struct {
		name   string
		offset time.Duration
		want   []liveCell
	}{
		{"at t", -time.Second, podLineLit},
		{"at t+4s", 3 * time.Second, podLineLit},
		{"at t+5s", 4 * time.Second, podLineDark},
	} {
		t.Run(tc.name, func(t *testing.T) {
			requireCells(t, liveModel(t, "LivePod", at(tc.offset)), tc.want...)
		})
	}
}

// Decision 67: a new message on the same line starts the 5 seconds again.
func TestLiveNewMessageKeepsTheLineLit(t *testing.T) {
	again := func(offset time.Duration) func(*demoFixture) {
		return func(f *demoFixture) {
			sent := f.now.Add(-time.Second)
			f.Data.Logs = append(f.Data.Logs, promptEvent(t, sent.Add(3*time.Second), "manager-hive", "t6-worker-1"))
			f.now = sent.Add(offset)
		}
	}
	requireCells(t, liveModel(t, "LivePod", again(7*time.Second)), podLineLit...)
	requireCells(t, liveModel(t, "LivePod", again(8*time.Second)), podLineDark...)
}

// Decision 66: a reply on the same line turns the arrow to the new receiver.
func TestLiveReplyTurnsTheArrow(t *testing.T) {
	m := liveModel(t, "LivePod", func(f *demoFixture) {
		f.Data.Logs = append(f.Data.Logs, promptEvent(t, f.now.Add(-500*time.Millisecond), "t6-worker-1", "manager-hive"))
	})
	requireCells(t, m, liveCell{60, 16, '▲', cellgrid.Info}, liveCell{75, 18, '┃', cellgrid.Info}, liveCell{75, 19, '┸', cellgrid.Info})
}

// The real TUI draws again when a line goes dark: the expiry message
// clears the cached view.
func TestLiveLineGoesDarkOnTheExpiryMessage(t *testing.T) {
	m := liveModel(t, "LivePod", nil)
	requireCells(t, m, podLineLit...)
	later := m.now().Add(5 * time.Second)
	m.now = func() time.Time { return later }
	m, _ = raise(t, m, agentsLitMsg{})
	requireCells(t, m, podLineDark...)
}

// Decision 66: an inbox record lights the line from its agent to its role.
// A sender or a receiver with no card lights nothing.
func TestLiveInboxRecordLightsALine(t *testing.T) {
	record := func(from, to string) records.Inbox {
		return records.Inbox{Seq: 1, Role: to, Kind: "message", Agent: from, Text: "hi", CreatedAt: "2026-10-21T14:21:58Z"}
	}
	inbox := func(items ...records.Inbox) func(*demoFixture) {
		return func(f *demoFixture) {
			f.Data.Inbox = map[string][]records.Inbox{}
			for _, item := range items {
				f.Data.Inbox[item.Role] = append(f.Data.Inbox[item.Role], item)
			}
		}
	}
	dark := []liveCell{{60, 8, '┬', cellgrid.Working}, {60, 9, '┼', cellgrid.Border}, {20, 10, '┷', cellgrid.Accent}, {22, 9, '─', cellgrid.Border}}
	t.Run("no card", func(t *testing.T) {
		m := liveModel(t, "LiveOverview", inbox(record("user", "manager-hive"), record("supervisor", "manager-hive"), record("ghost-1", "manager-hive"), record("coordinator", "ghost-2")))
		requireCells(t, m, dark...)
	})
	t.Run("coordinator to manager", func(t *testing.T) {
		m := liveModel(t, "LiveOverview", inbox(record("coordinator", "manager-hive")))
		requireCells(t, m, liveCell{60, 8, '┰', cellgrid.Info}, liveCell{60, 9, '╃', cellgrid.Info}, liveCell{22, 9, '◀', cellgrid.Info},
			liveCell{21, 9, '━', cellgrid.Info}, liveCell{20, 9, '┏', cellgrid.Info}, liveCell{20, 10, '┻', cellgrid.Info}, liveCell{61, 9, '─', cellgrid.Border})
	})
	// The pods have no direct line: the path goes on the bus.
	t.Run("manager to manager", func(t *testing.T) {
		m := liveModel(t, "LiveOverview", inbox(record("manager-hive", "manager-api")))
		requireCells(t, m, liveCell{20, 10, '┻', cellgrid.Info}, liveCell{20, 9, '┏', cellgrid.Info}, liveCell{40, 9, '━', cellgrid.Info},
			liveCell{58, 9, '▶', cellgrid.Info}, liveCell{59, 9, '━', cellgrid.Info}, liveCell{60, 9, '╅', cellgrid.Info},
			liveCell{60, 10, '┸', cellgrid.Info}, liveCell{60, 8, '┬', cellgrid.Working}, liveCell{61, 9, '─', cellgrid.Border})
	})
	// The rail of the hive pod in the overview: heavy arms on the lit side.
	t.Run("rail down", func(t *testing.T) {
		m := liveModel(t, "LiveOverview", inbox(record("manager-hive", "t6-worker-1")))
		requireCells(t, m, liveCell{4, 15, '┟', cellgrid.Info}, liveCell{4, 16, '┃', cellgrid.Info}, liveCell{4, 18, '┠', cellgrid.Info},
			liveCell{5, 18, '┤', cellgrid.Working}, liveCell{4, 21, '┃', cellgrid.Info}, liveCell{4, 22, '▼', cellgrid.Info},
			liveCell{4, 23, '┗', cellgrid.Info}, liveCell{5, 23, '┥', cellgrid.Info}, liveCell{60, 8, '┬', cellgrid.Working})
	})
	t.Run("rail up", func(t *testing.T) {
		m := liveModel(t, "LiveOverview", inbox(record("t5-worker-1", "manager-hive")))
		requireCells(t, m, liveCell{5, 18, '┥', cellgrid.Info}, liveCell{4, 18, '┡', cellgrid.Info}, liveCell{4, 17, '┃', cellgrid.Info},
			liveCell{4, 16, '▲', cellgrid.Info}, liveCell{4, 15, '┟', cellgrid.Info}, liveCell{4, 19, '│', cellgrid.Border})
	})
	// Two children of one manager in the tree: the path goes on the tree bus.
	t.Run("tree child to child", func(t *testing.T) {
		m := liveModel(t, "LivePod", func(f *demoFixture) {
			f.Data.Logs = nil
			inbox(record("t5-worker-1", "t6-worker-1"))(f)
		})
		requireCells(t, m, liveCell{45, 19, '┻', cellgrid.Info}, liveCell{45, 18, '┃', cellgrid.Info}, liveCell{45, 17, '┏', cellgrid.Info},
			liveCell{60, 17, '┷', cellgrid.Info}, liveCell{75, 17, '┓', cellgrid.Info}, liveCell{75, 18, '▼', cellgrid.Info},
			liveCell{75, 19, '┸', cellgrid.Info}, liveCell{60, 15, '┬', cellgrid.Working})
	})
}

// Decision 62: in the overview you select a full pod with ← and →, and the
// coordinator card with ↑. ↓ selects the last pod again. enter opens the
// one-pod view, and esc goes back with the same pod selected.
func TestLivePodSelection(t *testing.T) {
	hive := []liveCell{{2, 10, '┏', cellgrid.Accent}, {41, 10, '╭', cellgrid.Border}, {48, 5, '╭', cellgrid.Working}}
	api := []liveCell{{2, 10, '╭', cellgrid.Border}, {41, 10, '┏', cellgrid.Accent}, {48, 5, '╭', cellgrid.Working}}
	coordinator := []liveCell{{2, 10, '╭', cellgrid.Border}, {41, 10, '╭', cellgrid.Border}, {48, 5, '┏', cellgrid.Accent}, {60, 8, '┯', cellgrid.Accent}}
	m := liveModel(t, "LiveOverview", func(f *demoFixture) { f.Data.Inbox = nil })
	requireCells(t, m, hive...)
	m, _ = raise(t, m, key(tea.KeyRight, ""))
	requireCells(t, m, api...)
	m, _ = raise(t, m, key(tea.KeyLeft, ""))
	requireCells(t, m, hive...)
	m, _ = raise(t, m, key(tea.KeyLeft, ""))
	requireCells(t, m, hive...)
	m, _ = raise(t, m, key(tea.KeyRight, ""))
	m, _ = raise(t, m, key(tea.KeyUp, ""))
	requireCells(t, m, coordinator...)
	m, _ = raise(t, m, key(tea.KeyEscape, ""))
	requireCells(t, m, coordinator...)
	m, _ = raise(t, m, key(tea.KeyDown, ""))
	requireCells(t, m, api...)
	m, _ = raise(t, m, key(tea.KeyEnter, ""))
	requireView(t, m, "project agents · example-api", "esc pods")
	if m.popupOpen {
		t.Fatal("enter on a pod opened the details popup")
	}
	m, _ = raise(t, m, key(tea.KeyEscape, ""))
	requireView(t, m, "project agents · 2 projects")
	requireCells(t, m, api...)
	m, _ = raise(t, m, key(tea.KeyUp, ""))
	m, _ = raise(t, m, key(tea.KeyEnter, ""))
	if !m.popupOpen {
		t.Fatal("enter on the coordinator card did not open its details")
	}
	requireView(t, m, "coordinator-g8 · esc closes")
}

// Decision 62: in the one-pod view you select single agents. The arrows
// stay in the pod.
func TestLivePodViewSelectsCards(t *testing.T) {
	m := liveModel(t, "LivePod", nil)
	requireCells(t, m, liveCell{33, 19, '┏', cellgrid.Accent})
	m, _ = raise(t, m, key(tea.KeyRight, ""))
	if m.graph.selected != "t6-worker-1" {
		t.Fatalf("selected = %q; want t6-worker-1", m.graph.selected)
	}
	m, _ = raise(t, m, key(tea.KeyRight, ""))
	m, _ = raise(t, m, key(tea.KeyUp, ""))
	m, _ = raise(t, m, key(tea.KeyUp, ""))
	if m.graph.selected != "manager-hive-g2" {
		t.Fatalf("selected = %q; want manager-hive-g2: the coordinator card is not selectable in the one-pod view", m.graph.selected)
	}
	requireView(t, m, "project agents · hive")
}

// Decision 64: with one project and no home pod, the screen opens the one-pod
// view, esc does nothing, and the footer has no esc pods.
func TestLiveOneProjectOpensThePodView(t *testing.T) {
	m := newTestDemoModel(t, "Main.fixture.json", demoSize{})
	requireView(t, m, "project agents · hive")
	rejectView(t, m, "esc pods")
	m, _ = raise(t, m, key(tea.KeyEscape, ""))
	requireView(t, m, "project agents · hive")
}

// Decision 65: a pod drop 1 column or less from the coordinator joint moves
// to the column of the joint. The example-api pod center is column 59.
func TestLiveDropMergesWithTheJoint(t *testing.T) {
	m := liveModel(t, "LiveOverview", func(f *demoFixture) { f.Data.Inbox = nil })
	requireCells(t, m, liveCell{59, 10, '─', cellgrid.Border}, liveCell{60, 10, '┴', cellgrid.Border}, liveCell{60, 9, '┼', cellgrid.Border},
		liveCell{98, 9, '╮', cellgrid.Border}, liveCell{98, 10, '┴', cellgrid.Border}, liveCell{20, 9, '╭', cellgrid.Border})
}

// Decisions 69 and 70: the legend and the footers at 120 and at 80 columns.
func TestLiveFooters(t *testing.T) {
	for _, tc := range []struct {
		board string
		keys  []tea.KeyPressMsg
		want  string
	}{
		{"LiveOverview", nil, "←→ select pod · ↑ coordinator · enter open pod · h stopped · ? help"},
		{"CompactLiveOverview", nil, "←→ select pod · ↑ coordinator · enter open · ?"},
		{"LiveOverview", []tea.KeyPressMsg{key(tea.KeyEnter, "")}, "←↑↓→ move · enter details · a attach · p peek · t task · esc pods · h stopped · ? help"},
		{"CompactLiveOverview", []tea.KeyPressMsg{key(tea.KeyEnter, "")}, "←↑↓→ move · enter details · esc pods · ?"},
	} {
		t.Run(tc.board+"/"+tc.want, func(t *testing.T) {
			m := liveModel(t, tc.board, nil)
			for _, press := range tc.keys {
				m, _ = raise(t, m, press)
			}
			lines := strings.Split(m.View().Content, "\n")
			footer := strings.TrimSpace(strings.TrimPrefix(ansi.Strip(lines[len(lines)-1]), " [graph] "))
			if footer != tc.want {
				t.Fatalf("footer = %q; want %q", footer, tc.want)
			}
		})
	}
	requireView(t, liveModel(t, "LiveOverview", nil), "━ ▶ message, lit 5s")
}

// Decision 65 with a narrow pod: when the title is wider than the left half
// of the pod, the drop is 1 cell after the title, and the title stays whole.
// The 4 pods of AgentsMany are 28 cells wide.
func TestLiveDropClearsALongTitle(t *testing.T) {
	m := liveModel(t, "AgentsMany", nil)
	requireView(t, m, "╭─ example-api ─┴", "╭─ example-web ─┴", "╭─ example-docs ─┴")
	requireCells(t, m, liveCell{15, 10, '┷', cellgrid.Accent}, liveCell{47, 10, '┴', cellgrid.Border},
		liveCell{47, 9, '┬', cellgrid.Border}, liveCell{106, 9, '╮', cellgrid.Border})
}
