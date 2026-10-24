package tui

import (
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/Harrison-Blair/hive/internal/doctor"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/charmbracelet/x/ansi"
)

const (
	wantHoney = "38;2;245;185;66"
	wantText  = "38;2;232;228;218"
)

type sgrSpan struct {
	text       string
	foreground string
	bold       bool
}

func parseSGRSpans(raw string) []sgrSpan {
	var spans []sgrSpan
	var text strings.Builder
	foreground := ""
	bold := false
	flush := func() {
		if text.Len() == 0 {
			return
		}
		spans = append(spans, sgrSpan{text: text.String(), foreground: foreground, bold: bold})
		text.Reset()
	}
	for index := 0; index < len(raw); {
		if raw[index] == '\x1b' && index+1 < len(raw) && raw[index+1] == '[' {
			end := index + 2
			for end < len(raw) && raw[end] != 'm' {
				end++
			}
			if end == len(raw) {
				break
			}
			flush()
			parameters := strings.Split(raw[index+2:end], ";")
			for part := 0; part < len(parameters); part++ {
				switch parameters[part] {
				case "", "0":
					foreground, bold = "", false
				case "1":
					bold = true
				case "22":
					bold = false
				case "39":
					foreground = ""
				case "38":
					if part+4 < len(parameters) && parameters[part+1] == "2" {
						foreground = strings.Join(parameters[part:part+5], ";")
						part += 4
					}
				}
			}
			index = end + 1
			continue
		}
		_, size := utf8.DecodeRuneInString(raw[index:])
		text.WriteString(raw[index : index+size])
		index += size
	}
	flush()
	return spans
}

func requireHoneySelection(t *testing.T, raw, selected string) {
	t.Helper()
	want := "▶ " + selected
	for _, span := range parseSGRSpans(raw) {
		if span.foreground == wantHoney && span.bold && strings.HasPrefix(strings.TrimLeft(span.text, " \t"), want) {
			return
		}
	}
	t.Errorf("selection %q is not marked with honey bold text", want)
}

func requireHoneyBoldText(t *testing.T, raw, want string) {
	t.Helper()
	for _, span := range parseSGRSpans(raw) {
		if span.foreground == wantHoney && span.bold && strings.Contains(span.text, want) {
			return
		}
	}
	t.Fatalf("text %q is not honey bold", want)
}

func requireNoDefaultForeground(t *testing.T, raw string) {
	t.Helper()
	for _, span := range parseSGRSpans(raw) {
		if span.foreground == wantText {
			t.Fatal("raw output sets the default terminal foreground")
		}
	}
}

func selectionModel(t *testing.T, screen int) *Model {
	t.Helper()
	clock := fixedClock(t, time.Date(2026, 4, 6, 12, 0, 0, 0, time.Local))
	m := newTestModel(t, newTestHome(t), clock)
	m.width, m.height = 120, 40
	m.screenIndex = screen
	return m
}

func TestScreenSelectionsUseHoneyMarkerAndDefaultForeground(t *testing.T) {
	tests := []struct {
		name     string
		selected string
		also     []string
		honey    []string
		states   []string
		model    func(*testing.T) *Model
	}{
		{
			// The selected card has a bold accent name and a heavy border
			// (user decision 2026-10-21, option A).
			name: "Agents", honey: []string{"manager-alpha"}, states: []string{"◐ manager-alpha", "● worker-alpha", "┏━"},
			model: func(t *testing.T) *Model {
				m := agentsBoardModel(t)
				m.width, m.height = 120, 40
				m.graph.selected = "manager-alpha"
				// Decision 62: a card is selected in the one-pod view.
				m.graph.podView = true
				return m
			},
		},
		{
			// The Queue board draws the marker and the task in two runs.
			name: "Queue", honey: []string{"▶", "1-alpha"}, states: []string{"▶ 1-alpha", "assigned"},
			model: func(t *testing.T) *Model {
				m := selectionModel(t, 3)
				m.goalsQueue.data = goalsQueueData{
					projects: []records.Project{{ID: "demo"}},
					tasks:    []records.Task{{ID: "1-alpha", Project: "demo", Status: "assigned"}},
				}
				m.goalsQueue.selectedTaskID = "1-alpha"
				return m
			},
		},
		{
			name: "Tasks", honey: []string{"▶", "selected-task"}, states: []string{"assigned"},
			model: func(t *testing.T) *Model {
				m := selectionModel(t, 4)
				// Decision 8: the list panel has 13/30 of the width, so the
				// pane is 200 columns wide to show the Tasks columns.
				m.width = 200
				m.tasks.data = tasksData{items: []taskItem{{record: records.Task{ID: "selected-task", Title: "Selected", Project: "demo", Status: "assigned"}}}}
				m.tasks.selectedID = "selected-task"
				return m
			},
		},
		{
			name: "Friction", honey: []string{"▶", "issue", "[Open]"}, states: []string{"▶ issue", "status    Open"},
			model: func(t *testing.T) *Model {
				m := selectionModel(t, 5)
				m.friction.data.entries = []frictionEntry{{file: "issue.md", path: "entries/issue.md", title: "Open issue", status: "Open", opened: "2026-04-05", decision: "Keep", valid: true}}
				m.friction.selectedPath = "entries/issue.md"
				return m
			},
		},
		{
			// The Health board puts the check glyph between the marker and
			// the tile text, so the test reads the marker and the text
			// apart. health.md rule 2: the first tile that did not pass is
			// the first selection.
			name: "Health", honey: []string{"▶", "socket · unavailable"}, states: []string{"✓ /fixture", "▶ ! socket · unavailable"},
			model: func(t *testing.T) *Model {
				m := selectionModel(t, 6)
				m.health.accept(doctor.Result{Checks: []doctor.Check{
					{Name: "home", Status: doctor.StatusOK, Detail: "/fixture"},
					{Name: "herdr socket", Status: doctor.StatusFail, Detail: "unavailable"},
				}})
				return m
			},
		},
		{
			name: "Logs", honey: []string{"▶", "network retry"}, states: []string{"WARN  herdr     network retry"},
			model: func(t *testing.T) *Model {
				clock := fixedClock(t, time.Date(2026, 4, 6, 12, 0, 0, 0, time.Local))
				m := newTestLogsModel(t, newTestHome(t), clock)
				m.width, m.height = 120, 40
				m, _ = raise(t, m, screenLoadMsg{request: logsDataRequest, events: logsFixture(t, clock.now())})
				return m
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.model(t)
			m.invalidateView()
			raw := m.View().Content
			// The Logs board draws the space after ▶ in fg, so Logs
			// checks the marker and the message on their own.
			if tt.selected != "" {
				requireHoneySelection(t, raw, tt.selected)
			}
			for _, text := range tt.honey {
				requireHoneyBoldText(t, raw, text)
			}
			for _, selected := range tt.also {
				requireHoneySelection(t, raw, selected)
			}
			requireNoDefaultForeground(t, raw)
			plain := ansi.Strip(raw)
			for _, state := range tt.states {
				if !strings.Contains(plain, state) {
					t.Errorf("plain text omits state signal %q:\n%s", state, plain)
				}
			}
		})
	}
}
