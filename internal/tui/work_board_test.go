package tui

import (
	"bytes"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/Harrison-Blair/hive/internal/lib/cellgrid"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/tui/component"
)

// workTestNow is the clock of the Work board tests.
var workTestNow = time.Date(2026, 10, 23, 14, 22, 0, 0, time.UTC)

// workTestData gives two goals and one task with no goal. Goal g1 is older
// and has a stop. Goal g2 has the waves A, B, and C with one task of each
// card form of work.md rule 4.
func workTestData() demoData {
	landed := func(commit string) *records.Delivery {
		return &records.Delivery{Commit: commit, Pushed: true, Target: "dev"}
	}
	task := func(id, status string) records.Task {
		return records.Task{ID: id, Project: "hive", Title: "Task " + id, Feature: "board", Status: status}
	}
	a1, a2, a3 := task("1-a", "landed"), task("2-b", "landed"), task("3-c", "cancelled")
	a1.Delivery, a2.Delivery = landed("abcdef1234"), landed("123456789a")
	a3.CancelReason = "replaced"
	b4, b5, b6, b7 := task("4-d", "verified"), task("5-e", "assigned"), task("6-f", "created"), task("7-g", "created")
	b4.DependsOn = []string{"1-a"}
	b4.Verification = &records.Verification{Agent: "t4-verifier-1", Result: "pass", Reason: "all items pass"}
	b5.DependsOn = []string{"2-b"}
	b5.ActingAgent = "t5-worker-1"
	b5.Worktree = &records.Worktree{Branch: "task/5-e"}
	b6.NeedsUser = "mask ok?"
	b7.VerificationHistory = []records.VerificationHistory{{Verification: &records.Verification{Agent: "t7-verifier-1", Result: "fail", Reason: "V4 shows a diff"}}}
	c8, c9 := task("8-h", "created"), task("9-i", "created")
	c8.DependsOn = []string{"5-e", "6-f"}
	c9.Verification = &records.Verification{Agent: "t9-verifier-1", Result: "fail", Reason: "the gate fails"}
	old := task("20-x", "assigned")
	old.ActingAgent = "t20-worker-1"
	loose := task("30-z", "created")
	return demoData{
		Projects: []records.Project{{ID: "hive", Path: "/home/hive"}},
		Agents: []records.Agent{
			{Key: "t5-worker-1", Name: "t5-worker-1", Profile: records.ProfileBuilder, Status: "live", AgentStatus: "working", StatusChangedAt: "2026-10-23T14:10:00Z"},
			{Key: "t20-worker-1", Name: "t20-worker-1", Profile: records.ProfileBuilder, Status: "unresolved"},
		},
		Goals: []records.Goal{
			{ID: "g2-board", Project: "hive", Title: "The board", Grant: "grant line", Status: "active", CreatedAt: "2026-10-22T10:00:00Z", Waves: []records.GoalWave{
				{ID: "A", DoneWhen: "dev", Tasks: []string{"1-a", "2-b", "3-c"}},
				{ID: "B", DoneWhen: "dev", Tasks: []string{"4-d", "5-e", "6-f", "7-g"}},
				{ID: "C", DoneWhen: "main", Tasks: []string{"8-h", "9-i"}},
			}},
			{ID: "g1-old", Project: "hive", Title: "The old goal", Status: "active", CreatedAt: "2026-10-01T10:00:00Z", Waves: []records.GoalWave{
				{ID: "A", DoneWhen: "dev", Tasks: []string{"20-x"}},
			}},
		},
		Tasks: []records.Task{a1, a2, a3, b4, b5, b6, b7, c8, c9, old, loose},
		Files: map[string]string{
			".hive/spec/board/tasks/4-d/brief.md": "# 4-d\n\n### V1 Tests\n\n### V3 Live check\n",
			".hive/spec/board/tasks/5-e/brief.md": "# 5-e\n\n| ID | Criterion | Items |\n| --- | --- | --- |\n| AC1 | The card shows | V1 |\n",
		},
	}
}

