package tui

import (
	"encoding/json"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Harrison-Blair/hive/internal/lib/cellgrid"
	"github.com/charmbracelet/x/ansi"
)

const designDir = "../../.hive/spec/tui-v2/design"

func newTestDemoModel(t *testing.T, fixture string, size demoSize) *Model {
	t.Helper()
	loaded, err := loadDemoFixture(filepath.Join(designDir, fixture))
	if err != nil {
		t.Fatal(err)
	}
	m := newDemoModel(t.TempDir(), &demoState{fixture: loaded, size: size})
	t.Cleanup(m.Stop)
	return m
}

func TestDemoFixtureUnknownFieldNamesTheField(t *testing.T) {
	base, err := os.ReadFile(filepath.Join(designDir, "Main.fixture.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, old, new, field string
	}{
		{name: "top level", old: `"board":`, new: `"colour": 1, "board":`, field: `"colour"`},
		{name: "ui", old: `"screen": "agents"`, new: `"screen": "agents", "tab": 2`, field: `"tab"`},
		{name: "record", old: `"key": "coordinator-g2"`, new: `"key": "coordinator-g2", "nickname": "c"`, field: `"nickname"`},
		{name: "supervisor", old: `"binary_old": false`, new: `"binary_old": false, "uptime": 3`, field: `"uptime"`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data := strings.Replace(string(base), tc.old, tc.new, 1)
			if data == string(base) {
				t.Fatalf("fixture has no %s", tc.old)
			}
			_, err := parseDemoFixture([]byte(data))
			if err == nil || !strings.Contains(err.Error(), tc.field) {
				t.Fatalf("error = %v; want one that names %s", err, tc.field)
			}
		})
	}
}

func TestDemoFixtureBadValuesFail(t *testing.T) {
	base, err := os.ReadFile(filepath.Join(designDir, "Main.fixture.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		old, new, want string
	}{
		{old: `"screen": "agents"`, new: `"screen": "queue"`, want: `"ui.screen"`},
		{old: `"view": ""`, new: `"view": "goals"`, want: `"ui.view"`},
		{old: `"zone": "UTC"`, new: `"zone": "Mars/Olympus"`, want: `"clock.zone"`},
		{old: `"time": "2026-10-21T14:22:00Z"`, new: `"time": "14:22"`, want: `"clock.time"`},
		{old: `"width": 120`, new: `"width": 0`, want: `"terminal"`},
		{old: `"overlay": ""`, new: `"overlay": "brief"`, want: `"ui.overlay"`},
		{old: `"kind": ""`, new: `"kind": "quit"`, want: `"ui.prompt.kind"`},
	} {
		t.Run(tc.want, func(t *testing.T) {
			data := strings.Replace(string(base), tc.old, tc.new, 1)
			if data == string(base) {
				t.Fatalf("fixture has no %s", tc.old)
			}
			if _, err := parseDemoFixture([]byte(data)); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v; want one that names %s", err, tc.want)
			}
		})
	}
	if _, err := parseDemoFixture(append(append([]byte(nil), base...), []byte(`{}`)...)); err == nil {
		t.Fatal("data after the fixture object loaded without an error")
	}
}

func TestParseDemoSize(t *testing.T) {
	width, height, err := ParseDemoSize("120x40")
	if err != nil || width != 120 || height != 40 {
		t.Fatalf("ParseDemoSize(120x40) = %d, %d, %v", width, height, err)
	}
	for _, value := range []string{"", "120", "120x", "x40", "0x40", "120x-1", "120X40", "120x40x1"} {
		if _, _, err := ParseDemoSize(value); err == nil || !strings.Contains(err.Error(), `"`+value+`"`) {
			t.Fatalf("ParseDemoSize(%q) error = %v; want one that names the value", value, err)
		}
	}
}

