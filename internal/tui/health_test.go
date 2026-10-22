package tui

import (
	"fmt"
	"os"
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

// checkBoardGolden builds the demo model from the fixture of board and
// compares the whole view with the golden grid, cell by cell (decision 42).
func checkBoardGolden(t *testing.T, board string) {
	t.Helper()
	want := readBoardGrid(t, board)
	m := newTestDemoModel(t, board+".fixture.json", demoSize{})
	got := renderGrid(t, m, want.Width, want.Height)
	diffs, err := cellgrid.Compare(want, got, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(diffs) != 0 {
		t.Fatalf("%d cells differ from %s.grid:\n%s\nrendered:\n%s", len(diffs), board, cellgrid.FormatDiffs(diffs), m.View().Content)
	}
}

func TestHealthAgentsMatchesItsBoard(t *testing.T) {
	checkBoardGolden(t, "HealthAgents")
}

func TestHealthMatchesItsBoard(t *testing.T) {
	checkBoardGolden(t, "Health")
}

func TestCompactHealthMatchesItsBoard(t *testing.T) {
	checkBoardGolden(t, "CompactHealth")
}

// healthDemoModel builds the demo model of a Health board after change
// edits its fixture.
func healthDemoModel(t *testing.T, board string, change func(*demoFixture)) *Model {
	t.Helper()
	fixture, err := loadDemoFixture(filepath.Join(designDir, board+".fixture.json"))
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

// requireHealthFits checks the properties of decision 46: the view fills
// the pane, no line is wider than the pane, and no border breaks.
func requireHealthFits(t *testing.T, m *Model, width, height int) string {
	t.Helper()
	view := ansi.Strip(m.View().Content)
	lines := strings.Split(view, "\n")
	if len(lines) != height {
		t.Fatalf("view has %d lines, want %d:\n%s", len(lines), height, view)
	}
	for index, line := range lines {
		if got := ansi.StringWidth(line); got != width {
			t.Fatalf("line %d has width %d, want %d: %q", index+1, got, width, line)
		}
	}
	requireWholeBorders(t, view)
	return view
}

// TestHealthFitsEachSize renders both views at the sizes of decision 46,
// at the minimum height and at one row more, with the board data, with no
// data, with text longer than its column, and with more rows than fit.
func TestHealthFitsEachSize(t *testing.T) {
	long := strings.Repeat("long-", 14)
	cases := map[string]func(*demoFixture){
		"board": nil,
		"empty": func(f *demoFixture) { f.Data.Checks, f.Data.Agents = nil, nil },
		"long": func(f *demoFixture) {
			f.Data.Checks[0].Name, f.Data.Checks[0].Detail = long+"check", long+"detail"
			f.Data.Agents[2].Name, f.Data.Agents[2].Role = long+"name", long+"role"
		},
		"many": func(f *demoFixture) {
			for index := range 60 {
				f.Data.Checks = append(f.Data.Checks, doctor.Check{Name: fmt.Sprintf("extra %02d", index), Status: doctor.StatusOK, Detail: "ok"})
				f.Data.Agents = append(f.Data.Agents, records.Agent{Key: fmt.Sprintf("w%02d", index), Name: fmt.Sprintf("w%02d", index), Role: "worker", Status: records.AgentLive})
			}
			f.UI.Selection = "extra 59"
			if f.UI.View == "agents" {
				f.UI.Selection = "w59"
			}
		},
	}
	sizes := []demoSize{{79, 24}, {80, 24}, {80, 25}, {119, 40}, {119, 41}, {120, 40}, {120, 41}}
	for _, board := range []string{"Health", "CompactHealth"} {
		for name, change := range cases {
			for _, size := range sizes {
				t.Run(fmt.Sprintf("%s/%s/%dx%d", board, name, size.Width, size.Height), func(t *testing.T) {
					m := healthDemoModel(t, board, change)
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
					if name == "many" && !strings.Contains(view, "59") {
						t.Fatalf("the selected last row is not visible:\n%s", view)
					}
					if name == "long" && !strings.Contains(view, "…") {
						t.Fatalf("long text is not cut with …:\n%s", view)
					}
					if name == "empty" && !strings.Contains(view, "loading checks…") && !strings.Contains(view, "no agents") {
						t.Fatalf("empty data shows no notice:\n%s", view)
					}
					// Decisions 9 and 46: below 120 columns, enter opens the
					// details overlay.
					m, _ = raise(t, m, key(tea.KeyEnter, ""))
					view = requireHealthFits(t, m, size.Width, size.Height)
					if size.Width < 120 && name != "empty" && !strings.Contains(view, "esc back") {
						t.Fatalf("enter did not open the details overlay:\n%s", view)
					}
				})
			}
		}
	}
}

// footerLine gives the plain footer row of the view.
func footerLine(m *Model) string {
	lines := strings.Split(ansi.Strip(m.View().Content), "\n")
	return strings.TrimRight(lines[len(lines)-1], " ")
}

// TestHealthKeysAndFooterHints checks the keys of decisions 21 to 23 and the
// key hints of each focus state of the Health screen.
func TestHealthKeysAndFooterHints(t *testing.T) {
	m := healthDemoModel(t, "Health", nil)
	if got, want := footerLine(m), " [list] ↑↓ move · tab region · [ ] view · enter open · R refresh · ? help"; got != want {
		t.Fatalf("Home list footer = %q; want %q", got, want)
	}
	m, _ = raise(t, m, key('j', "j"))
	if m.health.selected != "git" {
		t.Fatalf("j selected %q; want git", m.health.selected)
	}
	requireView(t, m, "╭─ git ─")
	m, _ = raise(t, m, key('k', "k"))
	m, _ = raise(t, m, key('G', "G"))
	if m.health.selected != "gh" {
		t.Fatalf("G selected %q; want gh", m.health.selected)
	}
	m, _ = raise(t, m, key('g', "g"))
	if m.health.selected != "herdr socket" {
		t.Fatalf("g selected %q; want herdr socket", m.health.selected)
	}
	m, _ = raise(t, m, key(tea.KeyTab, ""))
	if got, want := footerLine(m), " [details] ↑↓ scroll · tab region · esc back to list · ? help"; got != want {
		t.Fatalf("details footer = %q; want %q", got, want)
	}
	m, _ = raise(t, m, key(tea.KeyTab, ""))
	m, _ = raise(t, m, key(']', "]"))
	if m.activeView() != healthAgents {
		t.Fatalf("] gave view %d; want Agents", m.activeView())
	}
	if got, want := footerLine(m), " [list] ↑↓ move · tab region · [ ] view · enter open · r restart · R refresh · ? help"; got != want {
		t.Fatalf("Agents list footer = %q; want %q", got, want)
	}
	requireView(t, m, "AGENT", "coordinator-g2", "supervisor-g4", "manager-hive", "cp5-worker-01", "t4-worker-1")
	m, _ = raise(t, m, key('h', "h"))
	rejectView(t, m, "t4-worker-1")
	m, _ = raise(t, m, key('h', "h"))
	requireView(t, m, "t4-worker-1")
	// Decision 35: restart, attach, and peek show the demo notice.
	for _, press := range []tea.KeyPressMsg{key('r', "r"), key('a', "a"), key('p', "p")} {
		m.demo.notice = ""
		m, _ = raise(t, m, press)
		if !strings.Contains(footerLine(m), demoNotice) {
			t.Fatalf("%s footer = %q; want %q", press.String(), footerLine(m), demoNotice)
		}
	}
	m, _ = raise(t, m, key('[', "["))
	if m.activeView() != healthHome {
		t.Fatalf("[ gave view %d; want Home", m.activeView())
	}

	// Below 120 columns, enter opens the details overlay and esc closes it.
	compact := healthDemoModel(t, "CompactHealth", nil)
	if got, want := footerLine(compact), " [list] ↑↓ move · enter details · [ ] view · r restart · R refresh · ?"; got != want {
		t.Fatalf("compact footer = %q; want %q", got, want)
	}
	compact, _ = raise(t, compact, key(tea.KeyEnter, ""))
	requireView(t, compact, "coordinator-g2 · esc back", "r restart · starts a new coordinator")
	if got, want := footerLine(compact), " [details] ↑↓ scroll · esc back to list"; got != want {
		t.Fatalf("overlay footer = %q; want %q", got, want)
	}
	compact, _ = raise(t, compact, key(tea.KeyEscape, ""))
	if compact.detailsFocused() {
		t.Fatal("esc did not close the details overlay")
	}
}

// TestHealthRestartPromptKeys checks the restart prompt of the HealthAgents
// board: the box in the details, the footer of decision 21, and the cancel.
func TestHealthRestartPromptKeys(t *testing.T) {
	m := healthDemoModel(t, "HealthAgents", nil)
	requireView(t, m, "restart supervisor-g4?", "y restart   esc cancel", "[restart] y restart · any other key cancels")
	m, _ = raise(t, m, key('n', "n"))
	rejectView(t, m, "restart supervisor-g4?", "[restart]")
	if !strings.Contains(footerLine(m), "restart cancelled") {
		t.Fatalf("footer = %q; want the cancel notice", footerLine(m))
	}
}

// TestHealthRIsOnlyRestartInAgentsView checks that r in Home starts nothing
// and shows no notice (decision 19).
func TestHealthRIsOnlyRestartInAgentsView(t *testing.T) {
	home := newTestHome(t)
	m := newTestModel(t, home, fixedClock(t, time.Date(2026, 10, 21, 12, 0, 0, 0, time.UTC)))
	m.width, m.height = 120, 40
	m.selectScreen(tabHealth, healthHome)
	m.health.acceptAgents(agentsData{nodes: []agentNode{{record: records.Agent{Key: "coordinator-g1", Role: "coordinator", Status: records.AgentLive}, current: true}}})
	m, cmd := raise(t, m, key('r', "r"))
	if cmd != nil || m.agents.restartRunning || m.agents.restartConfirm || m.agents.restartNotice != "" {
		t.Fatalf("r in Home acted: cmd %v, running %v, confirm %v, notice %q", cmd != nil, m.agents.restartRunning, m.agents.restartConfirm, m.agents.restartNotice)
	}
}

// TestHealthNormalModeRefreshKeepsSelection checks normal mode in a scratch
// home: R loads the checks and the agents with the record loaders, starts no
// hive process, and a refresh keeps the selected check (decision 38).
func TestHealthNormalModeRefreshKeepsSelection(t *testing.T) {
	home := newTestHome(t)
	clock := fixedClock(t, time.Date(2026, 10, 21, 12, 0, 0, 0, time.UTC))
	m := newTestModel(t, home, clock)
	m.width, m.height = 120, 40
	m.selectScreen(tabHealth, healthHome)
	bin := t.TempDir()
	marker := filepath.Join(bin, "hive-ran")
	if err := os.WriteFile(filepath.Join(bin, "hive"), []byte("#!/bin/sh\nprintf invoked > \"$HIVE_MARKER\"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("HIVE_MARKER", marker)

	m, cmd := raise(t, m, key('R', "R"))
	if cmd == nil || !m.screenLoading {
		t.Fatalf("R returned command %v with loading=%t; want a refresh", cmd != nil, m.screenLoading)
	}
	var loaded screenLoadMsg
	for _, message := range flattenCmd(t, cmd) {
		if load, ok := message.(screenLoadMsg); ok {
			loaded = load
		}
	}
	if loaded.request != healthDataRequest || len(loaded.health.Checks) == 0 {
		t.Fatalf("R gave %+v; want the doctor checks", loaded.request)
	}
	m, _ = raise(t, m, loaded)
	requireView(t, m, "CHECK", "RESULT", "checks · ", "checked   "+m.now().Format("15:04:05")+" · R to check again")
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("the TUI ran the hive binary; marker stat error = %v", err)
	}

	result := doctor.Result{Checks: []doctor.Check{
		{Name: "home", Status: doctor.StatusOK, Detail: "/scratch"},
		{Name: "git", Status: doctor.StatusWarn, Detail: "git version 2"},
	}}
	m, _ = raise(t, m, screenLoadMsg{request: healthDataRequest, health: result})
	m, _ = raise(t, m, key('j', "j"))
	result.Checks = append(result.Checks, doctor.Check{Name: "gh", Status: doctor.StatusFail, Detail: "missing"})
	m, _ = raise(t, m, screenLoadMsg{request: healthDataRequest, health: result})
	if m.health.selected != "git" {
		t.Fatalf("refresh selected %q; want git", m.health.selected)
	}
	requireView(t, m, "checks · 3 · 1 warn · 1 fail", "× gh", "! git")
}

// flattenCmd runs cmd and each command of a batch, and gives each message.
func flattenCmd(t *testing.T, cmd tea.Cmd) []tea.Msg {
	t.Helper()
	if cmd == nil {
		return nil
	}
	message := cmd()
	if batch, ok := message.(tea.BatchMsg); ok {
		var out []tea.Msg
		for _, child := range batch {
			out = append(out, flattenCmd(t, child)...)
		}
		return out
	}
	return []tea.Msg{message}
}

func TestHealthProgramUsesStringReader(t *testing.T) {
	home := newTestHome(t)
	var out strings.Builder
	// Decision 3: 5 selects Health, and the sub-tab row marks the Home view.
	if err := RunHome(home, strings.NewReader("5R\x03\x03"), &out); err != nil {
		t.Fatalf("RunHome: %v", err)
	}
	for _, want := range []string{"[Home]", "checks ·"} {
		if !strings.Contains(ansi.Strip(out.String()), want) {
			t.Fatalf("program output omitted %q:\n%s", want, ansi.Strip(out.String()))
		}
	}
}
