package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/charmbracelet/x/ansi"
)

func TestLoadTasksFindsLongTaskID(t *testing.T) {
	home := newTestHome(t)
	const id = "96-tasks-evidence-view"
	writeRecord(t, home, []string{"tasks"}, id, records.Task{ID: id, Title: "Long task", Status: "created"})

	data, err := loadTasks(home)
	if err != nil {
		t.Fatalf("loadTasks() error = %v", err)
	}
	for _, item := range data.items {
		if item.record.ID == id {
			return
		}
	}
	t.Fatalf("loadTasks() omitted task %q", id)
}

func TestTasksScreenRendersLandedTaskAndVerificationHistory(t *testing.T) {
	home := writeTasksFixture(t)
	clock := fixedClock(t, time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC))
	m := newTestModel(t, home, clock)
	m.screenIndex = 4
	cmd := m.startCurrentScreenLoad()
	if cmd == nil {
		t.Fatal("Tasks screen did not request data")
	}
	m, _ = raise(t, m, cmd())
	// Decision 6: tab moves the region, so the test gives tab to the
	// current Tasks screen.
	m, _ = raise(t, m, key('x', "x"))
	for range 2 {
		m, cmd = pressScreenKey(t, m, key(0, "tab"))
		if cmd == nil {
			t.Fatal("tab change did not log the action")
		}
	}
	view := ansi.Strip(m.currentScreen().View(120, 36))
	for _, line := range []string{
		"active 2  queued 2  ▶ [landed 1]  failed 1  all 5",
		"▶ 79-tui-shell   TUI shell                   hive      landed     t79-worker-2 · pass 7204d69",
		"needs: 81-agents-screen ✓",
		"blocks: -",
		"branch: task/79-tui-shell",
		"feature: hive-tui",
		"grant: grants.md 2026-10-20 hive TUI goal",
		"needs user: -",
		"evidence: VERIFIED.md, V1.md",
		"V1 fail · V1 failed: missing shell load · 2026-10-20T10:00:00Z",
	} {
		if !strings.Contains(view, line) {
			t.Errorf("Tasks view does not contain line %q:\n%s", line, view)
		}
	}
	// Decision 8: from 120 columns enter goes to the list.
	m, _ = raise(t, m, tea.WindowSizeMsg{Width: 120, Height: 40})
	m, cmd = raise(t, m, key(0, "enter"))
	m = deliverAgentCommands(t, m, cmd)
	if got := ansi.Strip(m.currentScreen().View(120, 36)); !strings.Contains(got, "# VERIFIED") {
		t.Fatalf("enter did not open VERIFIED.md:\n%s", got)
	}
	m, _ = raise(t, m, key(0, "esc"))
	m, cmd = raise(t, m, key('b', "b"))
	m = deliverAgentCommands(t, m, cmd)
	if got := ansi.Strip(m.currentScreen().View(120, 36)); !strings.Contains(got, "# TUI shell brief") {
		t.Fatalf("b did not open the task brief:\n%s", got)
	}
}

func TestTasksScreenArrowDownSelectsNextTask(t *testing.T) {
	home := writeTasksFixture(t)
	writeRecord(t, home, []string{"tasks"}, "78-landed-task", records.Task{
		ID: "78-landed-task", Project: "hive", Title: "Earlier landed task", Status: "verified",
		Delivery: &records.Delivery{Commit: "1234567890abcdef", Pushed: true},
	})
	clock := fixedClock(t, time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC))
	m := newTestModel(t, home, clock)
	m.screenIndex = 4
	cmd := m.startCurrentScreenLoad()
	if cmd == nil {
		t.Fatal("Tasks screen did not request data")
	}
	m, _ = raise(t, m, cmd())
	// Decision 6: tab moves the region, so the test gives tab to the
	// current Tasks screen.
	for range 2 {
		m, _ = pressScreenKey(t, m, key(0, "tab"))
	}
	view := ansi.Strip(m.currentScreen().View(120, 36))
	if !strings.Contains(view, "▶ 78-landed-task") {
		t.Fatalf("initial view does not select the first landed task:\n%s", view)
	}
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyDown})
	view = ansi.Strip(m.currentScreen().View(120, 36))
	if !strings.Contains(view, "▶ 79-tui-shell") {
		t.Fatalf("arrow-down did not select the next landed task:\n%s", view)
	}
}

