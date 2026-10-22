package tui

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	chatlib "github.com/Harrison-Blair/hive/internal/chat"
	"github.com/Harrison-Blair/hive/internal/doctor"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/logs"
	"github.com/charmbracelet/x/ansi"
)

const (
	demoNotice    = "demo mode: no action"
	demoNoFile    = "demo mode: no file"
	minimumWidth  = 80
	minimumHeight = 24
)

// demoFixture is the schema of a <Board>.fixture.json file. It holds the data
// and the UI state of one design board.
type demoFixture struct {
	Board    string      `json:"board"`
	Terminal demoSize    `json:"terminal"`
	Clock    demoClock   `json:"clock"`
	Home     string      `json:"home"`
	Data     demoData    `json:"data"`
	UI       demoUIState `json:"ui"`

	now time.Time
}

type demoSize struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

type demoClock struct {
	Time string `json:"time"`
	Zone string `json:"zone"`
}

// demoData holds the records that the screens show. Each list uses the record
// type that hive stores.
type demoData struct {
	Supervisor demoSupervisor             `json:"supervisor"`
	Projects   []records.Project          `json:"projects"`
	Agents     []records.Agent            `json:"agents"`
	Goals      []records.Goal             `json:"goals"`
	Tasks      []records.Task             `json:"tasks"`
	Inbox      map[string][]records.Inbox `json:"inbox"`
	Evidence   map[string][]string        `json:"evidence"`
	Friction   []demoFriction             `json:"friction"`
	Checks     []doctor.Check             `json:"checks"`
	Chat       []chatlib.Entry            `json:"chat"`
	Logs       []json.RawMessage          `json:"logs"`
	Files      map[string]string          `json:"files"`
}

// demoSupervisor replaces the supervisor claim and its process check.
type demoSupervisor struct {
	Live      bool   `json:"live"`
	Agent     string `json:"agent"`
	PID       int64  `json:"pid"`
	BinaryOld bool   `json:"binary_old"`
}

// demoFriction holds the front matter and the body of one friction entry.
type demoFriction struct {
	File     string `json:"file"`
	Resolved bool   `json:"resolved"`
	Title    string `json:"title"`
	Status   string `json:"status"`
	Opened   string `json:"opened"`
	Decision string `json:"decision"`
	Body     string `json:"body"`
}

// demoUIState is the UI state of a board.
type demoUIState struct {
	Screen      string                `json:"screen"`
	View        string                `json:"view"`
	Focus       string                `json:"focus"`
	Selection   string                `json:"selection"`
	Scroll      map[string]demoScroll `json:"scroll"`
	Filter      demoFilter            `json:"filter"`
	Overlay     string                `json:"overlay"`
	Popup       string                `json:"popup"`
	Prompt      demoPrompt            `json:"prompt"`
	HideStopped bool                  `json:"hide_stopped"`
	StopsOnly   bool                  `json:"stops_only"`
	Follow      bool                  `json:"follow"`
	LogDay      string                `json:"log_day"`
	ChatNoise   string                `json:"chat_noise"`
	ChatInput   string                `json:"chat_input"`
}

type demoScroll struct {
	Top  int `json:"top"`
	Left int `json:"left"`
}

type demoFilter struct {
	Open bool   `json:"open"`
	Text string `json:"text"`
}

type demoPrompt struct {
	Kind   string `json:"kind"`
	Target string `json:"target"`
}

var (
	demoScreens = []string{"agents", "chat", "work", "friction", "health", "logs"}
	demoViews   = map[string][]string{
		"agents":   {""},
		"chat":     {""},
		"work":     {"goals", "queue", "tasks"},
		"friction": {"open", "resolved"},
		"health":   {"home", "agents"},
		"logs":     {""},
	}
	demoFocuses    = []string{"graph", "feed", "input", "list", "details", "filter", "events", "json"}
	demoOverlays   = []string{"", "details", "help"}
	demoPopups     = []string{"", "agent-details"}
	demoPrompts    = []string{"", "restart"}
	demoChatNoises = []string{"", "quiet", "normal", "all"}
)

// demoState holds the demo fixture and the draw size of a demo run.
type demoState struct {
	fixture    *demoFixture
	size       demoSize
	paneWidth  int
	paneHeight int
	notice     string
}

