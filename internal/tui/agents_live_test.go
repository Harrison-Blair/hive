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

// bodyShift is the row offset of a screen body from the tui-v2 boards to
// the golden frames. Golden-frames decision 2 replaces the header and the
// two tab bar rows with the title row and the rule, so the body starts one
// row higher. The status row takes the place of the footer.
const bodyShift = -1

// boxShift is the row offset of the Agents body from the outer box of the
// tui-v2 boards. Golden-frames agents.md rule 1 removes the outer box, so
// its top border row goes and the body starts one more row higher.
const boxShift = -1

// tuiV2Body sizes m so that its body has the rows of the tui-v2 frame at the
// board size: the golden frame uses one row less than the tui-v2 frame, so
// the pane gets one row less. It gives the row shift of the body. At the
// minimum height the pane keeps its size, so the body has one row more at
// its end, and a view that keeps the selection at its end keeps its rows.
func tuiV2Body(m *Model) (shift int, restore func()) {
	if m.height <= minimumHeight {
		return boxShift, func() {}
	}
	m.height--
	if m.demo != nil {
		m.demo.paneHeight = m.height
		m.demo.size.Height = m.height
	}
	m.invalidateView()
	return bodyShift + boxShift, func() { m.height++ }
}

// requireCells fails when a cell of the view of m is not as want. The rows
// of want are the rows of the tui-v2 board, which bodyShift and boxShift
// move.
func requireCells(t *testing.T, m *Model, want ...liveCell) {
	t.Helper()
	shift, restore := tuiV2Body(m)
	defer restore()
	grid := renderGrid(t, m, m.width, m.height)
	for _, cell := range want {
		got := grid.At(cell.y+shift, cell.x)
		if got.Rune != cell.glyph || got.Token != cell.token {
			t.Errorf("cell %d,%d = %q %s; want %q %s", cell.x, cell.y+shift, string(got.Rune), got.Token, string(cell.glyph), cell.token)
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

// The agent_id of an agent send event is the record key, as
// `manager-hive-g2`. It lights the line of the card of that record.
func TestLiveSendEventGivesTheReceiverByKey(t *testing.T) {
	m := liveModel(t, "LivePod", func(f *demoFixture) {
		f.Data.Logs = append(f.Data.Logs, promptEvent(t, f.now.Add(-500*time.Millisecond), "t6-worker-1", "manager-hive-g2"))
	})
	requireCells(t, m, liveCell{60, 16, '▲', cellgrid.Info}, liveCell{75, 18, '┃', cellgrid.Info}, liveCell{75, 19, '┸', cellgrid.Info})
}

// A record key and an agent name of one card name one line: the newest
// message of the line wins, in each order of the messages.
func TestLiveMessagesGroupByCard(t *testing.T) {
	for _, tc := range []struct {
		name             string
		older, newer     [2]string
		wantFrom, wantTo string
	}{
		{"name then key", [2]string{"manager-hive", "t6-worker-1"}, [2]string{"t6-worker-1", "manager-hive-g2"}, "t6-worker-1", "manager-hive-g2"},
		{"key then name", [2]string{"t6-worker-1", "manager-hive-g2"}, [2]string{"manager-hive", "t6-worker-1"}, "manager-hive-g2", "t6-worker-1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := liveModel(t, "LivePod", func(f *demoFixture) {
				f.Data.Logs = []json.RawMessage{
					promptEvent(t, f.now.Add(-2*time.Second), tc.older[0], tc.older[1]),
					promptEvent(t, f.now.Add(-time.Second), tc.newer[0], tc.newer[1]),
				}
			})
			m.View()
			got := m.litMessages()
			if len(got) != 1 || got[0].from != tc.wantFrom || got[0].to != tc.wantTo {
				t.Fatalf("lit messages = %+v; want one message %s to %s", got, tc.wantFrom, tc.wantTo)
			}
		})
	}
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
	dark := []liveCell{{60, 8, '┬', cellgrid.Working}, {60, 9, '┴', cellgrid.Border}, {23, 10, '┷', cellgrid.Accent}, {25, 9, '─', cellgrid.Border}}
	t.Run("no card", func(t *testing.T) {
		m := liveModel(t, "LiveOverview", inbox(record("user", "manager-hive"), record("supervisor", "manager-hive"), record("ghost-1", "manager-hive"), record("coordinator", "ghost-2")))
		requireCells(t, m, dark...)
	})
	t.Run("coordinator to manager", func(t *testing.T) {
		m := liveModel(t, "LiveOverview", inbox(record("coordinator", "manager-hive")))
		requireCells(t, m, liveCell{60, 8, '┰', cellgrid.Info}, liveCell{60, 9, '┹', cellgrid.Info}, liveCell{25, 9, '◀', cellgrid.Info},
			liveCell{24, 9, '━', cellgrid.Info}, liveCell{23, 9, '┏', cellgrid.Info}, liveCell{23, 10, '┻', cellgrid.Info}, liveCell{61, 9, '─', cellgrid.Border})
	})
	// The pods have no direct line: the path goes on the bus.
	t.Run("manager to manager", func(t *testing.T) {
		m := liveModel(t, "LiveOverview", inbox(record("manager-hive", "manager-api")))
		requireCells(t, m, liveCell{23, 10, '┻', cellgrid.Info}, liveCell{23, 9, '┏', cellgrid.Info}, liveCell{40, 9, '━', cellgrid.Info},
			liveCell{60, 9, '┷', cellgrid.Info}, liveCell{68, 9, '▶', cellgrid.Info}, liveCell{69, 9, '━', cellgrid.Info}, liveCell{70, 9, '┱', cellgrid.Info},
			liveCell{70, 10, '┸', cellgrid.Info}, liveCell{60, 8, '┬', cellgrid.Working}, liveCell{71, 9, '─', cellgrid.Border})
	})
	// The rail of the hive pod in the overview: heavy arms on the lit side.
	// The selected card is in the other pod, so that no card of the rail
	// has the heavy border of agents.md rule 12.
	selectOther := func(m *Model) *Model {
		m.graph.pod, m.graph.selected = "example-api", ""
		m.invalidateView()
		return m
	}
	t.Run("rail down", func(t *testing.T) {
		m := selectOther(liveModel(t, "LiveOverview", inbox(record("manager-hive", "t6-worker-1"))))
		requireCells(t, m, liveCell{3, 15, '┟', cellgrid.Info}, liveCell{3, 16, '┃', cellgrid.Info}, liveCell{3, 18, '┠', cellgrid.Info},
			liveCell{4, 18, '┤', cellgrid.Working}, liveCell{3, 21, '┃', cellgrid.Info}, liveCell{3, 22, '▼', cellgrid.Info},
			liveCell{3, 23, '┗', cellgrid.Info}, liveCell{4, 23, '┥', cellgrid.Info}, liveCell{60, 8, '┬', cellgrid.Working})
	})
	t.Run("rail up", func(t *testing.T) {
		m := selectOther(liveModel(t, "LiveOverview", inbox(record("t5-worker-1", "manager-hive"))))
		requireCells(t, m, liveCell{4, 18, '┥', cellgrid.Info}, liveCell{3, 18, '┡', cellgrid.Info}, liveCell{3, 17, '┃', cellgrid.Info},
			liveCell{3, 16, '▲', cellgrid.Info}, liveCell{3, 15, '┟', cellgrid.Info}, liveCell{3, 19, '│', cellgrid.Border})
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
	hive := []liveCell{{1, 10, '┏', cellgrid.Accent}, {41, 10, '╭', cellgrid.Border}, {48, 5, '╭', cellgrid.Working}}
	api := []liveCell{{1, 10, '╭', cellgrid.Border}, {41, 10, '┏', cellgrid.Accent}, {48, 5, '╭', cellgrid.Working}}
	coordinator := []liveCell{{1, 10, '╭', cellgrid.Border}, {41, 10, '╭', cellgrid.Border}, {48, 5, '┏', cellgrid.Accent}, {60, 8, '┯', cellgrid.Accent}}
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
	// Golden-frames decision 17: esc with no popup open returns to Chat.
	// The chord opens Agents again with the same selection.
	m, _ = raise(t, m, key(tea.KeyEscape, ""))
	if m.activeTab() != tabChat {
		t.Fatalf("esc on the overview gave screen %d; want Chat", m.activeTab())
	}
	m, _ = raise(t, m, ctrl('g'))
	requireCells(t, m, coordinator...)
	m, _ = raise(t, m, key(tea.KeyDown, ""))
	if m.graph.selected != "t3-verifier-1" {
		t.Fatalf("down from the coordinator selected %q; want the last card t3-verifier-1 of example-api", m.graph.selected)
	}
	requireCells(t, m, api...)
	m, _ = raise(t, m, key('o', "o"))
	requireView(t, m, "agents · example-api")
	if !m.graph.podView {
		t.Fatal("o on a pod did not open the one-pod view")
	}
	if m.popupOpen {
		t.Fatal("o on a pod opened the details popup")
	}
	m, _ = raise(t, m, key(tea.KeyEscape, ""))
	requireView(t, m, "agents · 2 projects")
	requireCells(t, m, api...)
	for range 3 {
		m, _ = raise(t, m, key(tea.KeyUp, ""))
	}
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
	requireView(t, m, "agents · hive")
}

// Decision 64: with one project and no home pod, the screen opens the one-pod
// view and the keys have no esc pods. Golden-frames decision 17: esc returns
// to Chat, because the view has no overview to go back to.
func TestLiveOneProjectOpensThePodView(t *testing.T) {
	m := newTestDemoModel(t, "Main.fixture.json", demoSize{})
	requireView(t, m, "agents · hive")
	if hints := screenRegistry[tabAgents].keyHints("", m.hintKey(tabAgents, m.focusState()), true); strings.Contains(strings.Join(hints, " "), "esc pods") {
		t.Fatalf("keys = %v; want no esc pods", hints)
	}
	m, _ = raise(t, m, key(tea.KeyEscape, ""))
	if m.activeTab() != tabChat {
		t.Fatalf("esc gave screen %d; want Chat", m.activeTab())
	}
}

// Decision 65: a pod drop 1 column or less from the coordinator joint moves
// to the column of the joint. The example-api pod center is column 59. With
// no goal, no pod title has the wave of golden-frames agents.md rule 2, so
// each drop is at its pod center.
func TestLiveDropMergesWithTheJoint(t *testing.T) {
	m := liveModel(t, "LiveOverview", func(f *demoFixture) { f.Data.Inbox, f.Data.Goals = nil, nil })
	requireCells(t, m, liveCell{59, 10, '─', cellgrid.Border}, liveCell{60, 10, '┴', cellgrid.Border}, liveCell{60, 9, '┼', cellgrid.Border},
		liveCell{99, 9, '╮', cellgrid.Border}, liveCell{99, 10, '┴', cellgrid.Border}, liveCell{19, 9, '╭', cellgrid.Border})
}

// Decisions 69 and 70: the legend and the footers at 120 and at 80 columns.
func TestLiveFooters(t *testing.T) {
	for _, tc := range []struct {
		board string
		keys  []tea.KeyPressMsg
		want  string
	}{
		{"LiveOverview", nil, "←→ pod · ↑↓ card · enter details · o one pod · a attach · p peek · t task · h stopped"},
		{"CompactLiveOverview", nil, "←→ pod · ↑↓ card · enter details · o one pod · ?"},
		{"LiveOverview", []tea.KeyPressMsg{key('o', "o")}, "←↑↓→ move · enter details · a attach · p peek · t task · esc pods · h stopped"},
		{"CompactLiveOverview", []tea.KeyPressMsg{key('o', "o")}, "←↑↓→ move · enter details · esc pods · ?"},
	} {
		t.Run(tc.board+"/"+tc.want, func(t *testing.T) {
			// Golden-frames decisions 3 and 11: the status row takes the
			// place of the footer. Its screen segment shows the screen
			// name and the keys of the focus.
			m := liveModel(t, tc.board, nil)
			cfg := defaultTUI()
			cfg.StatusLines = [][]string{{"screen"}}
			m.setTUIConfig(cfg)
			for _, press := range tc.keys {
				m, _ = raise(t, m, press)
			}
			lines := strings.Split(m.View().Content, "\n")
			footer := strings.TrimSpace(strings.TrimPrefix(ansi.Strip(lines[len(lines)-1]), " agents · "))
			if footer != tc.want {
				t.Fatalf("status row = %q; want %q", footer, tc.want)
			}
		})
	}
	requireView(t, liveModel(t, "LiveOverview", nil), "━ ▶ message, lit 5s")
}

// Decision 65 with a narrow pod: when the title is wider than the left half
// of the pod, the drop is 1 cell after the title. The 4 pods of AgentsMany
// are 28 cells wide. The wave of golden-frames agents.md rule 2 makes each
// title too wide: the title is cut at its end, and the drop stays inside the
// corner.
func TestLiveDropClearsALongTitle(t *testing.T) {
	m := liveModel(t, "AgentsMany", nil)
	requireView(t, m, "╭─ example-api · wave B… ─┴╮", "╭─ example-web · wave A… ─┴╮", "╭─ example-docs · wave … ─┴╮")
	requireCells(t, m, liveCell{23, 10, '┷', cellgrid.Accent}, liveCell{57, 10, '┴', cellgrid.Border},
		liveCell{57, 9, '┬', cellgrid.Border}, liveCell{117, 9, '╮', cellgrid.Border})
}
