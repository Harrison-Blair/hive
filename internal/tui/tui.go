package tui

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"sync"
	"time"

	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	chatlib "github.com/Harrison-Blair/hive/internal/chat"
	"github.com/Harrison-Blair/hive/internal/doctor"
	"github.com/Harrison-Blair/hive/internal/lib/herdr"
	"github.com/Harrison-Blair/hive/internal/lib/home"
	hivelog "github.com/Harrison-Blair/hive/internal/lib/log"
	"github.com/Harrison-Blair/hive/internal/logs"
	xterm "github.com/charmbracelet/x/term"
)

const (
	defaultWidth  = 80
	defaultHeight = 24
	statusRows    = 3
)

type mode int

const (
	viewing mode = iota
	editing
)

// Model is the root Bubble Tea model for the eight-screen shell.
type Model struct {
	home   string
	now    func() time.Time
	reader *logs.Reader
	logger *slog.Logger
	closer io.Closer

	events   []logs.Event
	viewport viewport.Model
	input    textinput.Model
	mode     mode

	chatData           chatScreenData
	chatViewport       viewport.Model
	chatInput          textarea.Model
	chatInputActive    bool
	chatSending        bool
	chatCallerCwd      string
	chatPaneID         string
	chatCallerBlocked  bool
	chatEventsExpanded bool
	chatNoise          string
	chatFollow         bool
	chatErr            string
	level              string
	componentFilter    string
	projectFilter      string
	agentFilter        string
	requestFilter      string
	textFilter         string
	follow             bool
	logDate            time.Time
	logDatePinned      bool
	logSelection       int
	logSelectedKey     string
	logDetail          bool
	logErr             string

	header           headerState
	headerErr        string
	screenErr        string
	attachErr        string
	startupNotice    string
	startupErr       string
	headerLoader     func(string) (headerState, error)
	logLoader        func(*logs.Reader, time.Time) ([]logs.Event, error)
	healthLoader     func(string, *slog.Logger) doctor.Result
	agentsLoader     func(string, time.Time) (agentsData, error)
	goalsQueueLoader func(string) (goalsQueueData, error)
	paneReader       func(context.Context, string) (string, error)
	agents           agentsState
	health           healthState
	goalsQueue       goalsQueueState
	tasksLoader      func(string) (tasksData, error)
	tasks            tasksState
	frictionLoader   func(string) (frictionData, error)
	friction         frictionState
	chatLoader       func(string, time.Time) (chatScreenData, error)
	headerLoading    bool
	screenLoading    bool

	screens     []screen
	screenIndex int
	quitNotice  bool
	quitAt      time.Time

	inputEnded <-chan struct{}
	width      int
	height     int
}

// New builds the TUI for one resolved home and one clock.
func New(home string, now func() time.Time) *Model {
	if now == nil {
		now = time.Now
	}
	m := &Model{
		home:         home,
		now:          now,
		level:        "info",
		logDate:      startOfDay(now().Local()),
		follow:       true,
		chatNoise:    "messages",
		chatFollow:   true,
		chatLoader:   loadChatScreenData,
		width:        defaultWidth,
		height:       defaultHeight,
		header:       headerState{supervisor: "stale"},
		agents:       agentsState{hideStopped: true},
		headerLoader: readHeader,
		logLoader: func(_ *logs.Reader, day time.Time) ([]logs.Event, error) {
			dayReader := logs.NewReaderWithClock(home, func() time.Time { return day })
			events, _, err := dayReader.Read(context.Background(), time.Time{})
			return events, err
		},
		agentsLoader:     loadAgents,
		healthLoader:     loadHealth,
		goalsQueueLoader: loadGoalsQueue,
		paneReader:       readAgentPane,
		tasksLoader:      loadTasks,
		frictionLoader:   loadFriction,
	}
	if callerCwd, err := os.Getwd(); err != nil {
		m.chatCallerBlocked = true
		m.chatErr = boundedError("chat send is for the user only: rule 2 check failed: " + err.Error())
	} else {
		m.chatCallerCwd = callerCwd
		m.chatPaneID = os.Getenv("HERDR_PANE_ID")
		if err := chatlib.ValidateCaller(context.Background(), home, "", m.chatPaneID, callerCwd); err != nil {
			m.chatCallerBlocked = true
			m.chatErr = boundedError(err.Error())
		}
	}
	m.reader = logs.NewReaderWithClock(home, now)
	m.viewport = viewport.New(viewport.WithWidth(1), viewport.WithHeight(1))
	m.viewport.SoftWrap = false
	m.chatViewport = viewport.New(viewport.WithWidth(1), viewport.WithHeight(1))
	m.chatViewport.SoftWrap = true
	m.chatInput = newChatInput()
	m.input = textinput.New()
	m.input.Prompt = "filter: "
	m.input.CharLimit = 0
	m.input.Blur()
	m.screens = m.makeScreens()

	writer, err := hivelog.NewWriter(home, "tui", now)
	if err != nil {
		m.logErr = boundedError("log: " + err.Error())
	} else {
		m.closer = writer
		m.logger = slog.New(writer)
	}
	m.layout()
	return m
}