// newWorkModel gives a demo model of the Work screen with data and ui at
// width×height cells.
func newWorkModel(t *testing.T, data demoData, ui demoUIState, width, height int) *Model {
	t.Helper()
	ui.Screen = "work"
	fixture := &demoFixture{Terminal: demoSize{Width: width, Height: height}, Clock: demoClock{Time: "2026-10-23T14:22:00Z", Zone: "UTC"}, Home: "~/hive", Data: data, UI: ui, now: workTestNow}
	m := newDemoModel(t.TempDir(), &demoState{fixture: fixture})
	t.Cleanup(m.Stop)
	return m
}

// workRows gives the rows of the view without styles.
func workRows(m *Model) []string {
	return strings.Split(ansi.Strip(m.View().Content), "\n")
}

// pressWork sends one key to the model and runs its command.
func pressWork(t *testing.T, m *Model, press tea.KeyPressMsg) *Model {
	t.Helper()
	m, cmd := raise(t, m, press)
	return runWork(t, m, cmd)
}

// runWork runs cmd and each command of a batch, and gives each message to
// the model.
func runWork(t *testing.T, m *Model, cmd tea.Cmd) *Model {
	t.Helper()
	if cmd == nil {
		return m
	}
	switch msg := cmd().(type) {
	case nil:
	case tea.BatchMsg:
		for _, inner := range msg {
			m = runWork(t, m, inner)
		}
	default:
		m, _ = raise(t, m, msg)
	}
	return m
}

// requireWorkFits checks that the view fills width×height cells and that
// no box border breaks. The rail gaps of rule 5 hold ╭ and ╯ of the rails,
// so the check blanks them first. With popupWidth > 0, a popup of
// popupWidth×popupHeight covers the middle of the board, so the check reads
// only the popup cells.
func requireWorkFits(t *testing.T, m *Model, width, height, popupWidth, popupHeight int) {
	t.Helper()
	lines := strings.Split(ansi.Strip(m.View().Content), "\n")
	if len(lines) != height {
		t.Fatalf("view has %d lines, want %d", len(lines), height)
	}
	for index, line := range lines {
		if got := ansi.StringWidth(line); got != width {
			t.Fatalf("line %d has width %d, want %d: %q", index+1, got, width, line)
		}
		cells := []rune(line)
		if popupWidth > 0 {
			// The status row is below the dimmed area.
			w, h := min(popupWidth, width), min(popupHeight, height-1)
			x, y := (width-w)/2, (height-1-h)/2
			lines[index] = ""
			if index >= y && index < y+h {
				lines[index] = string(cells[x : x+w])
			}
			continue
		}
		for gap := 29; gap+1 < len(cells); gap += component.WaveColumnWidth + 2 {
			for _, col := range []int{gap, gap + 1} {
				if strings.ContainsRune("╭╯╮╰", cells[col]) {
					cells[col] = ' '
				}
			}
		}
		lines[index] = string(cells)
	}
	requireWholeBorders(t, strings.Join(lines, "\n"))
}

func arrowKey(code rune) tea.KeyPressMsg { return tea.KeyPressMsg{Code: code} }

// selectedCardPattern finds a card with the heavy border: the border, the
// dot, the ID, and the border again.
var selectedCardPattern = regexp.MustCompile(`┃ [●◐○!×] (\S+) +┃`)

// selectedCard gives the ID of the card with the heavy border.
func selectedCard(t *testing.T, m *Model) string {
	t.Helper()
	for _, row := range workRows(m) {
		if match := selectedCardPattern.FindStringSubmatch(row); match != nil {
			return match[1]
		}
	}
	t.Fatalf("no selected card:\n%s", strings.Join(workRows(m), "\n"))
	return ""
}

// Rule 4: each card shows the line 2 form of its state. waits on names each
// dependency by the number of its ID, as the boards show (GfWork row 7).
func TestWorkCardForms(t *testing.T) {
	m := newWorkModel(t, workTestData(), demoUIState{Selection: "1-a"}, 120, 40)
	requireView(t, m,
		"● 1-a", "landed · abcdef",
		"◐ 4-d", "verified · V3 pass",
		"○ 5-e", "t5-worke… · 00:12:00",
		"! 6-f", "needs you · mask ok?",
		"○ 7-g", "verify fail · V4",
		"○ 8-h", "waits on 5, 6 ",
		"○ 9-i", "verify fail ",
	)
	rejectView(t, m, "3-c")
}