func TestTasksScreenShowsDetailsAtDefaultTerminalSize(t *testing.T) {
	home := writeTasksFixture(t)
	clock := fixedClock(t, time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC))
	m := newTestModel(t, home, clock)
	m.screenIndex = 4
	m, _ = raise(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	cmd := m.startCurrentScreenLoad()
	if cmd == nil {
		t.Fatal("Tasks screen did not request data")
	}
	m, _ = raise(t, m, cmd())
	// Decision 6: tab moves the region, so the test gives tab to the
	// current Tasks screen.
	for range 2 {
		m, _ = pressScreenKey(t, m, key(0, "tab"))
	}
	view := ansi.Strip(m.View().Content)
	for _, text := range []string{"79-tui-shell", "landed", "needs:", "branch:", "feature:", "grant:", "needs user:", "evidence:"} {
		if !strings.Contains(view, text) {
			t.Errorf("default-size Tasks view omits %q:\n%s", text, view)
		}
	}
}

func TestTasksProjectFilterAndFailedTab(t *testing.T) {
	home := writeTasksFixture(t)
	clock := fixedClock(t, time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC))
	m := newTestModel(t, home, clock)
	m.screenIndex = 4
	cmd := m.startCurrentScreenLoad()
	if cmd == nil {
		t.Fatal("Tasks screen did not request data")
	}
	m, _ = raise(t, m, cmd())
	// Decision 6: tab moves the region, so the test gives tab to the
	// current Tasks screen.
	for range 3 {
		m, _ = pressScreenKey(t, m, key(0, "tab"))
	}
	view := ansi.Strip(m.currentScreen().View(120, 36))
	for _, line := range []string{"[failed 1]", "82-failed-task", "V1 fail"} {
		if !strings.Contains(view, line) {
			t.Errorf("failed Tasks view does not contain %q:\n%s", line, view)
		}
	}
	m, _ = raise(t, m, key('p', "p"))
	view = ansi.Strip(m.currentScreen().View(120, 36))
	if !strings.Contains(view, "project: hive") || strings.Contains(view, "90-qmeter-task") {
		t.Fatalf("project filter did not select hive:\n%s", view)
	}
	cmd = m.startCurrentScreenLoad()
	if cmd == nil {
		t.Fatal("Tasks refresh did not request data")
	}
	m, _ = raise(t, m, cmd())
	view = ansi.Strip(m.currentScreen().View(120, 36))
	if !strings.Contains(view, "▶ 82-failed-task") {
		t.Fatalf("refresh did not preserve task selection:\n%s", view)
	}
}

func TestFileViewsMatchTerminalFrames(t *testing.T) {
	for _, size := range []struct {
		width  int
		height int
	}{{width: 120, height: 40}, {width: 200, height: 57}} {
		t.Run(fmt.Sprintf("%dx%d", size.width, size.height), func(t *testing.T) {
			for _, keyText := range []string{"b", "v", "d"} {
				t.Run("Tasks/"+keyText, func(t *testing.T) {
					home := writeTasksFixture(t)
					m := newTestModel(t, home, fixedClock(t, time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)))
					m.screenIndex = 4
					m, _ = raise(t, m, tea.WindowSizeMsg{Width: size.width, Height: size.height})
					cmd := m.startCurrentScreenLoad()
					if cmd == nil {
						t.Fatal("Tasks screen did not request data")
					}
					m, _ = raise(t, m, cmd())
					// Decision 6: tab moves the region, so the test gives
					// tab to the current Tasks screen.
					for range 2 {
						m, _ = pressScreenKey(t, m, key(0, "tab"))
					}
					m, cmd = raise(t, m, key(rune(keyText[0]), keyText))
					if keyText != "d" && cmd == nil {
						t.Fatal(keyText + " did not request a file")
					}
					m = deliverAgentCommands(t, m, cmd)
					assertModelFrameSize(t, m, size.width, size.height)
				})
			}

			t.Run("Friction/file", func(t *testing.T) {
				home := newTestHome(t)
				writeText(t, filepath.Join(home, "reference", "dogfood", "friction", "FRICTION.md"), "# Friction\n")
				writeText(t, filepath.Join(home, "reference", "dogfood", "friction", "entries", "selected.md"), "---\ntitle: \"Selected issue\"\nstatus: \"Open\"\nopened: 2026-10-19\ndecision: \"Keep this decision.\"\n---\n\n**Summary:**\nA selected issue.\n")
				m := newTestModel(t, home, fixedClock(t, time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)))
				m.screenIndex = 5
				m, _ = raise(t, m, tea.WindowSizeMsg{Width: size.width, Height: size.height})
				cmd := m.startCurrentScreenLoad()
				if cmd == nil {
					t.Fatal("Friction screen did not request data")
				}
				m, _ = raise(t, m, cmd())
				m, cmd = raise(t, m, key(0, "enter"))
				if cmd == nil {
					t.Fatal("enter did not request the selected entry")
				}
				m = deliverAgentCommands(t, m, cmd)
				assertModelFrameSize(t, m, size.width, size.height)
			})
		})
	}
}

