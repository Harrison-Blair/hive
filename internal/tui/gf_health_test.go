package tui

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/Harrison-Blair/hive/internal/doctor"
	"github.com/Harrison-Blair/hive/internal/lib/cellgrid"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/charmbracelet/x/ansi"
)

// gfHealthBoards are the 4 Health screen boards of the golden frames.
var gfHealthBoards = []string{"GfHealth", "GfHealthGreen", "GfHealthRestart", "GfHealthChecking"}

// AC3, tui-v2 decisions 42 and 44: each Health board of the golden frames
// matches the View of its fixture cell for cell, with no mask.
func TestGfHealthBoardsMatch(t *testing.T) {
	for _, board := range gfHealthBoards {
		t.Run(board, func(t *testing.T) {
			want := readGfGrid(t, board)
			m := newGfDemoModel(t, board)
			got := renderGrid(t, m, want.Width, want.Height)
			diffs, err := cellgrid.Compare(want, got, nil)
			if err != nil {
				t.Fatal(err)
			}
			if len(diffs) != 0 {
				shown := diffs[:min(len(diffs), 12)]
				t.Fatalf("%d cells differ from %s.grid:\n%s\nrendered:\n%s", len(diffs), board, cellgrid.FormatDiffs(shown), m.View().Content)
			}
		})
	}
}

// selectHealthTable selects the Health screen and gives the focus to the
// agents table, where r, a, and p act (health.md keys).
func selectHealthTable(m *Model) {
	m.selectScreen(tabHealth, 0)
	m.regionFocus[tabHealth] = 1
}

// gfHealthModel gives the demo model of a Health board after change edits
// its fixture.
func gfHealthModel(t *testing.T, board string, change func(*demoFixture)) *Model {
	t.Helper()
	requireSpecFile(t, filepath.Join(gfDesignDir, board+".fixture.json"))
	fixture, err := loadDemoFixture(filepath.Join(gfDesignDir, board+".fixture.json"))
	if err != nil {
		t.Fatal(err)
	}
	if change != nil {
		change(fixture)
	}
	m := newDemoModel(t.TempDir(), &demoState{fixture: fixture})
	t.Cleanup(m.Stop)
	return m
}

// healthStatusRow gives the plain status row of the view.
func healthStatusRow(m *Model) string {
	rows := viewRows(m)
	return strings.TrimSpace(rows[len(rows)-1])
}

// detailsTitle gives the plain title of the details box.
func detailsTitle(t *testing.T, m *Model) string {
	t.Helper()
	for _, row := range viewRows(m) {
		if strings.HasPrefix(row, "╭─ ✓ ") || strings.HasPrefix(row, "╭─ ! ") {
			title, _, _ := strings.Cut(strings.TrimPrefix(row, "╭─ "), " ─")
			return title
		}
	}
	t.Fatalf("no details box:\n%s", ansi.Strip(m.View().Content))
	return ""
}

// AC2, health.md rules 1 and 5: ↑ and ↓ move between the area rows, ← and
// → move between the tiles of a row, and the details box shows the selected
// tile. Each move is a logged user action.
func TestGfHealthGridKeysMoveTheTile(t *testing.T) {
	m := gfHealthModel(t, "GfHealth", nil)
	steps := []struct {
		press tea.KeyPressMsg
		title string
	}{
		{key(tea.KeyRight, ""), "✓ ownerless inbox records"},
		{key(tea.KeyUp, ""), "✓ git"},
		{key(tea.KeyUp, ""), "✓ herdr pane"},
		{key(tea.KeyUp, ""), "✓ herdr pane"},
		{key(tea.KeyDown, ""), "✓ state store"},
		{key(tea.KeyDown, ""), "✓ unresolved agents"},
		{key(tea.KeyDown, ""), "✓ supervisor binary"},
		{key(tea.KeyRight, ""), "✓ supervisor binary"},
		{key(tea.KeyDown, ""), "✓ harness codex"},
		{key(tea.KeyLeft, ""), "✓ harness claude"},
		{key(tea.KeyLeft, ""), "✓ harness claude"},
	}
	if got := detailsTitle(t, m); got != "! unresolved agents · 1" {
		t.Fatalf("first details title = %q; want the first tile that did not pass", got)
	}
	for index, step := range steps {
		m, _ = raise(t, m, step.press)
		want := step.title
		if want == "✓ unresolved agents" {
			want = "! unresolved agents · 1"
		}
		if got := detailsTitle(t, m); got != want {
			t.Fatalf("step %d %s: details title = %q; want %q", index, step.press.String(), got, want)
		}
	}
	requireView(t, m, "▶ ✓ claude")
	if got, want := healthStatusRow(m), "health · ↑↓ ←→ tile · ^] agents · R check again"; !strings.Contains(got, want) {
		t.Fatalf("status row = %q; want %q", got, want)
	}
}