// Stop closes the TUI log writer.
func (m *Model) Stop() {
	if m.closer != nil {
		_ = m.closer.Close()
	}
}

// Init starts the first data loads and the refresh tick.
func (m *Model) Init() tea.Cmd {
	return batch(m.tick(), m.startHeaderLoad(), m.startCurrentScreenLoad())
}

// Update handles one message.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.layout()
		return m, nil
	case tickMsg:
		if m.follow && !m.logDatePinned && !sameDay(m.logDate, m.now()) {
			m.logDate = startOfDay(m.now().Local())
			m.events = nil
			m.logSelectedKey = ""
			m.logSelection = 0
			m.rebuild()
		}
		if m.inputEnded != nil {
			select {
			case <-m.inputEnded:
				return m, tea.Quit
			default:
			}
		}
		cmds := []tea.Cmd{m.tick(), m.startHeaderLoad(), m.startCurrentScreenLoad()}
		if m.follow && m.screenIndex == 7 && m.screenErr == "" {
			m.viewport.GotoBottom()
		}
		return m, batch(cmds...)
	case headerLoadMsg:
		m.headerLoading = false
		if msg.err != nil {
			m.headerErr = boundedError(msg.err.Error())
			m.layout()
			return m, m.logError("header", msg.err)
		}
		m.header = msg.state
		m.headerErr = ""
		m.layout()
		return m, nil
	case screenLoadMsg:
		m.screenLoading = false
		if msg.err != nil {
			m.screenErr = boundedError(msg.err.Error())
			return m, m.logError("screen", msg.err)
		}
		m.screenErr = ""
		switch msg.request {
		case agentsDataRequest:
			m.agents.accept(msg.agents)
		case goalsQueueDataRequest:
			m.goalsQueue.accept(msg.goalsQueue)
		case healthDataRequest:
			m.health.accept(msg.health)
		case chatDataRequest:
			m.chatData = msg.chat
			m.rebuildChat()
			if m.chatFollow {
				m.chatViewport.GotoBottom()
			}
		case tasksDataRequest:
			m.tasks.accept(msg.tasks)
		case frictionDataRequest:
			m.friction.accept(msg.friction)
		case logsDataRequest:
			if !msg.day.IsZero() && !sameDay(msg.day, m.logDate) {
				return m, m.startCurrentScreenLoad()
			}
			m.acceptLogEvents(msg.events)
			if m.follow {
				m.viewport.GotoBottom()
			}
		default:
			return m, nil
		}
		return m, nil
	case taskFileMsg:
		m.acceptTaskFile(msg)
		if msg.err != nil {
			return m, m.logError("task file read", msg.err)
		}
		return m, nil
	case frictionFileMsg:
		m.acceptFrictionFile(msg)
		if msg.err != nil {
			return m, m.logError("friction file read", msg.err)
		}
		return m, nil
	case chatSendMsg:
		return m, m.updateChatSend(msg)
	case quitNoticeExpiredMsg:
		if m.quitNotice && msg.at.Equal(m.quitAt) {
			m.quitNotice = false
		}
		return m, nil
	case agentsPeekMsg:
		m.agents.peekOpen = true
		m.agents.peekPane = msg.paneID
		m.agents.peekText = msg.text
		if msg.err != nil {
			m.agents.peekText = "pane read: " + boundedError(msg.err.Error())
			return m, m.logError("pane read", msg.err)
		}
		return m, nil
	case attachMsg:
		if msg.err != nil {
			m.attachErr = boundedError(msg.err.Error())
			return m, m.logError("attach", msg.err)
		}
		m.attachErr = ""
		return m, nil
	case logWriteErrMsg:
		if msg.err != nil && m.logErr == "" {
			m.logErr = boundedError(msg.err.Error())
		}
		return m, nil
	case tea.KeyPressMsg:
		return m, m.updateKey(msg)
	}
	return m, nil
}