func TestDemoMainShowsFixtureDataAtTheFixtureClock(t *testing.T) {
	m := newTestDemoModel(t, "Main.fixture.json", demoSize{})
	if m.agents.selectedKey != "manager-hive-g2" {
		t.Fatalf("selected agent = %q; want the fixture selection manager-hive-g2", m.agents.selectedKey)
	}
	first := ansi.Strip(m.View().Content)
	for _, want := range []string{"14:22", "~/source/hive", "manager-hive", "t4-worker-1", "t4-verifier-1"} {
		if !strings.Contains(first, want) {
			t.Fatalf("view does not contain %q:\n%s", want, first)
		}
	}
	m.Update(tickMsg{})
	m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	if got := m.now(); !got.Equal(time.Date(2026, 10, 21, 14, 22, 0, 0, time.UTC)) {
		t.Fatalf("clock = %v; want the fixture time", got)
	}
	if !strings.Contains(ansi.Strip(m.View().Content), "14:22") {
		t.Fatal("clock label changed after a tick")
	}
}

func TestDemoSizeDrawsAtTheSizeInALargerPane(t *testing.T) {
	m := newTestDemoModel(t, "Main.fixture.json", demoSize{Width: 120, Height: 40})
	m.Update(tea.WindowSizeMsg{Width: 150, Height: 50})
	lines := strings.Split(m.View().Content, "\n")
	if len(lines) != 40 {
		t.Fatalf("view has %d lines; want 40", len(lines))
	}
	widest := 0
	for _, line := range lines {
		widest = max(widest, lipgloss.Width(line))
	}
	if widest != 120 {
		t.Fatalf("widest line = %d cells; want 120", widest)
	}
}

func TestDemoSizeShowsTooSmallInASmallerPane(t *testing.T) {
	m := newTestDemoModel(t, "Main.fixture.json", demoSize{Width: 120, Height: 40})
	m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	view := ansi.Strip(m.View().Content)
	for _, want := range []string{"terminal too small", "this pane is 100×30", "hive needs at least 120×40", "ctrl+c ×2 quit"} {
		if !strings.Contains(view, want) {
			t.Fatalf("view does not contain %q:\n%s", want, view)
		}
	}
	if strings.Contains(view, "manager-hive") {
		t.Fatalf("too-small view shows screen content:\n%s", view)
	}
	m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	if view := ansi.Strip(m.View().Content); strings.Contains(view, "terminal too small") || !strings.Contains(view, "manager-hive") {
		t.Fatalf("pane of the exact size does not draw the TUI:\n%s", view)
	}
}

func TestDemoTooSmallMatchesItsBoardGrid(t *testing.T) {
	m := newTestDemoModel(t, "TooSmall.fixture.json", demoSize{})
	file, err := os.Open(filepath.Join(designDir, "TooSmall.grid"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = file.Close() }()
	want, err := cellgrid.ReadGrid(file)
	if err != nil {
		t.Fatal(err)
	}
	got, err := cellgrid.Convert(m.View().Content, want.Width, want.Height)
	if err != nil {
		t.Fatal(err)
	}
	diffs, err := cellgrid.Compare(want, got, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(diffs) != 0 {
		t.Fatalf("too-small view differs from TooSmall.grid:\n%s", cellgrid.FormatDiffs(diffs))
	}
}

func TestDemoActionKeysShowTheNotice(t *testing.T) {
	m := newTestDemoModel(t, "Main.fixture.json", demoSize{})
	// Decision 14: p peeks at the pane, and enter opens the details popup.
	// Decision 19: restart is in Health › Agents, so r on Agents does nothing.
	m.Update(tea.KeyPressMsg{Code: 'r', Text: "r"})
	if view := ansi.Strip(m.View().Content); strings.Contains(view, demoNotice) || m.agents.restartConfirm {
		t.Fatalf("r on the Agents tab acted:\n%s", view)
	}
	for _, key := range []tea.KeyPressMsg{
		{Code: 'a', Text: "a"},
		{Code: 'p', Text: "p"},
		{Code: 'r', Text: "r"},
	} {
		if key.Text == "r" {
			m.Update(tea.KeyPressMsg{Code: '5', Text: "5"})
			m.Update(tea.KeyPressMsg{Code: ']', Text: "]"})
			if m.activeTab() != tabHealth || m.screenIndex != 1 {
				t.Fatalf("5 ] did not select Health › Agents: tab %d, screen %d", m.activeTab(), m.screenIndex)
			}
		}
		m.Update(key)
		if view := ansi.Strip(m.View().Content); !strings.Contains(view, demoNotice) {
			t.Fatalf("key %q: view does not show %q:\n%s", key.String(), demoNotice, view)
		}
		if m.agents.peekOpen || m.agents.restartRunning || m.agents.restartConfirm {
			t.Fatalf("key %q started an action", key.String())
		}
		m.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})
		if view := ansi.Strip(m.View().Content); strings.Contains(view, demoNotice) {
			t.Fatalf("notice stays after the next key:\n%s", view)
		}
	}
	m.Update(tea.KeyPressMsg{Code: '2', Text: "2"})
	m.Update(tea.KeyPressMsg{Code: 'i', Text: "i"})
	m.Update(tea.KeyPressMsg{Code: 'h', Text: "h"})
	if cmd := m.updateKey(tea.KeyPressMsg{Code: tea.KeyEnter}); m.chatSending {
		t.Fatalf("chat send started; cmd %v", cmd)
	}
	if view := ansi.Strip(m.View().Content); !strings.Contains(view, demoNotice) {
		t.Fatalf("chat send: view does not show %q:\n%s", demoNotice, view)
	}
}