func assertModelFrameSize(t *testing.T, m *Model, width, height int) {
	t.Helper()
	lines := strings.Split(ansi.Strip(m.View().Content), "\n")
	if len(lines) != height {
		t.Errorf("frame has %d lines, want %d:\n%s", len(lines), height, strings.Join(lines, "\n"))
	}
	for index, line := range lines {
		if got := ansi.StringWidth(line); got != width {
			t.Errorf("frame line %d has display width %d, want %d: %q", index+1, got, width, line)
		}
	}
	// Decision 1: the footer is one row, below the content area.
	if !strings.HasSuffix(lines[height-2], "╯") {
		t.Errorf("main panel bottom border is not on line %d: %q", height-1, lines[height-2])
	}
}

func TestFrictionScreenShowsSelectedDetailsInRightPanel(t *testing.T) {
	for _, size := range []struct {
		width  int
		height int
	}{{width: 120, height: 40}, {width: 200, height: 57}} {
		t.Run(fmt.Sprintf("%dx%d", size.width, size.height), func(t *testing.T) {
			home := newTestHome(t)
			writeText(t, filepath.Join(home, "reference", "dogfood", "friction", "FRICTION.md"), "# Friction\n")
			writeText(t, filepath.Join(home, "reference", "dogfood", "friction", "entries", "selected.md"), "---\ntitle: \"Selected issue detail\"\nstatus: \"Open\"\nopened: 2026-10-19\ndecision: \"Keep this decision.\"\n---\n\n**Summary:**\nThe selected issue summary.\n")
			m := newTestModel(t, home, fixedClock(t, time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)))
			m.selectScreen(tabFriction, 0)
			m, _ = raise(t, m, tea.WindowSizeMsg{Width: size.width, Height: size.height})
			cmd := m.startCurrentScreenLoad()
			if cmd == nil {
				t.Fatal("Friction screen did not request data")
			}
			m, _ = raise(t, m, cmd())

			view := ansi.Strip(m.View().Content)
			listColumn, detailColumn := -1, -1
			for _, line := range strings.Split(view, "\n") {
				if index := strings.Index(line, "ENTRY"); index >= 0 && strings.Contains(line, "OPENED") {
					listColumn = ansi.StringWidth(line[:index])
				}
				if index := strings.Index(line, "Selected issue detail"); index >= 0 {
					detailColumn = ansi.StringWidth(line[:index])
				}
			}
			if listColumn < 0 {
				t.Fatalf("Friction list header is missing:\n%s", view)
			}
			// Decision 8: the details panel starts after the list, at 13/30
			// of the width.
			if detailColumn <= listColumn || detailColumn < size.width*13/30 {
				t.Fatalf("selected entry details are not in a right-hand panel (list column %d, detail column %d):\n%s", listColumn, detailColumn, view)
			}
			if !strings.Contains(view, "▶ selected") || !strings.Contains(view, "The selected issue summary.") {
				t.Fatalf("Friction view omits the selected entry or its summary:\n%s", view)
			}
		})
	}
}

