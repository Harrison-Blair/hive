package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/Harrison-Blair/hive/internal/doctor"
	"github.com/charmbracelet/x/ansi"
)

func TestHealthScreenFramesPanelsAndSeparatesWarnings(t *testing.T) {
	home := newTestHome(t)
	clock := fixedClock(t, time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC))
	m := newTestModel(t, home, clock)
	m.screenIndex = 6
	m, _ = raise(t, m, tea.WindowSizeMsg{Width: 120, Height: 40})
	result := doctor.Result{Checks: []doctor.Check{
		{Name: "home", Status: doctor.StatusOK, Detail: "/fixture"},
		{Name: "herdr socket", Status: doctor.StatusFail, Detail: "socket missing"},
		{Name: "supervisor binary", Status: doctor.StatusWarn, Detail: "old executable"},
		{Name: "starting agents", Status: doctor.StatusOK, Detail: "0"},
		{Name: "stale worktrees", Status: doctor.StatusWarn, Detail: "demo/1-landed"},
	}}
	m, _ = raise(t, m, screenLoadMsg{request: healthDataRequest, health: result})

	// Decision 8: the list panel is narrow from 120 columns, so the test
	// reads the current Health screen at the content width. The frame draws
	// the outer panels.
	view := ansi.Strip(m.currentScreen().View(120, 40-frameRows-1))
	for _, want := range []string{
		"┤ home ├", "┤ agents ├", "┤ warnings ├",
		"! supervisor binary: old executable", "✕ herdr socket: socket missing",
		"! stale worktrees: demo/1-landed", "later: capacity from qmeter. Not in this goal.",
	} {
		if !strings.Contains(view, want) {
			t.Errorf("health view omitted %q:\n%s", want, view)
		}
	}
	if got := strings.Count(view, "╭"); got != 3 {
		t.Errorf("health view has %d framed panels, want 3 health panels:\n%s", got, view)
	}
	if strings.Contains(view, "┤ capacity ├") {
		t.Errorf("health view shows an out-of-scope capacity panel:\n%s", view)
	}
}

func TestHealthScreenRendersChecksAndKeepsSelectionAcrossRefresh(t *testing.T) {
	home := newTestHome(t)
	clock := fixedClock(t, time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC))
	m := newTestModel(t, home, clock)
	m.width, m.height, m.screenIndex = 160, 40, 6

	initial := doctor.Result{Checks: []doctor.Check{
		{Name: "home", Status: doctor.StatusOK, Detail: "/fixture"},
		{Name: "state store", Status: doctor.StatusWarn, Detail: "lock held"},
		{Name: "herdr socket", Status: doctor.StatusFail, Detail: "socket missing"},
		{Name: "starting agents", Status: doctor.StatusOK, Detail: "0"},
		{Name: "stale worktrees", Status: doctor.StatusWarn, Detail: "demo/1-landed"},
	}}
	m, _ = raise(t, m, screenLoadMsg{request: healthDataRequest, health: initial})
	assertHealthLines(t, m, "✓ ▶ home", "! state store", "✕ herdr socket", "✓ starting agents", "! stale worktrees", "DETAIL home: /fixture", "later: capacity from qmeter. Not in this goal.")
	if !strings.Contains(m.currentScreen().View(120, 30), selectedStyle.Render("▶ home")) {
		t.Fatal("selected health check is not highlighted")
	}

	for range 4 {
		m, _ = raise(t, m, key('j', "j"))
	}
	assertHealthLines(t, m, "DETAIL stale worktrees: demo/1-landed")
	if !strings.Contains(m.currentScreen().View(120, 30), selectedStyle.Render("▶ stale worktrees")) {
		t.Fatal("moved selection is not highlighted")
	}

	refreshed := doctor.Result{Checks: []doctor.Check{
		{Name: "home", Status: doctor.StatusOK, Detail: "/fixture"},
		{Name: "state store", Status: doctor.StatusWarn, Detail: "lock held"},
		{Name: "herdr socket", Status: doctor.StatusFail, Detail: "socket missing"},
		{Name: "starting agents", Status: doctor.StatusOK, Detail: "0"},
		{Name: "git", Status: doctor.StatusOK, Detail: "git version 2"},
		{Name: "stale worktrees", Status: doctor.StatusWarn, Detail: "demo/1-landed, demo/2-landed"},
	}}
	m, _ = raise(t, m, screenLoadMsg{request: healthDataRequest, health: refreshed})
	assertHealthLines(t, m, "DETAIL stale worktrees: demo/1-landed, demo/2-landed")
}