func (m *Model) updateKey(press tea.KeyPressMsg) tea.Cmd {
	if press.String() == "ctrl+c" {
		now := m.now()
		if m.quitNotice && !now.Before(m.quitAt) && now.Sub(m.quitAt) <= 3*time.Second {
			m.quitNotice = false
			return tea.Quit
		}
		m.quitNotice = true
		m.quitAt = now
		expires := tea.Tick(3*time.Second, func(time.Time) tea.Msg { return quitNoticeExpiredMsg{at: now} })
		return batch(m.logAction("quit request"), expires)
	}
	m.quitNotice = false
	if m.screenIndex == 0 && m.chatInputActive {
		return m.currentScreen().Update(press)
	}
	if m.mode != editing && len(press.String()) == 1 && press.String()[0] >= '1' && press.String()[0] <= '8' {
		index := int(press.String()[0] - '1')
		m.screenIndex = index
		return batch(m.logAction("screen select", slog.Int("screen", index+1)), m.startCurrentScreenLoad())
	}
	return m.currentScreen().Update(press)
}

// View renders the header, sidebar, current screen, and footer.
func (m *Model) View() tea.View {
	view := tea.NewView(m.shellView())
	view.AltScreen = true
	return view
}

func (m *Model) layout() {
	width := m.width
	if width < 1 {
		width = defaultWidth
	}
	headerRows := lipgloss.Height(headerStyle.Width(width).Render(m.headerLine()))
	bodyHeight := m.height - headerRows - statusRows
	if bodyHeight < 1 {
		bodyHeight = 1
	}
	sidebarWidth := 24
	if width >= 100 {
		sidebarWidth = 30
	}
	if sidebarWidth > width-12 {
		sidebarWidth = width - 12
	}
	if sidebarWidth < 12 {
		sidebarWidth = 12
	}
	mainWidth := width - sidebarWidth
	contentWidth := mainWidth - 4
	contentHeight := bodyHeight - 3
	chatEventWidth := mainWidth / 3
	if chatEventWidth < 18 {
		chatEventWidth = 18
	}
	if chatEventWidth > 32 {
		chatEventWidth = 32
	}
	if chatEventWidth > mainWidth-12 {
		chatEventWidth = mainWidth - 12
	}
	chatWidth := mainWidth - chatEventWidth - 3 - 4
	chatHeight := bodyHeight - 7
	if chatWidth < 1 {
		chatWidth = 1
	}
	if chatHeight < 1 {
		chatHeight = 1
	}
	m.chatViewport.SetWidth(chatWidth)
	m.chatViewport.SetHeight(chatHeight)
	m.chatInput.SetWidth(chatWidth)
	m.chatInput.SetHeight(4)
	if contentWidth < 1 {
		contentWidth = 1
	}
	if contentHeight < 1 {
		contentHeight = 1
	}
	m.viewport.SetWidth(contentWidth)
	m.viewport.SetHeight(contentHeight)
	if m.follow {
		m.viewport.GotoBottom()
	}
}

func (m *Model) startHeaderLoad() tea.Cmd {
	if m.headerLoading {
		return nil
	}
	m.headerLoading = true
	loader, home := m.headerLoader, m.home
	return func() tea.Msg {
		state, err := loader(home)
		return headerLoadMsg{state: state, err: err}
	}
}