// Rule 4 and tui-v2 decision 60: line 2 of verify fail is accent, and the
// card of a working worker has the working border.
func TestWorkCardTokens(t *testing.T) {
	m := newWorkModel(t, workTestData(), demoUIState{Selection: "1-a"}, 120, 40)
	grid, err := cellgrid.Convert(m.View().Content, 120, 40)
	if err != nil {
		t.Fatal(err)
	}
	rows := workRows(m)
	find := func(text string) (int, int) {
		for r, row := range rows {
			if c := strings.Index(row, text); c >= 0 {
				return r, ansi.StringWidth(row[:c])
			}
		}
		t.Fatalf("no %q", text)
		return 0, 0
	}
	r, c := find("verify fail · V4")
	if got := grid.At(r, c).Token; got != cellgrid.Accent {
		t.Errorf("verify fail token %v, want accent", got)
	}
	r, c = find("○ 5-e")
	if got := grid.At(r, c-2).Token; got != cellgrid.Working {
		t.Errorf("border of the card of a working worker %v, want working", got)
	}
	r, c = find("! 6-f")
	if got := grid.At(r, c).Token; got != cellgrid.Accent {
		t.Errorf("stop dot token %v, want accent", got)
	}
}

// Rules 1 to 3 and the layout: the title row, the wave line, the column
// titles, and the feet.
func TestWorkTitleWaveLineAndColumns(t *testing.T) {
	m := newWorkModel(t, workTestData(), demoUIState{Selection: "5-e"}, 120, 40)
	rows := workRows(m)
	if want := " work · g2-board · The board"; !strings.HasPrefix(rows[0], want) || !strings.HasSuffix(rows[0], "[ ] goal 2 of 3 ") {
		t.Errorf("title row %q", rows[0])
	}
	if want := " wave B of 3 · 2/8 landed · 1 needs you · grant grant line  · 1 stops elsewhere"; strings.TrimRight(rows[1], " ") != want {
		t.Errorf("wave line %q, want %q", rows[1], want)
	}
	requireView(t, m, "╭─ A · done · 2/2 ─", "┏━ B · running · 0/4 ━", "╭─ C · planned · 0/2 ─", "done when dev ✓ done", "done when main", "h shows 1 cancelled")
	if !strings.HasPrefix(rows[38], " ● landed  ◐ verified  ○ assigned or created  ! needs you  × cancelled  ─▶ depends on") {
		t.Errorf("legend %q", rows[38])
	}
}

// Rule 5: a dependency between neighbouring columns draws a rail to the
// name line of the second card.
func TestWorkRails(t *testing.T) {
	m := newWorkModel(t, workTestData(), demoUIState{Selection: "1-a"}, 120, 40)
	rows := workRows(m)
	cells := func(row, col, n int) string { return ansi.Cut(rows[row], col, col+n) }
	// 1-a (row 6) to 4-d (row 6), 2-b (row 11) to 5-e (row 11).
	for _, row := range []int{6, 11} {
		// Column A holds the selected card, so its rail leaves from ┨.
		if got := cells(row, 28, 5); got != "┨──├▶" {
			t.Errorf("row %d columns 28 to 32 %q, want ┨──├▶", row, got)
		}
	}
	// 5-e (row 11) and 6-f (row 16) to 8-h (row 6).
	want := map[int]string{6: "│ ╭├▶", 7: "│ ││ ", 11: "┤─┤│ ", 15: "│ ││ ", 16: "┤─╯│ ", 17: "│  │ "}
	for row, text := range want {
		if got := cells(row, 58, 5); got != text {
			t.Errorf("row %d columns 58 to 62 %q, want %q", row, got, text)
		}
	}
}

