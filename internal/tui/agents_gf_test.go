package tui

import (
	"fmt"
	"os"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/charmbracelet/x/ansi"
)

// gfAgentsRows gives the plain rows of the View of the GfAgents fixture
// after keys.
func gfAgentsRows(t *testing.T, board string, keys ...tea.KeyPressMsg) (*Model, []string) {
	t.Helper()
	m := newGfDemoModel(t, board)
	for _, press := range keys {
		m, _ = raise(t, m, press)
	}
	return m, strings.Split(ansi.Strip(m.View().Content), "\n")
}

// agents.md rule 1: row 0 is the title row with the facts of the overview.
func TestGfAgentsTitleRowFacts(t *testing.T) {
	_, rows := gfAgentsRows(t, "GfAgents")
	want := " agents · 2 projects · 1 home · h shows 2 stopped"
	if got := strings.TrimRight(rows[0], " "); got != want {
		t.Fatalf("row 0 = %q, want %q", got, want)
	}
}

// agents.md rule 1: the outer box goes away. The body starts on row 2 and
// uses the full width with one cell of margin.
func TestGfAgentsHasNoOuterBox(t *testing.T) {
	_, rows := gfAgentsRows(t, "GfAgents")
	view := strings.Join(rows, "\n")
	if strings.Contains(view, "project agents") {
		t.Fatalf("the view keeps the outer box title:\n%s", view)
	}
	for index, row := range rows[2 : len(rows)-1] {
		if strings.HasPrefix(row, "│") || strings.HasPrefix(row, "╭") || strings.HasPrefix(row, "╰") {
			t.Fatalf("row %d starts with an outer box border: %q", index+2, row)
		}
	}
	if !strings.HasPrefix(rows[8], " ┏━ hive") {
		t.Fatalf("row 8 = %q, want the first pod on column 1", rows[8])
	}
	if !strings.HasSuffix(strings.TrimRight(rows[8], " "), "╮") || len([]rune(strings.TrimRight(rows[8], " "))) != 119 {
		t.Fatalf("row 8 = %q, want the last pod to end on column 118", rows[8])
	}
}

// agents.md rule 2: the pod title is PROJECT · wave W L/T. A pod with no
// goal shows PROJECT only.
func TestGfAgentsPodTitleShowsWaveProgress(t *testing.T) {
	_, rows := gfAgentsRows(t, "GfAgents")
	for _, want := range []string{"━ hive · wave B 2/4 ━", "─ example-api · wave B 1/3 ─", "─ home ─"} {
		if !strings.Contains(rows[8], want) {
			t.Fatalf("row 8 = %q, want %q", rows[8], want)
		}
	}
}

// agents.md rule 2: a pod with no goal shows PROJECT only, and a goal with
// no open wave shows no wave.
func TestGfAgentsPodTitleWithNoWave(t *testing.T) {
	_, rows := gfAgentsRows(t, "GfAgents")
	if strings.Contains(rows[8], "home ·") {
		t.Fatalf("row 8 = %q, the home pod has no goal", rows[8])
	}
}

// agents.md rule 3: the legend ends with the wave text.
func TestGfAgentsLegendEndsWithWaveText(t *testing.T) {
	_, rows := gfAgentsRows(t, "GfAgents")
	want := " ── spawn   ● working   ◐ idle   ○ stopped   ━ ▶ message, lit 5s   one column per project   wave B 2/4 = landed tasks"
	if got := strings.TrimRight(rows[37], " "); got != want {
		t.Fatalf("row 37 = %q, want %q", got, want)
	}
}

// agents.md rule 4: the screen segment shows the keys of the view.
func TestGfAgentsScreenKeys(t *testing.T) {
	m := newGfDemoModel(t, "GfAgents")
	want := "←→ pod · ↑↓ card · enter details · o one pod · a attach · p peek · t task · h stopped"
	if got := strings.Join(m.focusHints(m.width), " · "); got != want {
		t.Fatalf("overview keys = %q, want %q", got, want)
	}
	m.graph.podView = true
	want = "←↑↓→ move · enter details · a attach · p peek · t task · esc pods · h stopped"
	if got := strings.Join(m.focusHints(m.width), " · "); got != want {
		t.Fatalf("one-pod keys = %q, want %q", got, want)
	}
}