func TestFrictionScreenLoadsEntriesAndOpensSelectedEntry(t *testing.T) {
	home := newTestHome(t)
	writeText(t, filepath.Join(home, "reference", "dogfood", "friction", "FRICTION.md"), "# Friction\n")
	writeText(t, filepath.Join(home, "reference", "dogfood", "friction", "entries", "open-one.md"), "---\ntitle: \"Open issue\"\nstatus: \"Open\"\nopened: 2026-10-19\ndecision: \"\"\n---\n\n**Summary:**\nA missing dashboard.\n\n**Reproduction steps:**\n1. Start hive.\n")
	writeText(t, filepath.Join(home, "reference", "dogfood", "friction", "entries", "invalid.md"), "No front matter\n")
	writeText(t, filepath.Join(home, "reference", "dogfood", "friction", "entries", "resolved", "fixed.md"), "---\ntitle: \"Fixed issue\"\nstatus: \"Resolved 2026-10-20\"\nopened: 2026-10-18\ndecision: \"Fixed in task 9.\"\n---\n\n**Summary:**\nThe fix works.\n")
	writeText(t, filepath.Join(home, "reference", "dogfood", "friction", "entries", "resolved", "newer.md"), "---\ntitle: \"Newest fixed issue\"\nstatus: \"Resolved 2026-10-21\"\nopened: 2026-10-19\ndecision: \"Fixed today.\"\n---\n\n**Summary:**\nThe newest fix works.\n")
	clock := fixedClock(t, time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC))
	m := newTestModel(t, home, clock)
	m.selectScreen(tabFriction, 0)
	m, _ = raise(t, m, tea.WindowSizeMsg{Width: 120, Height: 40})
	cmd := m.startCurrentScreenLoad()
	if cmd == nil {
		t.Fatal("Friction screen did not request data")
	}
	m, _ = raise(t, m, cmd())
	view := ansi.Strip(m.View().Content)
	for _, line := range []string{
		"open · 2",
		"▶ open-one",
		"  invalid ",
		"Open issue",
		"SUMMARY",
		"A missing dashboard.",
		"REPRODUCTION STEPS",
		"1. Start hive.",
	} {
		if !strings.Contains(view, line) {
			t.Errorf("Friction view does not contain line %q:\n%s", line, view)
		}
	}
	// An entry with no front matter shows ? in its fields.
	m, _ = raise(t, m, key(0, "down"))
	if view := ansi.Strip(m.View().Content); !strings.Contains(view, "? invalid.md") {
		t.Errorf("invalid entry details omit \"? invalid.md\":\n%s", view)
	}
	m, _ = raise(t, m, key(0, "up"))
	// Decision 9: below 120 columns enter opens the details overlay.
	m, _ = raise(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	m, _ = raise(t, m, key(0, "enter"))
	smallView := ansi.Strip(m.View().Content)
	for _, text := range []string{"Open issue", "status    Open", "file      entries/open-one.md", "SUMMARY", "REPRODUCTION STEPS"} {
		if !strings.Contains(smallView, text) {
			t.Errorf("default-size Friction overlay omits %q:\n%s", text, smallView)
		}
	}
	m, _ = raise(t, m, key(0, "esc"))
	// From 120 columns enter opens the entry file in the details region.
	m, _ = raise(t, m, tea.WindowSizeMsg{Width: 120, Height: 40})
	m, cmd = raise(t, m, key(0, "enter"))
	if cmd == nil {
		t.Fatal("enter did not request the selected entry")
	}
	m = deliverAgentCommands(t, m, cmd)
	if view := ansi.Strip(m.View().Content); !strings.Contains(view, "**Reproduction steps:**") {
		t.Fatalf("entry reader did not show the full file:\n%s", view)
	}
	m, _ = raise(t, m, key(0, "esc"))
	if strings.Contains(ansi.Strip(m.View().Content), "**Reproduction steps:**") {
		t.Fatal("esc did not close the entry reader")
	}
	// Decision 22: ] selects the next view of the Friction tab. The newest
	// resolved date is first.
	m, cmd = raise(t, m, key(']', "]"))
	if cmd == nil {
		t.Fatal("] did not log the friction category change")
	}
	m = deliverAgentCommands(t, m, cmd)
	view = ansi.Strip(m.View().Content)
	if !strings.Contains(view, "resolved · 2") || !strings.Contains(view, "▶ newer ") || !strings.Contains(view, "  fixed ") {
		t.Fatalf("resolved entries are missing or not sorted newest first:\n%s", view)
	}
	m, _ = raise(t, m, key(0, "down"))
	cmd = m.startCurrentScreenLoad()
	if cmd == nil {
		t.Fatal("Friction refresh did not request data")
	}
	m, _ = raise(t, m, cmd())
	if view := ansi.Strip(m.View().Content); !strings.Contains(view, "▶ fixed ") {
		t.Fatalf("refresh did not preserve friction selection:\n%s", view)
	}
}

