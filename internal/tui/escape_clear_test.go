package tui

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
)

// TestViewingEscapeClearsAppliedTextFilter drives the CP4 sequence at the model:
// filter, Enter, level, follow, top, bottom, and Escape. Escape in viewing mode
// must clear the applied text filter and the input. It must keep the level, the
// follow state, and the viewing mode.
func TestViewingEscapeClearsAppliedTextFilter(t *testing.T) {
	home := newTestHome(t)
	day := time.Date(2026, 5, 20, 12, 0, 0, 0, time.Local)
	clock := fixedClock(t, day)
	writeEvents(t, dailyPath(home, day),
		eventJSON(t, day, "INFO", "project alpha"),
		eventJSON(t, day, "INFO", "other beta"),
		eventJSON(t, day, "DEBUG", "project debug"),
		eventJSON(t, day, "WARN", "project gamma"),
		eventJSON(t, day, "INFO", "delta"),
	)
	m := newTestLogsModel(t, home, clock)
	m, _ = raise(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})

	m, _ = raise(t, m, key('/', "/"))
	for _, r := range "project" {
		m, _ = raise(t, m, key(r, string(r)))
	}
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyEnter})
	m, _ = raise(t, m, key('l', "l"))
	m, _ = raise(t, m, key('f', "f"))
	m, _ = raise(t, m, key('g', "g"))
	m, _ = raise(t, m, key('G', "G"))
	requireView(t, m, "level: info  text: project  follow: on")
	rejectView(t, m, "other beta", "delta")

	before := countLogLines(t, home)
	m, cmd := raise(t, m, tea.KeyPressMsg{Code: tea.KeyEscape})
	if cmd != nil {
		m, _ = raise(t, m, cmd())
	}
	after := countLogLines(t, home)
	if after != before+1 {
		t.Fatalf("log lines went from %d to %d after viewing Escape; want one event", before, after)
	}
	data, err := os.ReadFile(dailyPath(home, day))
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	var event map[string]any
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &event); err != nil {
		t.Fatal(err)
	}
	if event["msg"] != "tui action" || event["action"] != "filter clear" || event["component"] != "tui" {
		t.Fatalf("viewing Escape log event = %v", event)
	}

	requireView(t, m, "level: info  text: -  follow: on", "project alpha", "other beta", "project gamma", "delta")
	rejectView(t, m, "project debug")
	if m.textFilter != "" {
		t.Fatalf("textFilter=%q after viewing Escape; want empty", m.textFilter)
	}
	if got := m.input.Value(); got != "" {
		t.Fatalf("input=%q after viewing Escape; want empty", got)
	}
	if m.mode != viewing {
		t.Fatalf("mode=%v after viewing Escape; want viewing", m.mode)
	}
	if m.level != "info" {
		t.Fatalf("level=%q after viewing Escape; want info", m.level)
	}
	if !m.follow {
		t.Fatal("follow=off after viewing Escape; want on")
	}
}