// agents.md rule 5: esc in the one-pod view returns to the overview, and esc
// in the overview returns to Chat.
func TestGfAgentsEscReturns(t *testing.T) {
	m := newGfDemoModel(t, "GfAgents")
	m.graph.podView = true
	m, _ = raise(t, m, key(tea.KeyEscape, ""))
	if m.activeTab() != tabAgents || m.graph.podView {
		t.Fatalf("esc in the one-pod view: tab %d, pod view %v; want the overview", m.activeTab(), m.graph.podView)
	}
	m, _ = raise(t, m, key(tea.KeyEscape, ""))
	if m.activeTab() != tabChat {
		t.Fatalf("esc in the overview: tab %d, want Chat", m.activeTab())
	}
}

// The one-pod title of GfAgentsPod: the pod, one pod, esc pods, and the h
// hint (case A12).
func TestGfAgentsPodTitleRow(t *testing.T) {
	_, rows := gfAgentsRows(t, "GfAgents", key('o', "o"))
	want := " agents · hive · one pod · esc pods · h shows 2 stopped"
	if got := strings.TrimRight(rows[0], " "); got != want {
		t.Fatalf("row 0 = %q, want %q", got, want)
	}
}

// agents.md rule 1: one project pod gives the singular noun.
func TestGfAgentsTitleRowOneProject(t *testing.T) {
	m := newGfDemoModel(t, "GfAgents")
	groups := m.agents.data.groups[:0]
	for _, group := range m.agents.data.groups {
		if group.id != "example-api" {
			groups = append(groups, group)
		}
	}
	m.agents.data.groups = groups
	if got := m.agentsTitleFacts(); !strings.HasPrefix(got, "· 1 project ·") {
		t.Fatalf("facts = %q, want the prefix %q", got, "· 1 project ·")
	}
}

// agents.md rule 7: the legend is the last row of the body, also when the
// view scrolls. The scroll line of rule 8 is the row above it.
func TestGfAgentsLegendIsLastBodyRowWhenScrolling(t *testing.T) {
	m := newGfDemoModel(t, "GfAgents")
	var source agentsSource
	for index := range 9 {
		projects, agents := agentsProject(fmt.Sprintf("p%d", index+1), "pi", 1)
		source.projects = append(source.projects, projects...)
		source.agents = append(source.agents, agents...)
	}
	source.supervisorLive = true
	m.agents.accept(buildAgents(source))
	m.invalidateView()
	rows := strings.Split(ansi.Strip(m.View().Content), "\n")
	if !strings.HasPrefix(rows[37], " ── spawn") {
		t.Fatalf("row 37 = %q, want the legend", rows[37])
	}
	if !strings.Contains(rows[36], "projects 1–4 of 9") {
		t.Fatalf("row 36 = %q, want the scroll line", rows[36])
	}
}