// AC2, health.md rule 5 and the key table: ctrl+] gives the focus to the
// table, which gets the accent border. ↑, ↓, g, and G select a row, and the
// grid hides its ▶ (ask a28).
func TestGfHealthTableKeysSelectARow(t *testing.T) {
	m := gfHealthModel(t, "GfHealth", nil)
	m, _ = raise(t, m, ctrl(']'))
	if m.focusState() != "table" {
		t.Fatalf("ctrl+] gave focus %q; want table", m.focusState())
	}
	grid := renderGrid(t, m, 120, 40)
	if cell := grid.At(18, 0); cell.Rune != '╭' || cell.Token != cellgrid.Accent {
		t.Fatalf("table corner = %q %s; want ╭ accent", cell.Rune, cell.Token)
	}
	if cell := grid.At(8, 0); cell.Token != cellgrid.Border {
		t.Fatalf("details corner = %s; want border", cell.Token)
	}
	if got, want := healthStatusRow(m), "health · ↑↓ row · g/G top/end · pgup/pgdn page · ^] checks · r restart · a attach · p peek"; !strings.Contains(got, want) {
		// Decision 10 drops the keys on the right, before the clock, when
		// the row is wider than the pane.
		t.Fatalf("status row = %q; want %q", got, want)
	}
	rejectView(t, m, "▶ !")
	for _, step := range []struct {
		press tea.KeyPressMsg
		agent string
	}{
		{key(tea.KeyDown, ""), "manager-hive-g2"},
		{key('G', "G"), "plan-review-1"},
		{key(tea.KeyDown, ""), "plan-review-1"},
		{key(tea.KeyUp, ""), "t3-verifier-1"},
		{key('g', "g"), "coordinator-g2"},
		{key(tea.KeyUp, ""), "coordinator-g2"},
	} {
		m, _ = raise(t, m, step.press)
		if m.health.selectedAgent != step.agent {
			t.Fatalf("%s selected %q; want %q", step.press.String(), m.health.selectedAgent, step.agent)
		}
	}
	requireView(t, m, "▶ ● coordinator ")
	m, _ = raise(t, m, ctrl(']'))
	if m.focusState() != "grid" {
		t.Fatalf("second ctrl+] gave focus %q; want grid", m.focusState())
	}
	requireView(t, m, "▶ ! unresolved 1")
}

// AC2, H12: r, a, and p act only while the table has the focus. In the grid
// they do nothing and show no notice.
func TestGfHealthActionKeysActOnlyInTheTable(t *testing.T) {
	m := gfHealthModel(t, "GfHealth", nil)
	for _, press := range []tea.KeyPressMsg{key('r', "r"), key('a', "a"), key('p', "p")} {
		m, _ = raise(t, m, press)
		if m.demo.notice != "" || m.agents.restartConfirm || m.agents.restartNotice != "" {
			t.Fatalf("%s in the grid acted: notice %q, confirm %v, restart notice %q", press.String(), m.demo.notice, m.agents.restartConfirm, m.agents.restartNotice)
		}
	}
	// Rule 8 and decision 35: in the table, each one shows the demo notice.
	m, _ = raise(t, m, ctrl(']'))
	for _, press := range []tea.KeyPressMsg{key('r', "r"), key('a', "a"), key('p', "p")} {
		m.demo.notice = ""
		m, _ = raise(t, m, press)
		if got := footerLine(m); got != demoNotice {
			t.Fatalf("%s in the table: notice line = %q; want %q", press.String(), got, demoNotice)
		}
	}
}