// Rule 8: the arrow keys and g and G select cards.
func TestWorkArrowKeys(t *testing.T) {
	m := newWorkModel(t, workTestData(), demoUIState{Selection: "2-b"}, 120, 40)
	steps := []struct {
		press tea.KeyPressMsg
		want  string
	}{
		{arrowKey(tea.KeyRight), "5-e"},
		{arrowKey(tea.KeyDown), "6-f"},
		{arrowKey(tea.KeyDown), "7-g"},
		{arrowKey(tea.KeyDown), "7-g"},
		{arrowKey(tea.KeyRight), "9-i"},
		{arrowKey(tea.KeyRight), "9-i"},
		{arrowKey(tea.KeyLeft), "5-e"},
		{key('g', "g"), "4-d"},
		{key('G', "G"), "7-g"},
		{arrowKey(tea.KeyUp), "6-f"},
		{arrowKey(tea.KeyLeft), "2-b"},
	}
	for index, step := range steps {
		m = pressWork(t, m, step.press)
		if got := selectedCard(t, m); got != step.want {
			t.Fatalf("step %d (%s) selected %q, want %q", index, step.press.String(), got, step.want)
		}
	}
}

// Rule 1: [ and ] cycle the goals in creation order, with no goal last.
func TestWorkGoalCycle(t *testing.T) {
	m := newWorkModel(t, workTestData(), demoUIState{Selection: "5-e"}, 120, 40)
	for _, want := range []string{" work · no goal · tasks outside every goal", " work · g1-old · The old goal", " work · g2-board · The board", " work · no goal"} {
		m = pressWork(t, m, key(']', "]"))
		if rows := workRows(m); !strings.HasPrefix(rows[0], want) {
			t.Fatalf("after ] title row %q, want %q", rows[0], want)
		}
	}
	requireView(t, m, "┏━ no goal · 1 ━", "1 task · 0 stops · no grant", "no done when", "○ 30-z")
	m = pressWork(t, m, key('[', "["))
	requireView(t, m, " work · g2-board")
	// With a goal for each task, no goal is not in the cycle.
	data := workTestData()
	data.Tasks = data.Tasks[:len(data.Tasks)-1]
	m = newWorkModel(t, data, demoUIState{Selection: "5-e"}, 120, 40)
	if rows := workRows(m); !strings.HasSuffix(rows[0], "[ ] goal 2 of 2 ") {
		t.Fatalf("title row %q, want goal 2 of 2", rows[0])
	}
}

// Rule 11: s shows only the stop cards.
func TestWorkStopsOnly(t *testing.T) {
	m := newWorkModel(t, workTestData(), demoUIState{Selection: "5-e"}, 120, 40)
	m = pressWork(t, m, key('s', "s"))
	rows := workRows(m)
	if !strings.Contains(rows[0], "The board  · stops only") {
		t.Errorf("title row %q, want · stops only", rows[0])
	}
	requireView(t, m, "! 6-f")
	rejectView(t, m, "4-d", "5-e", "1-a")
	if got := selectedCard(t, m); got != "6-f" {
		t.Errorf("selected %q, want the stop 6-f", got)
	}
	m = pressWork(t, m, key('s', "s"))
	requireView(t, m, "◐ 4-d", "○ 5-e")
}

// Rule 12: h shows the cancelled cards at the end of their column.
func TestWorkCancelled(t *testing.T) {
	m := newWorkModel(t, workTestData(), demoUIState{Selection: "5-e"}, 120, 40)
	m = pressWork(t, m, key('h', "h"))
	requireView(t, m, "× 3-c", "cancelled · replaced", "╭─ A · done · 2/2 ─")
	rejectView(t, m, "h shows 1 cancelled")
	m = pressWork(t, m, key('h', "h"))
	rejectView(t, m, "3-c")
}

