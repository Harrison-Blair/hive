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

func TestFileViewsMatchTerminalFrames(t *testing.T) {
	for _, size := range []struct {
		width  int
		height int
	}{{width: 120, height: 40}, {width: 200, height: 57}} {
		t.Run(fmt.Sprintf("%dx%d", size.width, size.height), func(t *testing.T) {
			for _, keyText := range []string{"b", "v", "d"} {
				t.Run("Work/"+keyText, func(t *testing.T) {
					home := writeTasksFixture(t)
					m := newTestModel(t, home, fixedClock(t, time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)))
					m.screenIndex = 2
					m, _ = raise(t, m, tea.WindowSizeMsg{Width: size.width, Height: size.height})
					cmd := m.startCurrentScreenLoad()
					if cmd == nil {
						t.Fatal("Work screen did not request data")
					}
					m, _ = raise(t, m, cmd())
					// Work rule 9: enter opens the task popup, and b, v,
					// and d act in it.
					m, _ = raise(t, m, key(tea.KeyEnter, ""))
					m, cmd = raise(t, m, key(rune(keyText[0]), keyText))
					if keyText != "d" && cmd == nil {
						t.Fatal(keyText + " did not request a file")
					}
					m = deliverAgentCommands(t, m, cmd)
					width, height := workTaskPopupWidth, workTaskPopupHeight
					if keyText == "d" {
						width, height = workDepsPopupWidth, workDepsPopupHeight
					}
					requireWorkFits(t, m, size.width, size.height, width, height)
					requireView(t, m, "esc closes")
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
	// Golden-frames decision 16: ctrl+t selects the Work board.
	if err := RunHome(home, strings.NewReader("\x14"), &output); err != nil {
		t.Fatal(err)
	}
	view := ansi.Strip(output.String())
	// The renderer writes only the changed cells of a row, so the test looks
	// for text that the first Work frame writes in full.
	for _, text := range []string{"work · g1-tui", "wave-c · done", "79-tui-shell"} {
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
	writeText(t, filepath.Join(evidence, "VERIFIED.md"), "# VERIFIED\n\n| Item | Command | Result |\n| --- | --- | --- |\n| V1 | go test ./internal/tui/... | pass |\n")
	writeText(t, filepath.Join(home, ".hive", "spec", "hive-tui", "tasks", "79-tui-shell", "brief.md"), "# TUI shell brief\n\n| ID | Criterion | Proven by |\n| --- | --- | --- |\n| AC1 | the shell loads | V1 |\n")
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