func TestDemoLogsShowOnlyFixtureEvents(t *testing.T) {
	home := t.TempDir()
	logDir := filepath.Join(home, ".hive", "logs")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		t.Fatal(err)
	}
	live := `{"time":"2026-10-21T14:10:00Z","level":"DEBUG","msg":"live home event","component":"cli"}` + "\n"
	if err := os.WriteFile(filepath.Join(logDir, "2026-10-21.jsonl"), []byte(live), 0o644); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(designDir, "Main.fixture.json"))
	if err != nil {
		t.Fatal(err)
	}
	text := strings.Replace(string(data), `"logs": []`, `"logs": [{"time":"2026-10-21T14:11:00Z","level":"INFO","msg":"fixture event","component":"task"}]`, 1)
	text = strings.Replace(text, `"screen": "agents"`, `"screen": "logs"`, 1)
	text = strings.Replace(text, `"focus": "graph"`, `"focus": "events"`, 1)
	text = strings.Replace(text, `"selection": "manager-hive-g2"`, `"selection": ""`, 1)
	text = strings.Replace(text, `"scroll": {"graph": {"top": 0, "left": 0}}`, `"scroll": {}`, 1)
	fixture, err := parseDemoFixture([]byte(text))
	if err != nil {
		t.Fatal(err)
	}
	m := newDemoModel(home, &demoState{fixture: fixture})
	t.Cleanup(m.Stop)
	view := ansi.Strip(m.View().Content)
	if !strings.Contains(view, "fixture event") {
		t.Fatalf("logs view lacks the fixture event:\n%s", view)
	}
	if strings.Contains(view, "live home event") {
		t.Fatalf("logs view shows a live home event:\n%s", view)
	}
}

// demoFileFixture gives the Main fixture with the project at root, the given
// UI state, the given friction entries, and the given file texts.
func demoFileFixture(t *testing.T, root string, ui map[string]any, friction []any, files map[string]string) *demoFixture {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(designDir, "Main.fixture.json"))
	if err != nil {
		t.Fatal(err)
	}
	var fixture map[string]any
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	fixtureData := fixture["data"].(map[string]any)
	fixtureData["projects"].([]any)[0].(map[string]any)["path"] = root
	fixtureData["friction"] = friction
	if files != nil {
		fixtureData["files"] = files
	}
	if ui["view"] == "goals" {
		// Goals opens the design of the first wave that is not done.
		goal := fixtureData["goals"].([]any)[0].(map[string]any)
		goal["waves"].([]any)[2].(map[string]any)["status"] = "active"
	}
	state := fixture["ui"].(map[string]any)
	maps.Copy(state, ui)
	text, err := json.Marshal(fixture)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := parseDemoFixture(text)
	if err != nil {
		t.Fatal(err)
	}
	return parsed
}