func TestRunHomeShowsTasksFromStringReader(t *testing.T) {
	home := writeTasksFixture(t)
	var output strings.Builder
	// Decisions 5 and 22: 3 selects Work, and ] ] select its Tasks view.
	if err := RunHome(home, strings.NewReader("3]]"), &output); err != nil {
		t.Fatal(err)
	}
	view := ansi.Strip(output.String())
	// The renderer writes only the changed cells of a row, so the test looks
	// for text that the first Tasks frame writes in full.
	for _, text := range []string{"[Tasks]", "81-agents-scr", "79-tui-shell"} {
		if !strings.Contains(view, text) {
			t.Fatalf("program output does not contain %q:\n%s", text, view)
		}
	}
}

func writeTasksFixture(t *testing.T) string {
	t.Helper()
	home := newTestHome(t)
	writeRecord(t, home, []string{"projects"}, "hive", records.Project{ID: "hive", Path: home, Base: "dev"})
	writeRecord(t, home, []string{"projects"}, "qmeter", records.Project{ID: "qmeter", Path: filepath.Join(home, "qmeter"), Base: "dev"})
	writeRecord(t, home, []string{"goals"}, "g1-tui", records.Goal{
		ID: "g1-tui", Project: "hive", Title: "hive TUI", Grant: "grants.md 2026-10-20 hive TUI goal", Status: "active",
		Waves: []records.GoalWave{{ID: "wave-c", Tasks: []string{"79-tui-shell"}}},
	})
	writeRecord(t, home, []string{"tasks"}, "79-tui-shell", records.Task{
		ID: "79-tui-shell", Project: "hive", Title: "TUI shell", Feature: "hive-tui", Status: "verified", ActingAgent: "t79-worker-2",
		DependsOn: []string{"81-agents-screen"}, Verification: &records.Verification{Agent: "t79-verifier-2", Result: "pass", At: "2026-10-20T11:00:00Z"},
		VerificationHistory: []records.VerificationHistory{{Verification: &records.Verification{Agent: "t79-verifier-1", Result: "fail", Reason: "V1 failed: missing shell load", At: "2026-10-20T10:00:00Z"}, Reason: "V1 failed: missing shell load", At: "2026-10-20T10:00:00Z"}},
		Worktree:            &records.Worktree{Branch: "task/79-tui-shell"},
		Delivery:            &records.Delivery{Commit: "7204d6997a64daf81b4ca6eba98c462513fa299f", Pushed: true, Grant: "grants.md 2026-10-20 hive TUI goal"},
	})
	writeRecord(t, home, []string{"tasks"}, "80-queued-task", records.Task{ID: "80-queued-task", Project: "hive", Title: "Queued task", Status: "created"})
	writeRecord(t, home, []string{"tasks"}, "82-failed-task", records.Task{ID: "82-failed-task", Project: "hive", Title: "Failed task", Status: "assigned", ActingAgent: "worker-fail", Verification: &records.Verification{Agent: "verifier-fail", Result: "fail", Reason: "V1 fail: assertion", At: "2026-10-20T09:00:00Z"}})
	writeRecord(t, home, []string{"tasks"}, "90-qmeter-task", records.Task{ID: "90-qmeter-task", Project: "qmeter", Title: "Qmeter task", Status: "created"})
	writeRecord(t, home, []string{"tasks"}, "81-agents-screen", records.Task{ID: "81-agents-screen", Project: "hive", Title: "Agents screen", Status: "verified", Verification: &records.Verification{Agent: "t81-verifier-1", Result: "pass"}})
	evidence := filepath.Join(home, ".hive", "spec", "hive-tui", "tasks", "79-tui-shell", "evidence")
	writeText(t, filepath.Join(evidence, "V1.md"), "# V1\n")
	writeText(t, filepath.Join(evidence, "VERIFIED.md"), "# VERIFIED\n")
	writeText(t, filepath.Join(home, ".hive", "spec", "hive-tui", "tasks", "79-tui-shell", "brief.md"), "# TUI shell brief\n")
	return home
}

func writeText(t *testing.T, path, text string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}