// loadDemoFixture reads and checks one fixture file. An unknown field fails
// the load with an error that names the field.
func loadDemoFixture(path string) (*demoFixture, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("fixture: %w", err)
	}
	fixture, err := parseDemoFixture(data)
	if err != nil {
		return nil, fmt.Errorf("fixture %s: %w", path, err)
	}
	return fixture, nil
}

func parseDemoFixture(data []byte) (*demoFixture, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	fixture := &demoFixture{}
	if err := decoder.Decode(fixture); err != nil {
		return nil, err
	}
	if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
		return nil, errors.New("data after the fixture object")
	}
	if err := fixture.check(); err != nil {
		return nil, err
	}
	return fixture, nil
}

func (f *demoFixture) check() error {
	if f.Board == "" {
		return errors.New(`field "board" is empty`)
	}
	if f.Terminal.Width < 1 || f.Terminal.Height < 1 {
		return fmt.Errorf(`field "terminal" is %dx%d; want a width and a height of 1 or more`, f.Terminal.Width, f.Terminal.Height)
	}
	if f.Home == "" {
		return errors.New(`field "home" is empty`)
	}
	zone, err := time.LoadLocation(f.Clock.Zone)
	if err != nil || f.Clock.Zone == "" {
		return fmt.Errorf(`field "clock.zone": unknown time zone %q`, f.Clock.Zone)
	}
	clock, err := time.Parse(time.RFC3339Nano, f.Clock.Time)
	if err != nil {
		return fmt.Errorf(`field "clock.time": %q is not an RFC 3339 time`, f.Clock.Time)
	}
	f.now = clock.In(zone)
	ui := f.UI
	if !slices.Contains(demoScreens, ui.Screen) {
		return fmt.Errorf(`field "ui.screen": unknown value %q`, ui.Screen)
	}
	if !slices.Contains(demoViews[ui.Screen], ui.View) {
		return fmt.Errorf(`field "ui.view": unknown value %q for screen %q`, ui.View, ui.Screen)
	}
	if ui.Focus != "" && !slices.Contains(demoFocuses, ui.Focus) {
		return fmt.Errorf(`field "ui.focus": unknown value %q`, ui.Focus)
	}
	for region, scroll := range ui.Scroll {
		if !slices.Contains(demoFocuses, region) {
			return fmt.Errorf(`field "ui.scroll": unknown region %q`, region)
		}
		if scroll.Top < 0 || scroll.Left < 0 {
			return fmt.Errorf(`field "ui.scroll.%s": negative position`, region)
		}
	}
	if !slices.Contains(demoOverlays, ui.Overlay) {
		return fmt.Errorf(`field "ui.overlay": unknown value %q`, ui.Overlay)
	}
	if !slices.Contains(demoPopups, ui.Popup) {
		return fmt.Errorf(`field "ui.popup": unknown value %q`, ui.Popup)
	}
	if !slices.Contains(demoPrompts, ui.Prompt.Kind) {
		return fmt.Errorf(`field "ui.prompt.kind": unknown value %q`, ui.Prompt.Kind)
	}
	if !slices.Contains(demoChatNoises, ui.ChatNoise) {
		return fmt.Errorf(`field "ui.chat_noise": unknown value %q`, ui.ChatNoise)
	}
	if ui.LogDay != "" {
		if _, err := time.Parse("2006-01-02", ui.LogDay); err != nil {
			return fmt.Errorf(`field "ui.log_day": %q is not a date`, ui.LogDay)
		}
	}
	for index, raw := range f.Data.Logs {
		var fields map[string]any
		if err := json.Unmarshal(raw, &fields); err != nil {
			return fmt.Errorf(`field "data.logs[%d]": %w`, index, err)
		}
	}
	return nil
}

// ParseDemoSize parses a --size value of the form WxH.
func ParseDemoSize(value string) (int, int, error) {
	widthText, heightText, found := strings.Cut(value, "x")
	width, widthErr := strconv.Atoi(widthText)
	height, heightErr := strconv.Atoi(heightText)
	if !found || widthErr != nil || heightErr != nil || width < 1 || height < 1 {
		return 0, 0, fmt.Errorf("invalid --size %q: want WxH, for example 120x40", value)
	}
	return width, height, nil
}