// Rule 6: r on a row that is not the coordinator or the supervisor shows
// the notice and starts nothing. The table move, the grid move, and the
// request are logged user actions.
func TestGfHealthRestartOtherRowShowsTheNotice(t *testing.T) {
	m, dir := newFrameModel(t, 120, 40)
	m.selectScreen(tabHealth, 0)
	m, _ = raise(t, m, screenLoadMsg{request: healthDataRequest, health: doctor.Result{Checks: []doctor.Check{
		{Name: "home", Status: doctor.StatusOK, Detail: "/scratch"},
		{Name: "git", Status: doctor.StatusOK, Detail: "git version 2"},
	}}})
	m, cmd := raise(t, m, key(tea.KeyRight, ""))
	drain(cmd)
	m.health.acceptAgents(agentsData{nodes: []agentNode{
		{record: records.Agent{Key: "coordinator-g2", Name: "coordinator", Role: "coordinator", Status: records.AgentLive}, current: true},
		{record: records.Agent{Key: "manager-hive-g2", Name: "manager-hive", Role: "manager", Status: records.AgentLive}, current: true},
	}})
	m, cmd = raise(t, m, ctrl(']'))
	drain(cmd)
	m, cmd = raise(t, m, key(tea.KeyDown, ""))
	drain(cmd)
	m, cmd = raise(t, m, key('r', "r"))
	drain(cmd)
	if m.agents.restartConfirm || m.agents.restartRunning {
		t.Fatal("r on a manager row started a restart")
	}
	if got := footerLine(m); got != healthRestartOnly {
		t.Fatalf("notice line = %q; want %q", got, healthRestartOnly)
	}
	want := map[string]string{"health selection": "git", "agent selection": "manager-hive-g2", "agent restart request": "unsupported target"}
	for _, event := range tuiActions(t, dir) {
		action, _ := event["action"].(string)
		for _, field := range []string{"check", "agent_id", "result"} {
			if value, ok := event[field]; ok && want[action] == value {
				delete(want, action)
			}
		}
	}
	if len(want) != 0 {
		t.Fatalf("actions not logged: %v", want)
	}
}

// Rule 7: R shows checking in the title row, dims the tiles and the
// details, and the result ends it. A failed hive doctor shows the error on
// the notice line, and the tiles keep their last result.
func TestGfHealthCheckAgainShowsCheckingAndTheError(t *testing.T) {
	m, dir := newFrameModel(t, 120, 40)
	m.selectScreen(tabHealth, 0)
	result := doctor.Result{Checks: []doctor.Check{
		{Name: "home", Status: doctor.StatusOK, Detail: "/scratch"},
		{Name: "git", Status: doctor.StatusWarn, Detail: "git missing"},
	}}
	m, _ = raise(t, m, screenLoadMsg{request: healthDataRequest, health: result})
	checked := m.health.checkedAt.Format("15:04:05")
	if got := viewRows(m)[0]; !strings.HasSuffix(strings.TrimSpace(got), "· 2 checks · 1 ok · 1 warn · checked "+checked+" · R checks again") {
		t.Fatalf("title row = %q", got)
	}
	m.healthLoader = func(context.Context, string, *slog.Logger) (doctor.Result, error) {
		return doctor.Result{}, errors.New("state store locked")
	}
	m, cmd := raise(t, m, key('R', "R"))
	if !m.health.checking {
		t.Fatal("R did not start checking")
	}
	if got := strings.TrimSpace(viewRows(m)[0]); !strings.HasSuffix(got, "· checked "+checked+" · checking…") {
		t.Fatalf("title row while checking = %q", got)
	}
	requireView(t, m, "source hive doctor · last result "+checked+" · new result pending")
	grid := renderGrid(t, m, 120, 40)
	if cell := grid.At(3, 14); cell.Rune != '✓' || cell.Token != cellgrid.Dim {
		t.Fatalf("home tile mark = %q %s; want ✓ dim", cell.Rune, cell.Token)
	}
	for _, message := range flattenCmd(t, cmd) {
		if load, ok := message.(screenLoadMsg); ok && load.request == healthDataRequest {
			m, _ = raise(t, m, load)
		}
	}
	if m.health.checking {
		t.Fatal("the doctor error did not end checking")
	}
	if got, want := footerLine(m), "error: hive doctor: state store locked"; got != want {
		t.Fatalf("notice line = %q; want %q", got, want)
	}
	requireView(t, m, "✓ /scratch", "! git git missing", "R checks again")
	var refresh bool
	for _, event := range tuiActions(t, dir) {
		refresh = refresh || event["action"] == "refresh"
	}
	if !refresh {
		t.Fatal("R was not logged as a user action")
	}
}