// Rules 6 and 7: the board and a column scroll only as far as the selection
// needs.
func TestWorkScroll(t *testing.T) {
	data := workTestData()
	goal := &data.Goals[0]
	for _, id := range []string{"D", "E", "F"} {
		task := records.Task{ID: "40-" + strings.ToLower(id), Project: "hive", Status: "created"}
		data.Tasks = append(data.Tasks, task)
		goal.Waves = append(goal.Waves, records.GoalWave{ID: id, DoneWhen: "dev", Tasks: []string{task.ID}})
	}
	for index := range 5 {
		task := records.Task{ID: string(rune('a'+index)) + "-extra", Project: "hive", Status: "created"}
		data.Tasks = append(data.Tasks, task)
		goal.Waves[1].Tasks = append(goal.Waves[1].Tasks, task.ID)
	}
	m := newWorkModel(t, data, demoUIState{Selection: "9-i"}, 120, 40)
	rows := workRows(m)
	if !strings.HasPrefix(rows[36], " ←→ at the edge scrolls · waves A–D of 6") || !strings.Contains(rows[36], "━") {
		t.Fatalf("scroll row %q", rows[36])
	}
	for range 3 {
		m = pressWork(t, m, arrowKey(tea.KeyRight))
	}
	if rows = workRows(m); !strings.HasPrefix(rows[36], " ←→ at the edge scrolls · waves C–F of 6") {
		t.Fatalf("after → → → scroll row %q, want waves C–F", rows[36])
	}
	m = pressWork(t, m, arrowKey(tea.KeyLeft))
	if rows = workRows(m); !strings.HasPrefix(rows[36], " ←→ at the edge scrolls · waves C–F of 6") {
		t.Fatalf("after ← scroll row %q, want no move", rows[36])
	}
	// Column B has 9 cards and shows 6. G moves its top by 3.
	for range 3 {
		m = pressWork(t, m, arrowKey(tea.KeyLeft))
	}
	m = pressWork(t, m, key('G', "G"))
	requireView(t, m, "○ e-extra", "○ 7-g", "waves B–E of 6")
	rejectView(t, m, "◐ 4-d", "! 6-f")
}

// Rules 9 and 10: enter opens the task popup, b opens the brief in it, d
// opens the dependency popup, and esc closes each in turn.
func TestWorkPopups(t *testing.T) {
	m := newWorkModel(t, workTestData(), demoUIState{Selection: "5-e"}, 120, 40)
	m = pressWork(t, m, key(tea.KeyEnter, ""))
	requireView(t, m, "5-e · esc closes", "title      Task 5-e", "status     assigned", "goal       g2-board · wave B",
		"worker     t5-worker-1 · working · 00:12:00", "verifier   none yet", "branch     task/5-e", "merge      none",
		"depends    2-b · landed", "ACCEPTANCE", "○ AC1      The card shows", "EVIDENCE", "b brief.md    v VERIFIED.md   d deps graph",
		"b brief · v verified · d deps · ↑↓ pgup/pgdn scroll · esc closes")
	m = pressWork(t, m, key('b', "b"))
	requireView(t, m, "brief.md · esc closes", "| AC1 | The card shows | V1 |")
	m = pressWork(t, m, key(tea.KeyEscape, ""))
	requireView(t, m, "5-e · esc closes")
	m = pressWork(t, m, key('d', "d"))
	requireView(t, m, "5-e · dependencies · esc closes", "WAITS ON", "● 2-b", "wave A · landed · 123456", "WAITED ON BY", "○ 8-h", "wave C · created")
	m = pressWork(t, m, key(tea.KeyEscape, ""))
	requireView(t, m, "5-e · esc closes")
	m = pressWork(t, m, key(tea.KeyEscape, ""))
	rejectView(t, m, "esc closes")
	if m.activeTab() != tabWork {
		t.Fatalf("esc closed the popup and left Work")
	}
	m = pressWork(t, m, key(tea.KeyEscape, ""))
	if m.activeTab() != tabChat {
		t.Fatalf("esc with no popup gave tab %d, want Chat", m.activeTab())
	}
}

// Rule 10 with the board GfWorkDeps: the dependency popup shows the ready
// when line, ↑ and ↓ move its selection, and enter opens the task popup of
// the selected row.
func TestWorkDepsPopupEnterOpensTask(t *testing.T) {
	m := newWorkModel(t, workTestData(), demoUIState{Selection: "8-h"}, 120, 40)
	m = pressWork(t, m, key('d', "d"))
	requireView(t, m, "8-h · dependencies · esc closes", "○ 5-e", "! 6-f", "ready when 5-e, 6-f land", "enter opens the selected task · ↑↓ move · esc closes")
	m = pressWork(t, m, arrowKey(tea.KeyDown))
	m = pressWork(t, m, key(tea.KeyEnter, ""))
	requireView(t, m, "6-f · esc closes", "status     created")
	m = pressWork(t, m, key(tea.KeyEscape, ""))
	if got := selectedCard(t, m); got != "6-f" {
		t.Fatalf("selected %q after enter in the dependency popup, want 6-f", got)
	}
}