func (m *Model) startCurrentScreenLoad() tea.Cmd {
	if m.screenLoading {
		return nil
	}
	request := m.currentScreen().DataRequest()
	m.screenLoading = true
	switch request {
	case agentsDataRequest:
		loader, home, now := m.agentsLoader, m.home, m.now
		return func() tea.Msg {
			data, err := loader(home, now())
			return screenLoadMsg{request: agentsDataRequest, agents: data, err: err}
		}
	case chatDataRequest:
		loader, home, now := m.chatLoader, m.home, m.now
		return func() tea.Msg {
			data, err := loader(home, now())
			return screenLoadMsg{request: chatDataRequest, chat: data, err: err}
		}
	case goalsQueueDataRequest:
		loader, home := m.goalsQueueLoader, m.home
		return func() tea.Msg {
			data, err := loader(home)
			return screenLoadMsg{request: goalsQueueDataRequest, goalsQueue: data, err: err}
		}
	case healthDataRequest:
		loader, home, logger := m.healthLoader, m.home, m.logger
		return func() tea.Msg {
			return screenLoadMsg{request: healthDataRequest, health: loader(home, logger)}
		}
	case tasksDataRequest:
		loader, home := m.tasksLoader, m.home
		return func() tea.Msg {
			data, err := loader(home)
			return screenLoadMsg{request: tasksDataRequest, tasks: data, err: err}
		}
	case frictionDataRequest:
		loader, home := m.frictionLoader, m.home
		return func() tea.Msg {
			data, err := loader(home)
			return screenLoadMsg{request: frictionDataRequest, friction: data, err: err}
		}
	default:
		loader, reader, day := m.logLoader, m.reader, startOfDay(m.logDate.Local())
		return func() tea.Msg {
			events, err := loader(reader, day)
			if err == nil {
				events = eventsOnDay(events, day)
			}
			return screenLoadMsg{request: logsDataRequest, day: day, events: events, err: err}
		}
	}
}

func (m *Model) attachCmd(paneID string) tea.Cmd {
	return func() tea.Msg {
		_, err := herdr.AgentFocus(context.Background(), paneID)
		return attachMsg{err: err}
	}
}

func (m *Model) startAgentsPeek(paneID string) tea.Cmd {
	reader := m.paneReader
	return func() tea.Msg {
		text, err := reader(context.Background(), paneID)
		return agentsPeekMsg{paneID: paneID, text: text, err: err}
	}
}

func readAgentPane(ctx context.Context, paneID string) (string, error) {
	output, err := exec.CommandContext(ctx, "herdr", "pane", "read", paneID, "--lines", "20", "--format", "text").Output()
	if err != nil {
		return "", fmt.Errorf("herdr pane read %s: %w", paneID, err)
	}
	return string(output), nil
}

func (m *Model) acceptLogEvents(events []logs.Event) {
	m.events = events
	m.rebuild()
}

func (m *Model) logAction(action string, attrs ...slog.Attr) tea.Cmd {
	attrs = append(attrs, slog.String("action", action))
	return m.recordLogCmd("tui action", attrs...)
}

func (m *Model) logError(context string, err error) tea.Cmd {
	if err == nil {
		return nil
	}
	return m.recordLogCmd("tui error", slog.String("context", context), slog.String("error", boundedError(err.Error())))
}

func (m *Model) recordLogCmd(message string, attrs ...slog.Attr) tea.Cmd {
	if m.logger == nil {
		return nil
	}
	logger, now := m.logger, m.now
	return func() tea.Msg {
		record := slog.NewRecord(now(), slog.LevelDebug, message, 0)
		record.AddAttrs(attrs...)
		return logWriteErrMsg{err: logger.Handler().Handle(context.Background(), record)}
	}
}

func batch(cmds ...tea.Cmd) tea.Cmd {
	filtered := make([]tea.Cmd, 0, len(cmds))
	for _, cmd := range cmds {
		if cmd != nil {
			filtered = append(filtered, cmd)
		}
	}
	if len(filtered) == 0 {
		return nil
	}
	if len(filtered) == 1 {
		return filtered[0]
	}
	return tea.Batch(filtered...)
}

// Run starts the TUI on in and out and resolves the home itself.
func Run(in io.Reader, out io.Writer) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	resolved, err := home.Resolve(os.Getenv("HIVE_HOME"), cwd)
	if err != nil {
		return err
	}
	return RunHome(resolved, in, out)
}