// Rule 4 and tui-v2 decision 55: the dot and the word of each agent state,
// the pane column, and a title that counts only the states that it has.
func TestGfHealthTableStates(t *testing.T) {
	m, _ := newFrameModel(t, 120, 40)
	m.selectScreen(tabHealth, 0)
	changed := gfDay.Add(-7 * time.Minute).UTC().Format(time.RFC3339Nano)
	agent := func(key, role, status, agentStatus string) agentNode {
		return agentNode{record: records.Agent{Key: key, Name: key, Role: role, Status: status, AgentStatus: agentStatus, PaneID: "w1:p1", StatusChangedAt: changed}, current: true}
	}
	m.health.acceptAgents(agentsData{nodes: []agentNode{
		agent("w-blocked", "worker", records.AgentLive, "blocked"),
		agent("w-idle", "worker", records.AgentLive, "idle"),
		agent("w-plain", "worker", records.AgentLive, ""),
		agent("w-stale", "worker", records.AgentStale, ""),
		agent("coordinator", "coordinator", records.AgentStopped, ""),
	}})
	for _, want := range []string{
		"▶ ○ coordinator", "stopped        gone",
		"▲ w-blocked", "blocked 7m",
		"○ w-idle", "idle 7m",
		"● w-plain", "live 7m",
		"○ w-stale", "stale          gone",
		"agents · 5 · 2 live · 1 idle · 1 stopped · 1 stale",
	} {
		requireView(t, m, want)
	}
	rejectView(t, m, "unresolved", "starting")
}