// Ask a28 (W10), rule 10: the dependency popup has the ready when line. The
// selected row has the ▶ mark and its ID in accent bold, ↑↓ moves it, and
// the hint names enter.
func TestWorkDepsPopupSelection(t *testing.T) {
	m := newWorkModel(t, workTestData(), demoUIState{Selection: "8-h"}, 120, 40)
	m = pressWork(t, m, key('d', "d"))
	requireView(t, m, "ready when 5-e, 6-f land", "enter opens the selected task · ↑↓ move · esc closes", "▶ ○ 5-e", "  ! 6-f")
	m = pressWork(t, m, arrowKey(tea.KeyDown))
	requireView(t, m, "  ○ 5-e", "▶ ! 6-f")
	grid, err := cellgrid.Convert(m.View().Content, 120, 40)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for row, line := range workRows(m) {
		if col := strings.Index(line, "▶ ! 6-f"); col >= 0 {
			found = true
			col = ansi.StringWidth(line[:col]) + 4
			if cell := grid.At(row, col); cell.Token != cellgrid.Accent || !cell.Bold {
				t.Errorf("selected ID cell %+v, want accent bold", cell)
			}
		}
	}
	if !found {
		t.Fatal("no selected row ▶ ! 6-f")
	}
}

// Rule 9: the brief and VERIFIED.md of the task popup scroll, so each line
// of a long file is reachable.
func TestWorkTaskPopupFileScrolls(t *testing.T) {
	data := workTestData()
	var lines []string
	for i := 1; i <= 80; i++ {
		lines = append(lines, fmt.Sprintf("brief line %02d", i))
	}
	var verified []string
	for i := 1; i <= 80; i++ {
		verified = append(verified, fmt.Sprintf("verified line %02d", i))
	}
	data.Files[".hive/spec/board/tasks/5-e/brief.md"] = strings.Join(lines, "\n")
	data.Files[".hive/spec/board/tasks/5-e/evidence/VERIFIED.md"] = strings.Join(verified, "\n")
	m := newWorkModel(t, data, demoUIState{Selection: "5-e"}, 120, 40)
	m = pressWork(t, m, key(tea.KeyEnter, ""))
	m = pressWork(t, m, key('b', "b"))
	requireView(t, m, "brief.md · esc closes", "brief line 01")
	m = pressWork(t, m, key(tea.KeyPgDown, ""))
	requireView(t, m, "brief line 11")
	for range 7 {
		m = pressWork(t, m, key(tea.KeyPgDown, ""))
	}
	requireView(t, m, "brief line 80")
	for range 8 {
		m = pressWork(t, m, key(tea.KeyPgUp, ""))
	}
	m = pressWork(t, m, key(tea.KeyEscape, ""))
	m = pressWork(t, m, key('v', "v"))
	for range 8 {
		m = pressWork(t, m, key(tea.KeyPgDown, ""))
	}
	requireView(t, m, "VERIFIED.md · esc closes", "verified line 80")
}

// Rule 10: the dependency popup keeps the selected row on the screen when
// the list is longer than the popup.
func TestWorkDepsPopupSelectionStaysVisible(t *testing.T) {
	data := workTestData()
	for i := 40; i <= 60; i++ {
		id := fmt.Sprintf("%d-dependency", i)
		data.Tasks = append(data.Tasks, records.Task{ID: id, Project: "hive", Status: "created"})
		data.Tasks[7].DependsOn = append(data.Tasks[7].DependsOn, id)
	}
	m := newWorkModel(t, data, demoUIState{Selection: "8-h"}, 120, 40)
	m = pressWork(t, m, key('d', "d"))
	requireView(t, m, "▶ ○ 5-e")
	for range 22 {
		m = pressWork(t, m, arrowKey(tea.KeyDown))
	}
	requireView(t, m, "▶ ○ 60-dependency", "WAITED ON BY", "ready when")
	for range 4 {
		m = pressWork(t, m, arrowKey(tea.KeyUp))
	}
	requireView(t, m, "▶ ○ 56-dependency")
}

