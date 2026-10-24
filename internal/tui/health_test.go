package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/Harrison-Blair/hive/internal/doctor"
	"github.com/Harrison-Blair/hive/internal/lib/cellgrid"
	"github.com/charmbracelet/x/ansi"
)

// checkBoardGolden builds the demo model from the fixture of board and
// compares the whole view with the golden grid, cell by cell (decision 42).
func checkBoardGolden(t *testing.T, board string) {
	t.Helper()
	if skipRetired(t, board) {
		return
	}
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
	requireSpecFile(t, filepath.Join(designDir, board+".fixture.json"))
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

// footerLine gives the plain footer row of the view.
// footerLine gives the notice line of golden-frames decision 14: the row
// above the status row.
func footerLine(m *Model) string {
	lines := strings.Split(ansi.Strip(m.View().Content), "\n")
	return strings.TrimSpace(lines[len(lines)-2])
}

// TestHealthNormalModeRefreshKeepsSelection checks normal mode in a scratch
// home: R loads the checks and the agents with the record loaders, starts no
// hive process, and a refresh keeps the selected check (decision 38).
func TestHealthNormalModeRefreshKeepsSelection(t *testing.T) {
	home := newTestHome(t)
	clock := fixedClock(t, time.Date(2026, 10, 21, 12, 0, 0, 0, time.UTC))
	m := newTestModel(t, home, clock)
	m.width, m.height = 120, 40
	m.selectScreen(tabHealth, 0)
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
	requireView(t, m, "health · ", "checked "+m.now().Format("15:04:05")+" · R checks again")
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("the TUI ran the hive binary; marker stat error = %v", err)
	}

	result := doctor.Result{Checks: []doctor.Check{
		{Name: "home", Status: doctor.StatusOK, Detail: "/scratch"},
		{Name: "git", Status: doctor.StatusWarn, Detail: "git version 2"},
	}}
	m, _ = raise(t, m, screenLoadMsg{request: healthDataRequest, health: result})
	m, _ = raise(t, m, key(tea.KeyDown, ""))
	result.Checks = append(result.Checks, doctor.Check{Name: "gh", Status: doctor.StatusFail, Detail: "missing"})
	m, _ = raise(t, m, screenLoadMsg{request: healthDataRequest, health: result})
	if m.health.selected != "git" {
		t.Fatalf("refresh selected %q; want git", m.health.selected)
	}
	requireView(t, m, "3 checks · 1 ok · 2 warn", "! git", "! gh")
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
	// Golden-frames decision 16: ctrl+y selects Health, and the status row
	// shows the grid keys.
	if err := RunHome(home, strings.NewReader("\x19R\x03\x03"), &out); err != nil {
		t.Fatalf("RunHome: %v", err)
	}
	for _, want := range []string{"health · ", "↑↓ ←→ tile"} {
		if !strings.Contains(ansi.Strip(out.String()), want) {
			t.Fatalf("program output omitted %q:\n%s", want, ansi.Strip(out.String()))
		}
	}
}
