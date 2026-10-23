package tui

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/Harrison-Blair/hive/internal/lib/home"
)

var gfDay = time.Date(2026, 10, 23, 14, 22, 0, 0, time.Local)

// newFrameModel gives a model of the golden frames at width×height.
func newFrameModel(t *testing.T, width, height int) (*Model, string) {
	t.Helper()
	dir := newTestHome(t)
	m := newTestModel(t, dir, fixedClock(t, gfDay))
	m, _ = raise(t, m, tea.WindowSizeMsg{Width: width, Height: height})
	return m, dir
}

func ctrl(code rune) tea.KeyPressMsg { return tea.KeyPressMsg{Code: code, Mod: tea.ModCtrl} }

func escKey() tea.KeyPressMsg { return tea.KeyPressMsg{Code: tea.KeyEscape} }

func viewRows(m *Model) []string {
	return strings.Split(ansi.Strip(m.View().Content), "\n")
}

// drain runs cmd and each command of a batch, and drops each command that
// does not end within a short time, for example a tick.
func drain(cmd tea.Cmd) {
	if cmd == nil {
		return
	}
	done := make(chan tea.Msg, 1)
	go func() { done <- cmd() }()
	select {
	case msg := <-done:
		if b, ok := msg.(tea.BatchMsg); ok {
			for _, c := range b {
				drain(c)
			}
		}
	case <-time.After(300 * time.Millisecond):
	}
}

// tuiActions gives the action field of each tui action event of the day.
func tuiActions(t *testing.T, dir string) []map[string]any {
	t.Helper()
	data, err := os.ReadFile(dailyPath(dir, gfDay))
	if err != nil {
		t.Fatal(err)
	}
	var out []map[string]any
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		var event map[string]any
		if json.Unmarshal([]byte(line), &event) == nil && event["msg"] == "tui action" {
			out = append(out, event)
		}
	}
	return out
}

// Golden-frames decisions 1 to 3: Chat is home. It has no header and no tab
// bar. The last row is the status row.
func TestFrameChatIsHome(t *testing.T) {
	m, _ := newFrameModel(t, 120, 40)
	rows := viewRows(m)
	if len(rows) != 40 {
		t.Fatalf("view has %d rows; want 40", len(rows))
	}
	for i, row := range rows {
		if strings.Contains(row, "Agents") && strings.Contains(row, "Friction") {
			t.Errorf("row %d is a tab bar: %q", i, row)
		}
		if strings.Contains(row, "hive ·") {
			t.Errorf("row %d is a header: %q", i, row)
		}
	}
	last := rows[39]
	for _, want := range []string{" ○ none · ", "? 0 asks", "needs you 0", "^g agents  ^t work  ^f friction  ^y health  ^l logs", "14:22 "} {
		if !strings.Contains(last, want) {
			t.Errorf("status row %q does not contain %q", last, want)
		}
	}
}

// Decisions 1, 2, and 16: each chord opens its screen at full size with a
// title row and a rule. esc, or the same chord again, returns to Chat. Each
// chord and each screen change is logged.
func TestFrameChordsOpenScreens(t *testing.T) {
	for _, tc := range []struct {
		chord rune
		name  string
	}{{'g', "agents"}, {'t', "work"}, {'f', "friction"}, {'y', "health"}, {'l', "logs"}} {
		t.Run(tc.name, func(t *testing.T) {
			m, dir := newFrameModel(t, 120, 40)
			m, cmd := raise(t, m, ctrl(tc.chord))
			drain(cmd)
			rows := viewRows(m)
			if !strings.HasPrefix(rows[0], " "+tc.name) {
				t.Fatalf("row 0 = %q; want the title %q", rows[0], tc.name)
			}
			if rows[1] != " "+strings.Repeat("─", 118)+" " {
				t.Fatalf("row 1 = %q; want the rule", rows[1])
			}
			if !strings.Contains(rows[39], "esc chat") {
				t.Errorf("status row %q does not show esc chat", rows[39])
			}
			back := escKey()
			if tc.name == "work" {
				back = ctrl(tc.chord)
			}
			m, cmd = raise(t, m, back)
			drain(cmd)
			if rows := viewRows(m); strings.HasPrefix(rows[0], " "+tc.name) || m.activeTab() != tabChat {
				t.Fatalf("after %v the screen is %d, row 0 %q; want Chat", back, m.activeTab(), rows[0])
			}
			var chords, screens []string
			for _, event := range tuiActions(t, dir) {
				switch event["action"] {
				case "chord":
					chords = append(chords, event["key"].(string))
				case "screen select":
					screens = append(screens, event["screen"].(string))
				}
			}
			if len(chords) < 1 || chords[0] != "ctrl+"+string(tc.chord) {
				t.Errorf("chord events = %v; want ctrl+%c first", chords, tc.chord)
			}
			if len(screens) != 2 || !strings.EqualFold(screens[0], tc.name) || screens[1] != "Chat" {
				t.Errorf("screen events = %v; want %s then Chat", screens, tc.name)
			}
		})
	}
}