// Verification finding F1: a task with no dependencies and no dependents
// opens an empty dependency popup. Each navigation key, enter, and esc keep
// the popup, its visible state, and its action log, with no panic.
func TestWorkDepsPopupEmptyList(t *testing.T) {
	m := newWorkModel(t, workTestData(), demoUIState{Selection: "30-z"}, 120, 40)
	var log bytes.Buffer
	m.logger = slog.New(slog.NewJSONHandler(&log, &slog.HandlerOptions{Level: slog.LevelDebug}))
	m = pressWork(t, m, key('d', "d"))
	requireView(t, m, "30-z · dependencies · esc closes", "WAITS ON", "WAITED ON BY", "none")
	for _, press := range []tea.KeyPressMsg{arrowKey(tea.KeyDown), key('j', "j"), arrowKey(tea.KeyUp), key('k', "k")} {
		log.Reset()
		m = pressWork(t, m, press)
		requireView(t, m, "30-z · dependencies · esc closes", "WAITS ON", "WAITED ON BY", "none")
		if !strings.Contains(log.String(), `"action":"work deps selection"`) {
			t.Errorf("%q wrote no dependency selection action: %s", press.String(), log.String())
		}
	}
	log.Reset()
	m = pressWork(t, m, key(tea.KeyEnter, ""))
	requireView(t, m, "30-z · dependencies · esc closes", "none")
	m = pressWork(t, m, key(tea.KeyEscape, ""))
	if m.tasks.popup != "" {
		t.Fatalf("esc left popup %q, want the board", m.tasks.popup)
	}
	// The task popup opens the same empty list, and esc returns to it.
	m = pressWork(t, m, key(tea.KeyEnter, ""))
	m = pressWork(t, m, key('d', "d"))
	m = pressWork(t, m, arrowKey(tea.KeyDown))
	requireView(t, m, "30-z · dependencies · esc closes", "none")
	m = pressWork(t, m, key(tea.KeyEscape, ""))
	requireView(t, m, "30-z · esc closes", "status     created")
}

// AC2: each scroll and each selection key of a popup writes one action log
// event.
func TestWorkPopupActionsLog(t *testing.T) {
	m := newWorkModel(t, workTestData(), demoUIState{Selection: "8-h"}, 120, 40)
	var log bytes.Buffer
	m.logger = slog.New(slog.NewJSONHandler(&log, &slog.HandlerOptions{Level: slog.LevelDebug}))
	m = pressWork(t, m, key(tea.KeyEnter, ""))
	log.Reset()
	m = pressWork(t, m, arrowKey(tea.KeyDown))
	if !strings.Contains(log.String(), `"action":"work task popup scroll"`) {
		t.Fatalf("task popup scroll wrote no action log event: %s", log.String())
	}
	log.Reset()
	m = pressWork(t, m, key('d', "d"))
	log.Reset()
	pressWork(t, m, arrowKey(tea.KeyDown))
	if !strings.Contains(log.String(), `"action":"work deps selection"`) {
		t.Fatalf("dependency selection wrote no action log event: %s", log.String())
	}
}

// Rule 13: no goal and no task shows the notice, and a goal with no waves
// shows one empty column.
func TestWorkEmptyStates(t *testing.T) {
	m := newWorkModel(t, demoData{Projects: []records.Project{{ID: "hive"}}}, demoUIState{}, 120, 40)
	rows := workRows(m)
	if !strings.HasPrefix(rows[0], " work · hive · no goals") || !strings.HasPrefix(rows[1], " ───") {
		t.Errorf("title rows %q %q", rows[0], rows[1])
	}
	if got := strings.TrimSpace(rows[19]); got != "no goals · hive goal create starts one" {
		t.Errorf("row 19 %q, want the notice", rows[19])
	}
	data := demoData{Projects: []records.Project{{ID: "hive"}}, Goals: []records.Goal{{ID: "g1-new", Project: "hive", Title: "New", Status: "planned"}}}
	m = newWorkModel(t, data, demoUIState{}, 120, 40)
	requireView(t, m, " work · g1-new · New", "╭─ no waves ─")
}

