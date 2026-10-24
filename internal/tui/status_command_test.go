package tui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/Harrison-Blair/hive/internal/lib/records"
)

// statusCommandModel gives a Chat model whose status row has the command
// segment and the command text.
func statusCommandModel(t *testing.T, command string) (*Model, *testClock) {
	t.Helper()
	home := newTestHome(t)
	writeRecord(t, home, []string{"agents"}, "coordinator-g1", records.Agent{
		Key: "coordinator-g1", Name: "coordinator", Role: "coordinator", Status: records.AgentLive,
		AgentStatus: "working", StatusChangedAt: "2026-10-20T11:58:00Z",
	})
	clock := fixedClock(t, time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC))
	m := newTestModel(t, home, clock)
	t.Cleanup(m.Stop)
	m, _ = raise(t, m, tea.WindowSizeMsg{Width: 120, Height: 40})
	state, err := readHeader(home)
	if err != nil {
		t.Fatal(err)
	}
	m.header = state
	cfg := defaultTUI()
	cfg.StatusLines = [][]string{{"state", "command", "clock"}}
	cfg.StatusCommand = command
	m.setTUIConfig(cfg)
	return m, clock
}

// Golden-frames decision 12: the command gets the JSON object on its
// standard input, and the segment shows the first line of its output with
// the ANSI colors kept. The process starts in the command, not on the loop.
func TestStatusCommandShowsTheFirstLineWithColors(t *testing.T) {
	stdin := filepath.Join(t.TempDir(), "stdin.json")
	m, _ := statusCommandModel(t, "cat > "+stdin+"; printf '\\033[31mbuild red\\033[0m\\nsecond line\\n'")
	cmd := m.startStatusCommand()
	if cmd == nil {
		t.Fatal("no status command started")
	}
	if _, err := os.Stat(stdin); err == nil {
		t.Fatal("the process ran on the input loop")
	}
	m, _ = raise(t, m, cmd())
	row := m.View().Content
	if got := viewRow(m, 39); !strings.HasPrefix(got, " ● working 2m · build red ") || strings.Contains(got, "second line") {
		t.Fatalf("status row = %q", got)
	}
	if !strings.Contains(row, "\x1b[31mbuild red") {
		t.Fatalf("status row lost the ANSI color:\n%q", row)
	}
	data, err := os.ReadFile(stdin)
	if err != nil {
		t.Fatal(err)
	}
	var input map[string]any
	if err := json.Unmarshal(data, &input); err != nil {
		t.Fatalf("stdin %q: %v", data, err)
	}
	for _, key := range []string{"home", "screen", "coordinator", "supervisor", "asks", "needs_you", "project", "clock"} {
		if _, ok := input[key]; !ok {
			t.Errorf("stdin %s lacks %q", data, key)
		}
	}
	coordinator, _ := input["coordinator"].(map[string]any)
	if coordinator["state"] != "working" || coordinator["changed_at"] != "2026-10-20T11:58:00Z" {
		t.Errorf("coordinator = %v, want working at 2026-10-20T11:58:00Z", coordinator)
	}
	if supervisor, _ := input["supervisor"].(map[string]any); supervisor["state"] != "stale" {
		t.Errorf("supervisor = %v, want stale", supervisor)
	}
	if input["screen"] != "chat" || input["asks"] != float64(0) || input["clock"] != "2026-10-20T12:00:00Z" {
		t.Errorf("stdin = %s", data)
	}
}

// Decision 12: a line with cursor moves keeps only its colors, so the row
// stays one row.
func TestStatusCommandDropsControlSequences(t *testing.T) {
	m, _ := statusCommandModel(t, "printf 'a\\033[2Jb\\033[1;1Hc\\033[32md\\a\\r'")
	m, _ = raise(t, m, m.startStatusCommand()())
	if got := viewRow(m, 39); !strings.HasPrefix(got, " ● working 2m · abcd ") {
		t.Fatalf("status row = %q", got)
	}
	if row := m.View().Content; strings.Contains(row, "\x1b[2J") || strings.Contains(row, "\x1b[1;1H") || !strings.Contains(row, "\x1b[32md") {
		t.Fatalf("status row kept a control sequence or lost the color:\n%q", row)
	}
}