// useDemoLoaders replaces each loader with one that reads only the fixture.
// The screen renderers stay the same.
func (m *Model) useDemoLoaders(fixture *demoFixture) {
	data := fixture.Data
	m.headerLoader = func(string) (headerState, error) {
		return buildHeader(data.Supervisor.Live, data.Supervisor.BinaryOld, data.Projects, data.Agents, data.Tasks), nil
	}
	m.agentsLoader = func(_ string, now time.Time) (agentsData, error) {
		unread := make(map[string][]records.Inbox, len(data.Inbox))
		for name, items := range data.Inbox {
			for _, item := range items {
				if item.ReadAt == "" {
					unread[name] = append(unread[name], item)
				}
			}
		}
		return buildAgents(agentsSource{
			projects:        data.Projects,
			goals:           data.Goals,
			tasks:           data.Tasks,
			agents:          data.Agents,
			supervisorLive:  data.Supervisor.Live,
			supervisorClaim: records.Claim{Key: "supervisor", Agent: data.Supervisor.Agent, PID: data.Supervisor.PID},
			unread:          unread,
		}, now), nil
	}
	m.goalsQueueLoader = func(string) (goalsQueueData, error) {
		return buildGoalsQueue(slices.Clone(data.Goals), slices.Clone(data.Projects), slices.Clone(data.Tasks), slices.Clone(data.Agents)), nil
	}
	m.tasksLoader = func(string) (tasksData, error) {
		return buildTasks(data.Tasks, data.Projects, data.Goals, func(item taskItem) ([]string, error) {
			return slices.Clone(data.Evidence[item.record.ID]), nil
		})
	}
	m.frictionLoader = func(string) (frictionData, error) {
		entries := make([]frictionEntry, 0, len(data.Friction))
		for _, item := range data.Friction {
			path := "entries/" + item.File
			if item.Resolved {
				path = "entries/resolved/" + item.File
			}
			entries = append(entries, frictionEntry{
				file: item.File, path: path, title: item.Title, status: item.Status, opened: item.Opened,
				decision: item.Decision, body: item.Body, sections: frictionSections(strings.Split(item.Body, "\n")),
				valid: true, resolved: item.Resolved,
			})
		}
		sortFrictionEntries(entries)
		return frictionData{entries: entries}, nil
	}
	m.healthLoader = func(string, *slog.Logger) doctor.Result {
		return doctor.Result{Checks: slices.Clone(data.Checks)}
	}
	m.chatLoader = func(_ string, now time.Time) (chatScreenData, error) {
		coordinators := []records.Agent{}
		for _, record := range data.Agents {
			if record.Role == "coordinator" {
				coordinators = append(coordinators, record)
			}
		}
		return buildChatScreenData(data.Chat, slices.Clone(data.Inbox["coordinator"]), coordinators, now), nil
	}
	m.logLoader = func(*logs.Reader, time.Time) ([]logs.Event, error) {
		events := make([]logs.Event, 0, len(data.Logs))
		for _, raw := range data.Logs {
			var fields map[string]any
			if err := json.Unmarshal(raw, &fields); err != nil {
				return nil, err
			}
			events = append(events, logs.Event{Raw: slices.Clone([]byte(raw)), Fields: fields})
		}
		return events, nil
	}
}

