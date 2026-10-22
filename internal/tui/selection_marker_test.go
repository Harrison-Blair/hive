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
		states   []string
		// bold is honey bold text of a selection that has the space after
		// ▶ in fg (user decision 7 of 2026-10-21).
		bold  []string
		model func(*testing.T) *Model
	}{
		{
			name: "Chat", selected: "noise: landings and decisions", states: []string{"not live"},
			model: func(t *testing.T) *Model {
				m := selectionModel(t, 0)
				return m
			},
		},
		{
			name: "Agents", selected: "● coordinator", also: []string{"[ GRAPH ]"}, states: []string{"● coordinator", "○ supervisor"},
			model: func(t *testing.T) *Model {
				m := agentsBoardModel(t)
				m.width, m.height = 120, 40
				m.screenIndex = 1
				m.agents.tab = agentsGraph
				m.agents.selectedKey = "coordinator"
				return m
			},
		},
		{
			name: "Goals", selected: "demo · Selected goal", states: []string{"done"},
			model: func(t *testing.T) *Model {
				m := selectionModel(t, 2)
				m.goalsQueue.data.goals = []records.Goal{{ID: "g1-demo", Project: "demo", Title: "Selected goal", Status: "done"}}
				m.goalsQueue.selectedGoal = "g1-demo"
				return m
			},
		},
		{
			name: "Queue", selected: "now  1-alpha", states: []string{"▶ assigned"},
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
			name: "Tasks", selected: "selected-task", also: []string{"[all 1]", "project: all"}, states: []string{"assigned"},
			model: func(t *testing.T) *Model {
				m := selectionModel(t, 4)
				// Decision 8: the list panel has 13/30 of the width, so the
				// pane is 200 columns wide to show the Tasks columns.
				m.width = 200
				m.tasks.data = tasksData{items: []taskItem{{record: records.Task{ID: "selected-task", Title: "Selected", Project: "demo", Status: "assigned"}}}, projects: []string{"demo"}}
				m.tasks.tab = tasksAll
				m.tasks.selectedID = "selected-task"
				return m
			},
		},
		{
			name: "Friction", bold: []string{"▶", "issue", "[Open]"}, states: []string{"▶ issue", "status    Open"},
			model: func(t *testing.T) *Model {
				m := selectionModel(t, 5)
				m.friction.data.entries = []frictionEntry{{file: "issue.md", path: "entries/issue.md", title: "Open issue", status: "Open", opened: "2026-04-05", decision: "Keep", valid: true}}
				m.friction.selectedPath = "entries/issue.md"
				return m
			},
		},
		{
			name: "Health", selected: "home", states: []string{"✓", "✕ herdr socket"},
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
			name: "Logs", selected: "12:00:03 WARN herdr network retry", also: []string{"level: info"}, states: []string{"WARN herdr network retry"},
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
			if tt.selected != "" {
				requireHoneySelection(t, raw, tt.selected)
			}
			for _, text := range tt.bold {
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

func TestAgentBoardShowsDistinctStatusColorAndDimWorkingCount(t *testing.T) {
	m := agentsBoardModel(t)
	lines, _ := m.agents.explorerLines()
	joined := strings.Join(lines, "\n")
	for _, text := range []string{"alpha · 1 working", "● coordinator", "○ supervisor", "● worker-alpha"} {
		if !strings.Contains(ansi.Strip(joined), text) {
			t.Fatalf("Agents explorer omits plain-text signal %q:\n%s", text, ansi.Strip(joined))
		}
	}
	var dotLine, countLine string
	for _, line := range lines {
		plain := ansi.Strip(line)
		if strings.Contains(plain, "● worker-alpha") {
			dotLine = line
		}
		if strings.Contains(plain, "alpha · 1 working") {
			countLine = line
		}
	}
	plainDot := ansi.Strip(dotLine)
	if got := foregroundAtCell(dotLine, ansi.StringWidth(plainDot[:strings.Index(plainDot, "●")])); got != "38;2;110;230;154" {
		t.Errorf("live explorer status foreground = %q, want working green", got)
	}
	plainCount := ansi.Strip(countLine)
	countIndex := strings.Index(plainCount, "1 working")
	if got := foregroundAtCell(countLine, ansi.StringWidth(plainCount[:countIndex])); got != "38;2;163;157;144" {
		t.Errorf("explorer working count foreground = %q, want dim", got)
	}
}

func TestAgentGraphDetailStripNameIsHoneyBold(t *testing.T) {
	m := agentsBoardModel(t)
	m.width = 120
	m.agents.tab = agentsGraph
	m.agents.selectedKey = "coordinator"
	view := agentsScreen{model: m}.View(120, 40)
	lines := strings.Split(view, "\n")
	detail := lines[len(lines)-2]
	if !strings.Contains(ansi.Strip(detail), "coordinator") {
		t.Fatalf("Agents detail strip omits the selected agent name: %s", ansi.Strip(detail))
	}
	requireHoneyBoldText(t, detail, "coordinator")
}

func TestAgentGraphBoardRendersTabsLegendHeaderAndDetails(t *testing.T) {
	m := agentsBoardModel(t)
	m.width = 120
	// The AgentsMany board shows a needs-you count above 0 in accent. A
	// count of 0 uses fg, as the Main board does.
	m.header = headerState{supervisor: "fresh", projects: 2, agents: 6, working: 6, needsUser: 1}
	m.agents.tab = agentsGraph
	m.agents.selectedKey = "coordinator"
	view := agentsScreen{model: m}.View(120, 40)
	lines := strings.Split(view, "\n")
	plain := ansi.Strip(lines[0])
	markerIndex := strings.Index(plain, "▶")
	if markerIndex < 0 || !strings.Contains(plain, "▶ [ GRAPH ]") || !strings.Contains(plain, "TREE") || !strings.Contains(plain, "TABLE") {
		t.Fatalf("Agents tab line lacks separate selected state:\n%s", lines[0])
	}
	if got := foregroundAtCell(lines[0], ansi.StringWidth(plain[:markerIndex])); got != wantHoney {
		t.Errorf("selected tab foreground = %q, want honey", got)
	}
	if got := foregroundAtCell(lines[0], ansi.StringWidth(plain[:strings.Index(plain, "TREE")])); got != "38;2;163;157;144" {
		t.Errorf("unselected tab foreground = %q, want dim", got)
	}
	legend := "edges: ─ delegates · ┄ nudges · ⋯ replaced"
	if !strings.Contains(plain, legend) || strings.Index(plain, legend) < strings.Index(plain, "TABLE") {
		t.Errorf("Agents legend does not follow and align after the tabs:\n%s", lines[0])
	}
	if ansi.StringWidth(plain[:strings.Index(plain, legend)])+ansi.StringWidth(legend) != ansi.StringWidth(plain) {
		t.Errorf("Agents legend is not right-aligned:\n%s", lines[0])
	}
	if got := foregroundAtCell(lines[0], ansi.StringWidth(plain[:strings.Index(plain, "edges:")])); got != "38;2;163;157;144" {
		t.Errorf("Agents legend foreground = %q, want dim", got)
	}
	// Decision 2: the header shows the state dots and the needs-you count.
	header := m.headerView().View(m.width)
	plainHeader := ansi.Strip(header)
	if got := foregroundAtCell(header, ansi.StringWidth(plainHeader[:strings.Index(plainHeader, "●")])); got != "38;2;110;230;154" {
		t.Errorf("Agents live supervisor dot foreground = %q, want working green", got)
	}
	needsIndex := strings.Index(plainHeader, "needs you ") + len("needs you ")
	if got := foregroundAtCell(header, ansi.StringWidth(plainHeader[:needsIndex])); got != wantHoney {
		t.Errorf("Agents needs-you count foreground = %q, want honey", got)
	}
	plainView := ansi.Strip(view)
	if !strings.Contains(plainView, "coordinator  parent ?") {
		t.Errorf("Agents detail strip omits selected agent state:\n%s", plainView)
	}
	_, positions := m.agents.graphLayout()
	selected := positions[m.agents.selectedKey]
	if got := foregroundAtCell(lines[1+selected.row-1-m.agents.scrollOffset], selected.col-m.agents.graphOffset); got != wantHoney {
		t.Errorf("selected node border foreground = %q, want honey", got)
	}
	if got := foregroundAtCell(lines[len(lines)-3], 0); got != wantHoney {
		t.Errorf("Agents detail strip border foreground = %q, want honey", got)
	}
	detail := lines[len(lines)-2]
	plainDetail := ansi.Strip(detail)
	parent := strings.Index(plainDetail, "parent")
	if got := foregroundAtCell(detail, ansi.StringWidth(plainDetail[:parent])); got != "38;2;163;157;144" {
		t.Errorf("Agents detail label foreground = %q, want dim", got)
	}
	if got := foregroundAtCell(detail, ansi.StringWidth(plainDetail[:parent+len("parent ")])); got != "" {
		t.Errorf("Agents detail value foreground = %q, want terminal default", got)
	}
	requireHoneyBoldText(t, view, "coordinator")
}