// Decision 12: at most one run each 300 ms, and one run at a time.
func TestStatusCommandRunsAtMostOnceEach300ms(t *testing.T) {
	m, clock := statusCommandModel(t, "echo ok")
	first := m.startStatusCommand()
	if first == nil {
		t.Fatal("no first run")
	}
	clock.set(clock.now().Add(time.Second))
	if m.startStatusCommand() != nil {
		t.Fatal("a second run started while the first runs")
	}
	m, _ = raise(t, m, first())
	clock.set(clock.now().Add(299 * time.Millisecond))
	if m.startStatusCommand() == nil {
		t.Fatal("no run 1.299 s after the first start")
	}
	m, _ = raise(t, m, statusCommandMsg{line: "ok"})
	clock.set(clock.now().Add(299 * time.Millisecond))
	if m.startStatusCommand() != nil {
		t.Fatal("a run started 299 ms after the last start")
	}
	clock.set(clock.now().Add(time.Millisecond))
	if m.startStatusCommand() == nil {
		t.Fatal("no run 300 ms after the last start")
	}
}

// Decision 12: a failure empties the segment and logs once. A success after
// it shows the line again, and the next failure logs again.
func TestStatusCommandFailureLogsOnce(t *testing.T) {
	m, clock := statusCommandModel(t, "echo before")
	m, _ = raise(t, m, m.startStatusCommand()())
	if got := viewRow(m, 39); !strings.Contains(got, " · before ") {
		t.Fatalf("status row = %q", got)
	}
	cfg := m.tuiConfig
	cfg.StatusCommand = "echo partial; exit 3"
	m.setTUIConfig(cfg)
	logs := 0
	for range 3 {
		clock.set(clock.now().Add(time.Second))
		cmd := m.startStatusCommand()
		var logCmd tea.Cmd
		m, logCmd = raise(t, m, cmd())
		if logCmd != nil {
			logs++
		}
		if got := ansi.Strip(m.statusSegment("command")); got != "" {
			t.Fatalf("segment after a failure = %q, want empty", got)
		}
	}
	if logs != 1 {
		t.Fatalf("%d log events for 3 failures, want 1", logs)
	}
	cfg.StatusCommand = "echo after"
	m.setTUIConfig(cfg)
	clock.set(clock.now().Add(time.Second))
	m, _ = raise(t, m, m.startStatusCommand()())
	cfg.StatusCommand = "exit 1"
	m.setTUIConfig(cfg)
	clock.set(clock.now().Add(time.Second))
	if _, logCmd := raise(t, m, m.startStatusCommand()()); logCmd == nil {
		t.Fatal("a failure after a success logged nothing")
	}
}

// Decisions 12 and 13: no command, a row with no command segment, and demo
// mode run no process. The tick starts the command.
func TestStatusCommandRunsOnlyWhenItApplies(t *testing.T) {
	m, _ := statusCommandModel(t, "")
	if m.startStatusCommand() != nil {
		t.Fatal("a run started with no command")
	}
	m, _ = statusCommandModel(t, "echo ok")
	cfg := m.tuiConfig
	cfg.StatusLines = [][]string{{"state", "clock"}}
	m.setTUIConfig(cfg)
	if m.startStatusCommand() != nil {
		t.Fatal("a run started with no command segment")
	}
	m, _ = statusCommandModel(t, "echo ok")
	m.demo = &demoState{}
	if m.startStatusCommand() != nil {
		t.Fatal("a run started in demo mode")
	}
	m, _ = statusCommandModel(t, "echo ticked")
	_, cmd := raise(t, m, tickMsg{})
	if !m.statusCommandRunning {
		t.Fatal("the tick started no status command")
	}
	_ = cmd
}