// Decisions 6 and 19: there are no number keys and no tab region move.
// ctrl+] moves the box focus on a screen with two boxes, and does nothing on
// Chat.
func TestFrameBoxFocusKeys(t *testing.T) {
	m, _ := newFrameModel(t, 120, 40)
	m, _ = raise(t, m, ctrl(']'))
	if m.activeTab() != tabChat || !m.chatInputActive {
		t.Fatalf("ctrl+] on Chat changed the screen %d or the input focus %v", m.activeTab(), m.chatInputActive)
	}
	m, _ = raise(t, m, ctrl('t'))
	for _, k := range []tea.KeyPressMsg{key('1', "1"), key('2', "2"), {Code: tea.KeyTab}} {
		m, _ = raise(t, m, k)
		if m.activeTab() != tabWork || m.regionFocus[tabWork] != 0 {
			t.Fatalf("%v moved to screen %d, region %d; want Work, region 0", k, m.activeTab(), m.regionFocus[tabWork])
		}
	}
	m, _ = raise(t, m, ctrl(']'))
	if m.regionFocus[tabWork] != 1 {
		t.Fatalf("ctrl+] gave region %d; want 1", m.regionFocus[tabWork])
	}
	m, _ = raise(t, m, ctrl(']'))
	if m.regionFocus[tabWork] != 0 {
		t.Fatalf("second ctrl+] gave region %d; want 0", m.regionFocus[tabWork])
	}
}

// Decisions 16 and 19: there is no normal mode. The composer has the focus
// at start, and i, j, k, g, and G type. The chords do not type.
func TestFrameChatHasNoNormalMode(t *testing.T) {
	m, _ := newFrameModel(t, 120, 40)
	if !m.chatInputActive {
		t.Fatal("the composer has no focus at start")
	}
	for _, r := range "ijkgG" {
		m, _ = raise(t, m, key(r, string(r)))
	}
	m, _ = raise(t, m, ctrl('t'))
	m, _ = raise(t, m, escKey())
	if got := m.chatInput.Value(); got != "ijkgG" {
		t.Fatalf("input = %q; want ijkgG", got)
	}
}

// Decision 18: [tui.keys] rebinds a chord. The old key then does nothing.
func TestFrameKeysRebind(t *testing.T) {
	m, _ := newFrameModel(t, 120, 40)
	cfg := home.DefaultTUI()
	cfg.Keys["agents"] = "ctrl+a"
	m.setTUIConfig(cfg)
	m, _ = raise(t, m, ctrl('g'))
	if m.activeTab() != tabChat {
		t.Fatalf("ctrl+g opened screen %d after the rebind; want Chat", m.activeTab())
	}
	m, _ = raise(t, m, ctrl('a'))
	if m.activeTab() != tabAgents {
		t.Fatalf("ctrl+a opened screen %d; want Agents", m.activeTab())
	}
	if rows := viewRows(m); !strings.Contains(rows[39], "esc chat  ^t work") {
		t.Errorf("status row %q; want esc chat first", rows[39])
	}
	m, _ = raise(t, m, escKey())
	if rows := viewRows(m); !strings.Contains(rows[39], "^a agents  ^t work") {
		t.Errorf("status row %q; want the rebound chord ^a agents", rows[39])
	}
}