func TestHealthRStartsBackgroundRefreshWithoutOtherActions(t *testing.T) {
	home := newTestHome(t)
	clock := fixedClock(t, time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC))
	m := newTestModel(t, home, clock)
	m.screenIndex = 6
	m.logger = nil
	bin := t.TempDir()
	marker := filepath.Join(bin, "hive-ran")
	fakeHive := filepath.Join(bin, "hive")
	if err := os.WriteFile(fakeHive, []byte("#!/bin/sh\nprintf invoked > \"$HIVE_MARKER\"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("HIVE_MARKER", marker)

	m, cmd := raise(t, m, key('r', "r"))
	if cmd == nil || !m.screenLoading {
		t.Fatalf("r returned command %v with loading=%t, want a background refresh", cmd != nil, m.screenLoading)
	}
	message := cmd()
	loaded, ok := message.(screenLoadMsg)
	if !ok || loaded.request != healthDataRequest {
		t.Fatalf("r returned %T, want health data from a shell request", message)
	}
	m, _ = raise(t, m, loaded)
	// Decision 21: the footer shows the hints of the registry. The check of
	// KeyHints below keeps "ask in Chat" for the current Health screen.
	assertHealthLines(t, m, "supervisor binary", "stale worktrees")
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("TUI ran the hive binary; marker stat error = %v", err)
	}
	// Decision 22: R refreshes on each screen.
	m.screenLoading = false
	if _, cmd = raise(t, m, key('R', "R")); cmd == nil || !m.screenLoading {
		t.Fatalf("R returned command %v with loading=%t, want a refresh", cmd != nil, m.screenLoading)
	}
	for _, press := range []tea.KeyPressMsg{key('w', "w"), key('c', "c")} {
		m.screenLoading = false
		m, cmd = raise(t, m, press)
		if cmd != nil || m.screenLoading {
			t.Fatalf("key %q started an action: command=%v loading=%t", press.String(), cmd != nil, m.screenLoading)
		}
	}
	hints := m.currentScreen().KeyHints()
	if !strings.Contains(hints, "ask in Chat") || !strings.Contains(hints, "r") {
		t.Fatalf("health hints = %q, want re-check and ask in Chat", hints)
	}
	for _, forbidden := range []string{"R restart", "w clean", "c copy"} {
		if strings.Contains(hints, forbidden) {
			t.Fatalf("health hints = %q, must omit %q", hints, forbidden)
		}
	}
}

func TestHealthProgramUsesStringReader(t *testing.T) {
	home := newTestHome(t)
	var out strings.Builder
	// Decisions 3 and 5: 5 selects Health, and the sub-tab row marks the Home
	// view. The renderer writes only the changed cells of a row, so the footer
	// region is checked at the model by TestTabMovesTheAccentBorder.
	if err := RunHome(home, strings.NewReader("5r\x03\x03"), &out); err != nil {
		t.Fatalf("RunHome: %v", err)
	}
	for _, want := range []string{"[Home]", "capacity"} {
		if !strings.Contains(ansi.Strip(out.String()), want) {
			t.Fatalf("program output omitted %q:\n%s", want, ansi.Strip(out.String()))
		}
	}
}

func assertHealthLines(t *testing.T, model *Model, wants ...string) {
	t.Helper()
	view := ansi.Strip(model.View().Content)
	lines := strings.Split(view, "\n")
	for _, want := range wants {
		found := false
		for _, line := range lines {
			if strings.Contains(strings.TrimSpace(line), want) {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("health view omitted rendered line %q:\n%s", want, view)
		}
	}
}