// AC2 of task 147: each key of agents.md rule 4 that changes the screen logs
// a user action.
func TestGfAgentsKeysLogUserActions(t *testing.T) {
	home, clock := hideHome(t, betaManager(records.AgentLive))
	m := openHide(t, home, clock, hideSizes[1])
	for _, press := range []tea.KeyPressMsg{key(tea.KeyRight, ""), key('o', "o"), key(tea.KeyEscape, "esc"), key('h', "h")} {
		var cmd tea.Cmd
		m, cmd = raise(t, m, press)
		m = deliverAgentCommands(t, m, cmd)
	}
	logData, err := os.ReadFile(dailyPath(home, clock.now()))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"action":"agent pod select"`, `"action":"agent pod open"`, `"action":"agent pod close"`, `"action":"agent hide stopped"`} {
		if !strings.Contains(string(logData), want) {
			t.Fatalf("Agents log omits %q:\n%s", want, logData)
		}
	}
}

// Ask a27 (1): on a screen other than Chat, the default first status row is
// state, the screen segment, and the clock. While the details popup is open,
// the screen segment names the popup.
func TestGfAgentsStatusRow(t *testing.T) {
	for _, tc := range []struct {
		name string
		keys []tea.KeyPressMsg
		want string
	}{
		{"overview", nil, " ● working 00:02:00 · agents · ←→ pod · ↑↓ card · enter details · o one pod · a attach · p peek · t task"},
		{"one pod", []tea.KeyPressMsg{key('o', "o")}, " ● working 00:02:00 · agents · ←↑↓→ move · enter details · a attach · p peek · t task · esc pods · h stopped"},
		{"details", []tea.KeyPressMsg{key(tea.KeyEnter, "")}, " ● working 00:02:00 · details · a attach · p peek · t task · ↑↓ scroll · esc close"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, rows := gfAgentsRows(t, "GfAgents", tc.keys...)
			want := tc.want + strings.Repeat(" ", 120-1-5-ansi.StringWidth(tc.want)) + "14:22 "
			if got := rows[39]; got != want {
				t.Fatalf("row 39 = %q, want %q", got, want)
			}
		})
	}
}

// agents.md rule 12 (ask a29, case A2): in the overview, ↑↓ select a card of
// the selected pod, ↑ from its first card selects the coordinator, ←→
// select a pod and its first card, enter opens the details of the card, and
// o opens the one-pod view with the same card selected. esc returns to the
// overview and keeps the card.
func TestGfAgentsOverviewSelectsCards(t *testing.T) {
	m := newGfDemoModel(t, "GfAgents")
	m.graph.selected, m.graph.pod = "manager-hive-g2", "hive"
	steps := []struct {
		press tea.KeyPressMsg
		want  string
	}{
		{key(tea.KeyDown, ""), "t5-worker-1"},
		{key(tea.KeyDown, ""), "t6-worker-1"},
		{key(tea.KeyDown, ""), "t6-worker-1"},
		{key(tea.KeyUp, ""), "t5-worker-1"},
		{key(tea.KeyUp, ""), "manager-hive-g2"},
		{key(tea.KeyUp, ""), "coordinator-g8"},
		{key(tea.KeyDown, ""), "t6-worker-1"},
		{key(tea.KeyRight, ""), "manager-api-g1"},
		{key(tea.KeyLeft, ""), "manager-hive-g2"},
		{key(tea.KeyDown, ""), "t5-worker-1"},
	}
	for index, step := range steps {
		m, _ = raise(t, m, step.press)
		if node, _, ok := m.selectedAgent(); !ok || node.record.Key != step.want || m.graph.podView || m.popupOpen {
			t.Fatalf("step %d: selected %q (ok %v), pod view %v, popup %v; want %q in the overview", index, node.record.Key, ok, m.graph.podView, m.popupOpen, step.want)
		}
	}
	m, _ = raise(t, m, key('o', "o"))
	if node, _, _ := m.selectedAgent(); !m.graph.podView || node.record.Key != "t5-worker-1" {
		t.Fatalf("o: pod view %v, selected %q; want the one-pod view of hive with t5-worker-1", m.graph.podView, node.record.Key)
	}
	m, _ = raise(t, m, key(tea.KeyEscape, ""))
	if node, _, _ := m.selectedAgent(); m.graph.podView || node.record.Key != "t5-worker-1" {
		t.Fatalf("esc: pod view %v, selected %q; want the overview with t5-worker-1", m.graph.podView, node.record.Key)
	}
	m, _ = raise(t, m, key(tea.KeyEnter, ""))
	if title, _, _ := m.agentDetailsPopup(); !m.popupOpen || !strings.HasPrefix(title, "t5-worker-1 ") {
		t.Fatalf("enter: popup %v, title %q; want the details of t5-worker-1", m.popupOpen, title)
	}
}

// Case A2: the selected card of the overview has the heavy accent border,
// and its pod keeps the heavy border.
func TestGfAgentsOverviewShowsSelectedCard(t *testing.T) {
	m := newGfDemoModel(t, "GfAgents")
	m.graph.selected, m.graph.pod = "t6-worker-1", "hive"
	m.invalidateView()
	rows := strings.Split(ansi.Strip(m.View().Content), "\n")
	if !strings.HasPrefix(rows[8], " ┏━ hive") {
		t.Fatalf("row 8 = %q, want the heavy border of the selected pod", rows[8])
	}
	if got := string([]rune(rows[20])[4:]); !strings.HasPrefix(got, "┏━━") || !strings.Contains(rows[21], "┨ ● t6-worker-1") {
		t.Fatalf("rows 20 and 21 = %q, %q; want the heavy border of the selected card", rows[20], rows[21])
	}
}

// agents.md rule 13 (ask a29, case A15): with no pod on the screen, the
// coordinator card is at the top and selected, the notice is at the center
// of the rows above the legend, the title leaves out the home count, and
// the keys are those of this state. enter opens the details of the
// coordinator.
func TestGfAgentsNoPodState(t *testing.T) {
	m, rows := gfAgentsRows(t, "GfAgentsNone")
	if got := strings.TrimRight(rows[0], " "); got != " agents · 0 projects · h shows 3 stopped" {
		t.Fatalf("row 0 = %q", got)
	}
	if got := strings.TrimRight(rows[3], " "); got != strings.Repeat(" ", 48)+"┏━━━━━━━━━━━━━━━━━━━━━━┓" {
		t.Fatalf("row 3 = %q, want the selected coordinator card on column 48", got)
	}
	if got := strings.TrimRight(rows[19], " "); got != strings.Repeat(" ", 51)+"no project agents" {
		t.Fatalf("row 19 = %q, want the notice at the center", got)
	}
	for index, row := range rows[7:37] {
		if index+7 != 19 && strings.TrimSpace(row) != "" {
			t.Fatalf("row %d = %q, want a blank row", index+7, row)
		}
	}
	if got := strings.Join(m.focusHints(m.width), " · "); got != "↑ coordinator · enter details · a attach · p peek · h stopped" {
		t.Fatalf("keys = %q", got)
	}
	m, _ = raise(t, m, key(tea.KeyEnter, ""))
	if title, _, _ := m.agentDetailsPopup(); !m.popupOpen || !strings.HasPrefix(title, "coordinator-g8 ") {
		t.Fatalf("enter: popup %v, title %q; want the details of the coordinator", m.popupOpen, title)
	}
}

// agents.md rule 14 (ask a29, case A13): the details popup has the fields
// of the board from stored data, the last message that the agent sent, and
// the key row on its last inner row. It covers the rows above the status
// row, at their center.
func TestGfAgentsDetailsPopup(t *testing.T) {
	_, rows := gfAgentsRows(t, "GfAgentsDetails")
	want := []string{
		"╭─ t6-worker-1 · esc closes ───────────────────────────────╮",
		"│ profile   builder                                        │",
		"│ project   hive                                           │",
		"│ task      6-starting-status · assigned                   │",
		"│ harness   pi · openai-codex/gpt-6.1-sol                  │",
		"│ state     working · 00:01:00 · changed 14:21:02          │",
		"│ parent    manager-hive                                   │",
		"│ pane      w6A:p1 · tab t6-worker-1                       │",
		"│ worktree  projects/hive/worktrees/6-starting-status      │",
		"│ branch    task/6-starting-status                         │",
		"│ started   14:19:33 · by manager-hive                     │",
		"│ inbox     0 unread                                       │",
		"│                                                          │",
		"│ LAST MESSAGE                                             │",
		"│ 14:21     → manager-hive · gate green, writing V2        │",
		"│                                                          │",
		"│                                                          │",
		"│                                                          │",
		"│ a attach · p peek · t task · esc closes                  │",
		"╰──────────────────────────────────────────────────────────╯",
	}
	for index, line := range want {
		if got := string([]rune(rows[9+index])[30:90]); got != line {
			t.Fatalf("row %d = %q, want %q", 9+index, got, line)
		}
	}
	if !strings.Contains(rows[39], "details · a attach") {
		t.Fatalf("row 39 = %q, want the status row of the details popup", rows[39])
	}
}
