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

// loadWorkTasks opens the Tasks view in normal mode at width×height and
// loads the records of home (decision 38).
func loadWorkTasks(t *testing.T, home string, width, height int) *Model {
	t.Helper()
	m := newTestModel(t, home, fixedClock(t, time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)))
	m, _ = raise(t, m, tea.WindowSizeMsg{Width: width, Height: height})
	m, cmd := openWorkView(t, m, 2)
	return deliverAgentCommands(t, m, cmd)
}

// AC3: normal mode reads the task, goal, project, and agent records and the
// files of the task, and the same renderer shows them.
func TestWorkTasksNormalModeShowsRecordDetails(t *testing.T) {
	home := writeTasksFixture(t)
	writeRecord(t, home, []string{"agents"}, "t79-worker-2", records.Agent{Key: "t79-worker-2", Name: "t79-worker-2", Role: "worker", Status: "stopped"})
	m := loadWorkTasks(t, home, 120, 40)
	view := ansi.Strip(m.View().Content)
	for _, line := range []string{
		"tasks · 5",
		"▶ 79-tui-shell              landed    wave-c",
		"80-queued-task            created   —",
		"90-qmeter-task",
		"title     TUI shell",
		"status    landed",
		"goal      g1-tui · wave wave-c",
		"worker    t79-worker-2 · stopped",
		"verifier  t79-verifier-2 · pass",
		"branch    task/79-tui-shell",
		"merge     7204d699 ",
		"depends   81-agents-screen",
		"✓ AC1 the shell loads",
		"V1 go test ./internal/tui/...",
	} {
		if !strings.Contains(view, line) {
			t.Errorf("Tasks view does not contain %q:\n%s", line, view)
		}
	}
	m, cmd := raise(t, m, key('b', "b"))
	m = deliverAgentCommands(t, m, cmd)
	if got := ansi.Strip(m.View().Content); !strings.Contains(got, "╭─ brief.md ─") || !strings.Contains(got, "# TUI shell brief") {
		t.Fatalf("b did not open the task brief in the details:\n%s", got)
	}
	m, _ = raise(t, m, key(0, "esc"))
	if got := ansi.Strip(m.View().Content); strings.Contains(got, "# TUI shell brief") || !strings.Contains(got, "title     TUI shell") {
		t.Fatalf("esc did not close the brief:\n%s", got)
	}
	m, cmd = raise(t, m, key('v', "v"))
	m = deliverAgentCommands(t, m, cmd)
	if got := ansi.Strip(m.View().Content); !strings.Contains(got, "# VERIFIED") {
		t.Fatalf("v did not open VERIFIED.md:\n%s", got)
	}
}

// AC4: ↓ selects the next task and its details show.
func TestWorkTasksArrowDownSelectsNextTask(t *testing.T) {
	home := writeTasksFixture(t)
	writeRecord(t, home, []string{"tasks"}, "78-landed-task", records.Task{
		ID: "78-landed-task", Project: "hive", Title: "Earlier landed task", Status: "verified",
		Delivery: &records.Delivery{Commit: "1234567890abcdef", Pushed: true},
	})
	m := loadWorkTasks(t, home, 120, 40)
	if view := ansi.Strip(m.View().Content); !strings.Contains(view, "▶ 78-landed-task") {
		t.Fatalf("initial view does not select the first task:\n%s", view)
	}
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyDown})
	view := ansi.Strip(m.View().Content)
	if !strings.Contains(view, "▶ 79-tui-shell") || !strings.Contains(view, "╭─ 79-tui-shell ─") {
		t.Fatalf("arrow-down did not select the next task and show its details:\n%s", view)
	}
}

// AC1: at 80×24 the list fills the content area, and enter opens the
// details overlay with the short form of the Compact board.
func TestWorkTasksShowsDetailsOverlayAtDefaultTerminalSize(t *testing.T) {
	m := loadWorkTasks(t, writeTasksFixture(t), 80, 24)
	if view := ansi.Strip(m.View().Content); !strings.Contains(view, "▶ 79-tui-shell") || strings.Contains(view, "ACCEPTANCE") {
		t.Fatalf("80x24 Tasks view does not show the list alone:\n%s", view)
	}
	m, _ = raise(t, m, key(tea.KeyEnter, "enter"))
	view := ansi.Strip(m.View().Content)
	for _, text := range []string{"╭─ 79-tui-shell · esc back", "list › details", "status    landed", "goal      g1-tui · wave wave-c", "verifier  t79-verifier-2 · pass", "merge     7204d699", "ACCEPTANCE", "EVIDENCE"} {
		if !strings.Contains(view, text) {
			t.Errorf("80x24 overlay omits %q:\n%s", text, view)
		}
	}
	if m.focusState() != "details" {
		t.Errorf("focus = %q; want details", m.focusState())
	}
	if strings.Contains(view, "title     ") || strings.Contains(view, "branch    ") {
		t.Errorf("80x24 overlay shows a field of the wide form:\n%s", view)
	}
	m, _ = raise(t, m, key(0, "esc"))
	if view := ansi.Strip(m.View().Content); strings.Contains(view, "ACCEPTANCE") || m.focusState() != "list" {
		t.Fatalf("esc did not close the overlay:\n%s", view)
	}
}

// AC3: a refresh keeps the selected task, and the list has no project
// filter.
func TestWorkTasksRefreshKeepsSelection(t *testing.T) {
	m := loadWorkTasks(t, writeTasksFixture(t), 120, 40)
	for range 3 {
		m, _ = raise(t, m, key('j', "j"))
	}
	if m.tasks.selectedID != "82-failed-task" {
		t.Fatalf("j j j selected %q, want 82-failed-task", m.tasks.selectedID)
	}
	m, cmd := raise(t, m, key('R', "R"))
	m = deliverAgentCommands(t, m, cmd)
	view := ansi.Strip(m.View().Content)
	if !strings.Contains(view, "▶ 82-failed-task") || !strings.Contains(view, "verifier  verifier-fail · fail") || !strings.Contains(view, "90-qmeter-task") {
		t.Fatalf("refresh did not keep the selection with each project listed:\n%s", view)
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

func TestRunHomeShowsTasksFromStringReader(t *testing.T) {
	home := writeTasksFixture(t)
	var output strings.Builder
	// Decisions 5 and 22: 3 selects Work, and ] ] select its Tasks view.
	if err := RunHome(home, strings.NewReader("\x14]]"), &output); err != nil {
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