// RunHome starts the TUI for one resolved home. It checks the input and the
// output at this boundary. A finite input end or a stream fault settles the
// program and the log writer.
func RunHome(home string, in io.Reader, out io.Writer) error {
	model := New(home, time.Now)
	defer model.Stop()
	notice, startupErr := setupSession(home, model.logger)
	model.startupNotice = notice
	if startupErr != nil {
		model.startupErr = boundedError(startupErr.Error())
	}

	inputEnded := make(chan struct{})
	var endOnce sync.Once
	model.inputEnded = inputEnded

	faults := make(chan struct{}, 1)
	signal := func() {
		select {
		case faults <- struct{}{}:
		default:
		}
	}
	inputGuard := &streamGuard{onFault: func(err error) {
		if errors.Is(err, io.EOF) {
			endOnce.Do(func() { close(inputEnded) })
			return
		}
		signal()
	}}
	outputGuard := &streamGuard{onFault: func(error) { signal() }}
	guardedIn := guardReader(in, inputGuard)
	guardedOut := guardWriter(out, outputGuard)

	program := tea.NewProgram(model,
		tea.WithInput(guardedIn),
		tea.WithOutput(guardedOut),
		tea.WithWindowSize(defaultWidth, defaultHeight),
	)
	watched := make(chan struct{})
	go func() {
		select {
		case <-faults:
			program.Quit()
		case <-watched:
		}
	}()

	_, runErr := program.Run()
	close(watched)

	if err := outputGuard.Err(); err != nil {
		return fmt.Errorf("tui: output: %w", err)
	}
	if err := inputGuard.Err(); err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return fmt.Errorf("tui: input: %w", err)
	}
	if runErr != nil {
		return fmt.Errorf("tui: %w", runErr)
	}
	return nil
}

type streamGuard struct {
	mu      sync.Mutex
	err     error
	onFault func(error)
}

func (g *streamGuard) fault(err error) {
	if err == nil {
		return
	}
	g.mu.Lock()
	first := g.err == nil
	if first {
		g.err = err
	}
	g.mu.Unlock()
	if first && g.onFault != nil {
		g.onFault(err)
	}
}

func (g *streamGuard) Err() error {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.err
}

type guardedReader struct {
	in    io.Reader
	guard *streamGuard
}

func (r *guardedReader) Read(data []byte) (int, error) {
	n, err := r.in.Read(data)
	if err != nil {
		r.guard.fault(err)
	}
	return n, err
}

type guardedOutput struct {
	out   io.Writer
	guard *streamGuard
}

func (w *guardedOutput) Write(data []byte) (int, error) {
	n, err := w.out.Write(data)
	if err == nil && n != len(data) {
		err = io.ErrShortWrite
	}
	if err != nil {
		w.guard.fault(err)
	}
	if n < 0 {
		n = 0
	}
	if n > len(data) {
		n = len(data)
	}
	return n, err
}

// Read and Close keep the term.File shape for the installed renderer. The
// renderer uses Fd to detect a terminal and to get its size.
func (w *guardedOutput) Read(data []byte) (int, error) {
	if reader, ok := w.out.(io.Reader); ok {
		return reader.Read(data)
	}
	return 0, errors.New("tui: output is not readable")
}

func (w *guardedOutput) Close() error {
	if closer, ok := w.out.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

func (w *guardedOutput) Fd() uintptr {
	if file, ok := w.out.(fdHolder); ok {
		return file.Fd()
	}
	return ^uintptr(0)
}

type fdHolder interface {
	Fd() uintptr
}

// guardReader keeps a real terminal unwrapped, so the input raw mode of Bubble
// Tea stays unchanged. Any other reader gets the guard: Bubble Tea then uses
// its fallback reader, and the guard turns a finite end of input into a program
// quit. A raw Fd on a regular file fails in epoll, and a pipe end otherwise
// leaves the program active.
func guardReader(in io.Reader, guard *streamGuard) io.Reader {
	if file, ok := in.(fdHolder); ok && xterm.IsTerminal(file.Fd()) {
		return in
	}
	return &guardedReader{in: in, guard: guard}
}

func guardWriter(out io.Writer, guard *streamGuard) io.Writer {
	return &guardedOutput{out: out, guard: guard}
}