// Rule 2: a stop is a task with needs_user or a task whose worker is
// unresolved. The other goal of the test data has one.
func TestWorkStopsElsewhere(t *testing.T) {
	data := workTestData()
	m := newWorkModel(t, data, demoUIState{Selection: "20-x"}, 120, 40)
	requireView(t, m, " work · g1-old", "! 20-x", "stop · unresolved", "· 1 stops elsewhere")
	data.Tasks[5].NeedsUser = ""
	m = newWorkModel(t, data, demoUIState{Selection: "20-x"}, 120, 40)
	rejectView(t, m, "stops elsewhere")
}

// tui-v2 decision 46: at each breakpoint the board fills the pane, no
// border breaks, and the selected card shows. At 79 columns the too-small
// notice shows. A resize keeps an open popup.
func TestWorkBreakpoints(t *testing.T) {
	for _, size := range []demoSize{{79, 24}, {80, 24}, {119, 40}, {120, 40}} {
		for _, popup := range []bool{false, true} {
			m := newWorkModel(t, workTestData(), demoUIState{Selection: "9-i"}, 120, 40)
			if popup {
				m = pressWork(t, m, key(tea.KeyEnter, ""))
			}
			m, _ = raise(t, m, tea.WindowSizeMsg{Width: size.Width, Height: size.Height})
			if size.Width < minimumWidth {
				requireView(t, m, "terminal too small")
				continue
			}
			width, height := 0, 0
			if popup {
				width, height = workTaskPopupWidth, workTaskPopupHeight
			}
			requireWorkFits(t, m, size.Width, size.Height, width, height)
			if popup {
				requireView(t, m, "9-i · esc closes")
				continue
			}
			if got := selectedCard(t, m); got != "9-i" {
				t.Errorf("%dx%d selected %q, want 9-i", size.Width, size.Height, got)
			}
		}
	}
}

// Ask a27: with the default config, the first status row of each screen
// except Chat is state, the screen segment, and clock. The screen segment
// names the open popup (GfWorkTask row 39). Chat keeps the row of decision
// 9, and a first row that the user sets stays on each screen.
func TestWorkStatusRow(t *testing.T) {
	m := newWorkModel(t, workTestData(), demoUIState{Selection: "8-h"}, 120, 40)
	statusRow := func() string {
		rows := workRows(m)
		return rows[len(rows)-1]
	}
	if row := statusRow(); !strings.Contains(row, " · work · ←→ ↑↓ card · [ ] goal") || strings.Contains(row, "supervisor") {
		t.Fatalf("Work status row = %q; want state, work and its keys, clock", row)
	}
	m = pressWork(t, m, key(tea.KeyEnter, ""))
	if row := statusRow(); !strings.Contains(row, " · task · b brief · v verified") {
		t.Fatalf("task popup status row = %q; want the popup name task", row)
	}
	m = pressWork(t, m, key('d', "d"))
	if row := statusRow(); !strings.Contains(row, " · deps · ↑↓ move · enter task") {
		t.Fatalf("deps popup status row = %q; want the popup name deps", row)
	}
	m = pressWork(t, m, key(tea.KeyEscape, ""))
	m = pressWork(t, m, key(tea.KeyEscape, ""))
	m = pressWork(t, m, key(tea.KeyEscape, ""))
	if m.activeTab() != tabChat {
		t.Fatalf("tab %d after esc on the board; want Chat", m.activeTab())
	}
	if row := statusRow(); !strings.Contains(row, "supervisor") || strings.Contains(row, " · chat · ") {
		t.Fatalf("Chat status row = %q; want the row of decision 9", row)
	}
	m.tuiConfig.StatusLines = [][]string{{"state", "asks", "clock"}}
	m.tuiConfig.StatusLinesExplicit = true
	m = pressWork(t, m, ctrl('t'))
	if row := statusRow(); !strings.Contains(row, "? 0 asks") || strings.Contains(row, " · work · ") {
		t.Fatalf("Work status row with a user row = %q; want the user row", row)
	}
}