// deliverFileReads runs cmd and gives each file read result to the model.
func deliverFileReads(m *Model, cmd tea.Cmd) {
	if cmd == nil {
		return
	}
	switch msg := cmd().(type) {
	case tea.BatchMsg:
		for _, next := range msg {
			deliverFileReads(m, next)
		}
	case taskFileMsg, frictionFileMsg:
		m.Update(msg)
	}
}

func TestDemoOpensFilesOnlyFromTheFixture(t *testing.T) {
	const taskDir = ".hive/spec/hive-tui/tasks/4-agents-equal-columns/"
	const frictionPath = "reference/dogfood/friction/entries/demo-entry.md"
	tasksUI := map[string]any{"screen": "work", "view": "tasks", "focus": "list", "selection": "4-agents-equal-columns", "scroll": map[string]any{}}
	goalsUI := map[string]any{"screen": "work", "view": "goals", "focus": "list", "selection": "g1-tui-restart", "scroll": map[string]any{}}
	frictionUI := map[string]any{"screen": "friction", "view": "open", "focus": "list", "selection": "demo-entry.md", "scroll": map[string]any{}}
	friction := []any{map[string]any{"file": "demo-entry.md", "title": "Demo entry", "status": "Open", "opened": "2026-10-21", "body": "## Summary\n\nfixture body"}}
	for _, tc := range []struct {
		name, key, livePath string
		ui                  map[string]any
		files               map[string]string
		want                string
	}{
		{name: "brief from fixture", key: "b", livePath: taskDir + "brief.md", ui: tasksUI,
			files: map[string]string{taskDir + "brief.md": "fixture brief text"}, want: "fixture brief text"},
		{name: "brief not in fixture", key: "b", livePath: taskDir + "brief.md", ui: tasksUI, want: demoNoFile},
		{name: "evidence from fixture", key: "v", livePath: taskDir + "evidence/VERIFIED.md", ui: tasksUI,
			files: map[string]string{taskDir + "evidence/VERIFIED.md": "fixture evidence text"}, want: "fixture evidence text"},
		{name: "verification not in fixture", key: "v", livePath: taskDir + "evidence/VERIFIED.md", ui: tasksUI, want: demoNoFile},
		{name: "design from fixture", key: "o", livePath: ".hive/spec/hive-tui/README.md", ui: goalsUI,
			files: map[string]string{".hive/spec/hive-tui/README.md": "fixture design text"}, want: "fixture design text"},
		{name: "design not in fixture", key: "o", livePath: ".hive/spec/hive-tui/README.md", ui: goalsUI, want: demoNoFile},
		{name: "friction entry from fixture", key: "enter", livePath: frictionPath, ui: frictionUI,
			files: map[string]string{frictionPath: "fixture friction text"}, want: "fixture friction text"},
		{name: "friction entry not in fixture", key: "enter", livePath: frictionPath, ui: frictionUI, want: demoNoFile},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			live := filepath.Join(root, filepath.FromSlash(tc.livePath))
			if err := os.MkdirAll(filepath.Dir(live), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(live, []byte("live file text"), 0o644); err != nil {
				t.Fatal(err)
			}
			m := newDemoModel(root, &demoState{fixture: demoFileFixture(t, root, tc.ui, friction, tc.files)})
			t.Cleanup(m.Stop)
			key := tea.KeyPressMsg{Code: rune(tc.key[0]), Text: tc.key}
			if tc.key == "enter" {
				key = tea.KeyPressMsg{Code: tea.KeyEnter}
			}
			deliverFileReads(m, m.updateKey(key))
			view := ansi.Strip(m.View().Content)
			if !strings.Contains(view, tc.want) {
				t.Fatalf("view does not show %q:\n%s", tc.want, view)
			}
			if strings.Contains(view, "live file text") {
				t.Fatalf("view shows the live file:\n%s", view)
			}
		})
	}
}