// AC2 and V3, tui-v2 decision 46: the Health view fills the pane at each
// size, no line is wider than the pane, and no border breaks. Each case runs
// with the grid and with the table focus: the board data, no data, text
// longer than its column, and more agent rows than fit with the last row
// selected.
func TestGfHealthFitsEachSize(t *testing.T) {
	long := strings.Repeat("long-", 14)
	cases := map[string]func(*demoFixture){
		"board": nil,
		"empty": func(f *demoFixture) { f.Data.Checks, f.Data.Agents = nil, nil },
		"long": func(f *demoFixture) {
			f.Data.Checks[0].Detail = long + "detail"
			f.Data.Agents[2].Name, f.Data.Agents[2].Role = long+"name", long+"role"
		},
		"many": func(f *demoFixture) {
			for index := range 60 {
				name := fmt.Sprintf("w%02d", index)
				f.Data.Agents = append(f.Data.Agents, records.Agent{Key: name, Name: name, Role: "worker", Status: records.AgentLive, PaneID: "w9:p1"})
			}
			f.UI.Selection = "w59"
		},
	}
	sizes := []demoSize{{79, 24}, {80, 24}, {80, 25}, {119, 40}, {119, 41}, {120, 40}, {120, 41}}
	for name, change := range cases {
		for _, focus := range []string{"grid", "table"} {
			for _, size := range sizes {
				t.Run(fmt.Sprintf("%s/%s/%dx%d", name, focus, size.Width, size.Height), func(t *testing.T) {
					m := gfHealthModel(t, "GfHealth", func(f *demoFixture) {
						if change != nil {
							change(f)
						}
						f.UI.Focus = focus
					})
					m, _ = raise(t, m, tea.WindowSizeMsg{Width: size.Width, Height: size.Height})
					if size.Width < minimumWidth {
						if view := ansi.Strip(m.View().Content); !strings.Contains(view, "terminal too small") {
							t.Fatalf("view at %d columns is not the too-small notice:\n%s", size.Width, view)
						}
						return
					}
					view := requireHealthFits(t, m, size.Width, size.Height)
					if name != "empty" && !strings.Contains(view, "▶") {
						t.Fatalf("the selected item is not visible:\n%s", view)
					}
					if name == "many" && (!strings.Contains(view, "▶ ● w59") || !strings.Contains(view, "┃")) {
						t.Fatalf("the selected last row or the ┃ thumb is not visible:\n%s", view)
					}
					if name == "long" && !strings.Contains(view, "✓ socket · long-long-long-long-…") {
						t.Fatalf("long text is not cut with …:\n%s", view)
					}
					if name == "long" && size.Height >= 40 && !strings.Contains(view, "long-long-long-l… ") {
						t.Fatalf("a long agent name is not cut with … in its column:\n%s", view)
					}
					if name == "board" && 1+ansi.StringWidth(healthNote) > size.Width {
						for _, row := range strings.Split(view, "\n") {
							if strings.HasPrefix(row, " restart: coordinator") && !strings.HasSuffix(row, "…") {
								t.Fatalf("the note row is not cut with …: %q", row)
							}
						}
					}
					if name == "empty" && (!strings.Contains(view, "no checks") || !strings.Contains(view, "no agents")) {
						t.Fatalf("empty data shows no notice:\n%s", view)
					}
				})
			}
		}
	}
}

// health.md states: an old supervisor binary gives the supervisor binary
// tile ! and old, and the title counts it as a warn.
func TestGfHealthOldSupervisorBinaryTile(t *testing.T) {
	m, _ := newFrameModel(t, 120, 40)
	m.selectScreen(tabHealth, 0)
	m, _ = raise(t, m, screenLoadMsg{request: healthDataRequest, health: doctor.Result{Checks: []doctor.Check{
		{Name: "supervisor", Status: doctor.StatusOK, Detail: "supervisor live: pid 7"},
		{Name: "supervisor binary", Status: doctor.StatusWarn, Detail: "supervisor /a modified 1; hive /b modified 2; supervisor binary is older"},
	}}})
	requireView(t, m, "✓ live · pid 7", "▶ ! binary old", "2 checks · 1 ok · 1 warn", "! supervisor binary · old")
}

// health.md rules 1, 4, and 7 and ask a28: the grid shows its ▶ only while
// the grid has the focus. The table keeps its ▶ at all times, also while
// hive doctor runs.
func TestGfHealthGridSelectionNeedsGridFocus(t *testing.T) {
	m := gfHealthModel(t, "GfHealth", nil)
	requireView(t, m, "▶ ! unresolved 1", "▶ ● supervisor")
	m, _ = raise(t, m, ctrl(']'))
	rejectView(t, m, "▶ ! unresolved 1")
	requireView(t, m, "▶ ● supervisor")
	m, _ = raise(t, m, key('R', "R"))
	if !m.health.checking {
		t.Fatal("R did not start checking")
	}
	rejectView(t, m, "▶ ! unresolved 1")
	requireView(t, m, "▶ ● supervisor")
	m, _ = raise(t, m, ctrl(']'))
	requireView(t, m, "▶ ! unresolved 1")
}