// applyDemoUI sets the parts of the fixture UI state that the current screens
// have.
func (m *Model) applyDemoUI(ui demoUIState) {
	switch ui.Screen {
	case "chat":
		m.screenIndex = 0
	case "agents":
		m.screenIndex = 1
		m.agents.selectedKey = ui.Selection
	case "work":
		switch ui.View {
		case "goals":
			m.screenIndex = 2
			m.goalsQueue.selectedGoal = ui.Selection
		case "queue":
			m.screenIndex = 3
			m.goalsQueue.selectedTaskID = ui.Selection
		default:
			m.screenIndex = 4
			m.tasks.selectedID = ui.Selection
		}
	case "friction":
		m.screenIndex = 5
		path := "entries/" + ui.Selection
		if ui.View == "resolved" {
			m.friction.tab = frictionResolved
			path = "entries/resolved/" + ui.Selection
		}
		if ui.Selection != "" {
			m.friction.selectedPath = path
		}
	case "health":
		m.screenIndex = 6
		m.health.hideStopped = ui.HideStopped
		if ui.View == "agents" {
			m.health.selectedAgent = ui.Selection
		} else {
			m.health.selected = ui.Selection
		}
	case "logs":
		m.screenIndex = 7
		m.textFilter = ui.Filter.Text
		m.logPendingTime = ui.Selection
		m.logOffset = ui.Scroll["events"].Top
		m.detailsOffset[tabLogs] = ui.Scroll["json"].Top
		if ui.Filter.Open {
			m.mode = editing
			m.filterDraft = ui.Filter.Text
		}
	}
	m.agents.showHistory = !ui.HideStopped
	m.goalsQueue.showStops = ui.StopsOnly
	m.follow = ui.Follow
	if ui.LogDay != "" {
		day, _ := time.ParseInLocation("2006-01-02", ui.LogDay, m.now().Location())
		m.logDate = day
		m.logDatePinned = true
	}
	if ui.ChatInput != "" {
		m.chatInput.SetValue(ui.ChatInput)
	}
	if ui.Screen == "chat" && ui.Focus == "input" {
		m.chatInputActive = true
		m.chatInput.Focus()
	}
	m.applyDemoFrame(ui)
}

// applyDemoFrame sets the tab, the view, the focus, the overlay, the popup,
// and the restart prompt of the frame from the fixture UI state.
func (m *Model) applyDemoFrame(ui demoUIState) {
	tab, view := m.activeTab(), 0
	switch {
	case tab == tabWork:
		view = m.screenIndex - 2
	case tab == tabFriction && ui.View == "resolved":
		view = 1
	case ui.Screen == "health" && ui.View == "agents":
		tab, view = tabHealth, 1
	}
	m.selectScreen(tab, view)
	if index := slices.Index(screenRegistry[tab].regions, ui.Focus); index > 0 {
		m.regionFocus[tab] = index
	}
	m.helpOpen = ui.Overlay == "help"
	m.popupOpen = ui.Popup == "agent-details"
	if ui.Prompt.Kind == "restart" {
		m.agents.restartConfirm = true
		m.agents.restartTarget = records.Agent{Key: ui.Prompt.Target}
		m.agents.restartNotice = "restart " + ui.Prompt.Target + "? y/n"
	}
}

// errDemoNoFile tells that the fixture has no text for a file.
var errDemoNoFile = errors.New(demoNoFile)

// readScreenFile reads one file that a screen opens. Demo mode reads only the
// fixture text. The fixture key is the path relative to the home or to a
// project directory.
func (m *Model) readScreenFile(path string) (string, error) {
	if m.demo == nil {
		data, err := os.ReadFile(path)
		return string(data), err
	}
	roots := []string{m.home}
	for _, project := range m.demo.fixture.Data.Projects {
		roots = append(roots, project.Path)
	}
	for _, root := range roots {
		if rel, err := filepath.Rel(root, path); err == nil && filepath.IsLocal(rel) {
			if text, ok := m.demo.fixture.Data.Files[filepath.ToSlash(rel)]; ok {
				return text, nil
			}
		}
	}
	return "", errDemoNoFile
}

// demoBlocked shows the demo notice in place of an action.
func (m *Model) demoBlocked(action string) tea.Cmd {
	m.demo.notice = demoNotice
	m.invalidateView()
	return m.logAction("demo action blocked", slog.String("blocked", action))
}

// demoResize keeps the pane size and draws at the --size value when the pane
// can hold it.
func (m *Model) demoResize(width, height int) {
	m.demo.paneWidth, m.demo.paneHeight = width, height
	if m.demo.size.Width > 0 && width >= m.demo.size.Width && height >= m.demo.size.Height {
		width, height = m.demo.size.Width, m.demo.size.Height
	}
	m.width, m.height = width, height
}

// demoTooSmall gives the size to show and the size that hive needs, when the
// demo cannot draw the TUI.
func (m *Model) demoTooSmall() (demoSize, demoSize, bool) {
	pane := demoSize{Width: m.demo.paneWidth, Height: m.demo.paneHeight}
	if size := m.demo.size; size.Width > 0 && (pane.Width < size.Width || pane.Height < size.Height) {
		return pane, size, true
	}
	draw := demoSize{Width: m.width, Height: m.height}
	if draw.Width < minimumWidth || draw.Height < minimumHeight {
		return draw, demoSize{Width: minimumWidth, Height: minimumHeight}, true
	}
	return demoSize{}, demoSize{}, false
}

// tooSmallView draws the too-small notice of the TooSmall board.
func tooSmallView(pane, need demoSize) string {
	width, height := pane.Width, pane.Height
	plain := lipgloss.NewStyle()
	accent := lipgloss.NewStyle().Foreground(lipgloss.Color(colorAccent))
	type cell struct {
		r     rune
		style *lipgloss.Style
	}
	grid := make([][]cell, height)
	for y := range grid {
		grid[y] = make([]cell, width)
		for x := range grid[y] {
			grid[y][x] = cell{' ', &plain}
		}
	}
	put := func(x, y int, text string, style *lipgloss.Style) {
		for _, r := range text {
			if y >= 0 && y < height && x >= 0 && x < width {
				grid[y][x] = cell{r, style}
			}
			x++
		}
	}
	center := func(y int, text string, style *lipgloss.Style) {
		put((width-lipgloss.Width(text))/2, y, text, style)
	}
	const boxWidth, boxHeight = 42, 8
	left, top := (width-boxWidth)/2, (height-boxHeight)/2
	put(left, top, "╭"+strings.Repeat("─", boxWidth-2)+"╮", &accent)
	for y := 1; y < boxHeight-1; y++ {
		put(left, top+y, "│", &accent)
		put(left+boxWidth-1, top+y, "│", &accent)
	}
	put(left, top+boxHeight-1, "╰"+strings.Repeat("─", boxWidth-2)+"╯", &accent)
	put(left+2, top, " terminal too small ", &titleStyle)
	center(top+2, fmt.Sprintf("this pane is %d×%d", pane.Width, pane.Height), &plain)
	center(top+3, fmt.Sprintf("hive needs at least %d×%d", need.Width, need.Height), &plain)
	center(top+5, "resize or zoom the herdr pane", &dimStyle)
	put(0, 0, " hive", &headerStyle)
	put(0, height-1, " ctrl+c ×2 quit", &dimStyle)
	rows := make([]string, height)
	for y, row := range grid {
		var line strings.Builder
		for x := 0; x < len(row); {
			end := x
			var run strings.Builder
			for end < len(row) && row[end].style == row[x].style {
				run.WriteRune(row[end].r)
				end++
			}
			line.WriteString(row[x].style.Render(run.String()))
			x = end
		}
		rows[y] = line.String()
	}
	return strings.Join(rows, "\n")
}

// clipView cuts a view to the top-left width×height cells.
func clipView(view string, width, height int) string {
	lines := strings.Split(view, "\n")
	if len(lines) > height {
		lines = lines[:height]
	}
	for index, line := range lines {
		lines[index] = ansi.Truncate(line, width, "")
	}
	return strings.Join(lines, "\n")
}

// RunDemo starts the TUI in demo mode. It reads all screen data from the
// fixture file, uses the fixture clock, and starts no agent and no process.
func RunDemo(home, fixturePath, size string, in io.Reader, out io.Writer) error {
	state := &demoState{}
	if size != "" {
		width, height, err := ParseDemoSize(size)
		if err != nil {
			return err
		}
		state.size = demoSize{Width: width, Height: height}
	}
	fixture, err := loadDemoFixture(fixturePath)
	if err != nil {
		return err
	}
	state.fixture = fixture
	model := newDemoModel(home, state)
	defer model.Stop()
	for _, step := range []string{"current tab label", "coordinator", "supervisor"} {
		logStartupStep(model.logger, step, "skipped", nil, "reason", "demo")
	}
	return runProgram(model, in, out, fixture.Terminal.Width, fixture.Terminal.Height)
}

// newDemoModel builds the demo model with the fixture loaders, the fixed
// fixture clock, and the fixture UI state.
func newDemoModel(home string, state *demoState) *Model {
	now := state.fixture.now
	m := newModel(home, func() time.Time { return now }, state)
	m.demoResize(state.fixture.Terminal.Width, state.fixture.Terminal.Height)
	m.layout()
	// The demo loaders read only memory, so the model holds its data before
	// the first frame and the first key.
	for _, load := range []tea.Cmd{m.startHeaderLoad(), m.startCurrentScreenLoad()} {
		m.Update(load())
	}
	return m
}
